package main

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/ctfer-io/recipes"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	Image              string            `form:"image"`
	Ports              []PortArgs        `form:"ports"`
	Hostname           string            `form:"hostname"`
	Files              map[string]string `form:"files"`
	IngressAnnotations map[string]string `form:"ingressAnnotations"`
	IngressNamespace   string            `form:"ingressNamespace"`
	IngressLabels      map[string]string `form:"ingressLabels"`
	ConnectionInfo     string            `form:"connectionInfo"`
}

type PortArgs struct {
	Port       int            `form:"port"`
	Protocol   string         `form:"protocol"`
	ExposeType k8s.ExposeType `form:"exposeType"`
}

// Values are used as part of the templating of Config.ConnectionInfo.
type Values struct {
	URLs map[string]string
}

func main() {
	recipes.Run(func(req *recipes.Request[Config], resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		// Build template ASAP -> fail fast
		citmpl, err := template.New("connectionInfo").
			Funcs(sprig.FuncMap()).
			Parse(req.Config.ConnectionInfo)
		if err != nil {
			return errors.Wrap(err, "building connection info template")
		}

		// Deploy k8s.ExposedMonopod
		cm, err := k8s.NewExposedMonopod(req.Ctx, "recipe-e1p", &k8s.ExposedMonopodArgs{
			Identity: pulumi.String(req.Identity),
			Hostname: pulumi.String(req.Config.Hostname),
			Label:    pulumi.String(req.Ctx.Stack()),
			Container: k8s.ContainerArgs{
				Image: pulumi.String(req.Config.Image),
				Ports: func() k8s.PortBindingArray {
					out := make([]k8s.PortBindingInput, 0, len(req.Config.Ports))
					for _, port := range req.Config.Ports {
						out = append(out, k8s.PortBindingArgs{
							Port:       pulumi.Int(port.Port),
							Protocol:   pulumi.String(port.Protocol),
							ExposeType: port.ExposeType,
						})
					}
					return out
				}(),
				Files: pulumi.ToStringMap(req.Config.Files),
			},
			IngressAnnotations: pulumi.ToStringMap(req.Config.IngressAnnotations),
			IngressNamespace:   pulumi.String(req.Config.IngressNamespace),
			IngressLabels:      pulumi.ToStringMap(req.Config.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		// Template connection info
		resp.ConnectionInfo = cm.URLs.ApplyT(func(urls map[string]string) (string, error) {
			values := &Values{
				URLs: urls,
			}
			buf := &bytes.Buffer{}
			if err := citmpl.Execute(buf, values); err != nil {
				return "", err
			}
			return buf.String(), nil
		}).(pulumi.StringOutput)

		return nil
	})
}
