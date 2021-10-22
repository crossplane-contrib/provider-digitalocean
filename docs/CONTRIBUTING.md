# Local Environment Setup

We have included some makefile targets to ease the local development process. You will need the following tools installed to get started.

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Crossplane CLI](https://crossplane.io/docs/v1.4/getting-started/install-configure.html#install-crossplane-cli) (Not required, but suggested)

You can then run `make` which will build the project and setup the build submodule. Once this is finished you can then run `make dev` which will boot up a kind cluster, install Crossplane, any CRDs for the project, and then start the provider.

You can run `make dev-clean` to then cleanup the cluster, or `make dev-restart` which will run the `dev-clean` and then the `dev` targets.
