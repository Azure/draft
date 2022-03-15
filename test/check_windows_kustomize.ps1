$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./base/deployment.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./base/ingress.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./base/kustomization.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./base/service.yaml -PathType Leaf)
echo "$file exists: $filesExist"
if (-not $filesExist) {Exit 1}
