package runtime

import (
	"fmt"
	"sync"

	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/orcaloop/data"
	"oss.nandlabs.io/orcaloop/models"
	"oss.nandlabs.io/orcaloop/storage"
	"oss.nandlabs.io/orcaloop/utils"
)

type WorkflowExecutor struct {
	store storage.Storage
}

// Execute runs the workflow
// Execute runs the workflow instance with the given instanceId.
// It retrieves the workflow state and pipeline, then processes each step in the workflow.
// If a step is not started, it starts the step. If a step is completed, it moves to the next step.
// If a step fails, it marks the workflow as failed and aborts further execution.
// If a step is running, it checks the status of its child steps and waits for them to complete if necessary.
// Returns an error if any operation fails.
//
// Parameters:
// - instanceId: The ID of the workflow instance to execute.
//
// Returns:
// - err: An error if any operation fails, otherwise nil.
func (w *WorkflowExecutor) Execute(instanceId string) (err error) {
	var workflowState *models.WorkflowState
	var pipeline *data.Pipeline
	// Get the workflow state
	workflowState, err = w.store.GetState(instanceId)
	if err != nil {
		return
	}
	pipeline, err = w.store.GetPipeline(workflowState.Workflow.Id)
	if err != nil {
		return
	}
	if workflowState == nil {
		err = ErrWorkflowStateNotFound(instanceId)
		return
	}
	if workflowState.Status == models.StatusCompleted || workflowState.Status == models.StatusFailed {
		return
	}
	// Get the next step
workFlowStepsLoop:
	for _, step := range workflowState.Workflow.Steps {
		stepState := workflowState.StepStates[step.Id]
		if stepState == nil {
			// Start the step
			err = w.executeStep(step, textutils.EmptyStr, workflowState, pipeline)
			if err != nil {
				return
			}
			break workFlowStepsLoop
		}
		switch stepState.Status {
		case models.StatusCompleted:
			logger.DebugF("Step %s is already completed moving to next step", step.Id)
			continue
		case models.StatusFailed:
			logger.DebugF("Step %s failed aborting the execution of workflow instance ", step.Id, instanceId)

			if workflowState.Status != models.StatusFailed {
				workflowState.Status = models.StatusFailed
				workflowState.Error = fmt.Sprintf("Step %s failed for instance id %s", step.Id, instanceId)
				// No Locking is requried here.
				err = w.store.SaveState(workflowState)
				if err != nil {
					return
				}
			}
			break workFlowStepsLoop
		case models.StatusSkipped:
			logger.DebugF("Step %s is skipped", step.Id)
			continue
		case models.StatusRunning:
			children := utils.GetDecendants(step)
			if len(children) == 0 {
				logger.DebugF("Step %s is running and has no children and has status Running will wait for it to be completed", step.Id)
				break workFlowStepsLoop
			}
			logger.DebugF("Step %s is running and has children checking for all decendants", step.Id)
			var child *models.Step
			var completedChildren int
		childStepLoop:
			for _, child = range children {
				logger.DebugF("Checking child step %s", child.Id)
				if childState, ok := workflowState.StepStates[child.Id]; ok {
					switch childState.Status {
					case models.StatusCompleted:
						logger.DebugF("Child step %s is completed moving to next child step", child.Id)
						completedChildren++
						continue
					case models.StatusFailed:
						logger.DebugF("Step %s failed no futher execution will be processed", child.Id)
						break childStepLoop
					case models.StatusSkipped:
						logger.DebugF("Step %s is skipped will continue to check next child", child.Id)
						completedChildren++
						continue
					case models.StatusRunning:
						logger.DebugF("Step %s is running will wait for it to be completed", child.Id)
						break childStepLoop
					default:
						logger.DebugF("Starting step with id %s as its pending", child.Id)
						err = w.executeStep(child, step.Id, workflowState, pipeline)
						if err != nil {
							return
						}
					}

				}
			}
			if completedChildren == stepState.ChildCount {
				logger.DebugF("All child steps of step %s are completed moving to next step", step.Id)
				stepState.Status = models.StatusCompleted
				err = w.store.SaveStepState(stepState)
				if err != nil {
					return
				}

			} else {
				logger.DebugF("Step %s  has child steps still in non completed state will wait for them to be completed", step.Id)
				break workFlowStepsLoop
			}

		}

	}
	return nil
}

// executeStep executes a given step within a workflow. It handles different types of steps
// such as action, if, parallel, for loop, and switch steps. It also manages the state of the
// step and updates the workflow state accordingly.
//
// Parameters:
//   - step: The step to be executed.
//   - parentStepId: The ID of the parent step.
//   - workflowState: The current state of the workflow.
//   - pipeline: The pipeline data associated with the workflow.
//
// Returns:
//   - err: An error if the step execution fails, otherwise nil.
func (w *WorkflowExecutor) executeStep(step *models.Step, parentStepId string, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {

	var stepState *models.StepState

	defer func() {
		if err != nil {
			respData := data.NewPipeline(workflowState.Id)
			respData.SetError(err.Error())
			err1 := w.store.SaveStepChangeEvent(&models.StepChangeEvent{
				InstanceId: workflowState.Id,
				StepId:     step.Id,
				Status:     models.StatusFailed,
				Data:       respData,
			})

			if err1 != nil {
				logger.ErrorF("Failed to execute step for instance %s and step %s", workflowState.Id, step.Id)
				multiErr := errutils.NewMultiErr(err)
				multiErr.Add(err1)
				err = multiErr
			}
		}
	}()

	stepState = &models.StepState{
		InstanceId: workflowState.Id,
		StepId:     step.Id,
		ParentStep: parentStepId,
		Status:     models.StatusRunning,
	}
	if step.Skip {
		stepState.Status = models.StatusSkipped
	}

	// Save the step state
	err = w.store.SaveStepState(stepState)
	if err != nil {
		return
	}
	if step.Skip {
		return
	}

	switch step.Type {
	case models.StepTypeAction:
		err = w.executeAction(step, workflowState, pipeline)
	case models.StepTypeIf:
		err = w.executeIfStep(step, workflowState, pipeline)
	case models.StepTypeParallel:
		err = w.executeParallelStep(step, workflowState, pipeline)
	case models.StepTypeForLoop:
		err = w.executeForLoopStep(step, workflowState, pipeline)
	case models.StepTypeSwitch:
		err = w.executeSwitchStep(step, workflowState, pipeline)
	default:
		err = fmt.Errorf("unsupported step type %s", step.Type)
	}
	return
}

// executeAction executes a specified action within a workflow step.
// It retrieves the action specification from the store and prepares the input parameters
// for the action based on the step configuration and the current pipeline state.
//
// Parameters:
//   - step: The workflow step containing the action to be executed.
//   - workflowState: The current state of the workflow.
//   - pipeline: The pipeline data used to resolve input parameters for the action.
//
// Returns:
//   - err: An error if the action configuration is missing, the action specification
//     cannot be retrieved, or if there is an issue resolving input parameters.
func (w *WorkflowExecutor) executeAction(step *models.Step, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {
	var actionStep *models.StepAction = step.Action
	if actionStep == nil {
		return fmt.Errorf("missing action configuration for step %s", step.Id)
	}
	actionSpec, err := w.store.ActionSpec(actionStep.Id)
	if err != nil {
		return
	}
	if actionSpec == nil {
		return fmt.Errorf("action %s not found", actionStep.Id)
	}

	// Execute the action
	actionInput := data.NewPipeline(workflowState.Id)
	for _, param := range actionStep.Parameters {
		if param.Name == "" {
			continue
		}
		if param.Var != "" {
			value, err := pipeline.Get(param.Var)
			if err != nil {
				return err
			}
			actionInput.Set(param.Name, value)
		} else if param.Value != nil {
			actionInput.Set(param.Name, param.Value)
		} else {
			// check if this
			actionInput.Set(param.Name, nil)
		}
	}

	return
}

// executeParallelStep executes a parallel step in the workflow.
// It takes a step, workflow state, and pipeline as arguments.
// The function checks if the parallel configuration is present for the step.
// If not, it returns an error. It then sets the child count for the step
// and initializes a wait group to manage the parallel execution of sub-steps.
// Each sub-step is executed in a separate goroutine, and any errors encountered
// are collected using a multi-error utility. The function waits for all sub-steps
// to complete before returning any errors encountered during execution.
//
// Parameters:
// - step: The step to be executed in parallel.
// - workflowState: The current state of the workflow.
// - pipeline: The pipeline data.
//
// Returns:
// - err: An error if the parallel step execution fails.
func (w *WorkflowExecutor) executeParallelStep(step *models.Step, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {
	parallel := step.Parallel
	if parallel == nil {
		return fmt.Errorf("missing parallel configuration for step %s", step.Id)
	}
	waitGroup := sync.WaitGroup{}
	workflowState.StepStates[step.Id].ChildCount = len(parallel.Steps)
	err = w.setChildrenCount(step.Id, len(parallel.Steps), workflowState)
	if err != nil {
		return
	}

	for _, subStep := range parallel.Steps {
		waitGroup.Add(1)
		multiError := errutils.NewMultiErr(nil)
		go func() {
			err = w.executeStep(subStep, step.Id, workflowState, pipeline)
			if err != nil {
				multiError.Add(err)
			}
			waitGroup.Done()
		}()
		waitGroup.Wait()
		if multiError.HasErrors() {
			err = multiError
		}
		if err != nil {
			return
		}
	}
	return
}

// executeForLoopStep executes a for-loop step within a workflow.
// It retrieves the items to iterate over from the pipeline, sets the loop and index variables,
// and executes each sub-step for each item in the loop.
//
// Parameters:
//   - step: The current step to be executed, which contains the for-loop configuration.
//   - workflowState: The current state of the workflow.
//   - pipeline: The data pipeline used to retrieve and set variables.
//
// Returns:
//   - err: An error if the execution fails, or nil if it succeeds.
func (w *WorkflowExecutor) executeForLoopStep(step *models.Step, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {
	var loop *models.For = step.For
	if loop == nil {
		return fmt.Errorf("missing for-loop configuration for step %s", step.Id)
	}

	items, err := pipeline.Get(loop.ItemsVar)
	if err != nil {
		return err
	}
	err = w.setChildrenCount(step.Id, len(items.([]any)), workflowState)
	if err != nil {
		return
	}

	for index, item := range items.([]any) {
		if loop.Loopvar != "" {
			pipeline.Set(loop.Loopvar, item)
		}
		if loop.IndexVar != "" {
			pipeline.Set(loop.IndexVar, index)
		}

		for _, subStep := range loop.Steps {
			if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
				return
			}

		}
	}

	return
}

// executeIfStep executes a conditional step within a workflow. It evaluates the condition
// specified in the step and executes the corresponding steps if the condition is met.
// If the condition is not met, it evaluates any ElseIf conditions and executes the corresponding
// steps if any of those conditions are met. If none of the conditions are met, it executes the
// steps specified in the Else block, if present.
//
// Parameters:
// - step: The current step to be executed, which contains the conditional logic.
// - workflowState: The current state of the workflow.
// - pipeline: The pipeline data used to evaluate conditions.
//
// Returns:
// - err: An error if any occurs during the execution of the step or its sub-steps.
func (w *WorkflowExecutor) executeIfStep(step *models.Step, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {
	var ifStep *models.If = step.If
	var conditionMet bool
	if ifStep == nil {
		return fmt.Errorf("missing if configuration for step %s", step.Id)
	}

	conditionMet, err = pipeline.EvaluateCondition(ifStep.Condition)
	if err != nil {
		return err
	}
	if conditionMet {
		err = w.setChildrenCount(step.Id, len(ifStep.Steps), workflowState)
		if err != nil {
			return
		}
		for _, subStep := range ifStep.Steps {
			if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
				return
			}
		}
	} else if ifStep.ElseIfs != nil {
		for _, elseIf := range ifStep.ElseIfs {
			conditionMet, err = pipeline.EvaluateCondition(elseIf.Condition)
			if err != nil {
				return err
			}
			if conditionMet {
				err = w.setChildrenCount(step.Id, len(elseIf.Steps), workflowState)
				if err != nil {
					return
				}

				for _, subStep := range elseIf.Steps {
					if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
						return err
					}
				}
				return
			}
		}
	} else if ifStep.Else != nil {
		err = w.setChildrenCount(step.Id, len(ifStep.Else.Steps), workflowState)
		if err != nil {
			return
		}

		for _, subStep := range ifStep.Else.Steps {
			if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
				return err
			}
		}
	}
	return
}

// executeSwitchStep executes a switch step within a workflow. It evaluates the
// value of a specified variable and executes the corresponding case block steps
// if a match is found. If no match is found, it executes the default case block
// steps if defined.
//
// Parameters:
//   - step: The current step to be executed, which contains the switch configuration.
//   - workflowState: The current state of the workflow.
//   - pipeline: The data pipeline used to retrieve the variable value.
//
// Returns:
//   - err: An error if the execution fails, otherwise nil.
func (w *WorkflowExecutor) executeSwitchStep(step *models.Step, workflowState *models.WorkflowState, pipeline *data.Pipeline) (err error) {

	var switchStep *models.Switch = step.Switch
	var value any
	if switchStep == nil {
		return fmt.Errorf("missing switch configuration for step %s", step.Id)
	}

	value, err = pipeline.Get(switchStep.Variable)
	if err != nil {
		return err
	}

	var defaultCase *models.Case

	for _, caseBlock := range switchStep.Cases {
		if caseBlock.Default {
			defaultCase = caseBlock
			continue
		}
		// Compare the value with the case block value
		if value == caseBlock.Value {
			err = w.setChildrenCount(step.Id, len(caseBlock.Steps), workflowState)
			if err != nil {
				return
			}
			// Execute the steps in the case block if the value matches
			for _, subStep := range caseBlock.Steps {
				if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
					return err
				}
			}
			// Exit the switch statement if a match is found
			return nil
		}
	}
	// Execute default case if no match found
	if defaultCase != nil {
		err = w.setChildrenCount(step.Id, len(defaultCase.Steps), workflowState)
		if err != nil {
			return
		}
		for _, subStep := range defaultCase.Steps {
			if err = w.executeStep(subStep, step.Id, workflowState, pipeline); err != nil {
				return err
			}
		}
	}
	return
}

// setChildrenCount sets the number of children for a given parent step in the workflow state.
// If the parent step ID is empty, the function returns without making any changes.
//
// Parameters:
//   - parentStepId: The ID of the parent step whose child count is to be set.
//   - count: The number of children to set for the parent step.
//   - workflowState: A pointer to the WorkflowState object that contains the state of the workflow.
//
// Returns:
//   - err: An error if there is an issue saving the step state, otherwise nil.
func (w WorkflowExecutor) setChildrenCount(parentStepId string, count int, workflowState *models.WorkflowState) (err error) {
	if parentStepId == "" {
		return
	}
	parentStepState := workflowState.StepStates[parentStepId]
	parentStepState.ChildCount = count
	err = w.store.SaveStepState(parentStepState)
	return

}

// OnStepChange handles the step change event for a workflow instance.
// It locks the instance, processes the step change event, and executes the workflow if needed.
// If there are pending step change events, it processes them in a loop until there are no more pending events.
//
// Parameters:
//   - stpChgEvt: A pointer to the StepChangeEvent model containing the details of the step change event.
//
// Returns:
//   - err: An error if any occurs during the processing of the step change event or execution of the workflow.
func (w *WorkflowExecutor) OnStepChange(stpChgEvt *models.StepChangeEvent) (err error) {

	var lock bool
	var resume bool

	// Lock the instance
	lock, err = w.store.LockInstance(stpChgEvt.InstanceId)
	if err != nil {
		return
	}
	if !lock {
		err = w.store.SaveStepChangeEvent(stpChgEvt)
		return
	} else {
		// Unlock the instance
		defer func(instanceId string) {
			err = w.store.UnlockInstance(instanceId)
			if err != nil {
				logger.ErrorF("Failed to unlock instance %s", instanceId)
			}
			// Do we need to check for pending events again here as there may be new events just before locking ???
		}(stpChgEvt.InstanceId)
	}
	// Handle the step change event
	resume, err = w.handleStepChange(stpChgEvt)
	if err != nil {
		return
	}
	if resume {
		err = w.Execute(stpChgEvt.InstanceId)
	}
	if err != nil {
		return
	}
	err = w.checkPendingStates(stpChgEvt.InstanceId)

	return

}

func (w WorkflowExecutor) checkPendingStates(instanceId string) (err error) {
	var pendingEvents []*models.StepChangeEvent
	resume := false
	for {
		// Get the pending events
		pendingEvents, err = w.store.GetStepChangeEvents(instanceId)
		if err != nil {
			return
		}
		if len(pendingEvents) == 0 {
			break
		}
		for _, evt := range pendingEvents {
			// Handle the step change event
			resume, err = w.handleStepChange(evt)
			if err != nil {
				return
			}
			if resume {
				err = w.Execute(evt.InstanceId)
			}
			if err != nil {
				return
			}
		}
	}
	return
}

// handleStepChange handles the step change event for a workflow executor.
// It updates the workflow and step states based on the status of the step change event.
//
// Parameters:
//   - stpChgEvt: A pointer to the StepChangeEvent containing the details of the step change.
//
// Returns:
//   - resume: A boolean indicating whether the workflow should resume execution.
//   - err: An error if any occurred during the handling of the step change event.
//
// The function handles the following step statuses:
//   - StatusCompleted: Marks the step as completed, merges the step output to the workflow output, and saves the step state.
//   - StatusFailed: Marks the step and workflow as failed, updates the workflow error message, and saves the workflow state.
//   - StatusSkipped: Marks the step as skipped and saves the step state.
//   - Default: Logs a debug message indicating that no action will be executed for the received status.
func (w *WorkflowExecutor) handleStepChange(stpChgEvt *models.StepChangeEvent) (resume bool, err error) {
	var workflowState *models.WorkflowState
	var stepState *models.StepState
	var worflowPipeline *data.Pipeline
	switch stpChgEvt.Status {

	case models.StatusCompleted:
		// Get the Step state
		workflowState, err = w.store.GetState(stpChgEvt.InstanceId)
		if err != nil {
			return
		}
		if workflowState == nil {
			err = ErrWorkflowStateNotFound(stpChgEvt.InstanceId)
			return
		}
		stepState = workflowState.StepStates[stpChgEvt.StepId]
		if stepState == nil {
			err = ErrStepStateNotFound(stpChgEvt.StepId)
			return
		}
		stepState.Status = models.StatusCompleted
		stepState.Output = stpChgEvt.Data
		worflowPipeline, err = w.store.GetPipeline(workflowState.Workflow.Id)
		if err != nil {
			return
		}
		if worflowPipeline == nil {
			err = ErrNoPipelineFound(workflowState.Workflow.Id)
			return
		}
		// Merge step output to workflow output
		worflowPipeline.Merge(stepState.Output)
		// Save the pipeline
		err = w.store.SavePipeline(worflowPipeline)
		// Save the step state
		err = w.store.SaveStepState(stepState)
		resume = true

	case models.StatusFailed:
		// Get the workflow state
		workflowState, err = w.store.GetState(stpChgEvt.InstanceId)
		if err != nil {
			return
		}
		if workflowState == nil {
			err = ErrWorkflowStateNotFound(stpChgEvt.InstanceId)
			return
		}
		stepState = workflowState.StepStates[stpChgEvt.StepId]
		if stepState == nil {
			err = ErrStepStateNotFound(stpChgEvt.StepId)
			return
		}
		stepState.Status = models.StatusFailed
		stepState.Output = stpChgEvt.Data
		// Update the step state
		workflowState.Status = models.StatusFailed

		errStr := stpChgEvt.Data.GetError()
		if errStr != "" {
			workflowState.Error = errStr
		} else {
			workflowState.Error = fmt.Sprintf("Step %s failed for instance id %s", stpChgEvt.StepId, stpChgEvt.InstanceId)
		}
		// check if the step has a parent step
		if parentStepState, ok := workflowState.StepStates[stepState.ParentStep]; ok {
			// Mark Parent step as failed
			parentStepState.Status = models.StatusFailed
		}
		// May be we can merge the pipeline here to the workflow pipeline ???

		// Save the step state
		err = w.store.SaveState(workflowState)
		if err != nil {
			return
		}
	case models.StatusSkipped:
		// Get the Step state
		stepState, err = w.store.GetStepState(stpChgEvt.InstanceId, stpChgEvt.StepId)
		if err != nil {
			return
		}

		if stepState == nil {
			err = ErrStepStateNotFound(stpChgEvt.StepId)
			return
		}
		stepState.Status = models.StatusSkipped
		// Save the step state
		err = w.store.SaveStepState(stepState)
		if err != nil {
			return
		}
		resume = true
	default:
		logger.DebugF("Received action for step %s with status %s no action will be executed.", stpChgEvt.StepId, stpChgEvt.Status)
	}

	return

}
