<div id="top"></div>

<br />
<div align="center">
  <h1 align="center">Draft</h1>
  <p align="center">
    A tool to help developers hit the ground running with Kubernetes.
    <br />
    <a href="https://github.com/Azure/draft"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/Azure/draft/issues">Report Bug</a>
    ·
    <a href="https://github.com/Azure/draft/issues">Request Feature</a>
  </p>
</div>


<!-- ABOUT THE PROJECT -->
## About The Project

Draft makes it easier for developers to get started building apps that run on Kubernetes by taking a non-containerized application and generating the Dockerfiles, Kubernetes manifests, Helm charts, Kustomize configuration, and other artifacts associated with a containerized application. Draft can also generate a GitHub Action workflow file to quickly build and deploy applications onto any Kubernetes cluster.

### Commands

* `draft create` adds the minimum required Dockerfile and manifest files for your deployment to the project directory.
  * Supported deployment types: Helm, Kustomize, Kubernetes manifest.
* `draft setup-gh` automates the Github OIDC setup process for your project.
* `draft generate-workflow` generates a Github Action workflow for automatic build and deploy to a Kubernetes cluster.
* `draft update` automatically make your application to be internet accessible.

Use `draft [command] --help` for more information about a command.

## Prerequisites
Draft requires Go version 1.18.x.
* Go
  ```sh
  go version
  ```

## Installation

1. Clone the repo
   ```sh
   git clone https://github.com/Azure/draft.git
   ```
2. Build the binary
   ```sh
   make
   ```
3. Add the binary to your path
   ```sh
   mv draft $GOPATH/bin/
   ```

## Contributing
Draft is fully compatible with [Azure Kubernetes Services](https://docs.microsoft.com/en-ca/azure/aks/draft). We strongly encourage contributions to make Draft available to other cloud providers 😊!

## License

Draft is under the MIT License. See [LICENSE](https://github.com/Azure/draft/blob/main/LICENSE) for more information.

## Trademark Notice
Authorized use of Microsoft trademarks or logos is subject to and must follow Microsoft’s Trademark & Brand Guidelines. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party’s policies.

