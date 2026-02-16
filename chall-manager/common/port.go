package common

import (
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
)

type PortArgs struct {
	Port        int               `form:"port"        json:"port"`
	Protocol    string            `form:"protocol"    json:"protocol"`
	ExposeType  k8s.ExposeType    `form:"exposeType"  json:"exposeType"`
	Annotations map[string]string `form:"annotations" json:"annotations,omitempty"`
}
