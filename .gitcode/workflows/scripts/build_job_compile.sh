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
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh clusterd
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh clusterd ${TARGET_BRANCH}
        ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/clusterd/output
    fi
    if grep -q "ascend-device-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ascend-device-plugin
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-device-plugin ${TARGET_BRANCH}
    fi
    if grep -q "ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && tar -zxvf makeself/makeself-2.4.2.tar.gz
        cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && cp -rf makeself-header/makeself-header.sh makeself-release-2.4.2/
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ascend-docker-runtime
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
        ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ascend-for-volcano v1.9.0
        mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
        ls -la ${ATOMGIT_WORKSPACE}/output/volcano-v1.9.0/
        rm -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano
        echo "***************start complie volcano 1.7***********************"
        cp -rf /opt/buildtools/volcano_opensource/volcano_1.7/volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/
        ls -la ./ &&  cp -rf ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/
        cd ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ascend-for-volcano ascend-volcano-plugin
        ls -la ${ATOMGIT_WORKSPACE} && cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ascend-for-volcano v1.7.0
        mkdir -p ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0 && cp -rf ${ATOMGIT_WORKSPACE}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${ATOMGIT_WORKSPACE}/output/volcano-v1.7.0/
        rm -rf ${ATOMGIT_WORKSPACE}/output/Dockerfile*
        cp -rf ${ATOMGIT_WORKSPACE}/output ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/
        cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output && chmod 550 volcano-v1.7.0 && chmod 550 volcano-v1.9.0
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-for-volcano ${TARGET_BRANCH}
    fi
    if grep -q "ascend-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ascend-operator
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh ascend-operator ${TARGET_BRANCH}
    fi
    if grep -q "noded" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh noded
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh noded ${TARGET_BRANCH}
    fi
    if grep -q "npu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh npu-exporter
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && bash -x build_package_CH.sh npu-exporter ${TARGET_BRANCH}
    fi
    if grep -q "taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh taskd
        cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh taskd ${TARGET_BRANCH}
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

        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-device-plugin/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-dynamic-resource-allocation/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-clusterops-agent/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/output/*.run ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-for-volcano/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-operator/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-faultdiag/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-faultdiag/output/*.whl ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/clusterd/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/container-manager/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/infer-operator/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/mindio/output/*.whl ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/noded/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/npu-exporter/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/dpu-exporter/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/k8s-rdma-shared-dev-plugin/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/ub-host-device-cni/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/taskd/output/*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        cp ${ATOMGIT_WORKSPACE}/${servicename}/component/taskd/output/*.whl ${ATOMGIT_WORKSPACE}/artifacts/
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
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} container-manager
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh container-manager ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/container-manager/output/Ascend-mindxdl-container-manager_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/clusterd" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} clusterd
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh clusterd ${TARGET_BRANCH}
            ls -al ${ATOMGIT_WORKSPACE}/${servicename}/component/clusterd/output
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/clusterd/output/Ascend-mindxdl-clusterd_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-device-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-device-plugin
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ascend-device-plugin ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-device-plugin/output/Ascend-mindxdl-device-plugin_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && tar -zxvf makeself/makeself-2.4.2.tar.gz
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/opensource && cp -rf makeself-header/makeself-header.sh makeself-release-2.4.2/
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-docker-runtime
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ascend-docker-runtime ${TARGET_BRANCH}
            cd ${ATOMGIT_WORKSPACE}/${servicename}/component/ascend-docker-runtime/output
            zip Ascend-docker-runtime_7.1.RC1_linux-${ostype}.zip Ascend-docker-runtime_*_linux-${ostype}.run
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/output/Ascend-docker-runtime_*.run ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-for-volcano" "${ATOMGIT_WORKSPACE}/$file_path"; then
            if ! command -v timeout >/dev/null 2>&1; then
                echo "timeout is required to build volcano" >&2
                exit 1
            fi
            if ! command -v setsid >/dev/null 2>&1; then
                echo "setsid is required to build volcano" >&2
                exit 1
            fi
            # Prefer a usable GOROOT from the job. Otherwise ask the go binary on PATH.
            # Clear GOROOT for that query so a bad value is not echoed back.
            if [ -z "${GOROOT}" ] || [ ! -x "${GOROOT}/bin/go" ]; then
                if command -v go >/dev/null 2>&1; then
                    GOROOT=$(GOROOT= go env GOROOT)
                fi
            fi
            if [ -z "${PYTHON_HOME}" ]; then
                py=$(command -v python3 2>/dev/null || true)
                if [ -n "${py}" ]; then
                    PYTHON_HOME=$(dirname "${py}")
                fi
            fi
            if [ -z "${GOROOT}" ] || [ ! -x "${GOROOT}/bin/go" ] || [ -z "${PYTHON_HOME}" ]; then
                echo "GOROOT and PYTHON_HOME are required in the environment" >&2
                exit 1
            fi
            export GOROOT PYTHON_HOME

            # Index i is one build. The image dir has no "v" and no ".0".
            # build.sh uses the version for patches and volcano-<version>.yaml, not the git branch.
            volcano_vers=(v1.9.0 v1.12.0 v1.15.0)
            volcano_srcs=(volcano_1.9 volcano_1.12 volcano_1.15)
            volcano_timeout=20m
            volcano_kill_after=30s
            volcano_poll_secs=5
            volcano_heartbeat_secs=60

            # Replace the version number only once. The three builds no longer call ci_build.sh,
            # so they do not edit the same ini at the same time and GOPATH is not pulled back
            # to one directory.
            # utils.sh is stored with CR. Strip it before source, or a blank line runs as $'\r'.
            utils_sh="${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build/utils.sh"
            if command -v dos2unix >/dev/null 2>&1; then
                dos2unix "${utils_sh}"
            else
                sed -i 's/\r$//' "${utils_sh}"
            fi
            source "${utils_sh}"
            VERSION_NUMBER=$(get_version_number "${TARGET_BRANCH}")
            if [ $? -ne 0 ]; then
                echo "get Version_Number failed, please check."
                exit 1
            fi
            sed -i "s/VERSION_NUMBER/${VERSION_NUMBER}/g" \
                "${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/config/service_config.ini"

            # Compile one Volcano version.
            # Arguments: version, source directory name, and the CPU count for this version.
            # After a successful build, copy the plugin output/ to ${ATOMGIT_WORKSPACE}/output/volcano-<version>/.
            compile_volcano() {
                local ver="$1"
                local src_name="$2"
                local slots="$3"
                # Own tree per version. build.sh writes under ${GOPATH}/src/volcano.sh/volcano.
                local root="${ATOMGIT_WORKSPACE}/volcano-build/${ver}"
                local plugin="${root}/src/volcano.sh/volcano/pkg/scheduler/plugins"
                set -e

                mkdir -p "${root}/src/volcano.sh"
                cp -rf "/opt/buildtools/volcano_opensource/${src_name}/volcano" "${root}/src/volcano.sh/"
                cp -rf "${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano" "${plugin}/"
                mv "${plugin}/ascend-for-volcano" "${plugin}/ascend-volcano-plugin"
                cp -f "${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/config/service_config.ini" \
                    "${root}/src/volcano.sh/volcano/"

                (
                    export GOPATH="${root}"
                    export GO111MODULE='auto'
                    export GOSUMDB=sum.golang.google.cn
                    export GOPROXY="https://repo.huaweicloud.com/repository/goproxy/"
                    export PATH="${PYTHON_HOME}:${GOROOT}/bin:${PATH}"
                    # go build -p defaults to GOMAXPROCS. Split the machine across the three versions.
                    export GOMAXPROCS="${slots}"
                    cd "${plugin}/ascend-volcano-plugin/build"
                    dos2unix *.sh && chmod +x *
                    # 124: timeout sent TERM. 137: KILL after --kill-after. A normal build is about 3 minutes.
                    status=0
                    timeout --kill-after="${volcano_kill_after}" "${volcano_timeout}" ./build.sh "${ver}" || status=$?
                    if [ "${status}" -ne 0 ]; then
                        exit "${status}"
                    fi
                )

                mkdir -p "${ATOMGIT_WORKSPACE}/output/volcano-${ver}"
                cp -rf "${plugin}/ascend-volcano-plugin/output/"* \
                    "${ATOMGIT_WORKSPACE}/output/volcano-${ver}/"
            }

            # Runs in its own session. $$ is the process group id used to stop the build.
            # The status path is baked into the trap: after a normal return the function's locals are gone.
            run_one_volcano() {
                local ver="$1"
                local src_name="$2"
                local slots="$3"
                local status_file="${ATOMGIT_WORKSPACE}/volcano-build/${ver}.status"
                local status_q
                status_q=$(printf '%q' "${status_file}")
                set -e
                # $$ is this session's process group id. The parent stops the build with it.
                printf '%s\n' "$$" > "${ATOMGIT_WORKSPACE}/volcano-build/${ver}.pgid"
                # EXIT runs after locals are gone, so the status path is already expanded in the trap.
                trap "code=\$?; printf '%s\n' \"\$code\" > ${status_q}" EXIT
                # 143: TERM. 130: INT. EXIT then writes that code into the status file.
                trap 'exit 143' TERM
                trap 'exit 130' INT
                set -x
                compile_volcano "${ver}" "${src_name}" "${slots}"
            }

            # Invoked as "bash build_job_compile.sh", then this script cds to the workspace.
            # dirname of that name is ".", so BASH_SOURCE cannot find the helper.
            # shellcheck source=volcano_job_control.sh
            source "${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/scripts/volcano_job_control.sh"

            # go build -p defaults to GOMAXPROCS, so each version gets about one third of the cores.
            volcano_cores=$(nproc 2>/dev/null || echo 1)
            volcano_slots=$(( volcano_cores / ${#volcano_vers[@]} ))
            if [ "${volcano_slots}" -lt 1 ]; then
                volcano_slots=1
            fi
            echo "volcano GOMAXPROCS=${volcano_slots} (${volcano_cores} cores, ${#volcano_vers[@]} versions)"

            # Drop a previous run's status files so a stale exit code cannot be collected.
            rm -rf "${ATOMGIT_WORKSPACE}/volcano-build"
            mkdir -p "${ATOMGIT_WORKSPACE}/volcano-build"
            volcano_pids=()
            volcano_started=()
            volcano_collected=()
            volcano_status=()
            volcano_elapsed=()

            # Monitor mode would make setsid fork, and $! would not be the build session.
            set +m
            export -f compile_volcano run_one_volcano
            trap 'on_volcano_interrupt 130' INT
            trap 'on_volcano_interrupt 143' TERM
            # set -e exits are not INT/TERM. The builds are in another session, so clean them up here.
            trap 'stop_volcano_sessions' EXIT

            for volcano_i in "${!volcano_vers[@]}"; do
                ver="${volcano_vers[$volcano_i]}"
                volcano_started[$volcano_i]=$(date +%s)
                volcano_collected[$volcano_i]=0
                # Per-version logs. The child records its own exit code; the parent reaps it below.
                ATOMGIT_WORKSPACE="${ATOMGIT_WORKSPACE}" servicename="${servicename}" \
                    volcano_timeout="${volcano_timeout}" volcano_kill_after="${volcano_kill_after}" \
                    GOROOT="${GOROOT}" PYTHON_HOME="${PYTHON_HOME}" \
                    setsid bash -c 'run_one_volcano "$@"' _ \
                    "${ver}" "${volcano_srcs[$volcano_i]}" "${volcano_slots}" \
                    > "${ATOMGIT_WORKSPACE}/volcano-build/${ver}.log" 2>&1 </dev/null &
                volcano_pids[$volcano_i]=$!
            done
            echo "volcano builds started: ${volcano_vers[*]}"

            # set +x turns off command tracing. Only the heartbeat, once every 60s, reports volcano build progress.
            set +x
            next_beat=$(( $(date +%s) + volcano_heartbeat_secs ))
            while :; do
                collect_finished_volcano_jobs
                running=$(volcano_still_running)
                if [ -z "${running}" ]; then
                    break
                fi
                now=$(date +%s)
                if [ "${now}" -ge "${next_beat}" ]; then
                    echo "volcano still running: ${running}"
                    next_beat=$(( now + volcano_heartbeat_secs ))
                fi
                sleep "${volcano_poll_secs}"
            done
            trap - EXIT INT TERM
            report_volcano_jobs 0
            set -x
            export -nf compile_volcano run_one_volcano 2>/dev/null || true
            # Skip chmod and packaging when any version failed.
            [ "${fail}" -eq 0 ]

            rm -rf ${ATOMGIT_WORKSPACE}/output/Dockerfile*
            cp -rf ${ATOMGIT_WORKSPACE}/output ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/
            cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output
            for ver in "${volcano_vers[@]}"; do
                chmod 550 "volcano-${ver}"
            done
            # TARGET_BRANCH is the git branch used for the package version, not the Volcano version.
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build \
                && bash -x build_package.sh ascend-for-volcano ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-for-volcano/output/Ascend-mindxdl-volcano_*.zip \
                ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-operator
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ascend-operator ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-operator/output/Ascend-mindxdl-ascend-operator_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/noded" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} noded
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh noded ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/noded/output/Ascend-mindxdl-noded_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ub-host-device-cni" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ub-host-device-cni
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ub-host-device-cni ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ub-host-device-cni/output/Ascend-mindxdl-ub-host-device-cni_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-dynamic-resource-allocation" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-dynamic-resource-allocation
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ascend-dynamic-resource-allocation ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-dynamic-resource-allocation/output/Ascend-mindxdl-ascend-dynamic-resource-allocation_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/dpu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} dpu-exporter
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh dpu-exporter ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/dpu-exporter/output/Ascend-mindxdl-dpu-exporter_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-clusterops-agent" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} ascend-clusterops-agent
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh ascend-clusterops-agent ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-clusterops-agent/output/Ascend-mindxdl-ascend-clusterops-agent_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/k8s-rdma-shared-dev-plugin" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} k8s-rdma-shared-dev-plugin
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh k8s-rdma-shared-dev-plugin ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/k8s-rdma-shared-dev-plugin/output/Ascend-mindxdl-k8s-rdma-shared-dev-plugin_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/npu-exporter" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} npu-exporter
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh npu-exporter ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/npu-exporter/output/Ascend-mindxdl-npu-exporter_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q -E "(component/.*\.yaml|helm-deploy-tool/)" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} helm-deploy-tool
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh helm-deploy-tool ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/helm-deploy-tool/output/Ascend-helm-deploy-tool_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} taskd
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh taskd ${TARGET_BRANCH}
            cp -p -f -r ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/output/Ascend-mindxdl-taskd_*.zip ${ATOMGIT_WORKSPACE}/artifacts/
        fi
        if grep -q "component/ascend-faultdiag-online" "${ATOMGIT_WORKSPACE}/$file_path"; then
            echo "******************************no build******************************"
        fi
        if grep -q "component/infer-operator" "${ATOMGIT_WORKSPACE}/$file_path"; then
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} infer-operator
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x build_package.sh infer-operator ${TARGET_BRANCH}
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

            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} mindio acp
            cd ${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/build && dos2unix *.sh 2>/dev/null || true && bash -x ci_build.sh ${servicename} ${TARGET_BRANCH} mindio tft
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
