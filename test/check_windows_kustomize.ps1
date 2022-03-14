$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/kustomize/draft.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/kustomize/base/deployment.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/kustomize/base/ingress.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/kustomize/base/kustomization.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/kustomize/base/service.yaml -PathType Leaf)
if (-not $filesExist) {Exit 1}
