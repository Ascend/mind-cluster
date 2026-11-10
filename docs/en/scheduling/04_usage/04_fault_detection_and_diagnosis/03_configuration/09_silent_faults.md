# Configuring Silent Fault Detection

## Overview

A silent fault is a special public fault comprehensively determined by ClusterD based on software fault and hardware fault information. This chapter describes how to enable silent fault detection and adjust detection and release parameters through the `silent_fault_policy.conf` item in the ConfigMap (`clusterd-config-cm`). For the definition, determination conditions, and detection principles of silent faults, see [Silent Faults](../02_fault_types/07_silent_faults.md).

## Configuration File Description

Silent fault detection is disabled by default and must be explicitly enabled by users. ClusterD performs silent fault detection based on the content of `silent_fault_policy.conf` in the ConfigMap (`clusterd-config-cm`). For parameter descriptions of silent_fault_policy.conf, see [Table 1](#silent_table01).

>[!NOTE]
>
>The fault code and level of a silent fault are configured in `publicFaultConfiguration.json`, and users can adjust the level through `publicCustomization.json`. For details, see [Configuring Public Faults](./08_public_faults.md).

**Table 1** Parameter description of silent_fault_policy.conf

<a name="silent_table01"></a>

|Primary Parameter|Secondary Parameter|Type|Description|
|--|--|--|--|
|enabled|-|bool|Silent fault detection switch. Values include:<ul><li>true: enables silent fault detection.</li><li>false: disables silent fault detection.</li></ul><p>The default value is false. When this switch is disabled, ClusterD will clear the silent faults in the clusterd-manual-info-cm and statistic-fault-info ConfigMaps.</p>|
|detect|min_task_cards|int|Total number of NPUs in a job required for silent fault determining. The value range is (0, 10000000), and the default value is 16.|
|detect|consecutive_times|int|Number of valid rescheduling attempts that must be reached within the determining period before a silent fault is determined. The value range is (0, 10000), and the default value is 3.|
|detect|hardware_fault_window_seconds|int|Window without hardware faults before and after rescheduling. The value range is (3, 86400), the default value is 30, and the unit is second.|
|detect|window_seconds|int|Determining period for silent faults. The value range is [30, 31536000), the default value is 10800 (3 hours), and the unit is second.|
|release|fault_free_seconds|int|Automatic release duration of a silent fault. Values include:<ul><li>-1: no automatic release.</li><li>(0, 31536000): explicit release duration, in s (second).</li></ul><p>The default value is 172800 (48 hours). After a silent fault is determined, a new hit refreshes the timer, and the release time is recalculated from the latest hit.</p>|

>[!NOTE]
>
>If the `enabled` field is missing, ClusterD recognizes it as `false`. If other fields are not configured or their values are invalid, ClusterD uses the default values of these fields.

## (Optional) Enabling Silent Fault Detection

**Procedure**

The following describes how to enable silent fault detection.

1. Log in to the environment and run the following command to query the current configuration.

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - If `clusterd-config-cm` exists, perform [step 3](#silent_li03) to edit it.
    - If `clusterd-config-cm` does not exist, perform [step 2](#silent_li02) to create it.

    >[!NOTE]
    >Normally `clusterd-config-cm` exists, because this ConfigMap is created by default when ClusterD is installed. If it does not exist, check whether an error occurred during the ClusterD installation.

2. <a name="silent_li02"></a>Create `clusterd-config-cm` required for silent fault detection.

    Save the following content as the file `cm.yaml`:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    data:
      silent_fault_policy.conf: |
        enabled: true
        detect:
          min_task_cards: 16
          consecutive_times: 3
          hardware_fault_window_seconds: 30
          window_seconds: 10800
        release:
          fault_free_seconds: 172800
    ```

    Run the following command:

    ```shell
    kubectl apply -f cm.yaml
    ```

    The following is the output example, indicating that the ConfigMap is created successfully.

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. <a name="silent_li03"></a>Run the following command to edit `clusterd-config-cm`.

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    Modify the silent fault detection configuration as required. For parameter descriptions, see [Table 1](#silent_table01).

    ```yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      silent_fault_policy.conf: |
        # Modify the silent fault detection switch
        enabled: true
        detect:
          # Modify the total number of NPUs in a job required for silent fault determining
          min_task_cards: 16
          # Modify the number of valid reschedulings that must be reached within the determining period
          consecutive_times: 3
          # Modify the window without hardware faults before and after rescheduling
          hardware_fault_window_seconds: 30
          # Modify the determining period for silent faults
          window_seconds: 10800
        release:
          # Modify the automatic release duration of a silent fault
          fault_free_seconds: 172800
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    ```

4. After the modification, press "Esc" and enter `:wq!` to save and exit.
5. After `clusterd-config-cm` is updated and takes effect (the detection period of ClusterD is 300s), check whether the operation is successful.

    1. Run the following command to query the ClusterD component name.

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        The following is the output example:

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. Query the ClusterD component logs by using the queried component name.

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        If "load silent fault policy config success" is displayed in the log, the silent fault detection configuration is loaded successfully.

## (Optional) Releasing a Silent Fault

A silent fault can be released in the following two ways.

### Automatic Release

After a silent fault is hit, the silent fault is automatically released after the `fault_free_seconds` duration (48 hours by default) is reached. If the silent fault is hit again during the isolation period, the timer is refreshed and the release time is postponed. If the silent fault is hit repeatedly, the release time is postponed continuously, and the silent fault is released only when no new hit occurs within the `fault_free_seconds` duration.

### Manual Release

A silent fault is isolated by the entire node, and it can be released only after all chips of the node are deleted from `clusterd-manual-info-cm`. A silent fault entry is identified by the `FaultLevel` value of `SilentFault` in `Detail`.

1. Run the following command to edit `clusterd-manual-info-cm`.

    ```shell
    kubectl edit cm -n cluster-system clusterd-manual-info-cm
    ```

2. Delete all chips in the value of the `Total` field of the node (that is, the list of all chips of the node), so that `Total` is cleared.

3. After the modification, press "Esc" and enter `:wq!` to save and exit.
4. After waiting for 15s, run the following command to check whether the silent fault entry of the node still exists in the Total and Detail fields of `clusterd-manual-info-cm`. If it does not exist, the silent fault is released successfully, and the node can be used normally.

    ```shell
    kubectl describe cm -n cluster-system clusterd-manual-info-cm
    ```

    >[!NOTE]
    >
    >- Deleting only some of the chips does not take effect. ClusterD automatically writes the complete node back in the next round.
    >- If both silent faults and manually isolated chips exist on a node, all chips must be cleared to release the silent fault, and the silent fault and the manual isolation are released together.
    >- For details about `clusterd-manual-info-cm`, see [clusterd-manual-info-cm](../../../06_api/04_clusterd/00_cluster_resources.md#clusterd-manual-info-cm).
