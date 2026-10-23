# Configuring Node Faults<a name="ZH-CN_TOPIC_0000002479226584"></a>

## Configuration File Description<a name="ZH-CN_TOPIC_0000002479226562"></a>

The fault detection feature performs level-based handling for different levels of **node hardware faults** among node faults. NodeD obtains the fault code of the current fault and handles the fault accordingly based on the fault level configured for the fault code in `NodeDConfiguration.json`. The fault levels and handling modes supported for node hardware faults are described as follows.

The configuration file `NodeDConfiguration.json` of NodeD is a system configuration file. Do not modify it at will unless you have special requirements. If you need to modify the fault level of a fault code, you can do so through the `mindx-dl-node-fault-config` file created from `NodeDConfiguration.json`. For details, see [(Optional) Configuring Node Hardware Fault Levels](#optional-configuring-node-hardware-fault-levels). For details about fault levels and node status, see [Customizing Node Faults](../../../06_api/03_noded.md#customizing-node-faults).

## (Optional) Configuring Node Hardware Fault Levels<a name="ZH-CN_TOPIC_0000002511346507"></a>

When the NodeD image is built, the fault level configuration file `NodeDConfiguration.json` is built into the image. When NodeD is started, it reads the default configuration in this file as the basis for current fault handling.

If you want to customize fault levels, you can create a ConfigMap file `(mindx-dl-node-fault-config`) in the cluster.

- If `mindx-dl-node-fault-config` already exists in the cluster when NodeD is started, NodeD preferentially uses the content configured in the existing `mindx-dl-node-fault-config` as the basis for current fault handling.
- If `mindx-dl-node-fault-config` already exists in the cluster after NodeD is reinstalled, the default NodeDConfiguration.json of NodeD does not take effect, and the existing `mindx-dl-node-fault-config` in the cluster is used. If you want to use the default configuration of `NodeDConfiguration.json`, you can delete m`indx-dl-node-fault-config` so that NodeD reads the default `NodeDConfiguration.json` file.
- If the content of `mindx-dl-node-fault-config` has problems such as format errors, NodeD reads the content of the `NodeDConfiguration.json` file built into the image by default as the basis for current fault handling.

**Procedure<a name="section25164134219"></a>**

The following uses fault code `0100001D` as an example to change the handling policy of the current fault from `NotHandleFault` (no handling required) to `PreSeparateFault` (if the node has job running, do not handle; subsequently, do not schedule jobs to the node).

1. Log in to the environment and go to the NodeD decompression directory.
2. Run the following command to create the ConfigMap file (`mindx-dl-node-fault-config`) required for dynamically configuring fault levels.

    ```shell
    kubectl create cm mindx-dl-node-fault-config -n mindx-dl  --from-file=./NodeDConfiguration.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-node-fault-config created
    ```

    **Table 1** Parameter description

    <a name="table1925220306444"></a>
    <table><thead align="left"><tr id="row172531430134411"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p16253163094420"><a name="p16253163094420"></a>Parameter</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p152534301443"><a name="p152534301443"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1325318306446"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15214952162210"><a name="p15214952162210"></a>mindx-dl-node-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p621417523229"><a name="p621417523229"></a>Name of the created <span id="ph188631730142314"><a name="ph188631730142314"></a>ConfigMap</span> file. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="row925343011442"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p82141952122212"><a name="p82141952122212"></a>mindx-dl</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p0214952142217"><a name="p0214952142217"></a>Name of the namespace. This namespace cannot be modified.</p>
    </td>
    </tr>
    <tr id="row1253183012444"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p182141521222"><a name="p182141521222"></a>NodeDConfiguration.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p22148525226"><a name="p22148525226"></a>Used to configure fault codes and the corresponding fault levels. The name must be the same as the NodeDConfiguration.json file name.</p>
    </td>
    </tr>
    </tbody>
    </table>

3. Run the following command to edit the `mindx-dl-node-fault-config` file.

    ```shell
    kubectl edit cm -n mindx-dl mindx-dl-node-fault-config
    ```

4. In the `mindx-dl-node-fault-config` file, find fault code `0100001D`.

    ```json
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
              "0100001D","03000009","03000013","0300000D","03000011"
            ],
    ...
      ],
    ...
    ```

    >[!NOTE]
    >When customizing fault levels, if the following problems occur by accident, this modification is invalid, and NodeD uses the last saved configuration for handling.
    >- The file format is abnormal or the fault code value is incorrect. A fault code can only be an 8-character string containing digits and letters.
    >- The same fault code is configured in multiple fault levels at the same time.

5. Delete fault code 0100001D from **NotHandleFaultCodes** and add it to **PreSeparateFaultCodes**.

    ```json
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
             "03000009","03000013","0300000D","03000011"
            ],
            "PreSeparateFaultCodes":[
              "28000037","00000011", "0100001D"
    ...
            ],
    ...
    ```

6. After the modification is complete, press the `Esc` key, type `:wq!`, and press `Enter` to save and exit.
7. After the `mindx-dl-node-fault-config` file is updated, check whether the operation is successful.
    1. Run the following command to query the NodeD component log name.

        ```shell
        kubectl get pods -A | grep noded
        ```

        Command output:

        ```ColdFusion
        mindx-dl      noded-c5f52   1/1     Running   0               2m16s
        ```

    2. Query the component log information of NodeD by using the queried component log name.

        ```shell
        kubectl logs noded-c5f52 -n mindx-dl -f
        ```

        If "update fault config success" appears in the log, the dynamic fault code configuration is successful.
