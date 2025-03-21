# Recipes

Recipes are [**pre-built scenarios**](https://ctfer.io/docs/chall-manager/challmaker-guides/create-scenario/#design-your-pulumi-factory), especially for [⁠chall-manager](https://github.com/ctfer-io/chall-manager), that provide a simple API to configure the complexity of a Pulumi Go-based scenarios.
It eases the adoption of Chall-Manager on the simplest and widely adopted use cases.

For instance, to deploy a single pod on Kubernetes, you need some basic data: the image name, its port, the hostname to expose on, the connection info template, etc. Out of these configuration elements, it is always a copy-pasta, which leads to compiling multiple times the same scenario with only some configuration slight changes.
With the (not so) recent adoption of the additional on challenges and instances we may pass these configuration variables to a pre-compiled scenario, thus guarantee the quality of a scenario (enhance security + ease debug), reduce chall-manager Time To Deploy (TTD contains time to compile, time in API, time executed, punctuated by network latencies), and as for every pre-built scenario, ease offline adoption (a binary can be contained in a Hauler archive).

The core engineering idea behind is toward Model-Based approach, in order to make these scenarios readable, especially machine-readable for documentation generation.
Maturity: **low**

## Example

Still with the example of the single pod, we wanna create a kubernetes.ExposedMonopod. Such scenario could be defined per the attached image, built in CI on release, attested at SLSA 3 level (highest reachable per v1.0 for FOSS organizations), and reused whenever required.

## Future

The recipes are an essential step toward CTFOps, as the JSON schema comes from these recipes.
Documentation generation, JSON schema, default values, Pulumi stack configuration, etc. remains work to do.
