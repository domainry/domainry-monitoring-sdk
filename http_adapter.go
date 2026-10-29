package monitoringsdk

import actioncontract "github.com/domainry/domainry-foundation/action"

const MonitoringHTTPAdapterContractVersion = "domainry-monitoring-http-adapter-v1"
const ActionMonitoringMetricsRead = "monitoring.metrics.read"

type HTTPRouteContract struct {
	Action actioncontract.ActionDefinition `json:"action"`
}

func (route HTTPRouteContract) Pattern() string {
	if route.Action.HTTP == nil {
		return ""
	}
	return route.Action.HTTP.Method + " " + route.Action.HTTP.RouteTemplate
}

type HTTPAdapterContract struct {
	ContractVersion string              `json:"contract_version"`
	Owner           string              `json:"owner"`
	Name            string              `json:"name"`
	Routes          []HTTPRouteContract `json:"routes"`
}

// MonitoringHTTPAdapterContract is the deployment-neutral product HTTP
// contract implemented by the embedded Monitoring module. Runtime mounts the
// route while Control Plane and Agents may discover it without importing the
// Monitoring implementation.
func MonitoringHTTPAdapterContract() HTTPAdapterContract {
	return HTTPAdapterContract{
		ContractVersion: MonitoringHTTPAdapterContractVersion,
		Owner:           "monitoring",
		Name:            "operations_metrics",
		Routes: []HTTPRouteContract{{
			Action: actioncontract.ActionDefinition{
				Key: ActionMonitoringMetricsRead, Owner: "module:monitoring", SourceKind: "module_http", CapabilityKey: "monitoring.metrics", CapabilityLabel: "Monitoring metrics",
				OperationKey: "read", OperationLabel: "Read monitoring metrics", Label: "Read monitoring metrics", Exposures: []actioncontract.Exposure{actioncontract.ExposureOps},
				Authorization: actioncontract.Authorization{Strategy: actioncontract.AuthorizationAuthenticated},
				HTTP:          &actioncontract.HTTPBinding{Method: "GET", RouteTemplate: "/monitoring/metrics"},
				Permission: &actioncontract.PermissionDefinition{
					Key: ActionMonitoringMetricsRead, Owner: "module:monitoring", ModuleKey: "monitoring", GroupKey: "monitoring.metrics",
					ResourceKey: "monitoring.metrics", OperationKey: "read", Label: "Read monitoring metrics", Category: "Monitoring",
					GrantMode: actioncontract.PermissionGrantInternal, ScopeKind: actioncontract.PermissionScopeNone, DefaultLocale: "en",
					I18n: actioncontract.PermissionLocalizedText{
						"en":    {ModuleLabel: "Monitoring", GroupLabel: "Runtime metrics", ResourceLabel: "Monitoring metrics", OperationLabel: "View", Label: "View monitoring metrics", Description: "Allows viewing Runtime health and performance metrics."},
						"zh-CN": {ModuleLabel: "监控", GroupLabel: "运行指标", ResourceLabel: "监控指标", OperationLabel: "查看", Label: "查看监控指标", Description: "允许查看 Runtime 的健康状态与性能指标。"},
					},
					LifecycleStatus: actioncontract.LifecycleActive,
				},
				EffectClass: actioncontract.EffectRead, RiskLevel: actioncontract.RiskLow, IdempotencyDecision: "not_applicable", AuditClass: "monitoring_owner_read", LifecycleStatus: actioncontract.LifecycleActive,
			},
		}},
	}
}
