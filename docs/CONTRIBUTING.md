# Contributing

We have included some makefile targets to ease the local development process. You will need the following tools
installed to get started.

## How-Tos

### Getting Started

1. Install [Kind][0].
2. Install [Kubectl][1].
3. Install [Crossplane CLI][2] (Not required, but suggested).
4. Install `go` version `1.18`.
5. Run `make submodules` to initialize the submodules.
6. Run `make`.
7. Run `make dev-kind` to start the cluster and install the provider.
8. Run `make dev-provider` to start the provider controller.
9. Rerun `make dev-provider` to restart the provider controller.

## Explanations

### Makefile

#### Working with the cluster

- `make dev-clean` destroy the `kind` cluster.
- `make dev-kind` will start the `kind` cluster and install `Crossplane` and any `CRDs` for the project.
- `make dev-provider` will start the provider controller.
- `make dev` will run `dev-kind` and `dev-provider`.

#### Working with CRDs

- `make crds.clean` will remove all `CRDs` from the cluster.
- `make crds.install` will install all `CRDs` from the project.

[0]: https://kind.sigs.k8s.io/docs/user/quick-start/

[1]: https://kubernetes.io/docs/tasks/tools/

[2]: https://docs.crossplane.io/v1.10/getting-started/install-configure#install-crossplane-cli
