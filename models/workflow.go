package models

const (
	StepTypeAction   = "Action"
	StepTypeParallel = "Parallel"
	StepTypeIf       = "If"
	StepTypeSwitch   = "Switch"
	StepTypeForLoop  = "ForLoop"
	ModeSync         = "sync"
	ModeAsync        = "async"
)

// Workflow is a struct that represents the Workflow workflow
type Workflow struct {
	Id          string  `yaml:"id" json:"id"`
	Name        string  `yaml:"name" json:"name"`
	Version     int     `yaml:"version" json:"version"`
	Description string  `yaml:"description" json:"description"`
	Steps       []*Step `yaml:"steps" json:"steps"`
}

// Parameter is a struct that represents a parameter in the workflow
type Parameter struct {
	Name  string `yaml:"name" json:"name"`
	Value any    `yaml:"value" json:"value"`
	Var   string `yaml:"var" json:"var"`
}

// SubStep is a struct that represents a substep in the workflow
type StepAction struct {
	Id         string       `yaml:"id" json:"id"`
	Name       string       `yaml:"name" json:"name"`
	Parameters []*Parameter `yaml:"parameters" json:"parameters"`
	Output     []string     `yaml:"output_names" json:"output_names"`
}

// Step is a struct that represents a step in the workflow
type Step struct {
	Id       string      `yaml:"id" json:"id"`
	Skip     bool        `yaml:"skip" json:"skip"`
	Type     string      `yaml:"type" json:"type"`
	Parallel *Parallel   `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	For      *For        `yaml:"for,omitempty" json:"for,omitempty"`
	If       *If         `yaml:"if,omitempty" json:"if,omitempty"`
	Switch   *Switch     `yaml:"switch,omitempty" json:"switch,omitempty"`
	Action   *StepAction `yaml:"action,omitempty" json:"action,omitempty"`
}

type Parallel struct {
	Steps []*Step `yaml:"steps" json:"steps"`
}

type For struct {
	Loopvar  string  `yaml:"loop_var" json:"loop_var"`
	IndexVar string  `yaml:"index_var" json:"index_var"`
	ItemsVar string  `yaml:"items_var" json:"items_var"`
	ItemsArr []any   `yaml:"items" json:"items"`
	Steps    []*Step `yaml:"steps" json:"steps"`
}

type If struct {
	Condition string    `yaml:"condition" json:"condition"`
	Steps     []*Step   `yaml:"steps" json:"steps"`
	ElseIfs   []*ElseIf `yaml:"else_ifs" json:"else_ifs"`
	Else      *Else     `yaml:"else" json:"else"`
}

type ElseIf struct {
	Condition string  `yaml:"condition" json:"condition"`
	Steps     []*Step `yaml:"steps" json:"steps"`
}

type Else struct {
	Steps []*Step `yaml:"steps" json:"steps"`
}

type Switch struct {
	Variable string  `yaml:"variable" json:"variable"`
	Cases    []*Case `yaml:"cases" json:"cases"`
}

type Case struct {
	Value   any     `yaml:"value" json:"value"`
	Default bool    `yaml:"default" json:"default"`
	Steps   []*Step `yaml:"steps" json:"steps"`
}
