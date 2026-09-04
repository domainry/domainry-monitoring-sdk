package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/domainry/domainry-foundation/modulecapability"
	monitoringsdk "github.com/domainry/domainry-monitoring-sdk"
	"github.com/domainry/domainry-monitoring-sdk/modulehost"
	"github.com/domainry/domainry-monitoring-sdk/saashost"
)

type Config struct {
	Endpoint                 string
	Token                    string
	Timeout                  time.Duration
	Client                   *http.Client
	CapabilityContractSHA256 string
}

func ConfigFromEnvironment() Config {
	timeout := 10 * time.Second
	if value := strings.TrimSpace(os.Getenv("DOMAINRY_MONITORING_TIMEOUT")); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return Config{
		Endpoint: strings.TrimRight(strings.TrimSpace(os.Getenv("DOMAINRY_MONITORING_ENDPOINT")), "/"), Token: strings.TrimSpace(os.Getenv("DOMAINRY_MONITORING_TOKEN")), Timeout: timeout,
		CapabilityContractSHA256: strings.TrimSpace(os.Getenv("DOMAINRY_MONITORING_CAPABILITY_CONTRACT_SHA256")),
	}
}

type Factory struct{ config Config }

func NewFactory(config Config) *Factory { return &Factory{config: config} }

func (*Factory) Open(context.Context, monitoringsdk.ApplicationRef) (monitoringsdk.Binding, error) {
	return nil, fmt.Errorf("Monitoring SaaS observation host is required")
}

func (f *Factory) OpenSaaS(ctx context.Context, application monitoringsdk.ApplicationRef, host modulehost.Host) (monitoringsdk.Binding, error) {
	if err := application.Validate(); err != nil {
		return nil, err
	}
	if err := modulecapability.ValidateRemoteExpectation("monitoring", f.config.CapabilityContractSHA256); err != nil {
		return nil, err
	}
	if host == nil || host.Storage() == nil || host.Migration() == nil || host.Scheduler() == nil || host.Lifecycle() == nil || host.Metrics() == nil {
		return nil, fmt.Errorf("Monitoring SaaS host is incomplete")
	}
	endpoint := strings.TrimRight(strings.TrimSpace(f.config.Endpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("Monitoring SaaS endpoint is required")
	}
	timeout := f.config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := f.config.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	binding := &remoteBinding{application: application, host: host, endpoint: endpoint, token: strings.TrimSpace(f.config.Token), client: client, timeout: timeout}
	descriptor, err := binding.fetchDescriptor(ctx)
	if err != nil {
		return nil, err
	}
	if err := descriptor.Validate(); err != nil {
		return nil, err
	}
	if descriptor.Mode != monitoringsdk.DeploymentModeSaaS {
		return nil, fmt.Errorf("Monitoring endpoint is not SaaS")
	}
	capability, err := modulecapability.OpenRemote(ctx, modulecapability.RemoteConfig{
		BaseURL: endpoint, Client: client, ExpectedModuleKey: "monitoring", ExpectedContractSHA256: f.config.CapabilityContractSHA256,
		Authorize: func(request *http.Request) error {
			if binding.token != "" {
				request.Header.Set("Authorization", "Bearer "+binding.token)
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	binding.descriptor = descriptor
	binding.capability = capability
	return binding, nil
}

type remoteBinding struct {
	application     monitoringsdk.ApplicationRef
	host            modulehost.Host
	endpoint, token string
	client          *http.Client
	timeout         time.Duration
	descriptor      monitoringsdk.Descriptor
	capability      modulecapability.Binding
}

func (b *remoteBinding) CapabilitySummary(ctx context.Context) (modulecapability.ModuleSummary, error) {
	return b.capability.CapabilitySummary(ctx)
}
func (b *remoteBinding) CapabilityCategory(ctx context.Context, key string) (modulecapability.CategoryDocument, error) {
	return b.capability.CapabilityCategory(ctx, key)
}
func (b *remoteBinding) ValidateCapabilityCandidate(ctx context.Context, request modulecapability.ValidationRequest) (modulecapability.ValidationResult, error) {
	return b.capability.ValidateCapabilityCandidate(ctx, request)
}

func (b *remoteBinding) Descriptor() monitoringsdk.Descriptor { return b.descriptor }
func (b *remoteBinding) Health(ctx context.Context) map[string]any {
	result := map[string]any{}
	if err := b.post(ctx, "/monitoring/v1/health", modulehost.CollectHealth(ctx, b.application.RuntimeID, b.host), &result); err != nil {
		return map[string]any{"status": "degraded", "runtime_id": b.application.RuntimeID, "checks": map[string]string{"monitoring_saas": "error"}, "errors": map[string]string{"monitoring_saas": err.Error()}}
	}
	return result
}
func (b *remoteBinding) Metrics(ctx context.Context) map[string]any {
	result := map[string]any{}
	if err := b.post(ctx, "/monitoring/v1/metrics", modulehost.CollectMetrics(ctx, b.application.RuntimeID, b.host), &result); err != nil {
		return map[string]any{"runtime_id": b.application.RuntimeID, "errors": map[string]string{"monitoring_saas": err.Error()}}
	}
	return result
}
func (b *remoteBinding) StorageReadiness(ctx context.Context) error {
	return b.host.Storage().Readiness(ctx)
}
func (b *remoteBinding) MigrationReadiness(ctx context.Context) error {
	return b.host.Migration().Readiness(ctx)
}
func (b *remoteBinding) MigrationTelemetry(ctx context.Context) (int, bool, error) {
	return b.host.Migration().Telemetry(ctx)
}
func (*remoteBinding) Close(context.Context) error { return nil }

func (b *remoteBinding) fetchDescriptor(ctx context.Context) (monitoringsdk.Descriptor, error) {
	var descriptor monitoringsdk.Descriptor
	requestCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	request, err := b.request(requestCtx, http.MethodGet, "/monitoring/v1/descriptor", nil)
	if err != nil {
		return descriptor, err
	}
	response, err := b.client.Do(request)
	if err != nil {
		return descriptor, fmt.Errorf("fetch Monitoring descriptor: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return descriptor, fmt.Errorf("fetch Monitoring descriptor: status %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(&descriptor); err != nil {
		return descriptor, fmt.Errorf("decode Monitoring descriptor: %w", err)
	}
	return descriptor, nil
}

func (b *remoteBinding) post(ctx context.Context, path string, input, output any) error {
	requestCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}
	request, err := b.request(requestCtx, http.MethodPost, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := b.client.Do(request)
	if err != nil {
		return fmt.Errorf("call Monitoring SaaS: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("call Monitoring SaaS: status %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode Monitoring SaaS response: %w", err)
	}
	return nil
}

func (b *remoteBinding) request(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, b.endpoint+path, body)
	if err != nil {
		return nil, err
	}
	if b.token != "" {
		request.Header.Set("Authorization", "Bearer "+b.token)
	}
	return request, nil
}

var _ monitoringsdk.Factory = (*Factory)(nil)
var _ saashost.Factory = (*Factory)(nil)
var _ monitoringsdk.Binding = (*remoteBinding)(nil)
