# Configuring Inference Job Switch Affinity

Currently, switch affinity can only be configured on Atlas 800I A2 inference servers. Enabling this feature helps avoid downstream traffic conflicts on Spine switches. For details about how this feature works, see [Switch Affinity Scheduling 1.0](../03_basic_scheduling/01_affinity_scheduling/04_node_based_affinity.md#switch-affinity-scheduling-10).

## Prerequisites

You have completed steps in [(Optional) Using Volcano Switch Affinity Scheduling](../../05_developer_guide/00_installation_deployment/00_manual_installation/05_volcano.md#optional-using-volcano-switch-affinity-scheduling).

## Procedure

Set the switch affinity `tor-affinity` to `normal-schema`. The following is a YAML example:

```Yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-server        # Indicates the role of MindIE Motor in Ascend Job. Do not modify it.
    jobID: mindie-ms-test        # The unique identifier of the current MindIE Motor inference job in the cluster. Configure it based on the actual situation.
    tor-affinity: normal-schema    # Enable switch affinity
    ring-controller.atlas: ascend-910b
```

## Viewing Switch Affinity Scheduling Results

1. Run the following command to check the ConfigMap content.

    ```shell
    kubectl describe cm -n kube-system basic-tor-node-cm
    ```

    Command output:

    ```text
    ====
    tor_info:
    ----
    {
      "version": "1.0",
      "tor_count": 4,
      "server_list":[
        {
          "tor_id": 0,
          "tor_ip": "192.168.0.x",
          "server": [
            {
              "server_ip": "192.168.1.x",
              "npu_count": 8,
              "slice_id": 0
            },
            {
              "server_ip": "192.168.1.x",
              "npu_count": 8,
              "slice_id": 2
            },
            ...
          ]
        },
            ...
      ]
    }
    ```

2. Run the following command to check Pod scheduling status.

    ```shell
    kubectl get pod --all-namespaces -owide
    ```

    Command output:

    ```ColdFusion
    NAMESPACE        NAME                                READY   STATUS    RESTARTS   AGE     IP            Node
    ...
    default          mindie-server-0-master-0            1/1     Running   0          10s     192.168.1.x   worker0
    default          mindie-server-0-worker-0            1/1     Running   0          10s     192.168.1.x   worker1
    ...
    ```

Compare the Pod IPs obtained in step 2 with the `basic-tor-node-cm` obtained in step 1. Confirm that multiple instances are distributed under the same ToR, which indicates that the switch affinity feature is running successfully.
