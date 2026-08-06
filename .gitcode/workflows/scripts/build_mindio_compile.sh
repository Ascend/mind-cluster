set -x

servicename_1=$(echo "${{ atomgit.repository }}" | cut -d '/' -f 2)
if [ "${{ inputs.TARGET_BRANCH }}" = "branch_v7.3.0" -o "${{ inputs.TARGET_BRANCH }}" = "master" -o "${{ inputs.TARGET_BRANCH }}" = "branch_v26.0.0" ];then
    echo "result=execute" >> $ATOMGIT_OUTPUT
    ls -al /opt/buildtools

    export GOROOT=/opt/buildtools/go_1.21.12
    export GO111MODULE='auto'
    export GONOSUMDB=*
    export GOPROXY="https://goproxy.cn,direct"
    export PATH=${PYTHON_HOME}:$GOROOT/bin:$PATH
    export GOPATH=${ATOMGIT_WORKSPACE}

    cd ${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/build && sh build_mindio.sh ${servicename_1} ${{ env.RUN_NUMBER }} ${{ inputs.pr_id }} ${ATOMGIT_WORKSPACE}
else
    echo "result=skip" >> $ATOMGIT_OUTPUT
    echo "****************************no build***************************************"
fi
cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/output/mindio_acp-*.whl ${ATOMGIT_WORKSPACE}/artifacts/
cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/output/mindio_ttp-*.whl ${ATOMGIT_WORKSPACE}/artifacts/
