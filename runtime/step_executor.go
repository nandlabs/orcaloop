package runtime

import (
	"strconv"
	"sync"

	"oss.nandlabs.io/golly/assertion"
	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type StepExecutor struct {
	storage Storage
}

func (se *StepExecutor) Execute(step *models.Step, pipeline *data.Pipeline) (err error) {

	// Start the execution of the step

	instanceId := pipeline.Id()
	parentId := pipeline.GetParent()

	stepState := &StepState{
		InstanceId: instanceId,
		StepId:     step.Id,
		ParentStep: parentId,
		Status:     models.StatusRunning,
	}

	switch step.Type {
	case models.StepTypeForLoop:
		var items []any = step.For.ItemsArr
		var idx_var string = step.For.IndexVar
		if idx_var == "" {
			idx_var = "idx-" + step.Id
		}
		if len(items) == 0 {
			items, err = data.ExtractValue[[]any](pipeline, step.For.ItemsVar)
			if err != nil {
				return
			}
		}
		stepState.ChildCount = len(items)
		err = se.storage.SaveStepState(stepState)
		if err != nil {
			return
		}
		// Execute the steps for each item in the array
		var pendingSteps []*PendingStep
		var childPipeline *data.Pipeline
		var firstChildStep *models.Step
		for idx, item := range items {
			for i, childStep := range step.For.Steps {
				if i == 0 && idx == 0 {
					childPipeline = cloneFor(pipeline, childStep, step.Id)
					childPipeline.Set(step.For.ItemsVar, item)
					childPipeline.Set(step.For.IndexVar, idx)
					firstChildStep = childStep

				} else {
					pendingSteps = append(pendingSteps, &PendingStep{
						StepId:   childStep.Id,
						VarName:  idx_var,
						VarValue: strconv.Itoa(idx),
					})
				}

			}
		}
		err = se.storage.AddPendingSteps(instanceId, pendingSteps...)
		if err != nil {
			return
		}

		err = se.Execute(firstChildStep, childPipeline)
		if err != nil {
			return
		}

	case models.StepTypeIf:
		var condition bool
		condition, err = pipeline.EvaluateCondition(step.If.Condition)
		if err != nil {
			return
		}
		var steps []*models.Step

		if condition {
			steps = step.If.Steps
		} else {
			if len(step.If.ElseIfs) > 0 {
				for _, elseIf := range step.If.ElseIfs {
					condition, err = pipeline.EvaluateCondition(elseIf.Condition)
					if err != nil {
						return
					}
					if condition {
						steps = elseIf.Steps
						break
					}

				}
			}
			if (!condition) && (step.If.Else != nil) {
				steps = step.If.Else.Steps
			}
		}
		if len(steps) > 0 {
			stepState.ChildCount = len(steps)
			err = se.storage.SaveStepState(stepState)
			for _, childStep := range steps {
				childPipeline := cloneFor(pipeline, childStep, step.Id)
				err = se.Execute(childStep, childPipeline)
				if err != nil {
					return
				}
			}
		}

	case models.StepTypeParallel:
		stepState.ChildCount = len(step.Parallel.Steps)
		err = se.storage.SaveStepState(stepState)
		wg := sync.WaitGroup{}
		var multiErr = errutils.NewMultiErr(nil)
		for _, childStep := range step.Parallel.Steps {
			childPipeline := cloneFor(pipeline, childStep, step.Id)
			wg.Add(1)
			go func(multiErr *errutils.MultiError) {
				defer wg.Done()
				err := se.Execute(childStep, childPipeline)
				if err != nil {
					multiErr.Add(err)
				}
			}(multiErr)
			wg.Wait()
			if multiErr.HasErrors() {
				err = multiErr
			}
		}

	case models.StepTypeSwitch:
		var value any
		value, err = data.ExtractValue[any](pipeline, step.Switch.Variable)
		if err != nil {
			return
		}
		var steps []*models.Step
		var found bool
		var defaultSteps []*models.Step
		for _, caseItem := range step.Switch.Cases {
			if caseItem.Default {
				defaultSteps = caseItem.Steps
				continue
			}
			if assertion.Equal(value, caseItem.Value) {
				steps = caseItem.Steps
				found = true
				break
			}
		}

		if !found {
			steps = defaultSteps
		}
		if len(steps) > 0 {
			stepState.ChildCount = len(steps)
			err = se.storage.SaveStepState(stepState)
			for _, childStep := range steps {
				childPipeline := cloneFor(pipeline, childStep, step.Id)
				err = se.Execute(childStep, childPipeline)
				if err != nil {
					return
				}
			}
		}

	case models.StepTypeAction:
		stepState.ChildCount = 0
		err = se.storage.SaveStepState(stepState)
		if err != nil {
			return
		}
		executor := NewActionExecutor(se.storage)
		err = executor.Execute(step, pipeline)
		if err != nil {
			return
		}
	}

	return
}

func cloneFor(pipeline *data.Pipeline, step *models.Step, parent string) (clone *data.Pipeline) {

	clone = pipeline.Clone()
	clone.Set(data.StepIdKey, step.Id)
	clone.Set(data.ParentIdKey, parent)

	return

}
