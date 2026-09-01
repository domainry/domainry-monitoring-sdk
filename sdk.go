// Package monitoringsdk defines deployment-neutral Runtime monitoring contracts.
package monitoringsdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/domainry/domainry-foundation/modulecapability"
)

type DeploymentMode string

const (
	DeploymentModeModule DeploymentMode = "module"
	DeploymentModeSaaS   DeploymentMode = "saas"
	ProtocolVersionV1                   = "domainry-monitoring-protocol-v1"
)

type ApplicationRef struct {
	RuntimeID string `json:"runtime_id"`
}

func (r ApplicationRef) Validate() error {
	if strings.TrimSpace(r.RuntimeID) == "" {
		return fmt.Errorf("monitoring runtime identity is required")
	}
	return nil
}

type Descriptor struct {
	ProtocolVersion string         `json:"protocol_version"`
	Mode            DeploymentMode `json:"mode"`
	Capabilities    []string       `json:"capabilities"`
}

func (d Descriptor) Validate() error {
	if d.ProtocolVersion != ProtocolVersionV1 {
		return fmt.Errorf("unsupported Monitoring protocol %q", d.ProtocolVersion)
	}
	if d.Mode != DeploymentModeModule && d.Mode != DeploymentModeSaaS {
		return fmt.Errorf("invalid Monitoring deployment mode %q", d.Mode)
	}
	return nil
}

type Factory interface {
	Open(context.Context, ApplicationRef) (Binding, error)
}

// Binding is the sole Runtime-facing monitoring capability. Implementations
// may run in-process or remotely without changing Runtime transports.
type Binding interface {
	modulecapability.Binding
	Descriptor() Descriptor
	Health(context.Context) map[string]any
	Metrics(context.Context) map[string]any
	StorageReadiness(context.Context) error
	MigrationReadiness(context.Context) error
	MigrationTelemetry(context.Context) (pending int, current bool, err error)
	Close(context.Context) error
}
