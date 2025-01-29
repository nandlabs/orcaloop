package service

import (
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/service/api"
)

var orcaloopServiceManager = lifecycle.NewSimpleComponentManager()

func Init(config *config.Orcaloop) (err error) {
	err = api.RegisterServer(config, orcaloopServiceManager)
	return
}

func StartService() (err error) {

	// Start Server
	orcaloopServiceManager.StartAll()
	return
}

func StartAndWait() (err error) {

	orcaloopServiceManager.StartAndWait()
	return
}

func StopService() {
	orcaloopServiceManager.StopAll()
}
