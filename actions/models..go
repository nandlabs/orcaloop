package actions

import (
	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

// ActionSpec represents the specification of an action.
// It contains metadata and configuration for the action.
//
// Fields:
//   - Id: The unique identifier of the action.
//   - Name: The name of the action.
//   - Description: A brief description of the action.
//   - Parameters: A list of schemas representing the parameters of the action.
//   - Returns: A list of schemas representing the return values of the action.
//   - Async: A flag indicating whether the action is asynchronous.
//   - Endpoint: The endpoint configuration for the action.
type ActionSpec struct {
	// Id is the id of the action
	Id string `json:"id" yaml:"id"`
	// Name is the name of the action
	Name string `json:"name" yaml:"name"`
	// Description is the description of the action
	Description string `json:"description" yaml:"description"`
	//Parameters is the parameters of the action
	Parameters []models.Schema `json:"parameters" yaml:"parameters"`
	// Returns is the returns of the action
	Returns []models.Schema `json:"returns" yaml:"returns"`
	// Async  is the async flag of the action
	Async bool `json:"async" yaml:"async"`
	// Endpoint is the endpoint of the action
	Endpoint *Endpoint `json:"endpoint" yaml:"endpoint"`
}

// Endpoint represents a network endpoint configuration.
//
// Fields:
//   - Type: Specifies the type of the Endpoint.
//   - Local: Represents the local endpoint configuration.
//   - Rest: Represents the REST endpoint configuration.
//   - Messaging: Represents the messaging endpoint configuration.
//   - Grpc: Represents the gRPC endpoint configuration.
//   - Qos: Represents the quality of service configuration.
type Endpoint struct {

	// type is the type of the Endpoint
	Type string `json:"type" yaml:"type"`
	//Local is the local endpoint
	Builtin *Builtin `json:"builtin" yaml:"builtin"`
	//Rest is the rest endpoint
	Rest *RestEndpoint `json:"rest" yaml:"rest"`
	//Messaging is the messaging endpoint
	Messaging *MessagingEndpoint `json:"messaging" yaml:"messaging"`
	//Grpc is the grpc endpoint
	Grpc *GrpcEndpoint `json:"grpc" yaml:"grpc"`
	//Qos is the quality of service
	Qos *Qos `json:"qos" yaml:"qos"`
}

// Qos represents the Quality of Service settings for a particular action.
// It includes the number of retries, timeout duration, and circuit breaker information.
type Qos struct {
	// Retries is the number of retries
	Retries int
	// Timeout is the timeout
	Timeout int
	// CircuitBreakerInfo is the circuit breaker info
	BreakerInfo *clients.BreakerInfo
}

type Builtin struct {
}

type RestEndpoint struct {
	// Url is the url of the rest endpoint
	// The format is http[s]://host:[port]/[pathname|$pathParam]*/?[name=value&]*
	// The $pathParam is the path param name from the pipeline
	Url string `json:"url" yaml:"url"`
	// AuthProvider
	AuthProvider string `json:"authProvider" yaml:"authProvider"`
}

// MessagingEndpoint represents the configuration for a messaging endpoint.
// It contains the URL, request headers, body, and body MIME type for the messaging provider.
// See https://pkg.go.dev/oss.nandlabs.io/golly/messaging#Provider for more information for the messaging provider.
type MessagingEndpoint struct {
	// Url is the URL of the messaging provider.
	// The format is <provider_scheme>://destination.
	// The provider_scheme is the scheme of the messaging provider.
	// The destination is the destination of the messaging provider.
	Url string `json:"url" yaml:"url"`
}

type GrpcEndpoint struct {
	// TODO
}
