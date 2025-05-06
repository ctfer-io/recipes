# Kubernetes ExposedMonopod

This recipe is based upon the official Chall-Manager ExposedMonopod SDK binding.
It does not add more values to it, but enable parametrizing inputs and outputs formats through additional values.

## Inputs

| Path | Description |
|---|---|
| `additional.image` | **Required**. The Docker image reference to deploy. |
| `additional.ports[x].port` | At least one port is required. Define the ports, protocols and expose type for the container. |
| `additional.hostname` | **Required**. The hostname to use as part of URLs in the connection info. |
| `additional.files` | The files to mount and on which path. |
| `additional.ingressAnnotations` | The ingress annotations to use. |
| `additional.ingressNamespace` | The namespace of the ingress controller to grant network access from. Required if any port is use `exposeType=Ingress`. |
| `additional.ingressLabels` | **Required**. The labels of the ingress controller pods to grant network access from. Required if any port is use `exposeType=Ingress`. |

## Outputs

| Path | Description |
|---|---|
| `additional.connectionInfo` | **Required**. The Go template to define the `connection_info` Chall-Manager must return for each instance. Example: `http://{{ index .Ports "8080/TCP"}}` returns a URL for a container that listens on port 8080 over TCP (e.g. gRPC or HTTP server). You can use the [`sprig`](https://masterminds.github.io/sprig/) functions. |

Notice that using Go templates and [`sprig`](https://masterminds.github.io/sprig/) you can extract specific parts of the output you want.
Follows an example that is used for SSH-based connections.

```gotmpl
{{- $hostport := index .Ports "8080/TCP" -}}
{{- $parts := splitList ":" $hostport -}}
{{- $host := index $parts 0 -}}
{{- $port := index $parts 1 -}}
ssh -p {{ $port }} {{ $host }}
```

## Example

```bash
# Init your stack
export PULUMI_CONFIG_PASSPHRASE=""
pulumi stack init example

# Configure your stack
# -> Global
pulmui config set identity a0b1c2d3
# -> Inputs
pulumi config set --path 'additional.hostname' 'demo.ctfer.io'
pulumi config set --path 'additional.image' 'pandatix/license-lvl1:latest'
pulumi config set --path 'additional.ports[0].port' '8080'
pulumi config set --path 'additional.ports[0].exposeType' 'Ingress'
# -> Outputs
pulumi config set --path 'additional.connectionInfo' 'http://{{ index .Ports "8080/TCP"}}'

# Preview plan
pulumi preview
```
