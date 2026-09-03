#!/bin/bash
# Perform build volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2026. All rights reserved
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e
set -x

# BASE_VER supports v1.9.0, and 1.10+
if [ ! -n "$1" ]; then
    BASE_VER=v1.12.0
else
    BASE_VER=$1
fi

function ver_at_least() {
    local ver="${1#v}"
    local major minor floor_major=$2 floor_minor=$3
    IFS='.' read -r major minor _ <<< "${ver}"
    if [ "${major:-0}" -gt ${floor_major} ] || { [ "${major:-0}" -eq ${floor_major} ] && [ "${minor:-0}" -ge ${floor_minor} ]; }; then
        return 0 # ver >= floor
    fi
    return 1 # ver < floor
}

if ver_at_least "${BASE_VER}" 1 15; then
    GO_BUILD_TAGS="-tags volcano_v115"
    echo ">> ${BASE_VER} >= v1.15.0: build with ${GO_BUILD_TAGS}"
else
    GO_BUILD_TAGS=""
    echo ">> ${BASE_VER} < v1.15.0: pre-v1.15 build (no build tags)"
fi

echo "===== Start Dual Build Mode ====="
echo "Build Version is ${BASE_VER}"
echo "Will generate both output(Alpine musl) and output-oe(openEuler glibc)"
echo ""

DEFAULT_VER='v26.1.0'
TOP_DIR=${GOPATH}/src/volcano.sh/volcano/
BASE_PATH=${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/
CMD_PATH=${GOPATH}/src/volcano.sh/volcano/cmd/
PKG_PATH=volcano.sh/volcano/pkg
DATE=$(date "+%Y-%m-%d %H:%M:%S")

function parse_version() {
    version_file="${TOP_DIR}"/service_config.ini
    if  [ -f "$version_file" ]; then
      line=$(sed -n '1p' "$version_file" 2>&1)
      version="v"${line#*=}
      echo "${version}"
      return
    fi
    echo ${DEFAULT_VER}
}

function parse_arch() {
   arch=$(arch 2>&1)
   echo "${arch}"
}

REL_VERSION=$(parse_version)
REL_ARCH=$(parse_arch)
REL_NPU_PLUGIN=volcano-npu_${REL_VERSION}

CLUSTER_PATH=/workspace/mind-cluster/component/ascend-for-volcano
GIT_COMMIT=$(git -C "$CLUSTER_PATH" rev-parse --verify HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git -C "$CLUSTER_PATH" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | awk '{print $3}')

function clean_dir() {
    local OUTPUT_DIR=$1
    rm -f "${OUTPUT_DIR}"/vc-controller-manager
    rm -f "${OUTPUT_DIR}"/vc-scheduler
    rm -f "${OUTPUT_DIR}"/*.so
}

function copy_resources() {
    local OUTPUT_DIR=$1
    cp "${BASE_PATH}"/build/volcano-"${BASE_VER}".yaml "${OUTPUT_DIR}/"
    # Copy the license agreement and replace version information
    cp "${BASE_PATH}"/build/agreement.txt "${OUTPUT_DIR}/agreement.txt"
    sed -i "s/Volcano Version .*/Volcano Version ${BASE_VER}/" "${OUTPUT_DIR}/agreement.txt"
}

# Automatically inject agreement file and ENTRYPOINT into Dockerfile
function df_print_agreement() {
    local OUTPUT_DIR=$1
    DOCKERFILES=("${OUTPUT_DIR}/Dockerfile*")
    for dockerfile in "${DOCKERFILES[@]}"; do
        if [ -f "$dockerfile" ]; then
            if ! grep -q "^COPY agreement.txt /usr/local/" "$dockerfile"; then
                sed -i '/^COPY /a COPY agreement.txt /usr/local/' "$dockerfile"
                echo "Inserted COPY agreement.txt after COPY in ${dockerfile}"
            fi

            if ! grep -q 'ENTRYPOINT \["/bin/sh", "-c", "cat /usr/local/agreement.txt; exec /bin/sh"\]' "$dockerfile"; then
                printf '\n%s\n' 'ENTRYPOINT ["/bin/sh", "-c", "cat /usr/local/agreement.txt; exec /bin/sh"]' >> "$dockerfile"
                echo "Added ENTRYPOINT to ${dockerfile}"
            fi
        fi
    done
}

# Source code patch, executed only once globally to avoid repeated execution
function patch_all_source() {
    echo "===== Apply source patch once ====="
    # Unified patch
    REPLACE_FILE="${GOPATH}/src/volcano.sh/volcano/pkg/controllers/job/state/running.go"
    SEARCH_STRING="Ignore"
    if ! grep -q "$SEARCH_STRING" "$REPLACE_FILE";then
      sed -i "s/switch action {/switch action { case \"Ignore\" : return nil/g" "$REPLACE_FILE"
    fi

    # Version-differentiated predicate patch
    if [[ "$BASE_VER" == "v1.9.0" ]]; then
        REPLACE_FILE="${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/npu.go"
        # 1. Change closure signature: error -> ([]*api.Status, error)
        sed -i "s/api.NodeInfo) error {/api.NodeInfo) (\[\]\*api.Status, error) {/g" "$REPLACE_FILE"
        # 2. Change convertToNPUFitError signature: error -> ([]*api.Status, error)
        sed -i "s/predicateErr error) error {/predicateErr error) (\[\]\*api.Status, error) {/g" "$REPLACE_FILE"
        # 3. Replace return api.NewFitErrWithStatus(...) with return []*api.Status{}, predicateErr
        #    Assumes the return statements are on a single line (as in your example).
        #    Matches up to the semicolon.
        sed -i '/return api\.NewFitErrWithStatus/,/})/c\       return []*api.Status{}, predicateErr' "$REPLACE_FILE"
        # 4. Change return nil to return nil, nil inside the addPredicateFn closure
        #    Uses range from the line containing "passed" to the line containing "return nil"
        sed -i '/predicateFn.*passed/,/return nil/s/return nil/return nil, nil/' "$REPLACE_FILE"
    fi

    # Node scoring logic patch
    REPLACE_FILE="${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/actions/allocate/allocate.go"

    sed -i '
    /case len(nodes) == 1:/ {
        N
        N
        s/case len(nodes) == 1:.*\n.*\n.*/            default:/
    }' "$REPLACE_FILE"

    # Evict data race patch: SchedulerCache.Evict hands the shared pod object
    # to an async evictor goroutine, which races with the scheduler goroutine.
    # Deep copy the pod before handing it to the evictor.
    REPLACE_FILE="${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/cache/cache.go"
    if ! grep -q "p := task.Pod.DeepCopy()" "$REPLACE_FILE"; then
      sed -i '/func (sc \*SchedulerCache) Evict/,/^}/s/^\([[:space:]]*\)p := task.Pod$/\1p := task.Pod.DeepCopy()/' "$REPLACE_FILE"
      if ! grep -q "p := task.Pod.DeepCopy()" "$REPLACE_FILE"; then
        echo "WARNING: Evict deepcopy patch failed on ${REPLACE_FILE}"
      fi
    fi
    echo "===== Source patch finished ====="
    echo ""
}

# Single build workflow, parameters:
# Param 1: Output directory
# Param 2: Whether to enable musl (true = Alpine, false = openEuler)
function build_workflow() {
    local OUTPUT_DIR=$1
    local USE_MUSL=$2

    echo "====================================="
    if [ "${USE_MUSL}" = "true" ]; then
        echo "Start build Alpine(musl) -> ${OUTPUT_DIR}"
    else
        echo "Start build openEuler(glibc) -> ${OUTPUT_DIR}"
    fi
    echo "====================================="

    clean_dir "${OUTPUT_DIR}"

    copy_resources "${OUTPUT_DIR}"

    df_print_agreement "${OUTPUT_DIR}"

    echo "Build Architecture is ${REL_ARCH}"
    export GO111MODULE=on
    export PATH=$GOPATH/bin:$PATH

    cd "${TOP_DIR}"

    cd "${OUTPUT_DIR}"

    export CGO_CFLAGS="-fstack-protector-all -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-all -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"

    # Musl compiler switch
    if [ "${USE_MUSL}" = "true" ]; then
        export CC=/usr/local/musl/bin/musl-gcc
    else
        unset CC
    fi

    export CGO_ENABLED=0
    go build -mod=mod ${GO_BUILD_TAGS} -buildmode=pie -ldflags "-s -bindnow
      -X '${PKG_PATH}/version.Built=${DATE}' -X '${PKG_PATH}/version.Version=${BASE_VER}'" \
      -o vc-controller-manager "${CMD_PATH}"/controller-manager

    export CGO_ENABLED=1
    go build -mod=mod ${GO_BUILD_TAGS} -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now
      -X '${PKG_PATH}/version.Built=${DATE}' -X '${PKG_PATH}/version.Version=${BASE_VER}'" \
      -o vc-scheduler "${CMD_PATH}"/scheduler

    go build -mod=mod ${GO_BUILD_TAGS} -buildmode=plugin -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin.PluginName=${REL_NPU_PLUGIN} \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.Version=${REL_VERSION} \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.GitCommit=${GIT_COMMIT} \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.GitBranch=${GIT_BRANCH} \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.BuildOS=linux \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.BuildArch=${REL_ARCH} \
      -X volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/version.GoVersion=${GO_VERSION}" \
      -o "${REL_NPU_PLUGIN}".so "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/

    if [ ! -f "${OUTPUT_DIR}/${REL_NPU_PLUGIN}.so" ]
    then
      echo "ERROR: Failed to find ${REL_NPU_PLUGIN}.so in ${OUTPUT_DIR}"
      exit 1
    fi

    # Modify the plugin name and image tag in YAML files
    sed -i "s/name: volcano-npu_.*/name: ${REL_NPU_PLUGIN}/" "${OUTPUT_DIR}"/volcano-*.yaml
    sed -i "s/:${BASE_VER}/:${BASE_VER}-${REL_VERSION}/g" "${OUTPUT_DIR}"/volcano-${BASE_VER}.yaml

    # New logic: Replace /bin/ash with /bin/sh for openEuler builds
    if [ "${USE_MUSL}" = "false" ]; then
      echo ">> openEuler yaml replace /bin/ash to /bin/sh"
      sed -i 's#/bin/ash#/bin/sh#g' "${OUTPUT_DIR}"/volcano-*.yaml
    fi

    chmod 400 "${OUTPUT_DIR}"/*.so
    chmod 500 vc-controller-manager vc-scheduler
    chmod 400 "${OUTPUT_DIR}"/Dockerfile*
    chmod 400 "${OUTPUT_DIR}"/volcano-*.yaml

    echo "Finish build -> ${OUTPUT_DIR}"
    echo ""
}

function main() {
  # Step 1: Apply source code patches only once to prevent repeated source code modifications via sed
  patch_all_source

  ALPINE_DIR="${BASE_PATH}/output/alpine"
  OE_DIR="${BASE_PATH}/output/openeuler"

  # Step 2: Build Alpine musl artifacts to output directory
  build_workflow "${ALPINE_DIR}" "true"

  # Step 3: Build openEuler glibc artifacts to output-oe directory
  build_workflow "${OE_DIR}" "false"
}

main "$1"

echo "==================== All Build Completed ===================="
echo "Alpine(musl) output dir: ${BASE_PATH}/output"
echo "openEuler(glibc) output dir: ${BASE_PATH}/output-oe"
echo "============================================================="
echo ""
