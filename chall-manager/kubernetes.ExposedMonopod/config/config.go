package config

import (
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
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
