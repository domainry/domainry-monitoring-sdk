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
