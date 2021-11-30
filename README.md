# draftv2 - electric boogaloo

Draftv2 (placeholder name) helps developers get started in the k8s space.
1. `draftv2 create` builds the necessary artifacts an application needs to run on k8s
## Building

run `make` in the repo dir

## Messing Around

After you have your binary built try the following commands in a directory with a webapp:
1. `draftv2 create` -- follow the prompts
2. If you chose `kustomize` run `skaffold init` then `skaffold dev`, otherwise just run `skaffold dev` for helm

-- tested with a go_app on port 8080 on a local minikube cluster with an ingress installed
