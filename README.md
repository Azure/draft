<div id="top"></div>

<br />
<div align="center">
  <h1 align="center">Draft</h1>
  <p align="center">
    A tool to help developers hit the ground running with Kubernetes.
    <br />
    <br />
    <a href="https://github.com/Azure/draft/issues">Report Bug</a>
    ·
    <a href="https://github.com/Azure/draft/issues">Request Feature</a>
  
  [![Draft Unit Tests](https://github.com/Azure/draft/actions/workflows/unit-tests.yml/badge.svg)](https://github.com/Azure/draft/actions/workflows/unit-tests.yml)
  [![GoDoc](https://godoc.org/github.com/Azure/draft?status.svg)](https://godoc.org/github.com/Azure/draft)
  [![Go Report Card](https://goreportcard.com/badge/github.com/Azure/draft)](https://goreportcard.com/report/github.com/Azure/draft)
  [![CodeQL](https://github.com/Azure/draft/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/Azure/draft/actions/workflows/codeql-analysis.yml)
  [![Draft Linux Integrations](https://github.com/Azure/draft/actions/workflows/integration-linux.yml/badge.svg)](https://github.com/Azure/draft/actions/workflows/integration-linux.yml)
  [![Draft Release & Publish](https://github.com/Azure/draft/actions/workflows/release-and-publish.yml/badge.svg)](https://github.com/Azure/draft/actions/workflows/release-and-publish.yml)
  </p>
</div>

## Getting Started

Draft is a tool made for users who are just getting started with Kubernetes, or users who want to simplify their experience with Kubernetes. This readme will give you a quick run down on Draft’s commands and what they do.

### `draft create`

In our directory that holds our application, we can run the CLI command ‘draft create’. Draft create will walk you through a series of questions prompting you on your application specification. At the end of it, you will have a Dockerfile as well as Kubernetes manifests to deploy your application. Below is a picture of running the Draft create command on our [Contoso Air repository](https://github.com/microsoft/ContosoAir).

![example of draft create command showing the prompt "select k8s deployment type" with three options "helm", "kustomize", and "manifests"](./ghAssets/draft-create.png)

### `generate-workflow`

Next up, we can run the ‘draft generate-workflow’ command.
This command will automatically build out a GitHub Action for us.
![screenshot of command line executing "draft generate-workflow" printing "Draft has successfully genereated a Github workflow for your project"](./ghAssets/generate-workflow.png)

### `setup-gh`

If you are using Azure, you can also run the ‘draft setup-gh’ command to automate the GitHub OIDC setup process. This process is needed to make sure your Azure account and your GitHub repository can talk to each other. If you plan on using the GitHub Action to deploy your application, this step must be completed.

![screenshot of command line executing "draft setup-gh" showing the prompt "Which account do you want to log into?" with two options "Github.com" and "Github Enterprise Server"](./ghAssets/setup-gh.png)

At this point, you have all the files needed to deploy your application onto a Kubernetes cluster!

If you don’t plan on using the GitHub Action, you can directly apply your deployment files by using the `kubectl apply -f` command.

If you plan on deploying your application through your GitHub Action, commit all the files to your repository and watch your application get deployed!

### `draft info`
The `draft info` command prints information about supported languages and deployment types.
Example output:
```
{
  "supportedLanguages": [
    {
      "name": "clojure",
      "displayName": "Clojure",
      "variableExampleValues": {
        "VERSION": [
          "8-jdk-alpine",
          "11-jdk-alpine"
        ]
      }
    }
  ],
  "supportedDeploymentTypes": [
    "helm",
    "kustomize",
    "manifests"
  ]
}
```
<!-- ABOUT THE PROJECT -->

## About The Project

Draft makes it easier for developers to get started building apps that run on Kubernetes by taking a non-containerized application and generating the Dockerfiles, Kubernetes manifests, Helm charts, Kustomize configuration, and other artifacts associated with a containerized application. Draft can also generate a GitHub Actions workflow file to quickly build and deploy applications onto any Kubernetes cluster.

### Commands

- `draft create` adds the minimum required Dockerfile and manifest files for your deployment to the project directory.
  - Supported deployment types: Helm, Kustomize, Kubernetes manifest.
- `draft setup-gh` automates the GitHub OIDC setup process for your project.
- `draft generate-workflow` generates a GitHub Actions workflow for automatic build and deploy to a Kubernetes cluster.
- `draft update` automatically make your application to be internet accessible.
- `draft info` print supported language and field information in json format.

Use `draft [command] --help` for more information about a command.

## Prerequisites

Draft requires Go version 1.18.x.

Check your go version.

```sh
go version
```

## Installation

1. Clone the repo

   ```sh
   git clone https://github.com/Azure/draft.git
   ```

2. Change to the `draft` directory and build the binary

   ```sh
   cd draft/
   make
   ```

3. Add the binary to your path (we use the same directory as [go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies))

   ```sh
   mv draft $HOME/go/bin/
   ```

### Install with HomeBrew

1. Run the following commands

   ```sh
   $ brew tap azure/draft
   $ brew install draft
   ```

### Installation using script

```sh
curl -fsSL https://raw.githubusercontent.com/Azure/draft/main/scripts/install.sh | bash
```

* Windows isn't currently supported (you can use WSL)

## Contributing

Draft is fully compatible with [Azure Kubernetes Service](https://docs.microsoft.com/azure/aks/draft). We strongly encourage contributions to make Draft available to other cloud providers 😊!

## Issues/Discussions

The Draft team will be monitoring both the [issues](https://github.com/Azure/draft/issues) and [discussions](https://github.com/Azure/draft/discussions) board. Please feel free to create issues for any problems you run into and the Draft team will be quick to respond. The discussions board will be used for community engagement. We look forward to see you there!

## License

Draft is under the MIT License. See [LICENSE](https://github.com/Azure/draft/blob/main/LICENSE) for more information.

## Trademark Notice

Authorized use of Microsoft trademarks or logos is subject to and must follow Microsoft’s Trademark & Brand Guidelines. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party’s policies.
