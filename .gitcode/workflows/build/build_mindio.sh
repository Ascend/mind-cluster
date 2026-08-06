#!/bin/sh

set -e
servicename=$1
BUILDNUMBER=$2
pr_id=$3
ATOMGIT_WORKSPACE=$4
ostype=`arch`
GOPATH=${ATOMGIT_WORKSPACE}

file_path='change.txt'
cd ${ATOMGIT_WORKSPACE}
if grep -q "mindio" "${ATOMGIT_WORKSPACE}/$file_path"; then
    pip3 show setuptools
    pip3 install wheel
    pip3 install --upgrade setuptools
    cd ${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename_1} ${env.ATOMGIT_BASE_REF} mindio acp
    cd ${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename_1} ${env.ATOMGIT_BASE_REF} mindio tft

    ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/acp/output
    chmod 755 ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/acp/output/*.whl
    ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/tft/output
    chmod 755 ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/tft/output/*.whl

fi
