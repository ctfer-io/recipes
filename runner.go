package recipes

import (
	"github.com/ctfer-io/chall-manager/sdk"
	challmanager "github.com/ctfer-io/recipes/chall-manager"
	"github.com/go-viper/mapstructure/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Request[T any] struct {
	Ctx      *pulumi.Context
	Identity string
	Config   *T
}

type Factory[T any] func(req *Request[T], resp *sdk.Response, opts ...pulumi.ResourceOption) error

func Run[T any](f Factory[T]) {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf := new(T)
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				challmanager.ExposeTypeHook,
			),
			WeaklyTypedInput: true,
			Result:           conf,
		})
		if err != nil {
			panic(err)
		}
		if err := dec.Decode(req.Config.Additional); err != nil {
			return err
		}

		return f(&Request[T]{
			Ctx:      req.Ctx,
			Identity: req.Config.Identity,
			Config:   conf,
		}, resp, opts...)
	})
}
