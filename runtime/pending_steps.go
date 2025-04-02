package runtime

type PendingStep struct {
	StepId   string `json:"step_id" yaml:"step_id"`
	VarName  string `json:"var_name" yaml:"var_name"`
	VarValue any    `json:"var_value" yaml:"var_value"`
}
