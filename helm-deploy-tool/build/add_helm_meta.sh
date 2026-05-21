#!/bin/bash

# Perform  add helm meta to resources
# Copyright @ Huawei Technologies CO., Ltd. 2026. All rights reserved
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================
RELEASE_NAME="mindcluster"
RELEASE_CRD_NAME="mindcluster-crds"
RELEASE_NAMESPACE="default"
LABEL="app.kubernetes.io/managed-by=Helm"
ANNOTATION_1="meta.helm.sh/release-name=${RELEASE_NAME}"
ANNOTATION_2="meta.helm.sh/release-namespace=${RELEASE_NAMESPACE}"
ANNOTATION_CRD="meta.helm.sh/release-name=${RELEASE_CRD_NAME}"

ALL_COMPONENTS="clusterd noded npu-exporter infer-operator ascend-operator ascend-device-plugin ascend-for-volcano"

add_helm_meta() {
    local resource_type=$1
    local resource_name=$2
    local namespace=$3

    if [ -n "$namespace" ]; then
        kubectl label ${resource_type} ${resource_name} -n ${namespace} ${LABEL} --overwrite
        kubectl annotate ${resource_type} ${resource_name} -n ${namespace} ${ANNOTATION_1} ${ANNOTATION_2} --overwrite
    else
        kubectl label ${resource_type} ${resource_name} ${LABEL} --overwrite
        kubectl annotate ${resource_type} ${resource_name} ${ANNOTATION_1} ${ANNOTATION_2} --overwrite
    fi
}

add_helm_meta_crds() {
    local resource_type=$1
    local resource_name=$2
    kubectl label ${resource_type} ${resource_name} ${LABEL} --overwrite
    kubectl annotate ${resource_type} ${resource_name} ${ANNOTATION_CRD} ${ANNOTATION_2} --overwrite
}

component_clusterd() {
    echo "========== clusterd =========="
    add_helm_meta configmap clusterd-config-cm cluster-system
    add_helm_meta sa clusterd mindx-dl
    add_helm_meta clusterrole pods-clusterd-role
    add_helm_meta clusterrolebinding pods-clusterd-rolebinding
    add_helm_meta deployment clusterd mindx-dl
    add_helm_meta svc clusterd-grpc-svc mindx-dl
}

component_noded() {
    echo "========== noded =========="
    add_helm_meta sa noded mindx-dl
    add_helm_meta clusterrole pods-noded-role
    add_helm_meta clusterrolebinding pods-noded-rolebinding
    add_helm_meta daemonset noded mindx-dl
}

component_npu-exporter() {
    echo "========== npu-exporter =========="
    add_helm_meta namespace npu-exporter
    add_helm_meta networkpolicy exporter-network-policy npu-exporter
    add_helm_meta daemonset npu-exporter npu-exporter
    add_helm_meta daemonset npu-exporter-310p-1usoc npu-exporter
}

component_infer-operator() {
    echo "========== infer-operator =========="
    add_helm_meta configmap infer-operator-config mindx-dl
    add_helm_meta deployment infer-operator-manager mindx-dl
    add_helm_meta sa infer-operator-manager mindx-dl
    add_helm_meta clusterrole infer-operator-manager-role
    add_helm_meta clusterrolebinding infer-operator-manager-rolebinding
    add_helm_meta_crds crd inferservices.mindcluster.huawei.com
    add_helm_meta_crds crd inferservicesets.mindcluster.huawei.com
    add_helm_meta_crds crd instancesets.mindcluster.huawei.com
}

component_ascend-operator() {
    echo "========== ascend-operator =========="
    add_helm_meta deployment ascend-operator-manager mindx-dl
    add_helm_meta sa ascend-operator-manager mindx-dl
    add_helm_meta clusterrole ascend-operator-manager-role
    add_helm_meta clusterrolebinding ascend-operator-manager-rolebinding
    add_helm_meta_crds crd ascendjobs.mindxdl.gitee.com
}

component_ascend-device-plugin() {
    echo "========== ascend-device-plugin (910) =========="
    add_helm_meta sa ascend-device-plugin-sa-910 kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-910
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-910
    add_helm_meta daemonset ascend-device-plugin-daemonset kube-system

    echo "========== ascend-device-plugin (npu) =========="
    add_helm_meta sa ascend-device-plugin-sa-npu kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-npu
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-npu
    add_helm_meta daemonset ascend-device-plugin-daemonset-npu kube-system

    echo "========== ascend-device-plugin (310P) =========="
    add_helm_meta sa ascend-device-plugin-sa-310p kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-310p
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-310p
    add_helm_meta daemonset ascend-device-plugin310p-daemonset kube-system

    echo "========== ascend-device-plugin (310) =========="
    add_helm_meta sa ascend-device-plugin-sa-310 kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-310
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-310
    add_helm_meta daemonset ascend-device-plugin2-daemonset kube-system

    echo "========== ascend-device-plugin (310P-1usoc) =========="
    add_helm_meta daemonset ascend-device-plugin3-daemonset-310p-1usoc kube-system

    echo "========== ascend-device-plugin (volcano-910) =========="
    add_helm_meta daemonset ascend-device-plugin-daemonset-910 kube-system

    echo "========== ascend-device-plugin (volcano-npu) =========="
    add_helm_meta daemonset ascend-device-plugin-daemonset-npu kube-system

    echo "========== ascend-device-plugin (volcano-310P) =========="
    add_helm_meta daemonset ascend-device-plugin3-daemonset-310p kube-system

    echo "========== ascend-device-plugin (volcano-310) =========="
    add_helm_meta daemonset ascend-device-plugin2-daemonset-310 kube-system

    echo "========== ascend-device-plugin (volcano-310P-1usoc) =========="
    add_helm_meta daemonset ascend-device-plugin3-daemonset-310p-1usoc kube-system
}

component_ascend-for-volcano() {
    echo "========== ascend-for-volcano =========="
    add_helm_meta namespace volcano-system
    add_helm_meta namespace volcano-monitoring
    add_helm_meta sa volcano-controllers volcano-system
    add_helm_meta clusterrole volcano-controllers
    add_helm_meta clusterrolebinding volcano-controllers-role
    add_helm_meta deployment volcano-controllers volcano-system
    add_helm_meta sa volcano-scheduler volcano-system
    add_helm_meta configmap volcano-scheduler-configmap volcano-system
    add_helm_meta clusterrole volcano-scheduler
    add_helm_meta clusterrolebinding volcano-scheduler-role
    add_helm_meta svc volcano-scheduler-service volcano-system
    add_helm_meta deployment volcano-scheduler volcano-system
    add_helm_meta_crds crd commands.bus.volcano.sh
    add_helm_meta_crds crd jobs.batch.volcano.sh
    add_helm_meta_crds crd numatopologies.nodeinfo.volcano.sh
    add_helm_meta_crds crd podgroups.scheduling.volcano.sh
    add_helm_meta_crds crd queues.scheduling.volcano.sh
    add_helm_meta_crds crd jobflows.flow.volcano.sh
    add_helm_meta_crds crd jobtemplates.flow.volcano.sh

}

all_namespace() {
    echo "========== namespace mindx-dl and cluster-system =========="
    add_helm_meta namespace mindx-dl
    add_helm_meta namespace cluster-system
    kubectl annotate namespace mindx-dl helm.sh/resource-policy=keep --overwrite
    kubectl annotate namespace cluster-system helm.sh/resource-policy=keep --overwrite
}

usage() {
    echo "Usage: $0 [component1] [component2] ..."
    echo ""
    echo "Available components:"
    echo "  clusterd"
    echo "  noded"
    echo "  npu-exporter"
    echo "  infer-operator"
    echo "  ascend-operator"
    echo "  ascend-device-plugin"
    echo "  ascend-for-volcano"
    echo ""
    echo "  all    - operate on all components"
    echo "  -h     - show this help"
    echo ""
    echo "Examples:"
    echo "  $0 all                              # operate on all components"
    echo "  $0 ns                              # operate on mindx-dl and cluster-system namespace"
    echo "  $0 clusterd noded                   # operate on clusterd and noded only"
    echo "  $0 ascend-device-plugin ascend-for-volcano  # operate on selected components"
}

run_component() {
    local comp=$1
    case ${comp} in
        ns)                     all_namespace ;;
        clusterd)              component_clusterd ;;
        noded)                 component_noded ;;
        npu-exporter)          component_npu-exporter ;;
        infer-operator)        component_infer-operator ;;
        ascend-operator)       component_ascend-operator ;;
        ascend-device-plugin)  component_ascend-device-plugin ;;
        ascend-for-volcano)    component_ascend-for-volcano ;;
        *)
            echo "Error: unknown component '${comp}'"
            echo "Available: ${ALL_COMPONENTS}"
            return 1
            ;;
    esac
}

if [ $# -eq 0 ]; then
    usage
    exit 0
fi

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    exit 0
fi

if [ "$1" = "all" ]; then
    for comp in ${ALL_COMPONENTS}; do
        run_component ${comp}
    done
else
    for comp in "$@"; do
        run_component ${comp}
    done
fi

echo "========== Done =========="
