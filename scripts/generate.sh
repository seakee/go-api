#!/usr/bin/env bash
# 生成一个go-api的项目
# $1 为新项目名，如"project-name"，如果不传此参数则项目名为"go-api"
# $2 为版本号，如"v1.0.0"，如果不传次参数则默认从master获取最新代码

projectName="go-api"
projectVersion="main"

if [ "$1" != "" ]; then
  projectName="$1"
fi

if [ "$2" != "" ]; then
  projectVersion="$2"
fi

projectDir=$(pwd)/$projectName

rm -rf "$projectDir"

git clone -b "$projectVersion" https://github.com/seakee/go-api.git "$projectDir"

rm -rf "$projectDir"/.git

grep -rl 'github.com/seakee/go-api' "$projectDir" | xargs sed -i "" "s/github.com\/seakee\/go-api/$projectName/g"

grep -rl 'go-api' "$projectDir" | xargs sed -i "" "s/go-api/$projectName/g"

echo "Success！"
