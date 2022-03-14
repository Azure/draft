$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/draft.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/Chart.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/helmignore -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/deployment.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/NOTES.txt -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/ingress.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/tests/test-connection.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/service.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/serviceaccount.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/templates/helpers.tpl -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/charts/values.yaml -PathType Leaf)
$filesExist=$filesExist -and (Test-Path -Path ./deployTypes/helm/skaffold.yaml -PathType Leaf)
if (-not $filesExist) {Exit 1}
