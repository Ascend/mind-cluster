# Agent Core<a name="ZH-CN_TOPIC_0000002524428761"></a>

- Users who use the [ClusterOps Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690) feature must install Agent Core.
- Agent Core is deployed on the management node as a Deployment. As the centralized control center for fault diagnosis, it receives diagnosis requests, schedules log collection on each node by job dimension, and performs centralized diagnosis.
- Agent Core is scheduled through `nodeSelector` to the management node with the `masterselector=dls-master-node` label. For details about how to create node labels, see [Creating Node Labels](./01_preparing_for_installation.md#creating-node-labels).
- Agent Core and Node Collector share the same image `ascend-clusterops-agent`. For details about how to obtain the image (build or pull), see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image).
- Before deploying Agent Core, complete the steps of creating a user, creating log directories, and creating a namespace in [Installation Preparation](./01_preparing_for_installation.md).

## Procedure<a name="section15023132772914"></a>

1. Log in to the K8s management node as the `root` user and run the following command to check whether the Agent Core image and version number are correct.

    ```shell
    docker images | grep ascend-clusterops-agent
    ```

    Command output:

    ```ColdFusion
    ascend-clusterops-agent       v26.2.0              c532e9d0889c        About an hour ago         698MB
    ```

    - If yes, go to [step 2](#li4121412112711).
    - If no, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to build and pull the image.

2. <a name="li4121412112711"></a>Copy the YAML file (`agent-core.yaml`) in the Agent Core software package decompression directory to any directory on the K8s management node.

3. If you do not need to modify the component startup parameters, skip this step. Otherwise, modify the Agent Core startup parameters in the YAML file as required.

4. In the path where the YAML file is stored on the management node, run the following command to start Agent Core.

    ```shell
    kubectl apply -f agent-core.yaml
    ```

    The following is an example:

    ```ColdFusion
    configmap/agent-core-task-crds created
    serviceaccount/agent-core created
    ...
    deployment.apps/agent-core created
    service/agent-core created
    ```

5. Run the following command to check whether the component is started successfully.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is the output example. **Running** indicates that the component is started successfully.

    ```ColdFusion
    NAME                          READY   STATUS    RESTARTS   AGE
    agent-core-7844cb867d-fwcj7   1/1     Running   0          45s
    ```
