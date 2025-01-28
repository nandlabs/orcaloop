package api

import (
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/orcaloop/config"
)

// Orcaloop Component
// implements lifcycle.Component
type orcaLoopComponent struct {
	*lifecycle.SimpleComponent
	options *config.Orcaloop
}

// GetApiServer creates a new Orcaloop component
func GetApiServer(options *config.Orcaloop) lifecycle.Component {
	return &orcaLoopComponent{
		SimpleComponent: &lifecycle.SimpleComponent{
			CompId: options.Name,
			StartFunc: func() (err error) {

				return
			},
			AfterStart: func(err error) {

			},

			StopFunc: func() (err error) {
				return
			},
		},
		options: options,
	}
}
