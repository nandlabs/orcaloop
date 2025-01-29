package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/runtime"
	"oss.nandlabs.io/orcaloop/service/api"
)

type WorkflowSvcHandler struct {
	storage         runtime.Storage
	workflowManager *runtime.WorkflowManager
}

func NewWorkflowSvcHandler(storage runtime.Storage) *WorkflowSvcHandler {
	return &WorkflowSvcHandler{storage: storage, workflowManager: runtime.NewWorkflowManager(storage)}
}

func (wsh *WorkflowSvcHandler) GetAllWorkflows(ctx rest.ServerContext) {
	var err error
	var workflows []*models.Workflow
	workflows, err = wsh.workflowManager.GetWorkflows()
	if err != nil {
		api.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get workflows", err)
		return
	}

	ctx.WriteJSON(&GetWorkflowsResponse{Workflows: workflows})
	ctx.SetStatusCode(http.StatusOK)
}

func (wsh *WorkflowSvcHandler) GetWorkflow(ctx rest.ServerContext) {
	var err error
	var id string
	var versionStr string
	var version int
	var workflow *models.Workflow
	id, err = ctx.GetParam("id", rest.PathParam)
	if err != nil || id == "" {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}

	versionStr, err = ctx.GetParam("version", rest.PathParam)
	if err != nil || versionStr == "" {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}

	version, err = strconv.Atoi(versionStr)
	if err != nil {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}

	workflow, err = wsh.workflowManager.GetWorkflow(id, version)
	if err != nil {
		api.RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Unable to fetch Workflow with id %s and version %d ", id, version), err)
		return
	}

	ctx.WriteJSON(&GetWorkflowResponse{Workflow: workflow})
	ctx.SetStatusCode(http.StatusOK)
}

func (wsh *WorkflowSvcHandler) Start(ctx rest.ServerContext) {
	var err error
	var req *StartWorkflowRequest = &StartWorkflowRequest{}
	var instanceId string
	err = ctx.Read(&req)
	if err != nil {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return
	}

	instanceId, err = wsh.workflowManager.Start(req.WorkflowId, req.Version, req.Input)
	if err != nil {
		api.RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to start workflow with id %s and version %d", req.WorkflowId, req.Version), err)
		return
	}

	ctx.WriteJSON(&StartWorkflowResponse{InstanceId: instanceId})
	ctx.SetStatusCode(http.StatusOK)

}

func (wsh *WorkflowSvcHandler) Status(ctx rest.ServerContext) {
	var err error
	var req *WorkflowStatusReqeust = &WorkflowStatusReqeust{}
	var pipeline *data.Pipeline
	var workflowState *runtime.WorkflowState
	err = ctx.Read(&req)
	if err != nil || req.InstanceId == "" {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return
	}
	workflowState, err = wsh.storage.GetState(req.InstanceId)
	if err != nil {
		api.RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get state for workflow instance %s", req.InstanceId), err)
		return
	}
	if workflowState == nil {
		api.RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Workflow instance %s not found", req.InstanceId), nil)
		return
	}

	pipeline, err = wsh.storage.GetPipeline(req.InstanceId)
	if err != nil {
		api.RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get pipeline for workflow instance %s", req.InstanceId), err)
		return
	}

	ctx.WriteJSON(&WorkflowStatusResponse{Status: workflowState.Status.String(), Pipeline: pipeline.Map()})
	ctx.SetStatusCode(http.StatusOK)

}

func (wsh *WorkflowSvcHandler) RegisterWorflow(ctx rest.ServerContext) {
	var err error
	var workflow *models.Workflow = &models.Workflow{}
	err = ctx.Read(workflow)
	if err != nil {
		api.RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return

	}
	err = wsh.workflowManager.Save(workflow)
	if err != nil {
		api.RespondWithError(ctx, http.StatusInternalServerError, "Unable to  to save workflow", err)
		return
	}
	ctx.SetStatusCode(http.StatusAccepted)

}
func (wsh *WorkflowSvcHandler) RegisterRoutes(server rest.Server) {
	server.Get("/workflows/defintions", wsh.GetAllWorkflows)
	server.Get("/workflows/defintions:id/:version", wsh.GetWorkflow)
	server.Post("/instances/instances/start", wsh.Start)
	server.Post("/instances/instances/status", wsh.Status)
}
