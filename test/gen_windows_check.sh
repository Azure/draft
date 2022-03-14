deployDir="./deployTypes"

deployTypes=("helm" "kustomize")
let count=0
for deploy in ${deployTypes[@]};do
    scriptName=./test/check_windows_$deploy.ps1
    echo "listing files for $deploy"
    files=$(find $deployDir/$deploy -type f)

    rm $scriptName
    touch $scriptName
    echo "\$filesExist=\$true" >> $scriptName
    for file in $files
    do
      echo "\$filesExist=\$filesExist -and (Test-Path -Path $file -PathType Leaf)" >> $scriptName
    done
    echo "if (-not \$filesExist) {Exit 1}" >> $scriptName

done