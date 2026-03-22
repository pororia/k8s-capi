package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k8s-cmp/k8s-cmp/internal/handler"
	"github.com/k8s-cmp/k8s-cmp/internal/k8s"
	"github.com/k8s-cmp/k8s-cmp/internal/middleware"
	"github.com/k8s-cmp/k8s-cmp/internal/openstack"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	ws "github.com/k8s-cmp/k8s-cmp/internal/websocket"
	"github.com/k8s-cmp/k8s-cmp/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// 설정 로드
	cfg := loadConfig()

	// 로거 초기화
	logger := initLogger(cfg.GetString("server.mode"))
	defer logger.Sync()

	// DB 연결
	db := initDB(cfg, logger)

	// Redis 연결
	rdb := initRedis(cfg, logger)

	// AES 암호화
	aes, err := crypto.NewAES(cfg.GetString("encryption.key"))
	if err != nil {
		logger.Fatal("failed to initialize AES", zap.Error(err))
	}

	// Repository 초기화
	userRepo := repository.NewUserRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	eventRepo := repository.NewClusterEventRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)

	// K8s Client 초기화
	k8sClient, err := k8s.NewDynamicClient(
		cfg.GetString("management_cluster.kubeconfig_path"),
		cfg.GetBool("management_cluster.in_cluster"),
	)
	if err != nil {
		logger.Warn("failed to initialize k8s client (some features may be unavailable)", zap.Error(err))
	}

	// Manifest Generator
	manifestGen, err := k8s.NewManifestGenerator()
	if err != nil {
		logger.Fatal("failed to initialize manifest generator", zap.Error(err))
	}

	// Cluster Manager (k8sClient가 nil일 수 있음)
	var clusterMgr *k8s.ClusterManager
	if k8sClient != nil {
		clusterMgr = k8s.NewClusterManager(k8sClient)
	}

	// Service 초기화
	accessTTL, _ := time.ParseDuration(cfg.GetString("auth.access_token_ttl"))
	refreshTTL, _ := time.ParseDuration(cfg.GetString("auth.refresh_token_ttl"))
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	authSvc := service.NewAuthService(userRepo, rdb, logger,
		cfg.GetString("auth.jwt_secret"), accessTTL, refreshTTL)
	tenantSvc := service.NewTenantService(tenantRepo, aes, logger)
	clusterSvc := service.NewClusterService(
		clusterRepo, tenantRepo, eventRepo,
		clusterMgr, manifestGen, tenantSvc, logger)
	osFactory := openstack.NewClientFactory()
	osSvc := service.NewOpenStackService(tenantRepo, tenantSvc, osFactory, rdb, logger)

	// WebSocket Hub
	hub := ws.NewHub(logger)
	go hub.Run()

	// Handler 초기화
	authHandler := handler.NewAuthHandler(authSvc)
	tenantHandler := handler.NewTenantHandler(tenantSvc)
	clusterHandler := handler.NewClusterHandler(clusterSvc)
	osHandler := handler.NewOpenStackHandler(osSvc)
	wsHandler := handler.NewWebSocketHandler(hub, authSvc, logger)

	// Gin 라우터 설정
	if cfg.GetString("server.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.GetStringSlice("cors.allowed_origins")))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// WebSocket 라우트
	wsGroup := r.Group("/ws/v1")
	{
		wsGroup.GET("/clusters/stream", wsHandler.StreamAll)
		wsGroup.GET("/clusters/:id/stream", wsHandler.StreamCluster)
	}

	// API v1 라우트
	v1 := r.Group("/api/v1")

	// 인증 (공개)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", middleware.Auth(authSvc), authHandler.Logout)
		auth.GET("/me", middleware.Auth(authSvc), authHandler.Me)
	}

	// 인증 필요 라우트
	protected := v1.Group("")
	protected.Use(middleware.Auth(authSvc))
	protected.Use(middleware.Audit(auditRepo, logger))

	// 테넌트 (Super Admin만)
	tenants := protected.Group("/tenants")
	tenants.Use(middleware.SuperAdminOnly())
	{
		tenants.GET("", tenantHandler.List)
		tenants.GET("/:id", tenantHandler.Get)
		tenants.POST("", tenantHandler.Create)
		tenants.PUT("/:id", tenantHandler.Update)
		tenants.DELETE("/:id", tenantHandler.Delete)
	}

	// 클러스터
	clusters := protected.Group("/clusters")
	{
		clusters.GET("", clusterHandler.List)
		clusters.GET("/:id", clusterHandler.Get)
		clusters.GET("/:id/events", clusterHandler.GetEvents)
		clusters.GET("/:id/kubeconfig", middleware.OperatorOrAbove(), clusterHandler.GetKubeconfig)
		clusters.POST("", middleware.OperatorOrAbove(), clusterHandler.Create)
		clusters.DELETE("/:id", middleware.OperatorOrAbove(), clusterHandler.Delete)
		clusters.POST("/:id/scale", middleware.OperatorOrAbove(), clusterHandler.Scale)
		clusters.POST("/:id/upgrade", middleware.OperatorOrAbove(), clusterHandler.Upgrade)
	}

	// OpenStack 리소스
	os := protected.Group("/openstack")
	{
		os.GET("/flavors", osHandler.GetFlavors)
		os.GET("/images", osHandler.GetImages)
		os.GET("/networks", osHandler.GetNetworks)
		os.GET("/networks/:id/subnets", osHandler.GetSubnets)
		os.GET("/security-groups", osHandler.GetSecurityGroups)
		os.GET("/keypairs", osHandler.GetKeypairs)
		os.GET("/external-networks", osHandler.GetExternalNetworks)
		os.GET("/quotas", osHandler.GetQuotas)
	}

	// HTTP 서버 시작
	addr := fmt.Sprintf("%s:%d", cfg.GetString("server.host"), cfg.GetInt("server.port"))
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful Shutdown
	go func() {
		logger.Info("starting api server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server exited")
}

func loadConfig() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/cmp")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Warning: config file not found: %v\n", err)
	}
	return v
}

func initLogger(mode string) *zap.Logger {
	var logger *zap.Logger
	var err error

	if mode == "release" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger
}

func initDB(cfg *viper.Viper, logger *zap.Logger) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"),
		cfg.GetString("database.user"),
		cfg.GetString("database.password"),
		cfg.GetString("database.name"),
		cfg.GetString("database.ssl_mode"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.max_open_conns"))
	sqlDB.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns"))

	logger.Info("database connected")
	return db
}

func initRedis(cfg *viper.Viper, logger *zap.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.GetString("redis.host"), cfg.GetInt("redis.port")),
		Password: cfg.GetString("redis.password"),
		DB:       cfg.GetInt("redis.db"),
	})

	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logger.Warn("failed to connect to redis", zap.Error(err))
	} else {
		logger.Info("redis connected")
	}
	return rdb
}
