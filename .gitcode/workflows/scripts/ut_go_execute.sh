set -ex

export TZ=Asia/Shanghai
go env

export GOROOT=/opt/buildtools/go_1.26.1
export GO111MODULE='auto'
export GONOSUMDB='*'
export GOPROXY="https://goproxy.cn,direct"
export PYTHON_HOME=/usr/local/python3.9.11/bin
export PATH=${PYTHON_HOME}:$GOROOT/bin:$PATH
export GOPATH=/opt/buildtools/
file_path="change.txt"
cd ${ATOMGIT_WORKSPACE}
if grep -q "component/mindio" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/test/clusterd.html
fi
if grep -q "component/container-manager" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/container-manager
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/container-manager/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/container-manager/test/container-manager.html
fi
if grep -q "component/clusterd" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/test/clusterd.html
fi
if grep -q "component/ascend-device-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-device-plugin
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-device-plugin/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-device-plugin/test/ascend-device-plugin.html
fi
if grep -q "component/ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/cli/test/dt_go
    dos2unix *.sh && chmod +x *
    bash -x build.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/cli/ascend-docker-runtime.html
    chmod 755 ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/cli/ascend-docker-runtime.html
fi
if grep -q "component/ascend-for-volcano" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}
    mkdir -p src/volcano.sh && cp -rf /opt/buildtools/volcano_opensource/volcano_1.7/volcano src/volcano.sh/
    ls -la ./ &&  cp -rf mind-cluster/component/ascend-for-volcano src/volcano.sh/volcano/pkg/scheduler/plugins/
    cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
    export GOPATH=${ATOMGIT_WORKSPACE}
    cp -rf /opt/buildtools/bin ${ATOMGIT_WORKSPACE}
    cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build/
    dos2unix *.sh && chmod +x *
    bash -x testBuild.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/_output/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/test/ascend-for-volcano.html
fi
if grep -q "component/ascend-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-operator
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-operator/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-operator/test/ascend-operator.html
fi
if grep -q "component/noded" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/noded
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/noded/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/noded/test/noded.html
fi
if grep -q "component/k8s-rdma-shared-dev-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/k8s-rdma-shared-dev-plugin
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/k8s-rdma-shared-dev-plugin/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/k8s-rdma-shared-dev-plugin/test/k8s-rdma-shared-dev-plugin.html
fi
if grep -q "component/npu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/npu-exporter
    sed -i "s#import#//import#g" "./platforms/inputs/all/npu.go"
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/npu-exporter/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/npu-exporter/test/npu-exporter.html
fi
if grep -q "ascend-hccl-controller" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-hccl-controller
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-hccl-controller/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-hccl-controller/test/ascend-hccl-controller.html
fi
if grep -q "component/taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/taskd/go
    go mod tidy
    cd ../../tests/ut/go
    dos2unix *.sh && chmod +x *
    bash -x run_test.sh
    pip3 install lxml
    mkdir test
    cd test
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/test/ut/go/* ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/tests/ut/go/
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/tests/ut/go/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/tests/ut/go/taskd.html
fi
if grep -q "component/infer-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/infer-operator
    go mod tidy
    cd build
    dos2unix *.sh && chmod +x *
    bash -x test.sh
    pip3 install lxml
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/infer-operator/test/api.html ${ATOMGIT_WORKSPACE}/mind-cluster/component/infer-operator/test/infer-operator.html
fi
