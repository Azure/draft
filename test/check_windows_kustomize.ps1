Test-Path -Path ./deployTypes/kustomize -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/draft.yaml -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/base -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/base/deployment.yaml -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/base/ingress.yaml -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/base/kustomization.yaml -PathType Leaf
Test-Path -Path ./deployTypes/kustomize/base/service.yaml -PathType Leaf
