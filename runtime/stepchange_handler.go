package runtime

import (
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/events"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type StepChangeHander struct {
	storage Storage
}

func (sh *StepChangeHander) Handle(stepChangeEvent *events.StepChangeEvent) (err error) {
	var lock bool
	if err != nil {
		return
	}
	// Lock the instance
	lock, err = sh.storage.LockInstance(stepChangeEvent.InstanceId)
	if err != nil {
		return
	}
	if lock {
		defer func() {
			logger.DebugF("Unlocking instance %s", stepChangeEvent.InstanceId)
			var pendingStepChangeEvents []*events.StepChangeEvent
			// Get all pending step change events
			for {
				pendingStepChangeEvents, err = sh.storage.GetStepChangeEvents(stepChangeEvent.InstanceId)
				if err != nil {
					return
				}
				if len(pendingStepChangeEvents) == 0 {
					break
				}
				for _, pendingStepChangeEvent := range pendingStepChangeEvents {
					err = sh.processStepChange(pendingStepChangeEvent)
					if err != nil {
						return
					}
					err = sh.storage.DeleteStepChangeEvent(pendingStepChangeEvent.InstanceId, pendingStepChangeEvent.EventId)
					if err != nil {
						return
					}
				}
				if err != nil {
					return
				}
			}
			// unlock instance at the end
			err = sh.storage.UnlockInstance(stepChangeEvent.InstanceId)
			logger.DebugF("Instance %s unlocked with error %v", stepChangeEvent.InstanceId, err)
		}()
		err = sh.processStepChange(stepChangeEvent)
	} else {
		// Save the event as the instance is already locked
		err = sh.storage.SaveStepChangeEvent(stepChangeEvent)
		// Return without processing
		return
	}
	return
}

func (sh *StepChangeHander) processStepChange(stepChangeEvent *events.StepChangeEvent) (err error) {
	logger.DebugF("Processing StepChangeEvent %v", stepChangeEvent)
	var pipeline *data.Pipeline
	var stepState *StepState
	var workflow *models.Workflow
	pipeline, err = sh.storage.GetPipeline(stepChangeEvent.InstanceId)
	if err != nil {
		return
	}
	workflow, err = sh.storage.GetWorkflowByInstance(stepChangeEvent.InstanceId)
	if err != nil {
		return
	}
	outputPipeline := data.NewPipelineFrom(stepChangeEvent.Data)
	iteration, err := data.ExtractValue[int](outputPipeline, data.StepIterationKey)
	if err != nil {
		iteration = 0
	}
	logger.DebugF("Fetching StepState for instance %s, step %s and iteration %d", stepChangeEvent.InstanceId, stepChangeEvent.StepId, iteration)
	stepState, err = sh.storage.GetStepState(stepChangeEvent.InstanceId, stepChangeEvent.StepId, iteration)
	if err != nil {
		return
	}
	stepState.Output = outputPipeline
	// step = utils.GetStepById(stepChangeEvent.StepId, workflow)
	// if err != nil {
	// 	return
	// }
	stepState.Status = stepChangeEvent.Status
	err = sh.storage.SaveStepState(stepState)
	if err != nil {
		return
	}
	pipeline.Merge(outputPipeline)
	err = sh.storage.SavePipeline(pipeline)
	if err != nil {
		return
	}
	switch stepChangeEvent.Status {
	case models.StatusCompleted, models.StatusSkipped:
		// Execute Next Step
		workfFlowExecutor := &WorkflowExecutor{
			storage: sh.storage,
		}
		err = workfFlowExecutor.Execute(workflow, pipeline)
		if err != nil {
			return
		}
	case models.StatusFailed:
		// Fail the instance
		var workflowState *WorkflowState
		workflowState, err = sh.storage.GetState(stepChangeEvent.InstanceId)
		if err != nil {
			return
		}
		errMsg := stepChangeEvent.Data[data.ErrorKey]
		if errMsg != nil {

			workflowState.Error = errMsg.(string)
		}
		workflowState.Status = models.StatusFailed
		err = sh.storage.SaveState(workflowState)
	}
	return
}
