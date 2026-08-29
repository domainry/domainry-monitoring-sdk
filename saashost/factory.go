package saashost

import (
	"context"

	monitoringsdk "github.com/domainry/domainry-monitoring-sdk"
	"github.com/domainry/domainry-monitoring-sdk/modulehost"
)

// Factory opens a remote Monitoring Binding with access to the local
// observation host. Raw owner facts cross the wire; Runtime stores do not.
type Factory interface {
	OpenSaaS(context.Context, monitoringsdk.ApplicationRef, modulehost.Host) (monitoringsdk.Binding, error)
}
