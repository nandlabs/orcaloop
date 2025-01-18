package service

import "oss.nandlabs.io/golly/lifecycle"

var serverManager = lifecycle.NewSimpleComponentManager()

func StartService() {
	// Start the service

}

func StopService() {
	serverManager.StopAll()
}
