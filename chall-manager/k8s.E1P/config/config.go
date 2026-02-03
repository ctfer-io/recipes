package config

import (
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	// Inputs

	Image            string            `form:"image"`
	Ports            []PortArgs        `form:"ports"`
	Envs             map[string]string `form:"envs"`
	Hostname         string            `form:"hostname"`
	Files            map[string]string `form:"files"`
	FromCIDR         string            `form:"fromCidr"`
	IngressNamespace string            `form:"ingressNamespace"`
	IngressLabels    map[string]string `form:"ingressLabels"`
	LimitCPU         *string           `form:"limitCpu"`
	LimitMemory      *string           `form:"limitMemory"`

	// Outputs

	ConnectionInfo string `form:"connectionInfo"`
}

type PortArgs struct {
	Port        int               `form:"port"`
	Protocol    string            `form:"protocol"`
	ExposeType  k8s.ExposeType    `form:"exposeType"`
	Annotations map[string]string `form:"annotations"`
}
