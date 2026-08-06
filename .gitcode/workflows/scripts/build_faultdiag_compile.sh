set -ex

servicename_1=$(echo "${{ atomgit.repository }}" | cut -d '/' -f 2)
source ${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/build/utils.sh
VERSION_NUMBER=$(get_version_number "$TARGET_BRANCH")
if [ $? -ne 0 ]; then
    echo "get Version_Number failed, please check."
    exit 1
fi
sed -i "s/VERSION_NUMBER/${VERSION_NUMBER}/g"  "${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/config/service_config.ini"

if [ -d "${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/" ]; then
    cp ${ATOMGIT_WORKSPACE}/${servicename_1}/.gitcode/workflows/config/service_config.ini ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/
fi

if grep -q "ascend-faultdiag-online" "${ATOMGIT_WORKSPACE}/change.txt"; then
    echo "result=skip" >> $ATOMGIT_OUTPUT
    echo "skip, no need to compile"
elif grep -q "faultdiag" "${ATOMGIT_WORKSPACE}/change.txt"; then
    echo "result=execute" >> $ATOMGIT_OUTPUT
    pip3 install  numpy==1.26.4
    export PATH=/opt/buildtools/python-3.11.4/bin:$PATH
    export CC=/opt/rh/devtoolset-7/root/usr/bin/gcc
    export CXX=/opt/rh/devtoolset-7/root/usr/bin/g++
    export GCC_HOME=/opt/rh/devtoolset-7/root/usr/bin

    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/build/ && dos2unix *.sh && chmod +x * && bash build.sh
    ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output

else
    echo "result=skip" >> $ATOMGIT_OUTPUT
    mkdir -p ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output && touch empty.whl
    echo "skip，no need to compile"
fi
