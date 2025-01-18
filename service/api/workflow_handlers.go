package service

// // Handler functions using server.Context
// func listWorkflows(ctx server.Context) {
// 	workflows := []context.Workflow{
// 		{ID: "1", Name: "Example models.Workflow", Version: "1.0.0", Description: "An example workflow"},
// 	}

// 	ctx.SetStatusCode(http.StatusOK)
// 	ctx.Write(workflows, ioutils.MimeApplicationJSON)
// }

// func createWorkflow(ctx server.Context) {
// 	var workflow models.Workflow

// 	ctx.SetStatusCode(http.StatusCreated)
// 	ctx.WriteJSON(workflow)
// }

// func getWorkflow(ctx server.Context) {
// 	workflowID, err := ctx.GetParam("workflowId", server.PathParam)
// 	if err != nil {
// 		ctx.SetStatusCode(http.StatusBadRequest)
// 		ctx.Write(ErrorResponse{Code: "InvalidData", Message: "Invalid workflow ID", Description: err.Error()})
// 		return
// 	}
// 	if workflowID == "" {
// 		ctx.SetStatusCode(http.StatusNotFound)
// 		ctx.WriteJSON(ErrorResponse{Code: "NotFound", Message: "Workflow not found"})
// 		return
// 	}

// 	workflow := models.Workflow{
// 		ID:   workflowID,
// 		Name: "Example models.Workflow",
// 	}
// 	ctx.SetStatusCode(http.StatusOK)
// 	ctx.WriteJSON(workflow)
// }

// func deleteWorkflow(ctx server.Context) {
// 	workflowID := ctx.GetPathParam("workflowId")
// 	if workflowID == "" {
// 		ctx.SetStatusCode(http.StatusNotFound)
// 		ctx.WriteJSON(ErrorResponse{Code: "NotFound", Message: "Workflow not found"})
// 		return
// 	}

// 	ctx.SetStatusCode(http.StatusNoContent)
// }

// func updateWorkflow(ctx server.Context) {
// 	var workflow models.Workflow
// 	if err := ctx.BindBody(&workflow); err != nil {
// 		ctx.SetStatusCode(http.StatusBadRequest)
// 		ctx.WriteJSON(ErrorResponse{Code: "InvalidData", Message: "Invalid workflow data", Description: err.Error()})
// 		return
// 	}

// 	ctx.SetStatusCode(http.StatusOK)
// 	ctx.WriteJSON(workflow)
// }

// func executeWorkflow(ctx server.Context) {
// 	status := ExecutionStatus{Status: "Running", Message: "Execution started"}
// 	ctx.SetStatusCode(http.StatusAccepted)
// 	ctx.WriteJSON(status)
// }

// func getWorkflowStatus(ctx server.Context) {
// 	status := ExecutionStatus{Status: "Running", Message: "Execution in progress"}
// 	ctx.SetStatusCode(http.StatusOK)
// 	ctx.WriteJSON(status)
// }
