package modulehost

import (
	"context"

	monitoringsdk "github.com/domainry/domainry-monitoring-sdk"
)

type Factory interface {
	OpenModule(context.Context, monitoringsdk.ApplicationRef, Host) (monitoringsdk.Binding, error)
}
