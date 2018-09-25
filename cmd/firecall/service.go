package main

import (
	"bitbucket.org/seibert-media/events/pkg/service"
	"github.com/seibert-media/golibs/log"
)

const (
	appName = "FireCall"
	appKey  = "firecall"
)

// Spec for the service
type Spec struct {
	service.BaseSpec
}

func main() {
	var svc Spec
	ctx := service.Init(appKey, appName, &svc)
	defer service.Defer(ctx)

	log.From(ctx).Info("finished")
}
