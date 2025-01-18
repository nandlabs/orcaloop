package actions

import "oss.nandlabs.io/golly/clients"

type Endpoint struct {

	// type is the type of the Endpoint
	Type string `json:"type" yaml:"type"`
	//Local is the local endpoint
	Local *LocalEndpoint `json:"local" yaml:"local"`
	//Rest is the rest endpoint
	Rest *RestEndpoint `json:"rest" yaml:"rest"`
	//Messaging is the messaging endpoint
	Messaging *MessagingEndpoint `json:"messaging" yaml:"messaging"`
	//Grpc is the grpc endpoint
	Grpc *GrpcEndpoint `json:"grpc" yaml:"grpc"`
	//Qos is the quality of service
	Qos *Qos `json:"qos" yaml:"qos"`
}

type Qos struct {
	// Retries is the number of retries
	Retries int
	// Timeout is the timeout
	Timeout int
	// CircuitBreakerInfo is the circuit breaker info
	BreakerInfo *clients.BreakerInfo
}

type LocalEndpoint struct {
}

type RestEndpoint struct {
}
type MessagingEndpoint struct {
}

type GrpcEndpoint struct {
	// TODO
}
