# Reporting Security Issues

Please report any security issues you discovered in a recipe to ctfer-io@protonmail.com.

We will assess the risk, plus make a fix available before we create a GitHub issue.

In case the vulnerability is into a dependency, please refer to their security policy directly.

Thank you for your contribution.

## Refering to this repository

To refer to this repository using a CPE v2.3, please use `cpe:2.3:a:ctfer-io:recipes:*:*:*:*:*:*:*:*`.
This mostly contains the [`runner.go`](./runner.go) file.

You could decline for each scenario, if required, using the following rule:
- an `environment` must be selected in the root directory ;
- a `recipe` must be selected from the `environment`.
Then you can use `cpe:2.3:a:ctfer-io:recipes-<environment>-<recipe>:*:*:*:*:*:*:*:*`.

For instance, with the `chall-manager` environment and `kubernetes.ExposedMonopod` recipe, you end up with `cpe:2.3:a:ctfer-io:recipes-chall-manager-kubernetes.ExposedMonopod:*:*:*:*:*:*:*:*`.

A security analyst could capture all refinement with `cpe:2.3:a:ctfer-io:recipes*:*:*:*:*:*:*:*:*`.

Use with the `version` set to the tag you are using.
