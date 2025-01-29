package runtime

import (
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type Executable interface {
	*models.Workflow | *models.Step
}

type Executor[T Executable] interface {
	//Execute
	Execute(item T, pipeline *data.Pipeline) error
}
