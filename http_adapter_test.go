package monitoringsdk

import "testing"

func TestMonitoringHTTPAdapterContractIsCompleteAndSourceOwned(t *testing.T) {
	contract := MonitoringHTTPAdapterContract()
	if contract.ContractVersion != MonitoringHTTPAdapterContractVersion || contract.Owner != "monitoring" || contract.Name == "" {
		t.Fatalf("incomplete Monitoring HTTP contract: %#v", contract)
	}
	if len(contract.Routes) != 1 {
		t.Fatalf("Monitoring HTTP routes=%d", len(contract.Routes))
	}
	route := contract.Routes[0]
	if route.Pattern() != "GET /monitoring/metrics" || len(route.Action.Exposures) != 1 || route.Action.Exposures[0] != "ops" {
		t.Fatalf("Monitoring HTTP route=%#v", route)
	}
	if route.Action.EffectClass != "read" || route.Action.IdempotencyDecision != "not_applicable" || route.Action.AuditClass == "" || route.Action.Permission == nil || route.Action.Permission.Key != route.Action.Key {
		t.Fatalf("incomplete Monitoring route policy: %#v", route)
	}
	operation := contract.OpenAPIOperations()[route.Pattern()]
	if operation["operationId"] != "getMonitoringMetrics" || operation["responses"] == nil {
		t.Fatalf("incomplete Monitoring OpenAPI operation: %#v", operation)
	}
}
