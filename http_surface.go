package monitoringsdk

import actioncontract "github.com/domainry/domainry-foundation/action"

const MonitoringHTTPSurfaceContractVersion = "domainry-monitoring-http-surface-v2"
const ActionMonitoringMetricsRead = "monitoring.metrics.read"

type HTTPRouteContract struct {
	Action           actioncontract.ActionDefinition `json:"action"`
	OpenAPIOperation map[string]any                  `json:"openapi_operation"`
}

func (route HTTPRouteContract) Pattern() string {
	if route.Action.HTTP == nil {
		return ""
	}
	return route.Action.HTTP.Method + " " + route.Action.HTTP.RouteTemplate
}

type HTTPSurfaceContract struct {
	ContractVersion string              `json:"contract_version"`
	Owner           string              `json:"owner"`
	Name            string              `json:"name"`
	Routes          []HTTPRouteContract `json:"routes"`
}

func (contract HTTPSurfaceContract) OpenAPIOperations() map[string]map[string]any {
	operations := make(map[string]map[string]any, len(contract.Routes))
	for _, route := range contract.Routes {
		operations[route.Pattern()] = route.OpenAPIOperation
	}
	return operations
}

// MonitoringHTTPSurfaceContract is the deployment-neutral product HTTP
// contract implemented by the embedded Monitoring module. Runtime mounts the
// route while Control Plane and Agents may discover it without importing the
// Monitoring implementation.
func MonitoringHTTPSurfaceContract() HTTPSurfaceContract {
	return HTTPSurfaceContract{
		ContractVersion: MonitoringHTTPSurfaceContractVersion,
		Owner:           "monitoring",
		Name:            "operations_metrics",
		Routes: []HTTPRouteContract{{
			Action: actioncontract.ActionDefinition{
				Key: ActionMonitoringMetricsRead, Owner: "module:monitoring", SourceKind: "module_surface", CapabilityKey: "monitoring.metrics", CapabilityLabel: "Monitoring metrics",
				OperationKey: "read", OperationLabel: "Read monitoring metrics", Label: "Read monitoring metrics", Exposures: []actioncontract.Exposure{actioncontract.ExposureOps},
				Authorization: actioncontract.Authorization{Strategy: actioncontract.AuthorizationExactRolePermission},
				HTTP:          &actioncontract.HTTPBinding{Method: "GET", RouteTemplate: "/operations/monitoring/metrics"},
				Permission:    &actioncontract.PermissionDefinition{Key: ActionMonitoringMetricsRead, Owner: "module:monitoring", ResourceKey: "monitoring.metrics", OperationKey: "read", Label: "Read monitoring metrics", Category: "Monitoring", LifecycleStatus: actioncontract.LifecycleActive},
				EffectClass:   actioncontract.EffectRead, RiskLevel: actioncontract.RiskLow, IdempotencyDecision: "not_applicable", AuditClass: "monitoring_owner_read", LifecycleStatus: actioncontract.LifecycleActive,
			},
			OpenAPIOperation: map[string]any{
				"operationId": "getMonitoringMetrics",
				"tags":        []string{"Monitoring"},
				"summary":     "Read Monitoring-owned aggregated Runtime metrics",
				"security":    []map[string]any{{"BearerAuth": []string{}}},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Monitoring metrics",
						"content": map[string]any{"application/json": map[string]any{
							"schema": map[string]any{"type": "object", "additionalProperties": true},
						}},
					},
				},
			},
		}},
	}
}
