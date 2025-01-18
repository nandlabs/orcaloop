package utils

import (
	"fmt"

	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop/actions"
	"oss.nandlabs.io/orcaloop/models"
)

// GetDecendantsById retrieves the IDs of all child steps for a given step ID within a workflow.
// It first fetches the step by its ID and then retrieves its child steps.
//
// Parameters:
//   - id: The ID of the step for which child steps are to be retrieved.
//   - workflow: The workflow containing the steps.
//
// Returns:
//   - ids: A slice of strings containing the IDs of the child steps.
func GetDecendantsById(id string, workflow *models.Workflow) (steps []*models.Step) {
	step := GetStepById(id, workflow)
	if step != nil {
		steps = GetDecendants(step)
	}
	return
}

// GetDecendants recursively retrieves the IDs of all child steps for a given step.
// It handles different types of steps including If, Parallel, ForLoop, and Switch steps.
//
// Parameters:
//   - step: A pointer to a models.Step object representing the current step.
//
// Returns:
//   - ids: A slice of strings containing the IDs of all child steps.
func GetDecendants(step *models.Step) (steps []*models.Step) {
	switch step.Type {
	case models.StepTypeIf:
		if step.If != nil {
			for _, subStep := range step.If.Steps {
				steps = append(steps, GetDecendants(subStep)...)
			}
			if step.If.ElseIfs != nil {
				for _, elseIf := range step.If.ElseIfs {

					for _, subStep := range elseIf.Steps {
						steps = append(steps, GetDecendants(subStep)...)
					}
				}
			}
			if step.If.Else != nil {
				for _, subStep := range step.If.Else.Steps {
					steps = append(steps, GetDecendants(subStep)...)
				}
			}
		}
	case models.StepTypeParallel:
		if step.Parallel != nil {
			for _, subStep := range step.Parallel.Steps {
				steps = append(steps, GetDecendants(subStep)...)
			}
		}
	case models.StepTypeForLoop:
		if step.For != nil {
			for _, subStep := range step.For.Steps {
				steps = append(steps, GetDecendants(subStep)...)
			}
		}
	case models.StepTypeSwitch:
		if step.Switch != nil {
			for _, caseBlock := range step.Switch.Cases {
				for _, subStep := range caseBlock.Steps {
					steps = append(steps, GetDecendants(subStep)...)
				}
			}
		}
	}

	return
}

// GetStepById retrieves a step from the workflow by its ID.
// It takes an ID as a string and a workflow of type models.Workflow.
// It returns a pointer to the models.Step that matches the given ID.
func GetStepById(id string, workflow *models.Workflow) *models.Step {
	return SearchSteps(id, workflow.Steps)
}

// SearchSteps searches for a step with the given id within a slice of steps.
// It recursively searches through nested steps in Parallel, If, For, and Switch blocks.
//
// Parameters:
//   - id: The identifier of the step to search for.
//   - steps: A slice of pointers to Step objects to search within.
//
// Returns:
//   - A pointer to the Step object with the matching id, or nil if no match is found.
func SearchSteps(id string, steps []*models.Step) *models.Step {
	for _, step := range steps {
		if step.Id == id {
			return step
		}
		if step.Parallel != nil {
			return SearchSteps(id, step.Parallel.Steps)
		}
		if step.If != nil {
			return SearchSteps(id, step.If.Steps)
		}
		if step.For != nil {
			return SearchSteps(id, step.For.Steps)
		}
		if step.Switch != nil {
			for _, caseBlock := range step.Switch.Cases {
				return SearchSteps(id, caseBlock.Steps)
			}
		}
	}
	return nil
}

// ValidateWorkflow validates the given workflow by checking if it has a name and steps.
// It returns an error if the workflow is invalid.
//
// Parameters:
//   - workflow: The workflow to be validated.
//
// Returns:
//   - err: An error if the workflow is invalid, otherwise nil.
func ValidateWorkflow(workflow models.Workflow) (err error) {
	if workflow.Name == "" {
		err = fmt.Errorf("missing name for workflow")
		return

	}
	if len(workflow.Steps) == 0 {
		err = fmt.Errorf("missing steps for workflow")
		return
	}
	for _, step := range workflow.Steps {
		err = ValidateStep(step)
		if err != nil {
			return
		}
	}
	return
}

// ValidateStep validates the configuration of a given step based on its type.
// It returns an error if any required configuration is missing or invalid.
//
// Parameters:
//   - step: A pointer to the Step object to be validated.
//
// Returns:
//   - err: An error if the step configuration is invalid, otherwise nil.
//
// Validation rules:
//   - For StepTypeAction: The Action field must not be nil.
//   - For StepTypeIf: The If field must not be nil, Condition must not be empty,
//     Steps must not be empty, and all sub-steps must be valid. ElseIf and Else
//     blocks, if present, must also be valid.
//   - For StepTypeParallel: The Parallel field must not be nil, Steps must not be
//     empty, and all sub-steps must be valid.
//   - For StepTypeForLoop: The For field must not be nil, ItemsVar or ItemsArr must
//     be provided, Loopvar or IndexVar must be provided, Steps must not be empty,
//     and all sub-steps must be valid.
//   - For StepTypeSwitch: The Switch field must not be nil, Variable must not be
//     empty, Cases must not be empty, and all case blocks must be valid.
//   - For any other step type: An error indicating an invalid step type is returned.
func ValidateStep(step *models.Step) (err error) {
	switch step.Type {
	case models.StepTypeAction:
		if step.Action == nil {
			return fmt.Errorf("missing action configuration for step %s", step.Id)
		}
	case models.StepTypeIf:
		if step.If == nil {
			return fmt.Errorf("missing if configuration for step %s", step.Id)

		}
		if step.If.Condition == "" {
			return fmt.Errorf("missing condition for if step %s", step.Id)
		}
		if len(step.If.Steps) == 0 {
			return fmt.Errorf("missing steps for if step %s", step.Id)
		}
		for _, subStep := range step.If.Steps {
			err = ValidateStep(subStep)
			if err != nil {
				return
			}
		}

		if step.If.ElseIfs != nil {

			for _, elseIf := range step.If.ElseIfs {
				if elseIf.Condition == "" {
					return fmt.Errorf("missing condition for else-if step %s", step.Id)
				}
				if len(elseIf.Steps) == 0 {
					return fmt.Errorf("missing steps for else-if step %s", step.Id)
				}
				for _, subStep := range elseIf.Steps {
					err = ValidateStep(subStep)
					if err != nil {
						return
					}
				}
			}
		}

		if step.If.Else != nil {
			if len(step.If.Else.Steps) == 0 {
				return fmt.Errorf("missing steps for else step %s", step.Id)
			}
			for _, subStep := range step.If.Else.Steps {
				err = ValidateStep(subStep)
				if err != nil {
					return
				}
			}
		}

	case models.StepTypeParallel:
		if step.Parallel == nil {
			return fmt.Errorf("missing parallel configuration for step %s", step.Id)
		}
		if len(step.Parallel.Steps) == 0 {
			return fmt.Errorf("missing steps for parallel step %s", step.Id)
		}
		for _, subStep := range step.Parallel.Steps {
			err = ValidateStep(subStep)
			if err != nil {
				return
			}
		}

	case models.StepTypeForLoop:
		if step.For == nil {
			return fmt.Errorf("missing for-loop configuration for step %s", step.Id)
		}
		if step.For.ItemsVar != "" && (len(step.For.ItemsArr) == 0 && step.For.ItemsVar == "") {
			return fmt.Errorf("missing items or itemsVar for for-loop step %s atleast one of them is required", step.Id)
		}
		if step.For.Loopvar != "" && step.For.IndexVar == "" {
			return fmt.Errorf("missing indexVar and index for for-loop step %s atleast one of them is required", step.Id)
		}

		if len(step.For.Steps) == 0 {
			return fmt.Errorf("missing steps for for-loop step %s", step.Id)
		}
		for _, subStep := range step.For.Steps {
			err = ValidateStep(subStep)
			if err != nil {
				return
			}
		}

	case models.StepTypeSwitch:
		if step.Switch == nil {
			return fmt.Errorf("missing switch configuration for step %s", step.Id)
		}
		if step.Switch.Variable == "" {
			return fmt.Errorf("missing variable for switch step %s", step.Id)
		}
		if len(step.Switch.Cases) == 0 {
			return fmt.Errorf("missing cases for switch step %s", step.Id)
		}
		for _, caseBlock := range step.Switch.Cases {
			if caseBlock.Default {
				for _, subStep := range caseBlock.Steps {
					err = ValidateStep(subStep)
					if err != nil {
						return
					}
				}

			} else {
				if caseBlock.Value == nil {
					return fmt.Errorf("missing value for case block in switch step %s", step.Id)
				}
				if len(caseBlock.Steps) == 0 {
					return fmt.Errorf("missing steps for case block in switch step %s", step.Id)
				}
				for _, subStep := range caseBlock.Steps {
					err = ValidateStep(subStep)
					if err != nil {
						return
					}
				}
			}
		}

	default:
		return fmt.Errorf("invalid Step Type %s", step.Id)
	}
	return
}

// ValidateInputs checks if all required inputs specified in the actionSpec are present in the pipeline.
// It returns a boolean indicating whether the inputs are valid and an error if any required inputs are missing.
//
// Parameters:
//   - actionSpec: A pointer to an ActionSpec struct containing the parameters to validate.
//   - pipeline: A pointer to a Pipeline struct where the inputs are checked.
//
// Returns:
//   - valid: A boolean indicating whether all required inputs are present.
//   - err: An error containing details of any missing required inputs, or nil if all inputs are valid.
func ValidateInputs(actionSpec *actions.ActionSpec, pipeline *data.Pipeline) (valid bool, err error) {
	var multiError *errutils.MultiError = errutils.NewMultiErr(nil)
	for _, input := range actionSpec.Parameters {
		if input.Required {
			if !pipeline.Has(input.Name) {
				err = fmt.Errorf("missing required input %s", input.Name)
				multiError.Add(err)
			}
		}
	}
	if multiError.HasErrors() {
		err = multiError
		valid = false
		return
	} else {
		valid = true
	}

	return
}
