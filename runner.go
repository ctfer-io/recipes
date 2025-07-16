package recipes

import (
	"net/url"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/go-playground/form/v4"
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

		dec := form.NewDecoder()
		if err := dec.Decode(conf, toValues(req.Config.Additional)); err != nil {
			return err
		}

		return f(&Request[T]{
			Ctx:      req.Ctx,
			Identity: req.Config.Identity,
			Config:   conf,
		}, resp, opts...)
	})
}

func toValues(additionals map[string]string) url.Values {
	vals := make(url.Values, len(additionals))
	for k, v := range additionals {
		vals[k] = []string{v}
	}
	return vals
}
