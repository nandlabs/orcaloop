package actions

import (
	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop/models"
)

// Action represents an interface for defining actions in the system.
// Each action is expected to have a unique identifier, name, description,
// input and output schemas, and an execution method. Additionally, actions
// can specify whether they are asynchronous and provide a specification
// and provider information.
type Action interface {
	// Id returns the id of the action. This is expected to be unique.
	Id() string
	// Name returns the name of the action
	Name() string
	// Description returns the description of the action
	Description() string
	// Inputs returns the inputs of the action
	Inputs() []models.Schema
	// Outputs returns the outputs of the action
	Outputs() []models.Schema
	// Execute function executes the action
	Execute(*data.Pipeline) error
	// Spec returns the spec of the action
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
