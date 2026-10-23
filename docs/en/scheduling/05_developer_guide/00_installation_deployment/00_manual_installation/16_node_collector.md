# Node Collector<a name="ZH-CN_TOPIC_0000002524428762"></a>

- Users who use the [ClusterOps Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690) feature must install Node Collector.
- Node Collector is deployed on each compute node as a DaemonSet. It collects job logs from the host and completes log parsing locally.
- Node Collector selects compute nodes through node labels during deployment, and is scheduled only to nodes with the `workerselector=dls-worker-node` label. For details about how to create node labels, see [Creating Node Labels](./01_preparing_for_installation.md#creating-node-labels).
- Node Collector and Agent Core share the same image `ascend-clusterops-agent`. For details about how to obtain the image (build or pull), see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image).
- Before deploying Node Collector, deploy Agent Core first. For details, see [Agent Core](./15_agent_core.md).

## Procedure<a name="section15023132772914"></a>

1. Log in to the K8s management node as the `root` user and run the following command to check whether the Node Collector image and version number are correct.

    ```shell
    docker images | grep ascend-clusterops-agent
    ```

    Command output:

    ```ColdFusion
    ascend-clusterops-agent       v26.2.0              c532e9d0889c        About an hour ago         112MB
    ```

    - If yes, go to [step 2](#li4121412112712).
    - If no, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to build and pull the image.

2. <a name="li4121412112712"></a>Copy the YAML file (`node-collector.yaml`) in the Node Collector software package decompression directory to any directory on the K8s management node.

3. If you do not need to modify the component startup parameters, skip this step. Otherwise, modify the Node Collector startup parameters in the YAML file as required.
   - Node label: Node Collector selects compute nodes through the `workerselector=dls-worker-node` label of the `nodeSelector` field. Modify it based on the actual node labels.
   - Updating the collection contract: The collection contract (`collect_manifest.yaml`) is built into the image by default. If you need to adjust it on site, run the following command through Kubectl Plugin to write the local collection contract to the cluster ConfigMap.

     ```shell
     kubectl ascend_diag --collect-manifest {collect_manifest.yaml}
     ```

4. In the path where the YAML file is stored on the management node, run the following command to start Node Collector.

    ```shell
    kubectl apply -f node-collector.yaml
    ```

    The following is an example:

    ```ColdFusion
    serviceaccount/node-collector created
    ...
    daemonset.apps/node-collector created
    service/node-collector created
    ```

5. Run the following command to check whether the component is started successfully.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is the output example. If a Pod in the **Running** state appears on each compute node, the component is started successfully.

    ```ColdFusion
    NAME                    READY   STATUS    RESTARTS   AGE
    node-collector-8j4kq    1/1     Running   0          45s
    node-collector-m4j4r    1/1     Running   0          45s
    ```
