deployDir="./deployTypes"

deployTypes=("helm" "kustomize")
let count=0
for deploy in ${deployTypes[@]};do
    echo "listing files for $deploy"
    files=$(find $deployDir/$deploy)

    touch ./test/check_windows_$deploy.ps1
    for file in $files
    do
      echo "Test-Path -Path $file -PathType Leaf" >> ./test/check_windows_$deploy.ps1
    done
done