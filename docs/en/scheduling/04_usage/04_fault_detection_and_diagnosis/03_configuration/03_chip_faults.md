# Configuring Chip Faults<a name="ZH-CN_TOPIC_0000002479226466"></a>

## Overview<a name="ZH-CN_TOPIC_0000002511346521_0101"></a>

Both Ascend Device Plugin and ClusterD provide the capability of manually isolating a chip based on the fault frequency. Their functions differ as follows:

- Ascend Device Plugin determines faults based on the node dimension and counts the frequency based on the faults that actually occur.
- ClusterD determines faults based on both the job dimension and the chip dimension:
    - If the faulty chip is not associated with a job, or if the job no longer exists when the fault is handled, the fault is not counted into the ClusterD manual isolation frequency.
    - For a fault that is associated with a valid job and is determined to be caused by hardware, each occurrence on a chip increments the fault frequency count of that chip by one.
    - If the same fault occurs on multiple cards of one job within 30s, the fault is considered not caused by hardware and is not counted into the fault frequency. This rule applies to most scenarios. However, the count may deviate in cases such as residual processes remaining after a Pod is deleted.
    - Only newly occurring faults trigger the upper limit determination. If you adjust the configured threshold to a value equal to or lower than the current count, the isolation determination is triggered only when the next fault occurs.
    - After ClusterD is restarted, the frequency count information is lost, and the fault frequency starts counting from zero again.
    - After the isolation is removed, if job scheduling does not meet expectations, check whether the `huawei.com/scheduler.chip1softsharedev.enable=false` label exists on the node, and delete it if it exists.

The manual chip isolation functions of Ascend Device Plugin and ClusterD theoretically do not need to cover the same fault codes. If you do not want to use the isolation function of Ascend Device Plugin, see [(Optional) Configuring the Chip Fault Frequency and Duration](#ZH-CN_TOPIC_0000002511426473) and delete the manual chip isolation related configuration from the `faultCustomization.json` file. If you do not want to use the isolation function of ClusterD, see [(Optional) Configuring the Chip Fault Frequency](#ZH-CN_TOPIC_0000002511426473_01) and turn off the manual chip isolation function switch.

If both Ascend Device Plugin and ClusterD have manually isolated the same chip, the isolation must be removed for each of them separately. For how to remove the isolation of Ascend Device Plugin, see the "Manually restore a forcibly isolated chip" step in [(Optional) Configuring the Chip Fault Frequency and Duration](#ZH-CN_TOPIC_0000002511426473). For how to remove the isolation of ClusterD, see the "Manually restore a manually isolated chip" step in [(Optional) Configuring the Chip Fault Frequency](#ZH-CN_TOPIC_0000002511426473_01).

## Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511346521_02"></a>

### Configuration File Description<a name="ZH-CN_TOPIC_0000002511346521"></a>

For **chip faults**, the fault detection feature supports processing based on the configuration of fault levels, fault frequencies, and fault durations.

- To handle chip faults **by level**, Ascend Device Plugin obtains the fault code of the current fault and handles the fault accordingly based on the fault level configured for the fault code in `faultCode.json`.
- To handle chip faults **by fault frequency and duration**, Ascend Device Plugin obtains the fault code of the current fault and handles the fault accordingly based on the fault frequency and duration configured for the fault in `faultCustomization.json`.

`faultCode.json` and `faultCustomization.json` are system configuration files. Do not modify them unless necessary. If the default frequency fault configuration of Ascend Device Plugin contains faults that can be triggered by software causes, you can delete the corresponding fault code yourself. (Software causes make a fault occur repeatedly in a large volume within a short period under one job, so Ascend Device Plugin perceives that the fault has reached the fault frequency and sets a large number of devices to the manually isolated state.)

If you need to change the fault level corresponding to a fault code, use the `mindx-dl-fault-config` file created from `faultCode.json` and `faultCustomization.json`.

>[!NOTE]
>
>- For the fault code of each fault, see [Chip Fault Code References](../../../07_references/05_appendix.md#ZH-CN_TOPIC_0000002511426289).
>- For the fault levels that can be configured for chip faults, see [Fault Levels](#zh-cn_topic_0000002171521445_section5245155017242).
>- For the fault frequencies and durations that can be configured for chip faults, see [Fault Frequency and Duration](#zh-cn_topic_0000002171521445_section115842029104220).

**Fault Levels in faultCode.json<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

The fault detection feature handles chip faults by level. If you need to change the fault level of a fault code, see [(Optional) Configuring the Chip Fault Level](#ZH-CN_TOPIC_0000002479226532).

After obtaining a chip fault code from the driver, Ascend Device Plugin classifies the fault into several levels based on the impact of the fault code on devices and jobs. For details, see [Table 1](../../../06_api/02_ascend_device_plugin.md#ZH-CN_TOPIC_0000002511346805).

>[!NOTE]
>
>- Stop the training process before resetting the chip; otherwise, the reset fails.
>- If Ascend Device Plugin receives an unrecognized fault code through subscription (one that is not saved in `faultCode.json`), the fault is handled by default according to the handling suggestion given by the subscription interface. If the fault level received by the subscription interface is "prompt" or "minor", the fault is handled at the NotHandleFault level. If the fault level is any other level, the fault is handled at the `SeparateNPU` level.

**Fault Frequency and Duration<a name="zh-cn_topic_0000002171521445_section115842029104220"></a>**

The fault detection feature handles chip faults by fault frequency and duration. Some hardware faults may occur repeatedly in one training job, causing the training job to be interrupted and rescheduled repeatedly. For the fault codes corresponding to these faults, cluster scheduling components provide the initialization configuration file faultCustomization.json, which upgrades the fault level.

- The relationship between the initialization configuration provided by the `faultCustomization.json` file and the fault types is described in [Initialization Configuration and Fault Types](#zh-cn_topic_0000002171521445_section13684172919539).
- For the default configuration (default values) of the `faultCustomization.json` file, see [Table 2](../../../06_api/02_ascend_device_plugin.md#ZH-CN_TOPIC_0000002511346805).
- If you need to change the fault frequency and duration configuration, see [(Optional) Configuring the Chip Fault Frequency and Duration](#ZH-CN_TOPIC_0000002511426473).

**Initialization Configuration and Fault Types<a name="zh-cn_topic_0000002171521445_section13684172919539"></a>**

Currently, the `faultCustomization.json` file provides only the initialization configuration that upgrades the fault level of identifiable hardware faults.

If a single node has fewer than 3 faults within 24 hours, the chip fault level is upgraded to `ManuallySeparateNPU`, the fault level that requires manual intervention. For details, see [faultCustomization.json Parameter Description](#zh-cn_topic_0000002171521445_section33036167576).

The following uses HBMC Ca Parity error with fault code 80E18005 as an example to show how to upgrade the current fault level to `ManuallySeparateNPU` (requiring manual intervention).

```json
  "FaultFrequency": [
    {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    },
    {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
    }
  ],
```

>[!NOTE]
>
>- When the handling policy of the fault is `ManuallySeparateNPU`, you can refer to the "Manually restore a forcibly isolated chip" step in [(Optional) Configuring the Chip Fault Frequency and Duration](#ZH-CN_TOPIC_0000002511426473).
>- In addition to identifiable hardware faults, the `faultCustomization.json` file also contains the following types of faults.
>     - Faults that require no handling: The occurrence of such faults does not affect training jobs or devices, and no initialization configuration is provided to upgrade the fault level.
>     - Faults that cannot be identified as hardware or software faults: Such faults cannot be accurately identified as hardware or software faults and affect training jobs. No initialization configuration is provided to upgrade the fault level. You are advised to manually configure, based on the actual situation, the maximum number of resumable training attempts supported by the job and the handling policy for the fault after the maximum number is reached. For details, see [(Optional) Configuring the Chip Fault Frequency and Duration](#ZH-CN_TOPIC_0000002511426473).
>     - Software configuration faults: Such faults are software configuration problems and do not occur under normal conditions. No initialization configuration is provided to upgrade the fault level. You are advised to check whether the software versions match each other.

**faultCustomization.json Parameter Description<a name="zh-cn_topic_0000002171521445_section33036167576"></a>**

If the user does not manually modify the `faultCustomization.json` file, Ascend Device Plugin handles faults based on the default configuration (default values) of `faultCustomization.json`. For details about the parameters of the `faultCustomization.json` file, see [Table 2](../../../06_api/02_ascend_device_plugin.md#ZH-CN_TOPIC_0000002511346805).

### (Optional) Configuring the Chip Fault Level<a name="ZH-CN_TOPIC_0000002479226532"></a>

When the Ascend Device Plugin image is built, the `faultCode.json` and `faultCustomization.json` configuration files are built into the image. When Ascend Device Plugin is started, the default configurations of these two files are read as the basis for current fault handling. For details about `faultCode.json` and `faultCustomization.json`, see [Configuration File Description](#ZH-CN_TOPIC_0000002511346521).

If you want to customize fault levels or graceful fault tolerance related configurations, you can create a ConfigMap file (`mindx-dl-fault-config`) in the cluster.

- If `mindx-dl-fault-config` already exists in the cluster when Ascend Device Plugin is started, Ascend Device Plugin preferentially uses the content configured in the existing `mindx-dl-fault-config` as the basis for current fault handling.
- If `mindx-dl-fault-config` already exists in the cluster after Ascend Device Plugin is reinstalled, the default `faultCode.jso`n of Ascend Device Plugin does not take effect, and the existing `mindx-dl-fault-config` in the cluster is used.
- If you want to use the default configuration of `faultCode.json` or `faultCustomization.json`, delete `mindx-dl-fault-config` so that Ascend Device Plugin reads the default `faultCode.json`, `SwitchFaultCode.json`, or `faultCustomization.json` file.
- If the content of the ConfigMap file has problems such as format errors, Ascend Device Plugin reads the content of the ConfigMap file built into the image by default as the basis for current fault handling.

**Configuring Fault Levels Using faultCode.json<a name="zh-cn_topic_0000001951258609_section112139052513"></a>**

The following uses "dmp\_daemon node status check exception" with fault code 80E21007, as an example to describe how to change the handling policy of the current fault from `NotHandleFaultCodes` (no handling required) to `RestartNPUCodes` (isolate the chip and perform job rescheduling).

1. Log in to the environment and go to the Ascend Device Plugin decompression directory.
2. Run the following command to create the ConfigMap file (`mindx-dl-fault-config`) required for dynamically configuring fault codes.

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **Table 1** Parameter description

    <a name="zh-cn_topic_0000001951258609_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001951258609_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951258609_p16631548171910"><a name="zh-cn_topic_0000001951258609_p16631548171910"></a>Parameter</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951258609_p1663144816197"><a name="zh-cn_topic_0000001951258609_p1663144816197"></a>Required</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951258609_p775918210209"><a name="zh-cn_topic_0000001951258609_p775918210209"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001951258609_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1863164816197"><a name="zh-cn_topic_0000001951258609_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1063194861910"><a name="zh-cn_topic_0000001951258609_p1063194861910"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p157595292015"><a name="zh-cn_topic_0000001951258609_p157595292015"></a>Name of the <span id="zh-cn_topic_0000001951258609_ph126311642183015"><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a>ConfigMap</span> file required for dynamically configuring fault codes. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1963194819195"><a name="zh-cn_topic_0000001951258609_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p76316488192"><a name="zh-cn_topic_0000001951258609_p76316488192"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p276092142019"><a name="zh-cn_topic_0000001951258609_p276092142019"></a>Namespace where mindx-dl-fault-config resides. This namespace name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p964144891914"><a name="zh-cn_topic_0000001951258609_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1164748191916"><a name="zh-cn_topic_0000001951258609_p1164748191916"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p876012211206"><a name="zh-cn_topic_0000001951258609_p876012211206"></a>If this parameter is not specified, the default value is 300s. It specifies the period for checking whether the mindx-dl-fault-config file is updated, in seconds, with a value range of 30 to 3600. The modification of PollInterval takes effect in the next period.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1964748141915"><a name="zh-cn_topic_0000001951258609_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p10641648191915"><a name="zh-cn_topic_0000001951258609_p10641648191915"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p147602211206"><a name="zh-cn_topic_0000001951258609_p147602211206"></a>Used to save fault codes and must be consistent with the faultCode.json file name.</p>
    </td>
    </tr>
    </tbody>
    </table>

3. Run the following command to edit the `mindx-dl-fault-config` file.

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

4. In the `mindx-dl-fault-config` file, find the fault code `80E21007`.

    ```json
    "NotHandleFaultCodes":[

    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE]
    >If the same fault code is configured under multiple fault levels, the setting is displayed as successful, but the fault is handled at the higher level by default.

5. Delete the fault code `80E21007` from `NotHandleFaultCodes` and add it to `RestartNPUCodes`.

    ```json
    "NotHandleFaultCodes":[
         "80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    "RestartNPUCodes":[
       "8C204E00","A8028802","A4302003","A4302004","A4302005","A4302006","A4302009","A430200A","80CF8009","80CF8008","80E21007",...
    ...
       ],
    ```

6. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
7. After the update of the `mindx-dl-fault-config` file takes effect (300s if `PollInterval` is not specified), check whether the operation is successful.
    1. Run the following command to query the name of the Ascend Device Plugin log.

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        Command output:

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. Use the queried component log name to query the Ascend Device Plugin log information.

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        If "load fault code from configmap success" appears in the log, the manual fault code configuration is successful.

### (Optional) Configuring the Chip Fault Frequency and Duration<a name="ZH-CN_TOPIC_0000002511426473"></a>

When the Ascend Device Plugin image is built, the `faultCode.json` and `faultCustomization.json` configuration files are built into the image. When Ascend Device Plugin is started, the default configurations of these two files are read as the basis for current fault handling. For details about `faultCode.json` and `faultCustomization.json`, see [Configuration File Description](#ZH-CN_TOPIC_0000002511346521).

If you want to customize the chip fault frequency and duration, you can create a ConfigMap file (`mindx-dl-fault-config`) in the cluster.

- If `mindx-dl-fault-config` already exists in the cluster when Ascend Device Plugin is started, Ascend Device Plugin preferentially uses the content configured in the existing `mindx-dl-fault-config` as the basis for current fault handling.
- If `mindx-dl-fault-config` already exists in the cluster after Ascend Device Plugin is reinstalled, the default `faultCustomization.json` of Ascend Device Plugin does not take effect, and the existing `mindx-dl-fault-config` in the cluster is used. If you want to use the default configuration of `faultCustomization.json`, delete `mindx-dl-fault-config` so that Ascend Device Plugin reads the default `faultCustomization.json` file.
- If the content of the ConfigMap file has problems such as format errors, Ascend Device Plugin reads the content of the ConfigMap file built into the image by default as the basis for current fault handling.

>[!CAUTION]
>Changing the fault frequency is a high-risk operation. Improper modification may cause chips to be isolated by mistake. For example, a software fault caused by a job error occurs repeatedly in a large volume within a short period, so Ascend Device Plugin perceives that the fault has reached the fault frequency and sets a large number of chips to the manually isolated state, making a large number of nodes unschedulable.

**Procedure<a name="section141902103110"></a>**

The following uses fault code `80CB8002` as an example. If the `80CB8002` fault occurs repeatedly on a chip and causes the training job to be rescheduled repeatedly, you can manually configure the maximum number of resumable training attempts supported within 24 hours as 2, and the handling policy for the fault as `ManuallySeparateNPU` after the maximum number is reached.

1. Log in to the environment and go to the Ascend Device Plugin decompression directory.
2. Run the following command to check whether `mindx-dl-fault-config` has been created based on the faultCode.json file.

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - If `mindx-dl-fault-config` already exists and contains the `faultCustomization.json` related field, perform [step 4](#zh-cn_topic_0000002136360238_li38432520129) to edit the file.
    - If `mindx-dl-fault-config` already exists but does not contain the `faultCustomization.json` related field, save the content of `mindx-dl-fault-config` first, delete the `mindx-dl-fault-config` file, and then perform [step 3](#zh-cn_topic_0000002136360238_li1946014413123) to create the file.
    - If mindx-dl-fault-config does not exist, perform [step 3](#zh-cn_topic_0000002136360238_li1946014413123) to create the file.

3. <a name="zh-cn_topic_0000002136360238_li1946014413123"></a>Run the following command to create the ConfigMap file (`mindx-dl-fault-config`) required for configuring the chip fault frequency and duration.

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json --from-file=./faultCustomization.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **Table 2** Parameter description

    <a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a>Parameter</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a>Required</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a>Name of the <span id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a>ConfigMap</span> file required for dynamically configuring fault codes. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a>Namespace where mindx-dl-fault-config resides. This namespace name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a>If this parameter is not specified, the default value is 300s. It specifies the period for checking whether the mindx-dl-fault-config file is updated, in seconds, with a value range of 30 to 3600. The modification of PollInterval takes effect in the next period.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a>Used to save fault codes and must be consistent with the faultCode.json file name.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_row9289716194614"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a>faultCustomization.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a>Used to customize the graceful fault tolerance time, fault frequency, and fault duration (supported only for parameter plane network faults). If this parameter is not specified, there is no fault frequency configuration and other configurations are processed using default values.</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002136360238_li38432520129"></a>Run the following command to edit the `mindx-dl-fault-config` file.

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

    Modify the fault frequency and duration of the chip based on the actual situation.

    ```yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
    PollInterval: "300"
    # Modify the fault level of chip faults
    faultCode.json: |
    {
    "NotHandleFaultCodes":[
    ...
    }
    # Modify the fault frequency and duration of chip faults
    faultCustomization.json: |
    {
     "GraceTolerance": {
     "WaitProcessReadCMTime": 30,
     "WaitDeviceResetTime": 150,
     "WaitFaultSelfHealingTime": 15
    },
    "FaultFrequency": [
     {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["81078603"],
      "TimeWindow": 86400,
      "Times": 5,
      "FaultHandling": "ManuallySeparateNPU",
      "ReleaseTimeWindow": 172800
     }
    ],
    "FaultDuration": [
     {
      "EventId": ["81078603"],
      "FaultTimeout": 20,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     },
     {
      "EventId": ["81B18603"],
      "FaultTimeout": 5,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     }
    ]
   }
    kind: ConfigMap
    metadata:
    creationTimestamp: "2024-06-20T10:12:07Z"
    name: mindx-dl-fault-config
    namespace: kube-system
    resourceVersion: "52893696"
    selfLink: /api/v1/namespaces/kube-system/configmaps/mindx-dl-fault-config
    uid: bba9e17f-41dd-43b3-848e-3d29cb8c595a
    ```

5. In the `mindx-dl-fault-config` file, add the following code under the FaultFrequency field to set the maximum number of resumable training attempts supported by the 80CB8002 fault within 24 hours to 2 and the handling policy for the fault to `ManuallySeparateNPU` after the maximum number is reached.

    ```json
    {
      "EventId": ["80CB8002"],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    }
    ```

6. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
7. After the update of the `mindx-dl-fault-config` file takes effect (300s if `PollInterval` is not specified), check whether the operation is successful.
    1. Run the following command to query the name of the Ascend Device Plugin log.

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        Command output:

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. Use the queried component log name to query the Ascend Device Plugin log information.

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        >[!NOTE]
        >- If "load fault customization from configmap complete" appears in the log, the manual fault frequency configuration is successful.
        >- If "modify  _xxx_  success" appears in the log, the <i>xxx</i> parameter in `faultCustomization.json` in the ConfigMap is set successfully.
        >- If "insert fault frequency success" appears in the log, a frequency fault occurrence time has been recorded. Within the frequency window, after the number of records of this fault on this card reaches the fault frequency trigger count, the fault level corresponding to the frequency fault is reported.

8. (Optional) Manually restore a forcibly isolated chip. When the handling policy of the fault is ManuallySeparateNPU, the chip remains isolated after the fault is recovered. You can manually restore a forcibly isolated chip when the release conditions are not met.
    1. Run the following command to find the `device-info-cm` reported by Ascend Device Plugin on the node.

        ```shell
        kubectl get cm -n kube-system | grep deviceinfo | grep {nodeName}
        ```

    2. Run the following command to edit the `device-info-cm`.

        ```shell
        kubectl edit cm -n kube-system {configMapName}
        ```

    3. Delete the names of the chips that have recovered to health after `ManuallySeparateNPU` under `data`.

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        data:
          DeviceInfoCfg: '{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1718702470},"CheckCode":"4f00cf1d220da26a8fdbeb5ba163a751d4b264c48b81d22149257e272ae3b413"}'
          ManuallySeparateNPU: Ascend910-0
        ```

        >[!NOTE]
        >Delete all chip names after the `ManuallySeparateNPU` field and set the value to `""`.

    4. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
    5. After waiting for one reporting period (if the device information changes, it is reported within the health status check period; if the device information does not change, the reporting period is fixed at 5 minutes), run the following command to check whether the chip name that was just deleted exists in `ManuallySeparateNPU` in `device-info-cm`. If it does not exist, the chip has recovered to health and you can continue to use it normally.

        ```shell
        kubectl describe cm -n kube-system {configMapName}
        ```

## ClusterD<a name="ZH-CN_TOPIC_0000002511346521_03"></a>

### Configuration Description<a name="ZH-CN_TOPIC_0000002511346521_04"></a>

For chip faults, the fault detection feature supports processing based on the fault frequency configuration.

To handle chip faults by level, ClusterD obtains the fault code and fault level of the current fault. For faults other than those of the `NotHandleFault` and `SubHealthFault` levels, the chip status is set to manually isolated based on the fault frequency configured in the ConfigMap (`clusterd-config-cm`). For details about the parameters of this ConfigMap, see [Table 1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md#table208901).

>[!NOTE]
>
>- The ConfigMap (`clusterd-config-cm`) is a system configuration. Do not modify it unless necessary. If you need to change the manual chip isolation detection switch, the fault frequency, or the isolation removal time, you can do so by modifying this ConfigMap. For details, see [(Optional) Configuring the Chip Fault Frequency](#ZH-CN_TOPIC_0000002511426473_01).
>- Configuring the fault code detection scope is not supported. ClusterD determines faults based on the fault level reported by Ascend Device Plugin. Faults other than those of the `NotHandleFault` and `SubHealthFault` levels all enter the manual chip isolation detection process.

### (Optional) Configuring the Chip Fault Frequency<a name="ZH-CN_TOPIC_0000002511426473_01"></a>

When ClusterD is installed, the ConfigMap (`clusterd-config-cm`) is automatically created as the basis for current manual chip isolation. For details about the parameters of this ConfigMap, see [Table 1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md#table208901).

If you want to customize the chip fault frequency, you can do so by modifying this ConfigMap. If the content of the modified ConfigMap has problems such as format errors, ClusterD retains the configuration that was last read successfully as the basis for current manual chip isolation. If the ConfigMap content read when ClusterD is started is incorrect, the manual chip isolation detection mechanism is disabled by default until the format and content are correct.

**Procedure<a name="section14190101"></a>**

The following uses adjusting the manual chip isolation threshold from the default of 3 occurrences within 24 hours to 5 occurrences within 24 hours as an example.

1. Log in to the environment and run the following command to query the current configuration.

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - If `clusterd-config-cm` exists, perform [step 3](#li01010203) to edit it.
    - If `clusterd-config-cm` does not exist, perform [step 2](#li010102) to create it.

    >[!NOTE]
    >Normally, `clusterd-config-cm` exists. If it does not exist, check whether an error occurred during the ClusterD installation.

2. <a name="li010102"></a>Create the ConfigMap (`clusterd-config-cm`) required for manual chip isolation detection.

    Save the following content as the file `cm.yaml`:

    ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: clusterd-config-cm
          namespace: cluster-system
        data:
          manually_separate_policy.conf: |
            enabled: true
            separate:
              fault_window_hours: 24
              fault_threshold: 3
            release:
              fault_free_hours: 48

    ```

    Run the following command:

    ```shell
    kubectl apply -f cm.yaml
    ```

    The following is a command output, which indicates that the creation is successful.

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. <a name="li01010203"></a>Run the following command to edit `clusterd-config-cm`.

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    Modify the fault frequency of manual chip isolation based on the actual situation. For details about the parameters, see [Table 1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md#table208901).

    ```Yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      manually_separate_policy.conf: |
        # Modify the manual chip isolation detection switch
        enabled: true
        separate:
          # Modify the fault frequency of manual chip isolation
          fault_window_hours: 24
          fault_threshold: 5   # Changed from 3 to 5
        release:
          # Modify the isolation removal time
          fault_free_hours: 48
    kind: ConfigMap
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"v1","data":{"manually_separate_policy.conf":"enabled: true\nseparate:\n  fault_window_hours: 24\n  fault_threshold: 3\nrelease:\n  fault_free_hours: 48\n"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"clusterd-config-cm","namespace":"cluster-system"}}
      creationTimestamp: "2026-02-24T11:25:19Z"
      name: clusterd-config-cm
      namespace: cluster-system
      resourceVersion: "3344125"
      selfLink: /api/v1/namespaces/cluster-system/configmaps/clusterd-config-cm
      uid: 68210bfc-f742-4765-a497-b61e9cc6b1a6
    ```

4. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
5. After the update of `clusterd-config-cm` takes effect (the ClusterD detection period is 300s), check whether the operation is successful.
    1. Run the following command to query the name of the ClusterD log.

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        Command output:

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. Use the queried component log name to query the ClusterD log information.

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        >[!NOTE]
        >- If "load manually separate policy config success" appears in the log, the manual modification of the fault frequency of manual chip isolation is successful.
        >- If "node: xx, dev: xx, code: xx is not found in manual fault cache, add" appears in the log, the fault triggers manual isolation.
        >- If "node: xx, dev: xx, code: xx is found in manual fault cache, update last separate time" appears in the log, a fault that has already triggered manual chip isolation has reached the manual isolation fault frequency again, and `LastSeparateTime` in `clusterd-manual-info-cm` is refreshed. For details about `clusterd-manual-info-cm`, see [clusterd-manual-info-cm](../../../06_api/04_clusterd/00_cluster_resources.md#section15483421165190).

6. (Optional) Manually restore a manually isolated chip. When the handling policy of the fault is `ManuallySeparateNPU`, the chip remains isolated after the fault is recovered, and you can manually restore a manually isolated chip.

    1. Run the following command to edit `clusterd-manual-info-cm`.

        ```shell
        kubectl edit cm -n cluster-system clusterd-manual-info-cm
        ```

    2. Delete the names of the chips whose manual isolation needs to be removed after the `Total` field under `Data`, for example, `Ascend910-2`.

        ```json
        Name:         clusterd-manual-info-cm
        Namespace:    cluster-system
        Labels:       <none>
        Annotations:  <none>

        Data
        ====
        localhost.localdomain:
        ----
        {"Total":["Ascend910-0","Ascend910-2","Ascend910-3"],"Detail":{"Ascend910-0":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-2":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-3":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}]}}

        Events:  <none>
        ```

    3. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
    4. After waiting for 15s, run the following command to check whether Ascend910-2 still exists in the `Total` and `Detail` fields in `clusterd-manual-info-cm`. In addition, check whether the `ManuallySeparateNPU` fault of the chip exists in `cluster-info-device-\${m}`. If it does not exist, the manual isolation of the chip is successfully removed and you can continue to use the chip normally.

        ```shell
        kubectl describe cm -n cluster-system clusterd-manual-info-cm
        ```

        >[!NOTE]
        >- Only deleting chips from the Total field is supported. Manual addition is not supported. Other content cannot be modified.
        >- After a manually isolated chip is manually restored, the fault count of the chip is cleared, and manual isolation is triggered again only when the frequency is reached again.
        >- To delete all manually isolated chips on a node, delete all chip names after the `Total` field and set the value to empty (`[]`). To remove the manual isolation of all chips at once, you can directly delete `clusterd-manual-info-cm`.
        >- Within 15s after ClusterD is started, do not modify `clusterd-manual-info-cm` to avoid data errors.
