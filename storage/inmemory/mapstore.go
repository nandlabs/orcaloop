package inmemory

import (
	"errors"
	"sync"

	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/actions"
)

type InMemoryStorage struct {
	mu               sync.RWMutex
	actionSpecs      map[string]*actions.ActionSpec
	workflows        map[string]map[int]*models.Workflow  // workflowId -> version -> Workflow
	instances        map[string]*data.Pipeline            // instanceId -> Pipeline
	workflowStates   map[string]*models.WorkflowState     // instanceId -> WorkflowState
	stepChangeEvents map[string][]*models.StepChangeEvent // instanceId -> StepChangeEvents
	lockedInstances  map[string]bool                      // instanceId -> locked (true/false)
}

// NewInMemoryStorage creates a new instance of InMemoryStorage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		actionSpecs:      make(map[string]*actions.ActionSpec),
		workflows:        make(map[string]map[int]*models.Workflow),
		instances:        make(map[string]*data.Pipeline),
		workflowStates:   make(map[string]*models.WorkflowState),
		stepChangeEvents: make(map[string][]*models.StepChangeEvent),
		lockedInstances:  make(map[string]bool),
	}
}

// Implementation of Storage interface methods

func (s *InMemoryStorage) ActionSpec(id string) (*actions.ActionSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	action, ok := s.actionSpecs[id]
	if !ok {
		return nil, errors.New("action not found")
	}
	return action, nil
}

func (s *InMemoryStorage) ActionSpecs() ([]*actions.ActionSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var specs []*actions.ActionSpec
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
	s.mu.Lock()
	defer s.mu.Unlock()

	s.instances[instanceId] = pipeline
	return nil
}

func (s *InMemoryStorage) DeleteAction(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.actionSpecs[id]; !ok {
		return errors.New("action not found")
	}
	delete(s.actionSpecs, id)
	return nil
}

func (s *InMemoryStorage) GetPipeline(id string) (*data.Pipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, ok := s.instances[id]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	return pipeline, nil
}

func (s *InMemoryStorage) GetState(instanceId string) (*models.WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.workflowStates[instanceId]
	if !ok {
		return nil, errors.New("workflow state not found")
	}
	return state, nil
}

func (s *InMemoryStorage) GetStepChangeEvents(instanceId string) ([]*models.StepChangeEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, ok := s.stepChangeEvents[instanceId]
	if !ok {
		return nil, errors.New("step change events not found")
	}
	return events, nil
}

func (s *InMemoryStorage) GetStepState(instanceId, stepId string) (*models.StepState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.workflowStates[instanceId]
	if !ok {
		return nil, errors.New("workflow state not found")
	}

	stepState, ok := state.StepStates[stepId]
	if !ok {
		return nil, errors.New("step state not found")
	}
	return stepState, nil
}

func (s *InMemoryStorage) GetWorkflow(workflowId string, version int) (*models.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

func (s *InMemoryStorage) GetWorkflowByInstance(id string) (*models.Workflow, error) {
	// Assume instanceId is somehow linked to a workflowId; not implemented here
	return nil, nil
}

func (s *InMemoryStorage) ListActions() ([]*actions.ActionSpec, error) {
	return s.ActionSpecs()
}

func (s *InMemoryStorage) LockInstance(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lockedInstances[id] {
		return false, errors.New("instance already locked")
	}
	s.lockedInstances[id] = true
	return true, nil
}

func (s *InMemoryStorage) SaveAction(action *actions.ActionSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.actionSpecs[action.Id] = action
	return nil
}

func (s *InMemoryStorage) SaveStepChangeEvent(stepEvent *models.StepChangeEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stepChangeEvents[stepEvent.InstanceId] = append(s.stepChangeEvents[stepEvent.InstanceId], stepEvent)
	return nil
}

func (s *InMemoryStorage) SavePipeline(pipeline *data.Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.instances[pipeline.Id()] = pipeline
	return nil
}

func (s *InMemoryStorage) SaveState(workflowState *models.WorkflowState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.workflowStates[workflowState.Id] = workflowState
	return nil
}

func (s *InMemoryStorage) SaveStepState(stepState *models.StepState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.workflowStates[stepState.InstanceId]
	if !ok {
		return errors.New("workflow state not found for instance")
	}

	state.StepStates[stepState.StepId] = stepState
	return nil
}

func (s *InMemoryStorage) SaveWorkflow(workflow *models.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.workflows[workflow.Id]; !ok {
		s.workflows[workflow.Id] = make(map[int]*models.Workflow)
	}
	s.workflows[workflow.Id][workflow.Version] = workflow
	return nil
}

func (s *InMemoryStorage) UnlockInstance(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lockedInstances[id] {
		return errors.New("instance is not locked")
	}
	delete(s.lockedInstances, id)
	return nil
}
