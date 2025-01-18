package actions

import (
	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/orcaloop/data"
	"oss.nandlabs.io/orcaloop/models"
)

// ActionSpec is a specification of Action
type ActionSpec struct {
	// Id is the id of the tool
	Id string `json:"id" yaml:"id"`
	// Name is the name of the tool
	Name string `json:"name" yaml:"name"`
	// Description is the description of the tool
	Description string `json:"description" yaml:"description"`
	//Parameters is the parameters of the tool
	Parameters []models.Schema `json:"parameters" yaml:"parameters"`
	// Returns is the returns of the tool
	Returns []models.Schema `json:"returns" yaml:"returns"`
	// Async  is the async flag of the tool
	Async bool `json:"async" yaml:"async"`
	// Endpoint is the endpoint of the tool
	Endpoint *Endpoint `json:"endpoint" yaml:"endpoint"`
}

// Action represents an interface for defining actions in the system.
// Each action is expected to have a unique identifier, name, description,
// input and output schemas, and an execution method. Additionally, actions
// can specify whether they are asynchronous and provide a specification
// and provider information.
type Action interface {
	// Id returns the id of the tool. This is expected to be unique.
	Id() string
	// Name returns the name of the tool
	Name() string
	// Description returns the description of the tool
	Description() string
	// Inputs returns the inputs of the tool
	Inputs() []models.Schema
	// Outputs returns the outputs of the tool
	Outputs() []models.Schema
	// Execute function executes the action
	Execute(*data.Pipeline) error
	// Spec returns the spec of the tool
	Spec() *ActionSpec
	// IsAsync returns true if the action is asynchronous
	IsAsync() bool
	// Provider returns the provider of the action
	Provider() string
}

// ActionSpecs retrieves the specifications of all actions managed by the given action manager.
// It iterates over the items in the action manager, calls the Specification method on each action,
// and appends the resulting ActionSpec to a slice, which is then returned.
//
// Parameters:
//   - actionManager: An instance of ItemManager that manages Action items.
//
// Returns:
//   - A slice of pointers to ActionSpec, each representing the specification of an action.
func ActionSpecs(actionManager managers.ItemManager[Action]) []*ActionSpec {
	var actions []*ActionSpec

	for _, action := range actionManager.Items() {
		actions = append(actions, action.Spec())
	}
	return actions
}

// DescribeActions encodes the specifications of actions managed by the given action manager
// into a string format specified by the provided format string.
//
// Parameters:
//   - format: A string specifying the desired encoding format.
//   - actionManager: An ItemManager instance managing Action items.
//
// Returns:
//   - val: A string containing the encoded action specifications.
//   - err: An error if the encoding process fails, otherwise nil.
func DescribeActions(format string, actionManager managers.ItemManager[Action]) (val string, err error) {
	var c codec.Codec

	c, err = codec.GetDefault(format)
	if err != nil {
		return
	}
	val, err = c.EncodeToString(ActionSpecs(actionManager))
	return
}
