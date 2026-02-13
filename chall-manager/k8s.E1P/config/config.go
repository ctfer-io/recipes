package config

import (
	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	// Inputs

	Image            string              `form:"image"            json:"images"`
	Ports            []PortArgs          `form:"ports"            json:"ports"`
	Envs             map[string]string   `form:"envs"             json:"envs,omitempty"`
	Hostname         string              `form:"hostname"         json:"hostname"`
	Files            map[string]Variable `form:"files"            json:"files,omitempty"`
	FromCIDR         string              `form:"fromCidr"         json:"fromCidr"`
	IngressNamespace string              `form:"ingressNamespace" json:"ingressNamespace"`
	IngressLabels    map[string]string   `form:"ingressLabels"    json:"ingressLabels"`
	LimitCPU         *string             `form:"limitCpu"         json:"limitCpu,omitempty"`
	LimitMemory      *string             `form:"limitMemory"      json:"limitMemory,omitempty"`

	// Outputs

	ConnectionInfo string `form:"connectionInfo" json:"connectionInfo"`
}

type PortArgs struct {
	Port        int               `form:"port"        json:"port"`
	Protocol    string            `form:"protocol"    json:"protocol"`
	ExposeType  k8s.ExposeType    `form:"exposeType"  json:"exposeType"`
	Annotations map[string]string `form:"annotations" json:"annotations,omitempty"`
}

// Variable represent a content that can be variated.
type Variable struct {
	// The content to set.
	Content string `form:"content" json:"content"`

	// Whether to variate the content according per a PRNG seeded by the instance's identity (reproducible).
	Variate bool `form:"variate" json:"variate"`

	// Variation functional options

	Lowercase *bool `form:"lowercase" json:"lowercase,omitempty"`
	Uppercase *bool `form:"uppercase" json:"uppercase,omitempty"`
	Numeric   *bool `form:"numeric"   json:"numeric,omitempty"`
	Special   *bool `form:"special"   json:"special,omitempty"`
}

// Produce the content given its configuration, and a seed (should be the instance identity
// for proper reproducibility).
func (v Variable) Produce(seed string) string {
	if !v.Variate {
		return v.Content
	}

	return sdk.Variate(seed, v.Content,
		sdk.WithLowercase(v.Lowercase == nil || *v.Lowercase),
		sdk.WithUppercase(v.Uppercase == nil || *v.Uppercase),
		sdk.WithNumeric(v.Numeric == nil || *v.Numeric),
		sdk.WithSpecial(v.Special != nil && *v.Special),
	)
}
