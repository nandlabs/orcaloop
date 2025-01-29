package api

import (
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/runtime"
)

// GetApiServer creates a new Orcaloop service Api Server
func RegisterServer(options *config.Orcaloop, manager lifecycle.ComponentManager) (err error) {
	var storage runtime.Storage
	var server rest.Server
	server, err = rest.NewServer(options.ApiSrvConfig)
	if err != nil {
		return
	}
	storage, err = runtime.GetStorage(options.StorageConfig)
	if err != nil {
		return
	}
	// Register the workflow service
	resthandler := NewRestHandler(storage, manager)
	resthandler.RegisterRoutes(server)
	manager.Register(server)
	return
}
