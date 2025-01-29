package service

import (
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/service/api"
)

var orcaloopServiceManager = lifecycle.NewSimpleComponentManager()

func StartService(config *config.Orcaloop) (err error) {
	//Register Server
	err = api.RegisterServer(config, orcaloopServiceManager)
	if err != nil {
		return
	}
	// Start Server
	orcaloopServiceManager.StartAll()
	return
}

func StopService() {
	orcaloopServiceManager.StopAll()
}
