# Domainry Monitoring SDK

Deployment-neutral contracts for Runtime monitoring.

- `modulehost.Host` exposes sanitized owner observations and live readiness probes.
- `modulehost.Factory` opens the in-process Module topology.
- `saashost.Factory` opens the SaaS topology while borrowing the same local observation host.
- `remote.Factory` performs protocol negotiation and sends bounded observation snapshots to the Monitoring SaaS service.

SaaS configuration:

- `DOMAINRY_MONITORING_ENDPOINT`: Monitoring service base URL.
- `DOMAINRY_MONITORING_TOKEN`: bearer credential shared with the service.
- `DOMAINRY_MONITORING_TIMEOUT`: optional Go duration, default `10s`.

Runtime stores and credentials are never exposed through the host contract. Only owner-produced monitoring observations cross the boundary.

## Package layout

- The root package is the stable Monitoring `Factory` and `Binding` entrypoint.
- `contract` contains monitoring observations and deployment-neutral values.
- `modulehost` describes sanitized observations, readiness, and embedded storage infrastructure.
- `remote` implements the SaaS client; `saashost` describes SaaS composition.

The SDK intentionally has no public `persistence` package. Monitoring storage is an embedded-host capability, not a Runtime-consumable repository.

Run `go test ./...` before publishing an immutable SDK version.
