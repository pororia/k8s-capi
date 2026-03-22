package k8s

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

//go:embed manifest_templates/*.tmpl
var templateFS embed.FS

// ClusterCreateParams는 클러스터 생성에 필요한 파라미터입니다.
type ClusterCreateParams struct {
	ClusterName       string
	Namespace         string
	KubernetesVersion string
	ControlPlaneCount int
	WorkerCount       int
	CPFlavor          string
	WorkerFlavor      string
	OSImage           string
	NetworkID         string
	SubnetID          string
	ExternalNetworkID string
	SSHKeyName        string
	PodCIDR           string
	ServiceCIDR       string
	CNIPlugin         string
	DNSNameservers    []string
	CloudName         string
	AuthURL           string
	ProjectID         string
	DomainName        string
	Username          string
	Password          string
	UserDomainName    string
}

// CloudsYAML은 OpenStack clouds.yaml 구조입니다.
type CloudsYAML struct {
	Clouds map[string]CloudConfig `yaml:"clouds"`
}

// CloudConfig는 clouds.yaml 내 개별 클라우드 설정입니다.
type CloudConfig struct {
	Auth           CloudAuth `yaml:"auth"`
	RegionName     string    `yaml:"region_name,omitempty"`
	Interface      string    `yaml:"interface,omitempty"`
	IdentityAPIVer string    `yaml:"identity_api_version,omitempty"`
}

// CloudAuth는 clouds.yaml 인증 정보입니다.
type CloudAuth struct {
	AuthURL            string `yaml:"auth_url"`
	ApplicationCredentialID     string `yaml:"application_credential_id,omitempty"`
	ApplicationCredentialSecret string `yaml:"application_credential_secret,omitempty"`
	UserDomainName     string `yaml:"user_domain_name,omitempty"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	ProjectID          string `yaml:"project_id,omitempty"`
}

// ManifestGenerator는 CAPI/CAPO 매니페스트를 생성합니다.
type ManifestGenerator struct {
	templates *template.Template
}

// NewManifestGenerator는 새로운 ManifestGenerator를 생성합니다.
func NewManifestGenerator() (*ManifestGenerator, error) {
	tmpl, err := template.New("").ParseFS(templateFS, "manifest_templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest templates: %w", err)
	}
	return &ManifestGenerator{templates: tmpl}, nil
}

// GenerateManifests는 CAPI/CAPO 리소스 매니페스트를 생성합니다.
func (g *ManifestGenerator) GenerateManifests(params ClusterCreateParams) ([]unstructured.Unstructured, error) {
	var manifests []unstructured.Unstructured

	// clouds.yaml Secret 생성
	cloudsSecret, err := g.generateCloudsSecret(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate clouds secret: %w", err)
	}
	manifests = append(manifests, *cloudsSecret)

	// 각 템플릿에서 매니페스트 생성
	templateNames := []string{
		"cluster.yaml.tmpl",
		"openstack_cluster.yaml.tmpl",
		"openstack_machine_template_cp.yaml.tmpl",
		"kubeadm_control_plane.yaml.tmpl",
		"openstack_machine_template_worker.yaml.tmpl",
		"kubeadm_config_template.yaml.tmpl",
		"machine_deployment.yaml.tmpl",
	}

	for _, tmplName := range templateNames {
		var buf bytes.Buffer
		if err := g.templates.ExecuteTemplate(&buf, tmplName, params); err != nil {
			return nil, fmt.Errorf("failed to execute template %s: %w", tmplName, err)
		}

		obj, err := parseYAMLToUnstructured(buf.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest from %s: %w", tmplName, err)
		}
		manifests = append(manifests, *obj)
	}

	return manifests, nil
}

func (g *ManifestGenerator) generateCloudsSecret(params ClusterCreateParams) (*unstructured.Unstructured, error) {
	cloudsYAML := fmt.Sprintf(`clouds:
  %s:
    auth:
      auth_url: %s
      username: %s
      password: %s
      project_id: %s
      user_domain_name: %s
    identity_api_version: 3
`, params.CloudName, params.AuthURL, params.Username, params.Password, params.ProjectID, params.UserDomainName)

	cloudsJSON := fmt.Sprintf(`{"clouds":{"%s":{"auth":{"auth_url":"%s","username":"%s","password":"%s","project_id":"%s","user_domain_name":"%s"},"identity_api_version":3}}}`,
		params.CloudName, params.AuthURL, params.Username, params.Password, params.ProjectID, params.UserDomainName)

	secret := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      params.ClusterName + "-cloud-config",
				"namespace": params.Namespace,
			},
			"stringData": map[string]interface{}{
				"clouds.yaml": cloudsYAML,
				"clouds.json": cloudsJSON,
			},
		},
	}

	return secret, nil
}

func parseYAMLToUnstructured(yamlStr string) (*unstructured.Unstructured, error) {
	jsonBytes, err := yaml.YAMLToJSON([]byte(yamlStr))
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &unstructured.Unstructured{Object: obj}, nil
}

// DNSServersStr은 DNS 서버 목록을 YAML 배열 형식으로 반환합니다.
func DNSServersStr(servers []string) string {
	if len(servers) == 0 {
		return "[]"
	}
	var sb strings.Builder
	for _, s := range servers {
		sb.WriteString(fmt.Sprintf("- %s\n", s))
	}
	return sb.String()
}
