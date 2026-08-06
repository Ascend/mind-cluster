#!/bin/sh
SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/utils.sh"

set -ex
reponame=$1
TARGET_BRANCH=$2
servicename=$3
targetname=$4

GOPATH=${WORKSPACE}

machine_arch=`uname -m`
if [ $machine_arch = "aarch64" ]; then
 REL_OSARCH="arm64"
fi

export GOROOT=/opt/buildtools/go_1.26.1
export GO111MODULE='auto'
export GOSUMDB=sum.golang.google.cn
export GOPROXY="https://repo.huaweicloud.com/repository/goproxy/"
export PYTHON_HOME=/usr/local/python3.7.5/bin
export PATH=${PYTHON_HOME}:$GOROOT/bin:$PATH
export GOPATH=${WORKSPACE}

function update_version_number() {
    set -x
    VERSION_NUMBER=$(get_version_number "$TARGET_BRANCH")
    if [ $? -ne 0 ]; then
        echo "get Version_Number failed, please check."
        exit 1
    fi
    echo "VERSION_NUMBER: ${VERSION_NUMBER}"
    cat "$GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini"
    sed -i "s/VERSION_NUMBER/${VERSION_NUMBER}/g"  "$GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini"

    cat "$GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini"
}

update_version_number

if [ "${servicename}" = "ascend-for-volcano" ]; then
    cp -rf $GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini $GOPATH/src/volcano.sh/volcano/
    ls -la $GOPATH/src/volcano.sh/volcano/
    echo "********$2*********"
    cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
    dos2unix *.sh && chmod +x *
    ./build.sh $2
elif [ "${servicename}" = "mindio" ]; then
    cp -rf $GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini $GOPATH/mind-cluster/component/${servicename}/${targetname}/
    ls -la $GOPATH/
    cd $GOPATH/mind-cluster/component/${servicename}/${targetname}/build
    dos2unix *.sh && chmod +x *
    ./build.sh
elif [ "${servicename}" = "helm-deploy-tool" ]; then
    cp -rf $GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini $GOPATH/mind-cluster/${servicename}/
    ls -la $GOPATH/
    cd $GOPATH/mind-cluster/${servicename}/build
    dos2unix *.sh && chmod +x *
    ./build.sh
else
    cp -rf $GOPATH/${reponame}/.gitcode/workflows/config/service_config.ini $GOPATH/mind-cluster/component/${servicename}/
    ls -la $GOPATH/
    cd $GOPATH/mind-cluster/component/${servicename}/build
    dos2unix *.sh && chmod +x *
    ./build.sh
fi
