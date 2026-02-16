package config

import (
	"github.com/ctfer-io/recipes/chall-manager/common"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	// Inputs

	Image            string                     `form:"image"            json:"image"`
	Ports            []common.PortArgs          `form:"ports"            json:"ports"`
	Envs             map[string]common.Variable `form:"envs"             json:"envs,omitempty"`
	Hostname         string                     `form:"hostname"         json:"hostname"`
	Files            map[string]common.Variable `form:"files"            json:"files,omitempty"`
	FromCIDR         string                     `form:"fromCidr"         json:"fromCidr"`
	IngressNamespace string                     `form:"ingressNamespace" json:"ingressNamespace"`
	IngressLabels    map[string]string          `form:"ingressLabels"    json:"ingressLabels,omitempty"`
	Requests         map[string]string          `form:"requests"         json:"requests,omitempty"`
	Limits           map[string]string          `form:"limits"           json:"limits,omitempty"`

	// Outputs

	ConnectionInfo string `form:"connectionInfo" json:"connectionInfo"`
}
