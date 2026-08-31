package monitoringsdk

import "testing"

func TestMonitoringHTTPSurfaceContractIsCompleteAndSourceOwned(t *testing.T) {
	contract := MonitoringHTTPSurfaceContract()
	if contract.ContractVersion != MonitoringHTTPSurfaceContractVersion || contract.Owner != "monitoring" || contract.Name == "" {
		t.Fatalf("incomplete Monitoring HTTP contract: %#v", contract)
	}
	if len(contract.Routes) != 1 {
		t.Fatalf("Monitoring HTTP routes=%d", len(contract.Routes))
	}
	route := contract.Routes[0]
	if route.Pattern != "GET /operations/monitoring/metrics" || len(route.Exposures) != 1 || route.Exposures[0] != "ops" {
		t.Fatalf("Monitoring HTTP route=%#v", route)
	}
	if route.EffectClass != "read" || route.IdempotencyDecision != "not_applicable" || route.AuditClass == "" || len(route.AnyPermissions) != 2 {
		t.Fatalf("incomplete Monitoring route policy: %#v", route)
	}
	operation := contract.OpenAPI[route.Pattern]
	if operation["operationId"] != "getMonitoringMetrics" || operation["responses"] == nil {
		t.Fatalf("incomplete Monitoring OpenAPI operation: %#v", operation)
	}
}
