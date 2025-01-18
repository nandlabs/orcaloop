package exec

import (
	"fmt"

	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop/actions"
	"oss.nandlabs.io/orcaloop/models"
	"oss.nandlabs.io/orcaloop/storage"
)

var ErrActionNotFound = fmt.Errorf("action not found")
var ErrUnsupportedAction = fmt.Errorf("unsupported action")

// ActionManager is a manager for all actions
var localActionManager = managers.NewItemManager[actions.Action]()

// ActionExecutor defines the interface for executing actions
type ActionExecutor interface {
	ExecuteAction(action *models.StepAction, pipeline *data.Pipeline) (bool, error)
}

type LocalActionExecutor struct{}

func (lae *LocalActionExecutor) ExecuteAction(storage storage.Storage, stepAction *models.StepAction, additionalValues map[string]any) (isAsync bool, err error) {
	action := localActionManager.Get(stepAction.Id)
	if action == nil {
		err = ErrActionNotFound
	} else {
		// 	err = action.Execute(state.Pipeline)
		// 	if err == nil {
		// 		isAsync = action.IsAsync()
		// 	}
		// }
	}
	return
}

// // GrpcActionExecutor defines the interface for executing gRPC actions
// type GrpcActionExecutor struct{}

// // ExecuteAction executes a gRPC action
// func (gae *GrpcActionExecutor) ExecuteAction(stepAction *StepAction, state *WorkflowState) (isAsync bool, err error) {
// 	return
// }

// // GetActionExecutor returns an ActionExecutor based on the action type
// func GetActionExecutor(stepAction *StepAction) (ae ActionExecutor, err error) {
// 	actionType := strings.Split(stepAction.Id, ".")[0]
// 	switch actionType {
// 	case "local":
// 		ae = &LocalActionExecutor{}
// 	case "grpc":
// 		ae = &GrpcActionExecutor{}
// 	default:
// 		err = ErrUnsupportedAction
// 	}
// 	return
// }
