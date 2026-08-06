set -ex
servicename=$1
TARGET_BRANCH=$2
# servicename=$(echo "${{ atomgit.repository }}" | cut -d '/' -f 2)
if [ "${TARGET_BRANCH}" = "v7.2rc1.ch" ];then

    ostype=`arch`
    GOPATH=${ATOMGIT_WORKSPACE}

    file_path='change.txt'
    cd ${ATOMGIT_WORKSPACE}
    if grep -q "clusterd" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh clusterd
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh clusterd ${TARGET_BRANCH}
        ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/clusterd/output
    fi
    if grep -q "ascend-device-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ascend-device-plugin
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-device-plugin ${TARGET_BRANCH}
    fi
    if grep -q "ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && tar -zxvf makeself/makeself-2.4.2.tar.gz
        cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && cp -rf makeself-header/makeself-header.sh makeself-release-2.4.2/
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ascend-docker-runtime
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-docker-runtime ${TARGET_BRANCH}
        cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/output
        zip Alan-docker-runtime_7.1.RC1_linux-${ostype}.zip Alan-docker-runtime_*_linux-${ostype}.run
    fi
    if grep -q "ascend-for-volcano" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}
        mkdir -p src/volcano.sh
        cp -rf /opt/buildtools/volcano_opensource/volcano_1.9/volcano src/volcano.sh/
        ls -la ./ &&  cp -rf mind-cluster/component/ascend-for-volcano src/volcano.sh/volcano/pkg/scheduler/plugins/
        cd src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
        ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ascend-for-volcano v1.9.0
        mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
        ls -la ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
        rm -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano
        echo "***************start complie volcano 1.7***********************"
        cp -rf /opt/buildtools/volcano_opensource/volcano_1.7/volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/
        ls -la ./ &&  cp -rf ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/
        cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
        ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ascend-for-volcano v1.7.0
        mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0/
        rm -rf ${ATOMGIT_WORKSPACE}/output/Dockerfile*
        cp -rf ${ATOMGIT_WORKSPACE}/output ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/
        cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output && chmod 550 volcano-v1.7.0 && chmod 550 volcano-v1.9.0
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-for-volcano ${TARGET_BRANCH}
    fi
    if grep -q "ascend-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ascend-operator
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-operator ${TARGET_BRANCH}
    fi
    if grep -q "noded" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh noded
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh noded ${TARGET_BRANCH}
    fi
    if grep -q "npu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh npu-exporter
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh npu-exporter ${TARGET_BRANCH}
    fi
    if grep -q "taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh taskd
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh taskd ${TARGET_BRANCH}
    fi
else
    mkdir -p ${ATOMGIT_WORKSPACE}/opensource
    cd ${ATOMGIT_WORKSPACE}/opensource
    echo "$(uname -m)"
    if [ "$(uname -m)" = "aarch64" ]; then
        wget https://get.helm.sh/helm-v3.17.3-linux-arm64.tar.gz --no-check-certificate
        tar -zxvf helm-v3.17.3-linux-arm64.tar.gz
        mv linux-arm64/helm /usr/local/bin/helm
    else
        wget -q https://get.helm.sh/helm-v3.17.3-linux-amd64.tar.gz --no-check-certificate
        tar -zxf helm-v3.17.3-linux-amd64.tar.gz
        mv linux-amd64/helm /usr/local/bin/helm
    fi
    helm version

    if [ "$(uname -m)" = "aarch64" ]; then
        wget -q https://mindcluster.obs.cn-north-4.myhuaweicloud.com/blueImageDependency/yq_linux_arm64 -O /usr/bin/yq
    else
        wget -q https://mindcluster.obs.cn-north-4.myhuaweicloud.com/blueImageDependency/yq_linux_amd64 -O /usr/bin/yq
    fi
    chmod +x /usr/bin/yq

    cd ..
    if grep -q "build/build_all.sh" "${ATOMGIT_WORKSPACE}/change.txt"; then
        echo 'build all'
        mkdir -p ${ATOMGIT_WORKSPACE}/artifacts
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.run
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.whl
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.zip

        export GOROOT=/opt/buildtools/go_1.26.1
        export GO111MODULE='auto'
        export GOSUMDB=sum.golang.google.cn
        export GOPROXY="https://mirrors.huaweicloud.com/repository/goproxy/"
        export PYTHON_HOME=/usr/local/python3.7.5/bin
        export PATH=${PYTHON_HOME}:$GOROOT/bin:$PATH
        export GOPATH=${ATOMGIT_WORKSPACE}
        pip3 show setuptools
        pip3 install wheel
        pip3 install --upgrade setuptools
        pip3 install joblib
        pip3 install numpy==1.26.4
        pip3 install pandas
        pip3 install scikit-learn
        pip3 install ply

        cd ${ATOMGIT_WORKSPACE}/${servicename}/build && dos2unix *.sh && chmod +x * && bash build_all.sh $GOPATH
        cd ${ATOMGIT_WORKSPACE}/artifacts/ && zip -r artifacts_$(uname -m).zip .
        ls -al ${ATOMGIT_WORKSPACE}/artifacts/
    else
        pip3 show setuptools
        pip3 install wheel
        pip3 install --upgrade setuptools
        pip3 install joblib

        ostype=`arch`
        GOPATH=${ATOMGIT_WORKSPACE}

        file_path='change.txt'
        cd ${ATOMGIT_WORKSPACE}
        mkdir -p ${ATOMGIT_WORKSPACE}/artifacts
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.run
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.whl
        touch ${ATOMGIT_WORKSPACE}/artifacts/empty.zip

        echo "env.ATOMGIT_BASE_REF: ${TARGET_BRANCH}"

        if grep -q "component/container-manager" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} container-manager
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh container-manager ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/container-manager/output/Ascend-mindxdl-container-manager_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/clusterd" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} clusterd
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh clusterd ${TARGET_BRANCH}
            ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/clusterd/output
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/output/Ascend-mindxdl-clusterd_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-device-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-device-plugin
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh ascend-device-plugin ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-device-plugin/output/Ascend-mindxdl-device-plugin_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && tar -zxvf makeself/makeself-2.4.2.tar.gz
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && cp -rf makeself-header/makeself-header.sh makeself-release-2.4.2/
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-docker-runtime
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh ascend-docker-runtime ${TARGET_BRANCH}
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/output
            zip Ascend-docker-runtime_7.1.RC1_linux-${ostype}.zip Ascend-docker-runtime_*_linux-${ostype}.run
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/output/Ascend-docker-runtime_*.run ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-for-volcano" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}
            mkdir -p src/volcano.sh
            cp -rf /opt/buildtools/volcano_opensource/volcano_1.9/volcano src/volcano.sh/
            ls -la ./ &&  cp -rf mind-cluster/component/ascend-for-volcano src/volcano.sh/volcano/pkg/scheduler/plugins/
            cd src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
            ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-for-volcano v1.9.0
            mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
            ls -la ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
            rm -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano

            echo "***************start complie volcano 1.7***********************"
            cp -rf /opt/buildtools/volcano_opensource/volcano_1.7/volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/
            ls -la ./ &&  cp -rf ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/
            cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
            ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-for-volcano v1.7.0
            mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0/

            echo "***************start complie volcano 1.12***********************"
            rm -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/*
            cp -rf /opt/buildtools/volcano_opensource/volcano_1.12/volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/
            ls -la ./ &&  cp -rf ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/
            cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
            ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-for-volcano v1.12.0
            mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.12.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.12.0/
            rm -rf ${ATOMGIT_WORKSPACE}/output/Dockerfile*
            cp -rf ${ATOMGIT_WORKSPACE}/output ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/
            cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output && chmod 550 volcano-v1.7.0 && chmod 550 volcano-v1.9.0 && chmod 550 volcano-v1.12.0
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh ascend-for-volcano ${TARGET_BRANCH}
            cp -p -f -r  ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output/Ascend-mindxdl-volcano_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-operator
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh ascend-operator ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-operator/output/Ascend-mindxdl-ascend-operator_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/noded" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} noded
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh noded ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/noded/output/Ascend-mindxdl-noded_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/k8s-rdma-shared-dev-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} k8s-rdma-shared-dev-plugin
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh k8s-rdma-shared-dev-plugin ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/k8s-rdma-shared-dev-plugin/output/Ascend-mindxdl-k8s-rdma-shared-dev-plugin_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/npu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} npu-exporter
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh npu-exporter ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/npu-exporter/output/Ascend-mindxdl-npu-exporter_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q -E "(component/.*\.yaml|helm-deploy-tool/)" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} helm-deploy-tool
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh helm-deploy-tool ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/helm-deploy-tool/output/Ascend-helm-deploy-tool_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} taskd
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh taskd ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/output/Ascend-mindxdl-taskd_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-faultdiag-online" "${ATOMGIT_WORKSPACE}/$file_path"; then
            echo "******************************no build******************************"
        fi
        if grep -q "component/infer-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} infer-operator
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package.sh infer-operator ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/infer-operator/output/Ascend-mindxdl-infer-operator_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/faultdiag" "${ATOMGIT_WORKSPACE}/$file_path"; then
            pip3 install  numpy==1.26.4
            pip3 install pandas
            pip3 install scikit-learn
            pip3 install ply

            export PATH=/opt/buildtools/python-3.11.4/bin:$PATH
            export CC=/opt/rh/devtoolset-7/root/usr/bin/gcc
            export CXX=/opt/rh/devtoolset-7/root/usr/bin/g++
            export GCC_HOME=/opt/rh/devtoolset-7/root/usr/bin

            cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/build/ && dos2unix *.sh && chmod +x * && bash build.sh
            ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output/alan_faultdiag-*.whl ${ATOMGIT_WORKSPACE}/artifacts/
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output/ascend_faultdiag_toolkit*.whl ${ATOMGIT_WORKSPACE}/artifacts/
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/output/ascend_faultdiag-*.whl ${ATOMGIT_WORKSPACE}/artifacts/

        fi
        if grep -q "component/mindio" "${ATOMGIT_WORKSPACE}/$file_path"; then
            SPDLOG_OBS_URL="https://mindcluster.obs.cn-north-4.myhuaweicloud.com/blueImageDependency/spdlog-1.12.0.zip"

            git clone -b v1.1.16 https://atomgit.com/openeuler/libboundscheck.git ${servicename}/component/mindio/acp/3rdparty/libboundscheck/libboundscheck
            wget -c --no-check-certificate -O ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/3rdparty/spdlog/spdlog-1.12.0.zip ${SPDLOG_OBS_URL}
            unzip -d ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/3rdparty/spdlog/ ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/3rdparty/spdlog/spdlog-1.12.0.zip && mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/3rdparty/spdlog/spdlog-1.12.0 ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/3rdparty/spdlog/spdlog
            git clone -b master https://atomgit.com/openeuler/ubs-comm.git ${servicename}/component/mindio/acp/3rdparty/ubs-comm/ubs-comm

            git clone -b v1.1.16 https://atomgit.com/openeuler/libboundscheck.git ${servicename}/component/mindio/tft/3rdparty/libboundscheck/libboundscheck
            wget -c --no-check-certificate -O ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/3rdparty/spdlog/spdlog-1.12.0.zip ${SPDLOG_OBS_URL}
            unzip -d ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/3rdparty/spdlog/ ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/3rdparty/spdlog/spdlog-1.12.0.zip && mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/3rdparty/spdlog/spdlog-1.12.0 ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/3rdparty/spdlog/spdlog

            pip3 show setuptools
            pip3 install wheel
            pip3 install --upgrade setuptools

            unset CC
            unset CXX

            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} mindio acp
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash +x ci_build.sh ${servicename} ${TARGET_BRANCH} mindio tft
            ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/acp/output
            chmod 755 ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/acp/output/*.whl
            ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/tft/output
            chmod 755 ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/tft/output/*.whl
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/output/mindio_acp-*.whl ${ATOMGIT_WORKSPACE}/artifacts/
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/output/mindio_ttp-*.whl ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        cd ${ATOMGIT_WORKSPACE}/artifacts/ && zip -r artifacts_$(uname -m).zip .
        ls -al ${ATOMGIT_WORKSPACE}/artifacts/
    fi
fi
