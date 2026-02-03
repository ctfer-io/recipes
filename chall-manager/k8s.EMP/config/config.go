package config

import (
	e1p "github.com/ctfer-io/recipes/chall-manager/k8s.E1P/config"
)

type Config struct {
	// Inputs

	Containers       map[string]ContainerArgs `form:"containers"`
	Rules            []RuleArgs               `form:"rules"`
	Hostname         string                   `form:"hostname"`
	FromCIDR         string                   `form:"fromCidr"`
	IngressNamespace string                   `form:"ingressNamespace"`
	IngressLabels    map[string]string        `form:"ingressLabels"`

	// Outputs

	ConnectionInfo string `form:"connectionInfo"`
}

type ContainerArgs struct {
	Image       string            `form:"image"`
	Ports       []e1p.PortArgs    `form:"ports"`
	Envs        map[string]string `form:"envs"`
	Files       map[string]string `form:"files"`
	LimitCPU    *string           `form:"limitCpu"`
	LimitMemory *string           `form:"limitMemory"`
}

type RuleArgs struct {
	From     string `form:"from"`
	To       string `form:"to"`
	On       int    `form:"on"`
	Protocol string `form:"protocol"`
}
