# draftv2 - electric boogaloo

Draftv2 (placeholder name) helps developers get started in the k8s space.
1. `draftv2 create` builds the necessary artifacts an application needs to run on k8s
## Building

run `make` in the repo dir

### Messing Around

After you have your binary built try the following commands in a directory with a webapp:
1. `draftv2 create` -- follow the prompts
2. If you chose `kustomize` run `skaffold init` then `skaffold dev`, otherwise just run `skaffold dev` for helm

-- tested with a go_app on port 8080 on a local minikube cluster with an ingress installed


## Developing

NOTE: Due to `embed` not allowing the embedding on `_` or `.` prefixed files, a config based workaround is being used. Once [this](https://github.com/golang/go/commit/36dbf7f7e63f3738795bb04593c3c011e987d1f3) is merged, the workaround will be removed.

### Builders
The `builders` directory is home to minimum viable Dockerfiles needed for containerizing an application in a given language.
Currently, we use a simple templating method: wrapping a variable in `{{}}` without inner spaces (will be swapped out eventually).

This is intended for use with skaffold.
### DeployTypes

The `deployTypes` directory contains the minimum k8s files needed to deploy the Dockerfile to a k8s cluster
