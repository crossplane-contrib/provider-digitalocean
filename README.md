# provider-digitalocean

## Archival Notice

**This repository, `crossplane-contrib/provider-digitalocean`, is now archived and is no longer actively maintained.**

Users are strongly encouraged to migrate to the officially supported [Crossplane provider for DigitalOcean based on Upjet](https://github.com/crossplane-contrib/provider-upjet-digitalocean).

**Reasons for Archival:**

* The `provider-upjet-digitalocean` offers improved maintainability and aligns with the Crossplane community's best practices.
* Upjet-based providers leverage code generation, ensuring better consistency and reducing manual maintenance overhead.
* The newer provider benefits from continuous updates and bug fixes.

**Migration Instructions:**

1.  Uninstall the existing `provider-digitalocean` from your Crossplane cluster.
2.  Install `provider-upjet-digitalocean` following the instructions in its repository.
3.  Update your Crossplane compositions and managed resources to use the API types and resource definitions provided by `provider-upjet-digitalocean`. Refer to the provider's documentation for details on API changes.

**Please update your bookmarks and references to the new repository.**

-------------

## Overview

`provider-digitalocean` is the Crossplane infrastructure provider for the
[DigitalOcean](https://www.digitalocean.com/). The provider that is built from the source
code in this repository can be installed into a Crossplane control plane and
adds the following new functionality:

* Custom Resource Definitions (CRDs) that model DigitalOcean infrastructure and services
  (e.g. [Droplets](https://www.digitalocean.com/products/droplets/), etc.)
* Controllers to provision these resources in DigitalOcean based on the users
  desired state captured in CRDs they create
* Implementations of Crossplane's portable resource abstractions, enabling DigitalOcean
  resources to fulfill a user's general need for cloud services

## Getting Started and Documentation

For getting started guides, installation, deployment and administration of Crossplane can be found on the official [Documentation](https://crossplane.io/docs/).
For the first steps with the digitalocean provider you can have a look at the [installation guide](./INSTALL.md).

## Contributing

provider-digitalocean is a community driven project and we welcome contributions. See
the Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

We also have a local [Contributing](docs/CONTRIBUTING.md) document to help you get your local environment setup to contribute.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-digitalocean/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Code of Conduct

provider-digitalocean adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-digitalocean is under the Apache 2.0 license.

[![FOSSA
Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-digitalocean.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-digitalocean?ref=badge_large)
