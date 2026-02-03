package main

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/ctfer-io/recipes"
	"github.com/ctfer-io/recipes/chall-manager/k8s.EMP/config"
)

// Values are used as part of the templating of Config.ConnectionInfo.
type Values struct {
	URLs map[string]map[string]string
}

func main() {
	recipes.Run(func(req *recipes.Request[config.Config], resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		// Build template ASAP -> fail fast
		citmpl, err := template.New("connectionInfo").
			Funcs(sprig.FuncMap()).
			Parse(req.Config.ConnectionInfo)
		if err != nil {
			return errors.Wrap(err, "building connection info template")
		}

		// Deploy k8s.ExposedMultipod
		cm, err := k8s.NewExposedMultipod(req.Ctx, "recipe-k8s-emp", &k8s.ExposedMultipodArgs{
			Identity: pulumi.String(req.Identity),
			Label:    pulumi.String(req.Ctx.Stack()),
			Hostname: pulumi.String(req.Config.Hostname),
			Containers: func() k8s.ContainerMap {
				out := map[string]k8s.ContainerInput{}
				for name, args := range req.Config.Containers {
					out[name] = k8s.ContainerArgs{
						Image: pulumi.String(args.Image),
						Ports: func() k8s.PortBindingArray {
							out := make([]k8s.PortBindingInput, 0, len(args.Ports))
							for _, port := range args.Ports {
								out = append(out, k8s.PortBindingArgs{
									Port:        pulumi.Int(port.Port),
									Protocol:    pulumi.String(port.Protocol),
									ExposeType:  port.ExposeType,
									Annotations: pulumi.ToStringMap(port.Annotations),
								})
							}
							return out
						}(),
						Envs: func() k8s.PrinterMap {
							out := map[string]k8s.PrinterInput{}
							for k, v := range args.Envs {
								out[k] = k8s.NewPrinter(v)
							}
							return out
						}(),
						Files:       pulumi.ToStringMap(args.Files),
						LimitCPU:    pulumi.StringPtrFromPtr(args.LimitCPU),
						LimitMemory: pulumi.StringPtrFromPtr(args.LimitMemory),
					}
				}
				return out
			}(),
			Rules: func() k8s.RuleArray {
				out := []k8s.RuleInput{}
				for _, rule := range req.Config.Rules {
					out = append(out, k8s.RuleArgs{
						From:     pulumi.String(rule.From),
						To:       pulumi.String(rule.To),
						On:       pulumi.Int(rule.On),
						Protocol: pulumi.String(rule.Protocol),
					})
				}
				return out
			}(),
			FromCIDR:         pulumi.String(req.Config.FromCIDR),
			IngressNamespace: pulumi.String(req.Config.IngressNamespace),
			IngressLabels:    pulumi.ToStringMap(req.Config.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		// Template connection info
		resp.ConnectionInfo = cm.URLs.ApplyT(func(urls map[string]map[string]string) (string, error) {
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
