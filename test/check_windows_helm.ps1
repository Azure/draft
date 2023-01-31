$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./charts/Chart.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/production.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/.helmignore -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/deployment.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/service.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/namespace.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/_helpers.tpl -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/values.yaml -PathType Leaf)
echo "$file exists: $filesExist"
if (-not $filesExist) {Exit 1}
