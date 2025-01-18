package actions

import (
	"fmt"

	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/orcaloop-sdk/data"
)

func ValidateInputs(actionSpec *ActionSpec, pipeline *data.Pipeline) (valid bool, err error) {
	var multiError *errutils.MultiError = errutils.NewMultiErr(nil)
	for _, input := range actionSpec.Parameters {
		if input.Required {
			if !pipeline.Has(input.Name) {
				err = fmt.Errorf("missing required input %s", input.Name)
				multiError.Add(err)
			}
		}
	}
	if multiError.HasErrors() {
		err = multiError
		valid = false
		return
	} else {
		valid = true
	}

	return
}
