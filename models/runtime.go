package models

import "oss.nandlabs.io/orcaloop/data"

const (
	StatusUnknown Status = iota
	StatusPending
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusSkipped

	StatusPendingStr   = "Pending"
	StatusRunningStr   = "Running"
	StatusCompletedStr = "Completed"
	StatusFailedStr    = "Failed"
	StatusSkippedStr   = "Skipped"
	StatusUnkonwnStr   = "Unknown"
)

// Status represents the execution status of the workflow
type Status int

func (s Status) String() string {
	switch s {
	case StatusPending:
		return StatusPendingStr
	case StatusRunning:
		return StatusRunningStr
	case StatusCompleted:
		return StatusCompletedStr
	case StatusFailed:
		return StatusFailedStr
	case StatusSkipped:
		return StatusSkippedStr
	default:
		return StatusUnkonwnStr
	}
}

type WorkflowState struct {
	Id         string                `json:"id" yaml:"id"`
	Version    int                   `json:"version" yaml:"version"`
	Status     Status                `json:"status" yaml:"status"`
	Workflow   *Workflow             `json:"workflow" yaml:"workflow"`
	StepStates map[string]*StepState `json:"step_states" yaml:"step_states"`
	Error      string                `json:"error" yaml:"error"`
}

type StepState struct {
	InstanceId string         `json:"instance_id" yaml:"instance_id"`
	StepId     string         `json:"step_id" yaml:"step_id"`
	ParentStep string         `json:"parent_step" yaml:"parent_step"`
	ChildCount int            `json:"child_count" yaml:"child_count"`
	Status     Status         `json:"status" yaml:"status"`
	Input      *data.Pipeline `json:"input" yaml:"input"`
	Output     *data.Pipeline `json:"output" yaml:"output"`
}

type StepChangeEvent struct {
	InstanceId string         `json:"instance_id" yaml:"instance_id"`
	StepId     string         `json:"step_id" yaml:"step_id"`
	Status     Status         `json:"status" yaml:"status"`
	Data       *data.Pipeline `json:"data" yaml:"data"`
}
