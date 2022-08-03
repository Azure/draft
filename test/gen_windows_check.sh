deployDir="./deployTypes"


deployTypes=("helm" "kustomize")

ignoredFiles=("./draft.yaml" "./skaffold.yaml" "./charts/templates/helpers.tpl" "./charts/helmignore")
let count=0
for deploy in ${deployments[@]};do
    scriptName=./test/check_windows_$deploy.ps1
    echo "listing files for $deploy"
    cd $deployDir/$deploy
    files=$(find . -type f)
    cd ../..

    rm $scriptName
    touch $scriptName
    echo "\$filesExist=\$true" >> $scriptName
    for file in $files
    do
      ignoreFile=0
      for ignore in ${ignoredFiles[@]};do
        echo "${ignore}"
        echo "${file}"
        if [[ "$ignore" == "$file" ]];then
          ignoreFile=1
        fi
      done

      if [[ $ignoreFile == 0 ]]; then

        echo "\$filesExist=\$filesExist -and (Test-Path -Path $file -PathType Leaf)" >> $scriptName
        echo 'echo "$file exists: $filesExist"' >> $scriptName

      fi
    done
    echo "if (-not \$filesExist) {Exit 1}" >> $scriptName

done