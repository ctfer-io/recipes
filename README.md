<div align="center">
    <h1>Recipes</h1>
    <p><b>A collection of pre-made scenarios entirely customisable with additional values, through a simple API.</b><p>
    <a href="https://pkg.go.dev/github.com/ctfer-io/recipes"><img src="https://shields.io/badge/-reference-blue?logo=go&style=for-the-badge" alt="reference"></a>
	<a href="https://goreportcard.com/report/github.com/ctfer-io/recipes"><img src="https://goreportcard.com/badge/github.com/ctfer-io/recipes?style=for-the-badge" alt="go report"></a>
	<a href="https://coveralls.io/github/ctfer-io/recipes?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/ctfer-io/recipes?style=for-the-badge" alt="Coverage Status"></a>
	<br>
	<a href=""><img src="https://img.shields.io/github/license/ctfer-io/recipes?style=for-the-badge" alt="License"></a>
	<a href="https://github.com/ctfer-io/recipes/actions/workflows/codeql-analysis.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/recipes/codeql-analysis.yaml?style=for-the-badge&label=CodeQL" alt="CodeQL"></a>
    <br>
    <a href="https://securityscorecards.dev/viewer/?uri=github.com/ctfer-io/recipes"><img src="https://img.shields.io/ossf-scorecard/github.com/ctfer-io/recipes?label=openssf%20scorecard&style=for-the-badge" alt="OpenSSF Scoreboard"></a>
</div>

> [!CAUTION]
> Recipes are **highly experimental** thus are subject to major refactoring and breaking changes.

The recipes avoid reinventing the wheel for common stuff, like deploying a container in Kubernetes, with no need for re-compiling: we distribute OCI scenarios as [release artifacts](https://github.com/ctfer-io/recipes/releases) ! 🎉

## Load into OCI registry

### From Docker Hub

You can [find the recipes images on our Docker Hub](https://hub.docker.com/u/ctferio?page=1&search=recipes).

Then, simply copy it where you need!
```bash
oras cp <image_from_docker_hub> <image_to_your_registry> 
```

Of course, you can directly use it for simplicity.

That's all :stuck_out_tongue_winking_eye::muscle:

### From GitHub release assets

The following example focus on how to download the `debug` recipe then load it into an OCI registry, from a GitHub release asset.
All commands should **run in the same terminal**.

Requirements:
- [`jq`](https://jqlang.org/) ;
- [ORAS](https://oras.land/).

> [!TIP]
> We don't explain how to start an OCI registry, perhaps if necessary you can use the [Docker registry](https://hub.docker.com/_/registry).

1. Download a recipe (here we use the `debug` recipe, change to your needs):
    ```bash
    export LATEST=$(curl -s "https://api.github.com/repos/ctfer-io/recipes/tags" | jq -r '.[0].name')
    wget "https://github.com/ctfer-io/recipes/releases/download/${LATEST}/recipes_debug_${LATEST}.oci.tar.gz"
    ```

2. Untar:
    ```bash
    export DIR="debug-oci-layout"
    mkdir -p "${DIR}"
    tar -xzf "recipes_debug_${LATEST}.oci.tar.gz" -C "${DIR}/"
    ```

3. Copy to registry:
    ```bash
    export REGISTRY="localhost:5000"
    oras cp --from-oci-layout "./${DIR}:${LATEST}" "${REGISTRY}/debug:${LATEST}"
    ```
