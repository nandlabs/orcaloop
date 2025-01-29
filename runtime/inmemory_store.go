package runtime

import (
	"errors"

	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/events"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/config"
)

type InMemoryStorage struct {
	actionSpecs      map[string]*models.ActionSpec
	workflows        map[string]map[int]*models.Workflow  // workflowId -> version -> Workflow
	instances        map[string]*data.Pipeline            // instanceId -> Pipeline
	workflowStates   map[string]*WorkflowState            // instanceId -> WorkflowState
	stepStates       map[string]map[string]*StepState     // instanceId -> stepId -> StepState
	stepChangeEvents map[string][]*events.StepChangeEvent // instanceId -> StepChangeEvents
	pendingSteps     map[string][]*PendingStep
	lockedInstances  map[string]bool // instanceId -> locked (true/false)
}

// NewInMemoryStorage creates a new instance of InMemoryStorage
func NewInMemoryStorage(c *config.StorageConfig) *InMemoryStorage {
	return &InMemoryStorage{
		actionSpecs:      make(map[string]*models.ActionSpec),
		workflows:        make(map[string]map[int]*models.Workflow),
		instances:        make(map[string]*data.Pipeline),
		workflowStates:   make(map[string]*WorkflowState),
		stepStates:       make(map[string]map[string]*StepState),
		stepChangeEvents: make(map[string][]*events.StepChangeEvent),
		pendingSteps:     make(map[string][]*PendingStep),
		lockedInstances:  make(map[string]bool),
	}
}

// Implementation of Storage interface methods

func (s *InMemoryStorage) ActionSpec(id string) (*models.ActionSpec, error) {

	action, ok := s.actionSpecs[id]
	if !ok {
		return nil, errors.New("action not found")
	}
	return action, nil
}
func (s *InMemoryStorage) AddPendingSteps(instanceId string, pendingStep ...*PendingStep) error {

	if _, ok := s.pendingSteps[instanceId]; !ok {
		s.pendingSteps[instanceId] = make([]*PendingStep, 0)
	}
	s.pendingSteps[instanceId] = append(pendingStep, s.pendingSteps[instanceId]...)
	return nil
}

func (s *InMemoryStorage) ActionEndpoint(id string) (*models.Endpoint, error) {
	action, err := s.ActionSpec(id)
	if err != nil {
		return nil, err
	}
	return action.Endpoint, nil
}

func (s *InMemoryStorage) ActionSpecs() ([]*models.ActionSpec, error) {

	var specs []*models.ActionSpec
	for _, spec := range s.actionSpecs {
		specs = append(specs, spec)
	}
	return specs, nil
}

func (s *InMemoryStorage) ArchiveInstance(workflowId string, archiveInstance bool) error {
	// Archive logic (in-memory doesn't support "archive" directly)
	return nil
}

func (s *InMemoryStorage) CreateNewInstance(workflowId string, instanceId string, pipeline *data.Pipeline) error {

	s.instances[instanceId] = pipeline
	return nil
}

func (s *InMemoryStorage) DeleteAction(id string) error {

	if _, ok := s.actionSpecs[id]; !ok {
		return errors.New("action not found")
	}
	delete(s.actionSpecs, id)
	return nil
}
func (s *InMemoryStorage) DeletePendingStep(instanceId string, pendingStep *PendingStep) (err error) {

	pSteps, ok := s.pendingSteps[instanceId]
	if !ok {
		return
	}
	for i, pStep := range pSteps {
		if pStep.StepId == pendingStep.StepId && pStep.VarName == pendingStep.VarName && pStep.VarValue == pendingStep.VarValue {
			s.pendingSteps[instanceId] = append(pSteps[:i], pSteps[i+1:]...)
			break
		}
	}
	return
}

func (s *InMemoryStorage) DeleteStepChangeEvent(instanceId, eventId string) (err error) {

	events, ok := s.stepChangeEvents[instanceId]
	if !ok {
		return
	}
	for i, event := range events {
		if event.EventId == eventId {
			s.stepChangeEvents[instanceId] = append(events[:i], events[i+1:]...)
			break
		}
	}

	return nil
}

func (s *InMemoryStorage) GetPipeline(id string) (*data.Pipeline, error) {

	pipeline, ok := s.instances[id]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	return pipeline, nil
}

func (s *InMemoryStorage) GetState(instanceId string) (*WorkflowState, error) {

	state, ok := s.workflowStates[instanceId]
	if !ok {
		return nil, errors.New("workflow state not found")
	}
	return state, nil
}

func (s *InMemoryStorage) GetNextPendingStep(instanceId string) (*PendingStep, error) {

	steps, ok := s.pendingSteps[instanceId]
	if !ok || len(steps) == 0 {
		return nil, nil
	}
	return steps[0], nil
}

func (s *InMemoryStorage) GetPendingSteps(instanceId string) (steps []*PendingStep, err error) {

	steps = s.pendingSteps[instanceId]
	if steps == nil {
		steps = make([]*PendingStep, 0)
	}
	return
}

func (s *InMemoryStorage) GetStepChangeEvents(instanceId string) (events []*events.StepChangeEvent, err error) {

	events = s.stepChangeEvents[instanceId]

	return
}

func (s *InMemoryStorage) GetStepStates(instanceId string) (map[string]*StepState, error) {

	stepStatesMap := s.stepStates[instanceId]
	if stepStatesMap == nil {
		stepStatesMap = make(map[string]*StepState)
		s.stepStates[instanceId] = stepStatesMap
	}
	return stepStatesMap, nil
}

func (s *InMemoryStorage) GetStepState(instanceId, stepId string) (*StepState, error) {

	stepStatesMap := s.stepStates[instanceId]
	if stepStatesMap == nil {
		stepStatesMap = make(map[string]*StepState)
		s.stepStates[instanceId] = stepStatesMap
	}

	return stepStatesMap[stepId], nil
}

func (s *InMemoryStorage) GetWorkflow(workflowId string, version int) (*models.Workflow, error) {

	versions, ok := s.workflows[workflowId]
	if !ok {
		return nil, errors.New("workflow not found")
	}
	workflow, ok := versions[version]
	if !ok {
		return nil, errors.New("workflow version not found")
	}
	return workflow, nil
}

func (s *InMemoryStorage) GetWorkflowByInstance(id string) (wf *models.Workflow, err error) {
	// Assume instanceId is somehow linked to a workflowId; not implemented here
	var workflowState *WorkflowState
	workflowState, err = s.GetState(id)
	if err != nil {
		return
	}
	return s.GetWorkflow(workflowState.WorkflowId, workflowState.WorkflowVersion)
}

func (s *InMemoryStorage) ListActions() ([]*models.ActionSpec, error) {
	return s.ActionSpecs()
}

func (s *InMemoryStorage) ListWorkflows() ([]*models.Workflow, error) {

	var workflows []*models.Workflow
	for _, versions := range s.workflows {
		for _, workflow := range versions {
			workflows = append(workflows, workflow)
		}
	}
	return workflows, nil
}

func (s *InMemoryStorage) ListWorkflowVersions(workflowID string) ([]*models.Workflow, error) {

	versions := make([]*models.Workflow, 0)
	for _, wf := range s.workflows[workflowID] {
		versions = append(versions, wf)
	}
	return versions, nil
}

func (s *InMemoryStorage) LockInstance(id string) (bool, error) {

	if s.lockedInstances[id] {
		return false, nil
	}
	s.lockedInstances[id] = true
	return true, nil
}

func (s *InMemoryStorage) SaveAction(action *models.ActionSpec) error {

	s.actionSpecs[action.Id] = action
	return nil
}

func (s *InMemoryStorage) SaveStepChangeEvent(stepEvent *events.StepChangeEvent) error {

	s.stepChangeEvents[stepEvent.InstanceId] = append(s.stepChangeEvents[stepEvent.InstanceId], stepEvent)
	return nil
}

func (s *InMemoryStorage) SavePipeline(pipeline *data.Pipeline) error {

	s.instances[pipeline.Id()] = pipeline
	return nil
}

func (s *InMemoryStorage) SaveState(workflowState *WorkflowState) error {

	s.workflowStates[workflowState.InstanceId] = workflowState
	return nil
}

func (s *InMemoryStorage) SaveStepState(stepState *StepState) error {

	if _, ok := s.stepStates[stepState.InstanceId]; !ok {
		s.stepStates[stepState.InstanceId] = make(map[string]*StepState)
	}
	s.stepStates[stepState.InstanceId][stepState.StepId] = stepState

	return nil
}

func (s *InMemoryStorage) SaveWorkflow(workflow *models.Workflow) error {

	if _, ok := s.workflows[workflow.Id]; !ok {
		s.workflows[workflow.Id] = make(map[int]*models.Workflow)
	}
	s.workflows[workflow.Id][workflow.Version] = workflow
	return nil
}

func (s *InMemoryStorage) UnlockInstance(id string) error {

	if !s.lockedInstances[id] {
		return errors.New("instance is not locked")
	}
	delete(s.lockedInstances, id)
	return nil
}

func (s *InMemoryStorage) DeleteWorkflow(workflowID string, version int) error {

	if _, ok := s.workflows[workflowID]; !ok {
		return errors.New("workflow not found")
	}
	delete(s.workflows[workflowID], version)
	return nil
}

func (s *InMemoryStorage) Config() *config.StorageConfig {
	return &config.StorageConfig{
		Type: config.InMemoryStorageType,
		Provider: &config.Provider{
			Local: &config.LocalStorage{},
		},
	}
}
