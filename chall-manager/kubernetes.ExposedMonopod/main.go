package main

import (
	"bytes"
	"text/template"

	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/ctfer-io/recipes"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	Image              string            `mapstructure:"image"`
	Port               int               `mapstructure:"port"`
	ExposeType         k8s.ExposeType    `mapstructure:"exposeType"`
	Hostname           string            `mapstructure:"hostname"`
	Files              map[string]string `mapstructure:"files,omitempty"`
	IngressAnnotations map[string]string `mapstructure:"ingressAnnotations,omitempty"`
	IngressNamespace   string            `mapstructure:"ingressNamespace,omitempty"`
	IngressLabels      map[string]string `mapstructure:"ingressLabels,omitempty"`
	ConnectionInfo     string            `mapstructure:"connectionInfo"`
}

// Values are used as part of the templating of Config.ConnectionInfo.
type Values struct {
	URL string `json:"url"`
}

func main() {
	recipes.Run(func(req *recipes.Request[*Config], resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		// Build template ASAP -> fail fast
		citmpl, err := template.New("connectionInfo").Parse(req.Config.ConnectionInfo)
		if err != nil {
			return errors.Wrap(err, "building connection info template")
		}

		// Deploy k8s.ExposedMonopod
		cm, err := k8s.NewExposedMonopod(req.Ctx, &k8s.ExposedMonopodArgs{
			Image:              pulumi.String(req.Config.Image),
			Port:               pulumi.Int(req.Config.Port),
			ExposeType:         req.Config.ExposeType,
			Hostname:           pulumi.String(req.Config.Hostname),
			Identity:           pulumi.String(req.Identity),
			Files:              pulumi.ToStringMap(req.Config.Files),
			IngressAnnotations: pulumi.ToStringMap(req.Config.IngressAnnotations),
			IngressNamespace:   pulumi.String(req.Config.IngressNamespace),
			IngressLabels:      pulumi.ToStringMap(req.Config.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		// Template connection info
		resp.ConnectionInfo = cm.URL.ApplyT(func(url string) (string, error) {
			values := &Values{
				URL: url,
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
