package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var clusterNameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{2,62}$`)

// RegisterCustomValidators는 커스텀 유효성 검증 규칙을 등록합니다.
func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("cluster_name", validateClusterName)
	_ = v.RegisterValidation("k8s_version", validateK8sVersion)
	_ = v.RegisterValidation("cp_replicas", validateCPReplicas)
}

// validateClusterName은 클러스터 이름 규칙을 검증합니다.
// 규칙: 소문자 알파벳으로 시작, 소문자/숫자/하이픈만 허용, 3~63자
func validateClusterName(fl validator.FieldLevel) bool {
	return clusterNameRegex.MatchString(fl.Field().String())
}

// validateK8sVersion은 Kubernetes 버전 형식을 검증합니다. (예: v1.29.2)
func validateK8sVersion(fl validator.FieldLevel) bool {
	k8sVersionRegex := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	return k8sVersionRegex.MatchString(fl.Field().String())
}

// validateCPReplicas는 Control Plane 레플리카 수가 홀수인지 검증합니다.
func validateCPReplicas(fl validator.FieldLevel) bool {
	v := fl.Field().Int()
	return v == 1 || v == 3 || v == 5
}
