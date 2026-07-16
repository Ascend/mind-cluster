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
LABEL_KEY="app.kubernetes.io/managed-by"
LABEL_VALUE="Helm"
ANNOTATION_KEY1="meta.helm.sh/release-name"
ANNOTATION_KEY2="meta.helm.sh/release-namespace"

resource_exists() {
    local resource_type=$1
    local resource_name=$2
    local namespace=$3

    if [ -n "$namespace" ]; then
        kubectl get ${resource_type} ${resource_name} -n ${namespace} &>/dev/null
    else
        kubectl get ${resource_type} ${resource_name} &>/dev/null
    fi
}

add_helm_meta() {
    local resource_type=$1
    local resource_name=$2
    local namespace=$3

    if ! resource_exists ${resource_type} ${resource_name} "${namespace}"; then
        echo "Skipping ${resource_type}/${resource_name}, not found."
        return
    fi

    if [ -n "$namespace" ]; then
        kubectl label ${resource_type} ${resource_name} -n ${namespace} ${LABEL_KEY}=${LABEL_VALUE} --overwrite
        kubectl annotate ${resource_type} ${resource_name} -n ${namespace} ${ANNOTATION_KEY1}=${RELEASE_NAME} ${ANNOTATION_KEY2}=${RELEASE_NAMESPACE} --overwrite
    else
        kubectl label ${resource_type} ${resource_name} ${LABEL_KEY}=${LABEL_VALUE} --overwrite
        kubectl annotate ${resource_type} ${resource_name} ${ANNOTATION_KEY1}=${RELEASE_NAME} ${ANNOTATION_KEY2}=${RELEASE_NAMESPACE} --overwrite
    fi
}

add_helm_meta_crds() {
    local resource_type=$1
    local resource_name=$2

    if ! resource_exists ${resource_type} ${resource_name}; then
        echo "Skipping ${resource_type}/${resource_name}, not found."
        return
    fi

    kubectl label ${resource_type} ${resource_name} ${LABEL_KEY}=${LABEL_VALUE} --overwrite
    kubectl annotate ${resource_type} ${resource_name} ${ANNOTATION_KEY1}=${RELEASE_CRD_NAME} ${ANNOTATION_KEY2}=${RELEASE_NAMESPACE} --overwrite
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
    add_helm_meta daemonset ascend-device-plugin-daemonset kube-system # 910
    add_helm_meta daemonset ascend-device-plugin-daemonset-910 kube-system # volcano-910

    echo "========== ascend-device-plugin (npu) =========="
    add_helm_meta sa ascend-device-plugin-sa-npu kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-npu
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-npu
    add_helm_meta daemonset ascend-device-plugin-daemonset-npu kube-system

    echo "========== ascend-device-plugin (310P) =========="
    add_helm_meta sa ascend-device-plugin-sa-310p kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-310p
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-310p
    add_helm_meta daemonset ascend-device-plugin310p-daemonset kube-system # 310P
    add_helm_meta daemonset ascend-device-plugin3-daemonset-310p kube-system  # volcano-310P
    add_helm_meta daemonset ascend-device-plugin3-daemonset-310p-1usoc kube-system # 310P-1usoc， volcano-310P-1usoc

    echo "========== ascend-device-plugin (310) =========="
    add_helm_meta sa ascend-device-plugin-sa-310 kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role-310
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding-310
    add_helm_meta daemonset ascend-device-plugin2-daemonset kube-system # 310
    add_helm_meta daemonset ascend-device-plugin2-daemonset-310 kube-system # volcano-310

    echo "========== ascend-device-plugin (new) =========="
    add_helm_meta sa ascend-device-plugin-sa kube-system
    add_helm_meta clusterrole pods-node-ascend-device-plugin-role
    add_helm_meta clusterrolebinding pods-node-ascend-device-plugin-rolebinding
    add_helm_meta daemonset ascend-device-plugin-daemonset kube-system
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

    add_helm_meta configmap volcano-controller-configmap volcano-system
    add_helm_meta svc volcano-controllers-service volcano-system
    add_helm_meta_crds crd hypernodes.topology.volcano.sh
}

component_k8s-rdma-shared-dev-plugin() {
    echo "========== k8s-rdma-shared-dev-plugin =========="
    add_helm_meta sa rdma-shared-dp-sa kube-system
    add_helm_meta role rdma-shared-dp-cm-role kube-system
    add_helm_meta rolebinding rdma-shared-dp-cm-rb kube-system
    add_helm_meta configmap rdma-devices kube-system
    add_helm_meta daemonset rdma-shared-dp-ds kube-system
}

all_namespace() {
    echo "========== namespace mindx-dl and cluster-system =========="
    add_helm_meta namespace mindx-dl
    add_helm_meta namespace cluster-system
    kubectl annotate namespace mindx-dl helm.sh/resource-policy=keep --overwrite
    kubectl annotate namespace cluster-system helm.sh/resource-policy=keep --overwrite
}

delete_recource() {
    local resource_type=$1
    local resource_name=$2
    local namespace=$3

    if ! resource_exists ${resource_type} ${resource_name} "${namespace}"; then
        echo "Skipping ${resource_type}/${resource_name}, not found."
        return
    fi
    if [ -n "$namespace" ]; then
        kubectl delete ${resource_type} ${resource_name} -n ${namespace}
    else
        kubectl delete ${resource_type} ${resource_name}
    fi

}

delete_old_device_plugin() {
    echo "========== Delete old ascend-device-plugin daemonset =========="
    delete_recource daemonset ascend-device-plugin-daemonset-910 kube-system
    delete_recource daemonset ascend-device-plugin-daemonset-npu kube-system
    delete_recource daemonset ascend-device-plugin310p-daemonset kube-system
    delete_recource daemonset ascend-device-plugin3-daemonset-310p kube-system
    delete_recource daemonset ascend-device-plugin2-daemonset kube-system
    delete_recource daemonset ascend-device-plugin2-daemonset-310 kube-system
}

add_helm_meta_all() {
    all_namespace
    component_clusterd
    component_noded
    component_npu-exporter
    component_infer-operator
    component_ascend-operator
    component_ascend-device-plugin
    component_ascend-for-volcano
    component_k8s-rdma-shared-dev-plugin
}

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --all                       Execute all operations (delete ascend-device-plugin daemonsets before version 26.1.0 and add helm meta to all)"
    echo "  --delete-old-demonset       Delete ascend-device-plugin daemonsets before version 26.1.0"
    echo "  --add-helm-meta-all         Only add helm meta to all resources"
    echo "  --namespace                 Only add helm meta to namespaces (mindx-dl, cluster-system)"
    echo "  --clusterd                  Only add helm meta for clusterd component"
    echo "  --noded                     Only add helm meta for noded component"
    echo "  --npu-exporter              Only add helm meta for npu-exporter component"
    echo "  --infer-operator            Only add helm meta for infer-operator component"
    echo "  --ascend-operator           Only add helm meta for ascend-operator component"
    echo "  --ascend-device-plugin      Only add helm meta for ascend-device-plugin component"
    echo "  --ascend-for-volcano        Only add helm meta for ascend-for-volcano component"
    echo "  --k8s-rdma-shared-dev-plugin Only add helm meta for k8s-rdma-shared-dev-plugin component"
    echo "  -h, --help                  Show this help message"
    echo ""
    echo "Multiple options can be combined, e.g.:"
    echo "  $0 --namespace --clusterd --noded"
    echo "  $0 --delete-old-device-plugin --add-helm-meta-all"
}

main() {
    if [ $# -eq 0 ]; then
        usage
        exit 1
    fi

    while [ $# -gt 0 ]; do
        case "$1" in
            --all)
                delete_old_device_plugin
                add_helm_meta_all
                ;;
            --delete-old-demonset)
                delete_old_device_plugin
                ;;
            --add-helm-meta-all)
                add_helm_meta_all
                ;;
            --namespace)
                all_namespace
                ;;
            --clusterd)
                component_clusterd
                ;;
            --noded)
                component_noded
                ;;
            --npu-exporter)
                component_npu-exporter
                ;;
            --infer-operator)
                component_infer-operator
                ;;
            --ascend-operator)
                component_ascend-operator
                ;;
            --ascend-device-plugin)
                component_ascend-device-plugin
                ;;
            --ascend-for-volcano)
                component_ascend-for-volcano
                ;;
            --k8s-rdma-shared-dev-plugin)
                component_k8s-rdma-shared-dev-plugin
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
        shift
    done
}

main "$@"

echo "========== Done =========="
