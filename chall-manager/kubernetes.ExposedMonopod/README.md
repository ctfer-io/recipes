# Kubernetes ExposedMonopod

This recipe is based upon [Chall-Manager ExposedMonopod SDK binding](https://github.com/ctfer-io/chall-manager/blob/main/sdk/kubernetes/exposed-monopod.go).
It does not add more values to it, but enable parametrizing inputs and outputs formats through additional values.

> [!NOTE]
> The inputs and outputs are technically both input, as they come from additional values from the challenge and the instance.
> These are separated in the two groups to better understand their impact on the deployment.

## Inputs

| Form Path | Description |
|---|---|
| `image` | **Required**. The Docker image reference to deploy. |
| `ports[x].port` | At least one port is required. Define the ports, protocols and expose type for the container. |
| `hostname` | **Required**. The hostname to use as part of URLs in the connection info. |
| `files` | The files to mount and on which path. |
| `ingressAnnotations` | The ingress annotations to use. |
| `ingressNamespace` | The namespace of the ingress controller to grant network access from. Required if any port is use `exposeType=Ingress`. |
| `ingressLabels` | **Required**. The labels of the ingress controller pods to grant network access from. Required if any port is use `exposeType=Ingress`. |

## Outputs

| Form Path | Description |
|---|---|
| `connectionInfo` | **Required**. The Go template to define the `connection_info` Chall-Manager must return for each instance. Example: `http://{{ index .URLs "8080/TCP"}}` returns a URL for a container that listens on port 8080 over TCP (e.g. gRPC or HTTP server). You can use the [`sprig`](https://masterminds.github.io/sprig/) functions. |

Notice that using Go templates and [`sprig`](https://masterminds.github.io/sprig/) you can extract specific parts of the output you want.
Follows an example that is used for SSH-based connections.

```gotmpl
{{- $hostport := index .URLs "8080/TCP" -}}
{{- $parts := splitList ":" $hostport -}}
{{- $host := index $parts 0 -}}
{{- $port := index $parts 1 -}}
ssh -p {{ $port }} {{ $host }}
```
