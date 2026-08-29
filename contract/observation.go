package contract

type Identity struct {
	TemplateID      string `json:"template_id"`
	TemplateVersion string `json:"template_version"`
}

type ComponentObservation struct {
	Payload map[string]any `json:"payload"`
	Error   string         `json:"error,omitempty"`
}

type MigrationObservation struct {
	Current bool   `json:"current"`
	Payload any    `json:"payload"`
	Error   string `json:"error,omitempty"`
}

type HealthRequest struct {
	RuntimeID string               `json:"runtime_id"`
	Identity  Identity             `json:"identity"`
	Storage   ComponentObservation `json:"storage"`
	Migration MigrationObservation `json:"migration"`
	Scheduler ComponentObservation `json:"scheduler"`
	Lifecycle ComponentObservation `json:"lifecycle"`
}

type MetricsRequest struct {
	RuntimeID string            `json:"runtime_id"`
	Identity  Identity          `json:"identity"`
	Sections  map[string]any    `json:"sections"`
	Errors    map[string]string `json:"errors,omitempty"`
}
