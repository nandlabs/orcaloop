package runtime

import (
	"fmt"

	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop/actions"
	"oss.nandlabs.io/orcaloop/storage"
)

var EndpointInvoker managers.ItemManager[ActionInvoker] = managers.NewItemManager[ActionInvoker]()

type ActionInvoker interface {
	Invoke(actionSpec *actions.ActionSpec, store storage.Storage, pipeline *data.Pipeline) error
}

var ErrActionNotFound = func(id string) error { return fmt.Errorf("action with id %s not found ", id) }

type LocalEndpointInvoker struct {
}

func (lei *LocalEndpointInvoker) Invoke(actionSpec *actions.ActionSpec, store storage.Storage, pipeline *data.Pipeline) error {
	return nil
}

type MsgEndopointInvoker struct {
}

func (mei *MsgEndopointInvoker) Invoke(actionSpec *actions.ActionSpec, store storage.Storage, pipeline *data.Pipeline) error {
	return nil
}

type RestEndpointInvoker struct {
}

func (rei *RestEndpointInvoker) Invoke(actionSpec *actions.ActionSpec, store storage.Storage, pipeline *data.Pipeline) error {
	return nil
}

type GrpcEndpointInvoker struct {
}

func (gei *GrpcEndpointInvoker) Invoke(actionSpec *actions.ActionSpec, store storage.Storage, pipeline *data.Pipeline) error {
	return nil
}
