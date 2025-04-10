package runtime

import (
	"errors"

	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop-sdk/utils"
)

type WorkflowExecutor struct {
	storage Storage
}

func NewWorkflowExecutor(storage Storage) *WorkflowExecutor {
	return &WorkflowExecutor{storage: storage}
}

func (wfe *WorkflowExecutor) Execute(workflow *models.Workflow, pipeline *data.Pipeline) (err error) {
	var instanceId = pipeline.Id()
	var workflowState *WorkflowState
	var se = StepExecutor{storage: wfe.storage}
	// GetWorkflowState
	workflowState, err = wfe.storage.GetState(instanceId)
	if err != nil {
		return
	}
	if workflowState.Status != models.StatusRunning {
		return
	}
	var stepStates map[string][]*StepState
	stepStates, err = wfe.storage.GetStepStates(instanceId)
	if err != nil {
		return
	}
	var pendingStep *PendingStep
	pendingStep, err = wfe.storage.GetAndRemoveNextPendingStep(instanceId)
	if err != nil {
		return
	}
	logger.DebugF("Pending Step %v", pendingStep)
	if pendingStep != nil {
		//GetFirst pending Step
		// if pendingStep.VarName != "" {
		// 	pipeline.Set(pendingStep.VarName, pendingStep.VarValue)
		// }
		if pendingStep.Vars != nil {
			for k, v := range pendingStep.Vars {
				pipeline.Set(k, v)
			}
		}
		logger.DebugF("Executing pending step %v", pendingStep)
		step := utils.GetStepById(pendingStep.StepId, workflow)
		if step == nil {
			err = errors.New("Unable to find step with id " + pendingStep.StepId)
			return
		}
		pipeline.Set(data.ParentIdKey, pendingStep.ParentId)
		pipeline.Set(data.StepIterationKey, pendingStep.Iteration)
		err = se.Execute(step, pipeline)
		// if err != nil {
		// 	return
		// }
		// err = wfe.storage.DeletePendingStep(instanceId, pendingStep)
		return
	} else {
		// No pending step
		logger.DebugF("No pending step found, setting iteration to 0")
		pipeline.Set(data.StepIterationKey, 0)
		// Save the pipeline
		err = wfe.storage.SavePipeline(pipeline)
		if err != nil {
			return
		}
	}
	// create empty array
	var workflowStepsArr = make([]*models.Step, 0)
	for _, step := range workflow.Steps {
		decendents := utils.GetDecendants(step)
		workflowStepsArr = append(workflowStepsArr, decendents...)
		workflowStepsArr = append(workflowStepsArr, step)
	}
	for _, step := range workflowStepsArr {
		stepStateArr, ok := stepStates[step.Id]
		if ok {
			for _, stepState := range stepStateArr {
				logger.DebugF("Checking for Step %s with state %v", step.Id, stepState)
				switch stepState.Status {
				case models.StatusCompleted, models.StatusSkipped:
					continue
				case models.StatusFailed:
					logger.DebugF("Step %s failed aborting the workflow", step.Id)
					return
				case models.StatusRunning:
					var completedChildren int
					// var stepState = stepStates[step.Id]
					var childError string
					for _, v := range stepStates {
						for _, potentialChildStep := range v {
							if potentialChildStep.ParentStep == step.Id && potentialChildStep.Status > models.StatusRunning {

								completedChildren++
								if potentialChildStep.Status == models.StatusFailed {
									childError = potentialChildStep.Output.GetError()
									break
								}
							}
						}
					}
					if completedChildren == stepState.ChildCount {
						logger.DebugF("All children of step %s completed proceeding to next step", step.Id)
						if childError != textutils.EmptyStr {
							logger.DebugF("Child failed for step %s aborting the workflow", step.Id)
							stepState.Status = models.StatusFailed
							err = wfe.storage.SaveStepState(stepState)
							if err != nil {
								return
							}
							workflowState.Status = models.StatusFailed
							err = wfe.storage.SaveState(workflowState)
							return
						} else {
							stepState.Status = models.StatusCompleted
							err = wfe.storage.SaveStepState(stepState)
							if err != nil {
								return
							}
							continue
						}
					} else {
						logger.DebugF("Not all children of step %s completed waiting for them to complete", step.Id)
						break
					}
				}
			}
		} else {
			err = se.Execute(step, pipeline)
			return
		}
	}
	// This is possible only if all steps are completed
	workflowState.Status = models.StatusCompleted
	err = wfe.storage.SaveState(workflowState)
	return
}
