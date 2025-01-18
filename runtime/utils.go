package runtime

import (
	"fmt"
	"strings"
)

var ErrWorkFlowNotFound = func(id string) error { return fmt.Errorf("workflow definition not found for workflow with id %s", id) }
var ErrNoPipelineFound = func(id string) error { return fmt.Errorf("pipeline not found for id %s", id) }
var ErrStepStateNotFound = func(id string) error { return fmt.Errorf("step state not found for step with id %s ", id) }
var ErrWorkflowStateNotFound = func(id string) error { return fmt.Errorf("workflow state not found for workflow with id %s", id) }
var ErrWorkflowAlreadyRegistered = func(id string, v int) error {
	return fmt.Errorf("workflow already registered with id %s and version %i", id, v)
}

func IsWorkflowNotFound(err error) bool {

	return err != nil && strings.HasPrefix(err.Error(), "workflow definition not found for workflow with id")
}

func IsWorkflowStateNotFound(err error) bool {

	return err != nil && strings.HasPrefix(err.Error(), "workflow state not found for workflow with id")
}

func IsStepStateNotFound(err error) bool {

	return err != nil && strings.HasPrefix(err.Error(), "step state not found for step with id")
}
