package service

import (
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/service/api"
)

var orcaloopServiceManager = lifecycle.NewSimpleComponentManager()

func StartService(config *config.Orcaloop) {
	// Start the service
	orcaloopServiceManager.Register(api.GetApiServer(config))

}

func StopService() {
	orcaloopServiceManager.StopAll()
}
