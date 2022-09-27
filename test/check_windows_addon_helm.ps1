$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./charts/templates/ingress.yaml -PathType Leaf)
echo "$file exists: $filesExist"
if (-not $filesExist) {Exit 1}
