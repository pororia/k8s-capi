package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 설정 로드
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/cmp")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: config file not found: %v\n", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// NATS 연결
	nc, err := nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		logger.Fatal("failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.Fatal("failed to create JetStream context", zap.Error(err))
	}

	// Stream 생성 (없으면)
	streamName := viper.GetString("nats.stream_name")
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamName + ".*"},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		logger.Warn("failed to create NATS stream", zap.Error(err))
	}

	logger.Info("worker started", zap.String("nats", viper.GetString("nats.url")))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 클러스터 프로비저닝 모니터링 작업 구독
	sub, err := js.QueueSubscribe(streamName+".cluster.provision", "workers", func(msg *nats.Msg) {
		logger.Info("received provision task", zap.String("data", string(msg.Data)))
		// TODO: 프로비저닝 상태 모니터링 로직
		msg.Ack()
	})
	if err != nil {
		logger.Fatal("failed to subscribe", zap.Error(err))
	}
	defer sub.Unsubscribe()

	// 종료 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("worker shutting down...")
	case <-ctx.Done():
	}

	logger.Info("worker exited")
}
