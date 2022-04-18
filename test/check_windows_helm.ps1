$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./charts/Chart.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/deployment.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/tests/test-connection.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/service.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/serviceaccount.yaml -PathType Leaf)
echo "$file exists: $filesExist"
$filesExist=$filesExist -and (Test-Path -Path ./charts/values.yaml -PathType Leaf)
echo "$file exists: $filesExist"
if (-not $filesExist) {Exit 1}
