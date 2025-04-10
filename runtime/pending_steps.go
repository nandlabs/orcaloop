package runtime

type PendingStep struct {
	Id        string         `json:"id" yaml:"id"`
	Iteration int            `json:"iteration" yaml:"iteration"`
	ParentId  string         `json:"parent_id" yaml:"parent_id"`
	StepId    string         `json:"step_id" yaml:"step_id"`
	Vars      map[string]any `json:"vars" yaml:"vars"`
	// VarName   string `json:"var_name" yaml:"var_name"`
	// VarValue  any    `json:"var_value" yaml:"var_value"`
}
