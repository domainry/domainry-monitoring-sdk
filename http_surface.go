package monitoringsdk

const MonitoringHTTPSurfaceContractVersion = "domainry-monitoring-http-surface-v1"

type HTTPRouteContract struct {
	Pattern             string   `json:"pattern"`
	Exposures           []string `json:"exposures"`
	Authentication      string   `json:"authentication"`
	Permission          string   `json:"permission,omitempty"`
	AnyPermissions      []string `json:"any_permissions,omitempty"`
	PrincipalOnly       bool     `json:"principal_only,omitempty"`
	EffectClass         string   `json:"effect_class"`
	HighRiskPolicy      string   `json:"high_risk_policy"`
	IdempotencyDecision string   `json:"idempotency_decision"`
	AuditClass          string   `json:"audit_class"`
}

type HTTPSurfaceContract struct {
	ContractVersion string                    `json:"contract_version"`
	Owner           string                    `json:"owner"`
	Name            string                    `json:"name"`
	Routes          []HTTPRouteContract       `json:"routes"`
	OpenAPI         map[string]map[string]any `json:"openapi_operations"`
}

// MonitoringHTTPSurfaceContract is the deployment-neutral product HTTP
// contract implemented by the embedded Monitoring module. Runtime mounts the
// route while Control Plane and Agents may discover it without importing the
// Monitoring implementation.
func MonitoringHTTPSurfaceContract() HTTPSurfaceContract {
	const pattern = "GET /operations/monitoring/metrics"
	return HTTPSurfaceContract{
		ContractVersion: MonitoringHTTPSurfaceContractVersion,
		Owner:           "monitoring",
		Name:            "operations_metrics",
		Routes: []HTTPRouteContract{{
			Pattern: pattern, Exposures: []string{"ops"}, Authentication: "authenticated",
			AnyPermissions: []string{"workspace.admin", "runtime_ops.capability_status.read"},
			EffectClass:    "read", HighRiskPolicy: "none", IdempotencyDecision: "not_applicable", AuditClass: "monitoring_owner_read",
		}},
		OpenAPI: map[string]map[string]any{
			pattern: {
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
		},
	}
}
