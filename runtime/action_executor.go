package runtime

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"oss.nandlabs.io/golly/messaging"
	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/events"
	"oss.nandlabs.io/orcaloop-sdk/handlers"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type ActionExecutor struct {
	storage Storage
}

func NewActionExecutor(storage Storage) Executor[*models.Step] {
	return &ActionExecutor{
		storage: storage,
	}
}

func (ae *ActionExecutor) Execute(step *models.Step, pipeline *data.Pipeline) (err error) {
	logger.DebugF("Executing action step %s with input %v", step.Id, pipeline)
	if step.Action == nil {
		err = errors.New("no action to execute in the step")
		return
	}
	var actionSpec *models.ActionSpec
	actionSpec, err = ae.storage.ActionSpec(step.Action.Id)
	if err != nil {
		return
	}
	actionPipeline := pipeline.Clone()
	parametersMap := make(map[string]*models.Schema)
	for _, param := range actionSpec.Parameters {
		parametersMap[param.Name] = param
	}
	if err != nil {
		return
	}

	if actionSpec == nil {
		err = errors.New("action action found by id " + step.Action.Name)
		return
	}
	stepChangeHandler := &StepChangeHander{storage: ae.storage}
	// Validate the parameters for any missing required parameters
	for _, param := range step.Action.Parameters {
		var inVal any
		if param.Value != nil {
			inVal = param.Value
		} else if actionPipeline.Has(param.Var) {
			inVal, err = actionPipeline.Get(param.Var)
			if err != nil {
				return
			}

		}
		_, ok := inVal.(int)
		if ok && inVal != nil && parametersMap[param.Name].Type == "number" {

			inVal = float64(inVal.(int))

		}

		logger.DebugF("Setting param %s with value :%v", param.Name, inVal)
		actionPipeline.Set(param.Name, inVal)

	}

	switch actionSpec.Endpoint.Type {
	case models.EndpointTypeLocal:
		handler := handlers.ActionRegistry.Get(step.Action.Id)
		if handler == nil {
			err = errors.New("action handler not found for action id " + step.Action.Id)
			return
		}
		err = handler.Handle(actionPipeline)
		if err != nil {
			return
		}
		retMap := make(map[string]any)

		for _, result := range step.Action.Results {
			var outVal any

			outVal, err = actionPipeline.Get(result.OutputVar)
			if err != nil {
				return
			}
			retMap[result.PipelineVar] = outVal
		}

		event := &events.StepChangeEvent{
			EventId:    CreateId(),
			InstanceId: actionPipeline.Id(),
			StepId:     step.Id,
			Status:     models.StatusCompleted,
			Data:       retMap,
		}
		err = stepChangeHandler.Handle(event)
		if err != nil {
			return
		}
		logger.DebugF("Event Sent to stepChangeHandler %v", event)

	case models.EndpointTypeRest:
		var res *rest.Response
		client := rest.NewClient()
		req := client.NewRequest(actionSpec.Endpoint.Rest.Url, http.MethodPost)
		req.SetBody(actionPipeline.Map())
		res, err = client.Execute(req)
		if err != nil {
			return
		}
		switch res.StatusCode() {
		case http.StatusOK:
			// This is a sync call, so we can expect the action to be available.
			resMap := make(map[string]any)
			err = res.Decode(&resMap)
			if err != nil {
				return errors.New("failed to decode response for action " + step.Action.Id + " with error " + err.Error())
			}

			status := models.StatusFailed
			if _, ok := resMap[data.ErrorKey]; !ok {

				status = models.StatusCompleted
			}
			event := &events.StepChangeEvent{
				EventId:    CreateId(),
				InstanceId: actionPipeline.Id(),
				StepId:     step.Id,
				Status:     status,
				Data:       resMap,
			}

			err = stepChangeHandler.Handle(event)
			if err != nil {
				return
			}
		case http.StatusAccepted:
			// This is an async call, we just fire and forget
			logger.InfoF("action %s accepted", step.Action.Id)

		case http.StatusInternalServerError:
			// try parsing the error message
			var errMessage *models.Error = &models.Error{}
			err = res.Decode(errMessage)
			if err != nil {
				err = errors.New("Unable to execute the rest action for  " + actionSpec.Id)
				return
			}

			err = fmt.Errorf("unable to execute the rest action for  %s with error %s", actionSpec.Id, errMessage.Message)

			return

		}

	case models.EndpointTypeMessaging:
		var u *url.URL
		var message messaging.Message
		var manager messaging.Manager = messaging.GetManager()
		u, err = url.Parse(actionSpec.Endpoint.Messaging.Url)
		if err != nil {
			return fmt.Errorf("invalid url %s for action %s", actionSpec.Endpoint.Messaging.Url, actionSpec.Id)
		}
		// Publish the message to the messaging system
		message, err = manager.NewMessage(u.Scheme)
		if err != nil {
			return
		}
		err = message.WriteJSON(actionPipeline.Map())
		if err != nil {
			return
		}
		err = manager.Send(u, message)
		return
	}

	return
}
