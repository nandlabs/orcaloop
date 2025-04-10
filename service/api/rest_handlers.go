package api

import (
	"fmt"
	"net/http"
	"strconv"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/runtime"
)

type RestHandler struct {
	storage        runtime.Storage
	wfm            *runtime.WorkflowManager
	serviceManager lifecycle.ComponentManager
}

func NewRestHandler(storage runtime.Storage, manager lifecycle.ComponentManager) *RestHandler {
	return &RestHandler{storage: storage, wfm: runtime.NewWorkflowManager(storage), serviceManager: manager}
}

func (rh *RestHandler) GetAllWorkflows(ctx rest.ServerContext) {
	var err error
	var workflows []*models.Workflow
	workflows, err = rh.wfm.GetWorkflows()
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, "Failed to get workflows", err)
		return
	}

	ctx.WriteJSON(&GetWorkflowsResponse{Workflows: workflows})
	ctx.SetStatusCode(http.StatusOK)
}

func (rh *RestHandler) GetWorkflowVersions(ctx rest.ServerContext) {
	var err error
	var id string
	var wfVersions []*models.Workflow
	id, err = ctx.GetParam("id", rest.PathParam)
	if err != nil || id == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}

	wfVersions, err = rh.storage.ListWorkflowVersions(id)
	if err != nil {
		RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Unable to fetch Workflow with id %s", id), err)
		return
	}

	ctx.WriteJSON(&GetWorkflowsResponse{Workflows: wfVersions})
	ctx.SetStatusCode(http.StatusOK)
}

func (rh *RestHandler) GetWorkflow(ctx rest.ServerContext) {
	var err error
	var id string
	var versionStr string
	var version int
	var workflow *models.Workflow
	id, err = ctx.GetParam("id", rest.PathParam)
	if err != nil || id == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}

	versionStr, err = ctx.GetParam("version", rest.PathParam)
	if err != nil || versionStr == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}

	version, err = strconv.Atoi(versionStr)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}

	workflow, err = rh.wfm.GetWorkflow(id, version)
	if err != nil {
		RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Unable to fetch Workflow with id %s and version %d ", id, version), err)
		return
	}

	ctx.WriteJSON(&GetWorkflowResponse{Workflow: workflow})
	ctx.SetStatusCode(http.StatusOK)
}

func (rh *RestHandler) Start(ctx rest.ServerContext) {
	var err error
	var req *StartWorkflowRequest = &StartWorkflowRequest{}
	var instanceId string
	err = ctx.Read(&req)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return
	}

	instanceId, err = rh.wfm.Start(req.WorkflowId, req.Version, req.Input)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to start workflow with id %s and version %d", req.WorkflowId, req.Version), err)
		return
	}

	ctx.WriteJSON(&StartWorkflowResponse{InstanceId: instanceId})
	ctx.SetStatusCode(http.StatusOK)

}

func (rh *RestHandler) Status(ctx rest.ServerContext) {
	var err error
	var req *WorkflowStatusReqeust = &WorkflowStatusReqeust{}
	var pipeline *data.Pipeline
	var workflowState *runtime.WorkflowState
	err = ctx.Read(&req)
	if err != nil || req.InstanceId == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return
	}
	workflowState, err = rh.storage.GetState(req.InstanceId)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get state for workflow instance %s", req.InstanceId), err)
		return
	}
	if workflowState == nil {
		RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Workflow instance %s not found", req.InstanceId), nil)
		return
	}

	pipeline, err = rh.storage.GetPipeline(req.InstanceId)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get pipeline for workflow instance %s", req.InstanceId), err)
		return
	}

	ctx.WriteJSON(&WorkflowStatusResponse{Status: workflowState.Status.String(), Pipeline: pipeline.Map()})
	ctx.SetStatusCode(http.StatusOK)

}

func (rh *RestHandler) RegisterWorflow(ctx rest.ServerContext) {
	var err error
	var workflow *models.Workflow = &models.Workflow{}
	err = ctx.Read(workflow)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return

	}
	err = rh.wfm.Save(workflow)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, "Unable to  to save workflow", err)
		return
	}
	ctx.SetStatusCode(http.StatusAccepted)

}

func (rh *RestHandler) RegisterAction(ctx rest.ServerContext) {
	actionSpec := &models.ActionSpec{}
	err := ctx.Read(actionSpec)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid Input", err)
		return
	}

	err = rh.storage.SaveAction(actionSpec)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, "Unable to save action", err)
		return
	}
	ctx.SetStatusCode(http.StatusAccepted)

}

func (rh RestHandler) GetAllActions(ctx rest.ServerContext) {
	var err error
	var actionSpecs []*models.ActionSpec
	actionSpecs, err = rh.storage.ListActions()
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, "Failed to get actions", err)
		return
	}

	ctx.WriteJSON(&GetActionsResponse{ActionSpec: actionSpecs})
	ctx.SetStatusCode(http.StatusOK)
}

func (rh RestHandler) GetAction(ctx rest.ServerContext) {
	var err error
	var id string
	var actionSpec *models.ActionSpec
	id, err = ctx.GetParam("id", rest.PathParam)
	if err != nil || id == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}

	actionSpec, err = rh.storage.ActionSpec(id)
	if err != nil {
		RespondWithError(ctx, http.StatusNotFound, fmt.Sprintf("Unable to fetch Action with id %s", id), err)
		return
	}

	ctx.WriteJSON(&GetActionResponse{ActionSpec: actionSpec})
	ctx.SetStatusCode(http.StatusOK)
}

func (rh *RestHandler) SystemAction(server rest.Server) {

}

func (rh *RestHandler) DeleteWorkflow(ctx rest.ServerContext) {
	var err error
	var id string
	id, err = ctx.GetParam("id", rest.PathParam)
	if err != nil || id == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}
	versionStr, err := ctx.GetParam("version", rest.PathParam)
	if err != nil || versionStr == "" {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, "Invalid version", err)
		return
	}
	err = rh.wfm.DeleteWorkflow(id, version)
	if err != nil {
		RespondWithError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to delete workflow with id %s", id), err)
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}

func (rh *RestHandler) RegisterRoutes(server rest.Server) {
	server.Post("/workflows", rh.RegisterWorflow)
	server.Get("/workflows", rh.GetAllWorkflows)
	server.Delete("/workflow/:id/:version", rh.DeleteWorkflow)
	server.Get("/workflows/:id", rh.GetWorkflowVersions)
	server.Get("/workflows/:id/:version", rh.GetWorkflow)
	server.Post("/instances/start", rh.Start)
	server.Post("/instances/status", rh.Status)
	server.Post("/actions", rh.RegisterAction)
	server.Get("/actions/:id", rh.GetAllActions)
	server.Post("/system/stop", rh.GetAllActions)

}
