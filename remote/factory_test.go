package remote

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/domainry/domainry-foundation/requestcontext"
	"github.com/domainry/domainry-foundation/telemetry"
	monitoringsdk "github.com/domainry/domainry-monitoring-sdk"
	monitoringcontract "github.com/domainry/domainry-monitoring-sdk/contract"
	monitoringmodulehost "github.com/domainry/domainry-monitoring-sdk/modulehost"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"
)

func TestRemoteFactoryPropagatesOneTraceIdentityWithoutBusinessCorrelation(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(previousProvider)
	})

	type observedHeaders struct{ traceparent, correlationID, requestID string }
	observed := make(chan observedHeaders, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		observed <- observedHeaders{
			traceparent:   request.Header.Get("traceparent"),
			correlationID: request.Header.Get("X-Correlation-ID"),
			requestID:     request.Header.Get(telemetry.RequestIDHeader),
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(monitoringsdk.Descriptor{ProtocolVersion: monitoringsdk.ProtocolVersionV1, Mode: monitoringsdk.DeploymentModeSaaS})
	}))
	t.Cleanup(server.Close)

	traceID, _ := apitrace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := apitrace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := apitrace.ContextWithSpanContext(t.Context(), apitrace.NewSpanContext(apitrace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: apitrace.FlagsSampled}))
	ctx = requestcontext.WithRequestID(ctx, "runtime-request")
	if _, err := NewFactory(Config{Endpoint: server.URL, Token: "secret", Client: server.Client()}).OpenSaaS(ctx, monitoringsdk.ApplicationRef{RuntimeID: "runtime"}, remoteTestHost{}); err != nil {
		t.Fatal(err)
	}
	headers := <-observed
	if len(headers.traceparent) != 55 || headers.traceparent[3:35] != traceID.String() {
		t.Fatalf("traceparent=%q", headers.traceparent)
	}
	if headers.correlationID != "" {
		t.Fatalf("correlation id=%q", headers.correlationID)
	}
	if headers.requestID != traceID.String() {
		t.Fatalf("request/trace alias=%q want=%q", headers.requestID, traceID.String())
	}
}

func TestRemoteFactoryRejectsHTTPFailureWithoutCopyingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusServiceUnavailable)
		_, _ = response.Write([]byte("secret-provider-body"))
	}))
	t.Cleanup(server.Close)
	_, err := NewFactory(Config{Endpoint: server.URL, Client: server.Client()}).OpenSaaS(t.Context(), monitoringsdk.ApplicationRef{RuntimeID: "runtime"}, remoteTestHost{})
	if err == nil || !strings.Contains(err.Error(), "status 503") || strings.Contains(err.Error(), "secret-provider-body") {
		t.Fatalf("error=%v", err)
	}
}

type remoteTestHost struct{}

func (remoteTestHost) Identity() monitoringcontract.Identity     { return monitoringcontract.Identity{} }
func (remoteTestHost) Storage() monitoringmodulehost.Storage     { return remoteTestStorage{} }
func (remoteTestHost) Migration() monitoringmodulehost.Migration { return remoteTestMigration{} }
func (remoteTestHost) Scheduler() monitoringmodulehost.Component { return remoteTestComponent{} }
func (remoteTestHost) Lifecycle() monitoringmodulehost.Component { return remoteTestComponent{} }
func (remoteTestHost) Metrics() monitoringmodulehost.Metrics     { return remoteTestMetrics{} }

type remoteTestStorage struct{}

func (remoteTestStorage) Status(context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}
func (remoteTestStorage) Readiness(context.Context) error { return nil }

type remoteTestMigration struct{}

func (remoteTestMigration) Status(context.Context) (monitoringmodulehost.Status, error) {
	return monitoringmodulehost.Status{Current: true}, nil
}
func (remoteTestMigration) Readiness(context.Context) error              { return nil }
func (remoteTestMigration) Telemetry(context.Context) (int, bool, error) { return 0, true, nil }

type remoteTestComponent struct{}

func (remoteTestComponent) Observe(context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}

type remoteTestMetrics struct{}

func (remoteTestMetrics) Observe(context.Context) (map[string]any, map[string]string) {
	return map[string]any{}, nil
}
