package challmanager

import (
	"fmt"
	"reflect"

	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/go-viper/mapstructure/v2"
)

// ExposeTypeHook transforms string to Chall-Manager SDK's Kubernetes
// ExposeType corresponding value.
var ExposeTypeHook mapstructure.DecodeHookFuncType = func(f reflect.Type, t reflect.Type, data any) (any, error) {
	if t != reflect.TypeOf(k8s.ExposeType(0)) {
		return data, nil
	}

	switch f.Kind() {
	case reflect.String:
		str := data.(string)
		switch str {
		case "NodePort":
			return k8s.ExposeNodePort, nil
		case "Ingress":
			return k8s.ExposeIngress, nil
		default:
			return data, fmt.Errorf("unsupported expose type: %s", str)
		}
	}
	return data, nil
}
