package modulehost

import (
	"context"

	"github.com/domainry/domainry-monitoring-sdk/contract"
)

// Host exposes owner-produced observations to an in-process Monitoring Module.
// Monitoring aggregates these facts but never reaches into Runtime stores.
type Host interface {
	Identity() Identity
	Storage() Storage
	Migration() Migration
	Scheduler() Component
	Lifecycle() Component
	Metrics() Metrics
}

type Identity = contract.Identity

type Storage interface {
	Status(context.Context) (map[string]any, error)
	Readiness(context.Context) error
}

type Migration interface {
	Status(context.Context) (Status, error)
	Readiness(context.Context) error
	Telemetry(context.Context) (pending int, current bool, err error)
}

type Status struct {
	Current bool
	Payload any
}

type Component interface {
	Observe(context.Context) (map[string]any, error)
}

// Metrics returns owner-shaped metric sections. Owners retain metric meaning;
// Monitoring owns consistent collection, error isolation, and aggregation.
type Metrics interface {
	Observe(context.Context) (map[string]any, map[string]string)
}

func CollectHealth(ctx context.Context, runtimeID string, host Host) contract.HealthRequest {
	storage, storageErr := host.Storage().Status(ctx)
	migration, migrationErr := host.Migration().Status(ctx)
	scheduler, schedulerErr := host.Scheduler().Observe(ctx)
	lifecycle, lifecycleErr := host.Lifecycle().Observe(ctx)
	return contract.HealthRequest{
		RuntimeID: runtimeID, Identity: host.Identity(),
		Storage:   component(storage, storageErr),
		Migration: contract.MigrationObservation{Current: migration.Current, Payload: migration.Payload, Error: errorText(migrationErr)},
		Scheduler: component(scheduler, schedulerErr), Lifecycle: component(lifecycle, lifecycleErr),
	}
}

func CollectMetrics(ctx context.Context, runtimeID string, host Host) contract.MetricsRequest {
	sections, errorsByOwner := host.Metrics().Observe(ctx)
	if sections == nil {
		sections = map[string]any{}
	}
	return contract.MetricsRequest{RuntimeID: runtimeID, Identity: host.Identity(), Sections: sections, Errors: errorsByOwner}
}

func component(payload map[string]any, err error) contract.ComponentObservation {
	if payload == nil {
		payload = map[string]any{}
	}
	return contract.ComponentObservation{Payload: payload, Error: errorText(err)}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
