package service

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

// NewOrcaloopComponent creates a new Orcaloop component
func NewOrcaloopComponent(options *config.Orcaloop) lifecycle.Component {
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
