# Kubernetes ExposedMultipod (k8s.EMP)

This recipe is based upon [Chall-Manager ExposedMultipod SDK binding](https://github.com/ctfer-io/chall-manager/blob/main/sdk/kubernetes/exposed-multipod.go).
It does not add more values to it, but enable parametrizing inputs and outputs formats through additional values.

> [!NOTE]
> The inputs and outputs are technically both input, as they come from additional values from the challenge and the instance.
> These are separated in the two groups to better understand their impact on the deployment.

## Inputs

| Form Path | Description |
|---|---|
| `containers[xxx].image` | **Required**. The Docker image reference to deploy. |
| `containers[xxx].ports[x].port` | At least one port is required. Define the ports, protocols and expose type for the container. |
| `containers[xxx].ports[x].protocol` | The protocol to expose the port on. |
| `containers[xxx].ports[x].exposeType` | The kind of exposure for this port/protocol couple. |
| `containers[xxx].ports[x].annotations` | A k=v map of annotations to pass to the exposing resource of this port/protocol couple. |
| `containers[xxx].envs` | A k=v map of environment variables to pass to the container. |
| `containers[xxx].files` | A k=v map of file path and content to mount in the container. |
| `containers[xxx].limitCpu` | The limit of CPU usage. Optional, yet recommended. |
| `containers[xxx].limitMemory` | The limit of memory usage. Optional, yet recommended. |
| `rules[x].from` | The container name from which to grant network interaction. |
| `rules[x].to` | The container name to which grant network interaction. |
| `rules[x].on` | The port to which grant network interaction. |
| `rules[x].protocol` | The protocol on which to grant 
| `hostname` | **Required**. The hostname to use as part of URLs in the connection info. |
| `fromCidr` | A CIDR from which to limit restrein access to the challenge. |
| `ingressNamespace` | The namespace of the ingress controller to grant network access from. Required if any port is use `exposeType=Ingress`. |
| `ingressLabels` | **Required**. The labels of the ingress controller pods to grant network access from. Required if any port is use `exposeType=Ingress`. |

## Outputs

| Form Path | Description |
|---|---|
| `connectionInfo` | **Required**. The Go template to define the `connection_info` Chall-Manager must return for each instance. Example: `http://{{ index .URLs "app" "8080/TCP"}}` returns a URL for the container "app" that listens on port 8080 over TCP (e.g. gRPC or HTTP server). You can use the [`sprig`](https://masterminds.github.io/sprig/) functions. |

Notice that using Go templates and [`sprig`](https://masterminds.github.io/sprig/) you can extract specific parts of the output you want.
Follows an example that is used for SSH-based connections, that is resilient to infrastructure errors.

```gotmpl
{{- $hostport := index .URLs "app" "8080/TCP" -}}
{{- $parts := splitList ":" $hostport -}}
{{- $host := "" -}}
{{- $port := "" -}}

{{- if ge (len $parts) 1 -}}
    {{- $host = index $parts 0 -}}
{{- end -}}

{{- if ge (len $parts) 2 -}}
    {{- $port = index $parts 1 -}}
{{- end -}}

{{- if and $host $port -}}
ssh -p {{ $host }} {{ $port }}
{{- else -}}
Host or port not available...
{{- end -}}
```
