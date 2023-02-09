$filesExist=$true
$filesExist=$filesExist -and (Test-Path -Path ./overlays/production/ingress.yaml -PathType Leaf)
echo "$file exists: $filesExist"
if (-not $filesExist) {Exit 1}
