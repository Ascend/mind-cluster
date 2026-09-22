# 配置节点故障<a name="ZH-CN_TOPIC_0000002479226584"></a>

## 配置文件说明<a name="ZH-CN_TOPIC_0000002479226562"></a>

故障检测特性针对节点故障中**节点硬件故障**的不同级别进行分级处理。NodeD组件会获取到当前故障的故障码，根据NodeDConfiguration.json中故障码配置的故障级别，对故障进行相应处理。节点硬件故障支持的故障级别和处理方式说明如下。

NodeD组件的配置文件NodeDConfiguration.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码的故障级别，可以通过由NodeDConfiguration.json创建的mindx-dl-node-fault-config文件实现，操作指导请参见[（可选）配置节点硬件故障级别](#可选配置节点硬件故障级别)。故障级别说明及节点状态说明请参见[自定义节点故障](../../../06_api/03_noded.md#自定义节点故障)。

## （可选）配置节点硬件故障级别<a name="ZH-CN_TOPIC_0000002511346507"></a>

在制作NodeD镜像时，会将故障级别配置文件NodeDConfiguration.json内置在镜像中，启动NodeD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在集群中创建ConfigMap文件（mindx-dl-node-fault-config）。

- 如果NodeD启动时，集群中已经存在该mindx-dl-node-fault-config，NodeD会优先按照已存在的mindx-dl-node-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装NodeD后，集群中已经存在mindx-dl-node-fault-config，NodeD的默认NodeDConfiguration.json将不会生效，使用集群中已经存在mindx-dl-node-fault-config。若想要使用NodeDConfiguration.json的默认配置，可以删除mindx-dl-node-fault-config，使NodeD读取默认的NodeDConfiguration.json文件。
- 如果mindx-dl-node-fault-config内容存在格式错误等问题，NodeD会默认读取镜像中内置的NodeDConfiguration.json文件的内容，作为当前故障处理依据。

**操作步骤<a name="section25164134219"></a>**

以故障码0100001D为例，将当前故障的处理策略NotHandleFault（无需处理）修改为PreSeparateFault（该节点上有任务则不处理，后续不调度任务到该节点）的操作示例如下。

1. 登录环境，进入NodeD解压目录。
2. 执行以下命令，创建动态配置故障级别所需ConfigMap文件（mindx-dl-node-fault-config）。

    ```shell
    kubectl create cm mindx-dl-node-fault-config -n mindx-dl  --from-file=./NodeDConfiguration.json
    ```

    回显示例如下：

    ```ColdFusion
    configmap/mindx-dl-node-fault-config created
    ```

    **表 1**  参数说明

    <a name="table1925220306444"></a>
    <table><thead align="left"><tr id="row172531430134411"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p16253163094420"><a name="p16253163094420"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p152534301443"><a name="p152534301443"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1325318306446"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15214952162210"><a name="p15214952162210"></a>mindx-dl-node-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p621417523229"><a name="p621417523229"></a>创建的<span id="ph188631730142314"><a name="ph188631730142314"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="row925343011442"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p82141952122212"><a name="p82141952122212"></a>mindx-dl</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p0214952142217"><a name="p0214952142217"></a>命名空间名称，不能修改该命名空间。</p>
    </td>
    </tr>
    <tr id="row1253183012444"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p182141521222"><a name="p182141521222"></a>NodeDConfiguration.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p22148525226"><a name="p22148525226"></a>用于配置故障码以及对应的故障级别，必须与NodeDConfiguration.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3. 执行以下命令，编辑mindx-dl-node-fault-config文件。

    ```shell
    kubectl edit cm -n mindx-dl mindx-dl-node-fault-config
    ```

4. 在mindx-dl-node-fault-config文件中，找到故障码0100001D。

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
    >自定义故障级别时，若不小心导致出现以下问题，则本次修改无效，NodeD将会使用上一次保存的配置进行处理。
    >- 文件格式异常或故障码取值错误，故障码只能为8位的包含数字和字母的字符串。
    >- 同一故障码同时配置在多个故障级别中。

5. 将故障码0100001D在**NotHandleFaultCodes**中删除，并添加到**PreSeparateFaultCodes**中。

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

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-node-fault-config文件更新后，查看操作是否成功。
    1. 执行以下命令，查询NodeD组件日志名称。

        ```shell
        kubectl get pods -A | grep noded
        ```

        回显示例如下：

        ```ColdFusion
        mindx-dl      noded-c5f52   1/1     Running   0               2m16s
        ```

    2. 通过查询到的组件日志名称，查询NodeD的组件日志信息。

        ```shell
        kubectl logs noded-c5f52 -n mindx-dl -f
        ```

        若日志出现“update fault config success”，表示动态配置故障码操作成功。
