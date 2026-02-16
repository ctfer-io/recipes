package config

import (
	"github.com/ctfer-io/recipes/chall-manager/common"

	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
)

type Config struct {
	// Inputs

	Containers       map[string]ContainerArgs `form:"containers"              json:"containers"`
	Rules            []RuleArgs               `form:"rules"                   json:"rules"`
	Hostname         string                   `form:"hostname"                json:"hostname"`
	FromCIDR         string                   `form:"fromCidr"                json:"fromCidr"`
	IngressNamespace string                   `form:"ingressNamespace"        json:"ingressNamespace"`
	IngressLabels    map[string]string        `form:"ingressLabels,omitempty" json:"ingressLabels,omitempty"`

	// Outputs

	ConnectionInfo string `form:"connectionInfo" json:"connectionInfo"`
}

type ContainerArgs struct {
	Image    string                     `form:"image"       json:"image"`
	Ports    []common.PortArgs          `form:"ports"       json:"ports"`
	Envs     map[string]Printable       `form:"envs"        json:"envs"`
	Files    map[string]common.Variable `form:"files"       json:"files"`
	Requests map[string]string          `form:"requests"    json:"requests"`
	Limits   map[string]string          `form:"limits"      json:"limits"`
}

type RuleArgs struct {
	From     string `form:"from"     json:"from"`
	To       string `form:"to"       json:"to"`
	On       int    `form:"on"       json:"on"`
	Protocol string `form:"protocol" json:"protocol"`
}

type Printable struct {
	Variable common.Variable `form:"variable" json:"variable"`

	Format   string   `form:"format"   json:"format"`
	Serivces []string `form:"services" json:"services"`
}

func (pr Printable) ToPrinter(seed string) k8s.PrinterArgs {
	if pr.Variable.Content != "" {
		return k8s.NewPrinter(pr.Variable.Produce(seed))
	}
	return k8s.NewPrinter(pr.Format, pr.Serivces...)
}
