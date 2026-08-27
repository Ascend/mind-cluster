# DPU Exporter<a name="ZH-CN_TOPIC_0000002511426332"></a>

- 使用**DPU资源监测**时，必须安装DPU Exporter，该组件支持对接Prometheus，通过镜像和二进制两种方式部署。
- 不使用**DPU资源监测**的用户，可以不安装DPU Exporter，请直接跳过本章节。

## 使用约束<a name="section1362795652417"></a>

在安装DPU Exporter前，需要提前了解相关约束，具体说明请参见[表1](#table1050718522712)。

**表 1**  约束说明

<a name="table1050718522712"></a>
<table><thead align="left"><tr id="row2050719520273"><th class="cellrowborder" valign="top" width="29.970000000000002%" id="mcps1.2.3.1.1"><p id="p1950795152712"><a name="p1950795152712"></a><a name="p1950795152712"></a>约束场景</p>
</th>
<th class="cellrowborder" valign="top" width="70.03%" id="mcps1.2.3.1.2"><p id="p75071151278"><a name="p75071151278"></a><a name="p75071151278"></a>约束说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row115077513272"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p17925222412"><a name="p17925222412"></a><a name="p17925222412"></a>DPU驱动与工具</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p450745142713"><a name="p450745142713"></a><a name="p450745142713"></a><span id="ph10112356112714"><a name="ph10112356112714"></a><a name="ph10112356112714"></a>DPU Exporter</span>会周期性调用<code>hinicadm5</code>工具查询DPU全局指标，并读取<code>/sys/class/net/</code>下的sysfs接口获取Interface级指标。如果要升级驱动或工具，请先停止业务任务，再停止<span id="ph154413248376"><a name="ph154413248376"></a><a name="ph154413248376"></a>DPU Exporter</span>服务。</p>
<div class="note" id="note1993172317416"><a name="note1993172317416"></a><a name="note1993172317416"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226379_p18934232419"><a name="zh-cn_topic_0000002479226379_p18934232419"></a><a name="zh-cn_topic_0000002479226379_p18934232419"></a>镜像部署时，需确保宿主机上<code>hinicadm5</code>位于<code>/usr/sbin/</code>目录下，且其依赖的动态链接库位于<code>/usr/lib64/</code>和<code>/lib/</code>目录下。YAML中已通过hostPath将这些目录挂载至容器中。</div>
</div></div>
</td>
</tr>
<tr id="row54685525283"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p5249201634115"><a name="p5249201634115"></a><a name="p5249201634115"></a><span id="ph1461172794117"><a name="ph1461172794117"></a><a name="ph1461172794117"></a>K8s</span>版本</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p5468852142814"><a name="p5468852142814"></a><a name="p5468852142814"></a>容器运行时为Docker的场景使用<span id="ph98079531287"><a name="ph98079531287"></a><a name="ph98079531287"></a>DPU Exporter</span>前需要确保环境的<span id="ph18807253152811"><a name="ph18807253152811"></a><a name="ph18807253152811"></a>K8s</span>版本，若<span id="ph6808453102814"><a name="ph6808453102814"></a><a name="ph6808453102814"></a>K8s</span>版本为1.24.x及以上版本，需要用户自行<a href="https://github.com/mirantis/cri-dockerd#build-and-install" target="_blank" rel="noopener noreferrer">安装cri-dockerd</a>依赖。</p>
</td>
</tr>
<tr id="row7507135142717"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p050719519272"><a name="p050719519272"></a><a name="p050719519272"></a>hinicadm5并发</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p6555101612382"><a name="p6555101612382"></a><a name="p6555101612382"></a><code>hinicadm5</code>工具不支持并发调用，DPU Exporter已采用串行方式逐卡采集全局指标，无需用户额外处理。</p>
</td>
</tr>
<tr id="row_dpu_sysfs"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p_dpu_sysfs_cat">sysfs文件接口</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p_dpu_sysfs_desc">DPU Exporter通过读取<code>/sys/class/net/&lt;interface_name&gt;/</code>目录下的文件获取Interface级指标（carrier、carrier_changes、operstate及statistics目录下所有统计项），需保证DPU-Exporter运行用户对该目录有读取权限。镜像部署时，YAML中已通过hostPath将<code>/sys</code>挂载至容器中（readOnly: true）。</p>
</td>
</tr>
</tbody>
</table>

## 操作步骤<a name="section83111543151613"></a>

DPU Exporter支持两种安装方式，用户可根据实际情况选择其中一种进行安装。该组件仅提供HTTP服务，如需使用更为安全的HTTPS服务，请自行修改源码进行适配。

- （推荐）以镜像方式运行，安装步骤参见[镜像方式运行](#section2035402135915)。
- 当安全要求较高时，建议在物理机上以二进制方式运行，安装步骤参见[二进制方式运行](#section103551921135918)。

## 镜像方式运行<a name="section2035402135915"></a>

1. 以root用户登录各计算节点。

2. （可选）修改config.json配置文件，配置采集周期和指标白名单。
    1. 进入DPU Exporter软件包解压目录。
    2. <a name="li11364381195"></a>打开config.json文件。

        ```shell
        vi config.json
        ```

    3. 按"i"进入编辑模式，根据实际需要配置采集周期和指标白名单。

        配置文件示例如下：

        ```json
        {
            "hinicadm5CollectorInterval": 40,
            "sysfsCollectorInterval": 20,
            "dpuListRefreshInterval": 60,
            "metricWhiteList": []
        }
        ```

        <a name="table192202574407"></a>

        |参数|说明|
        |---|---|
        |hinicadm5CollectorInterval|全局指标（hinicadm5 counter）采集周期，单位为秒。取值范围为1~86400，默认值40。|
        |sysfsCollectorInterval|Interface级指标（sysfs文件接口）采集周期，单位为秒。取值范围为1~86400，默认值20。|
        |dpuListRefreshInterval|DPU设备列表刷新周期，单位为秒。取值范围为1~86400，默认值60。|
        |metricWhiteList|自定义指标白名单，支持前缀匹配（以`*`结尾）。为空时使用[默认白名单](../../../06_api/16_dpu_exporter.md#默认白名单)。|

    4. 按"Esc"键，输入:wq!保存并退出。

    5. 将修改后的config.json文件所在的目录挂载到容器中，路径为`/etc/dpu-exporter/config.json`。或挂载至容器中的其他路径，则需要修改YAML中DPU Exporter的-config启动参数（详见[参数说明](#参数说明)）。

3. （可选）配置文件挂载。

    配置文件挂载情况有如下两种，具体的配置方法请参考[动态配置加载说明](#动态配置加载说明)章节。

    - **默认情况**：配置文件通过ConfigMap挂载到容器中，使用YAML中定义的默认配置，支持通过修改ConfigMap实现动态配置修改。
    - **HostPath挂载**：将宿主机上的配置文件挂载到容器中，路径为`/etc/dpu-exporter/config.json`。此方式支持每个节点独立配置，也可通过共享目录实现全局统一配置。

4. 请根据实际使用的容器运行时，查看DPU Exporter镜像是否存在、名称和版本号是否正确。

    - **Docker场景**：执行如下命令。

        ```shell
        docker images | grep dpu-exporter
        ```

        回显示例如下。

        ```ColdFusion
        dpu-exporter                         v26.2.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    - **Containerd场景**：执行如下命令。

        ```shell
        ctr -n k8s.io c ls | grep dpu-exporter
        ```

    若镜像存在且名称和版本号均正确，执行[步骤5](#li0640635114212)。若镜像不存在，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

5. <a name="li0640635114212"></a>将DPU Exporter软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。

6. 如不修改组件的其他启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中DPU Exporter的启动参数。启动参数如[表2](#table872410431915)所示，也可执行<b>./dpu-exporter -h</b>查看参数说明。

7. 在管理节点的YAML所在路径，执行以下命令，启动DPU Exporter。

    ```shell
    kubectl apply -f dpu-exporter-v26.2.0.yaml
    ```

    启动示例如下：

    ```ColdFusion
    namespace/dpu-exporter created
    configmap/dpu-exporter-config created
    daemonset.apps/dpu-exporter created
    service/dpu-exporter created
    ```

    >[!NOTE]
    >启动DPU Exporter时，若出现报错"Error from server (NotFound): error when creating "dpu-exporter-v6.0.0.yaml":namespaces "dpu-exporter" not found"，说明DPU Exporter的命名空间未创建成功，需执行以下命令手动创建。
    >
    >```shell
    >kubectl create ns dpu-exporter
    >```

8. 在任意节点执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n dpu-exporter
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                      READY   STATUS    RESTARTS   AGE
    dpu-exporter-xxxxx        1/1     Running   0          11s
    ```

    >[!NOTE]
    >
    >- DPU Exporter以镜像方式运行时，请确保"/sys"目录、"/usr/sbin"目录（hinicadm5工具）、"/usr/lib64"、"/lib"目录（动态链接库）和"/var/log/hinic5"目录（hinicadm5日志）已通过hostPath挂载至DPU Exporter容器中。YAML中已默认配置这些挂载。
    >- 安装组件后，组件的Pod状态不为Running，可查看Pod日志定位问题：
    >
    > ```bash
    > kubectl logs -n dpu-exporter <pod-name>
    > ```

9. 验证指标采集是否正常。

    在任意节点执行以下命令，访问metrics接口。其中&lt;node-ip&gt;为部署DPU Exporter的节点IP。

    ```shell
    curl http://<node-ip>:8080/metrics
    ```

    正常时应能看到以`dpu_`和`dpu_interface_`为前缀的指标。

## 二进制方式运行<a name="section103551921135918"></a>

DPU Exporter组件以镜像方式运行时需使用特权容器、root用户和挂载hostPath目录，如果容器被人恶意利用，有容器逃逸风险。当安全性要求较高时，可直接在物理机上通过二进制方式运行。

>[!NOTE]
>
>- 以二进制方式部署DPU Exporter时，可以使用非root用户（例如hwMindX）进行部署。请将日志目录权限修改为hwMindX，命令示例如下：**chown <i>hwMindX:hwMindX</i> /var/log/mindx-dl/dpu-exporter**。
>- 下文步骤中的用户均为hwMindX。
>- 以非root用户运行时，需确保`hinicadm5`工具对运行用户可执行。由于代码启动时会校验`hinicadm5`文件属主与当前进程用户的一致性，建议通过设置SUID权限解决：**chmod 4755 /usr/sbin/hinicadm5**，或直接以root用户运行。

1. 使用root用户登录服务器。

2. 将DPU Exporter软件包上传至服务器的任意目录（如"/home/ascend-dpu-exporter"）并进行解压操作。

3. （可选）创建配置文件目录并将配置文件拷贝到该目录下。

    ```shell
    mkdir -p /etc/dpu-exporter
    cp config.json /etc/dpu-exporter/
    ```

4. （可选）修改config.json配置文件，配置采集周期和指标白名单。
    1. 进入"/etc/dpu-exporter"目录。
    2. <a name="li1445835411479"></a>打开config.json文件。

        ```shell
        vi config.json
        ```

    3. 按"i"进入编辑模式，根据实际需要配置采集周期和指标白名单。

        配置文件示例如下：

        ```json
        {
            "hinicadm5CollectorInterval": 40,
            "sysfsCollectorInterval": 20,
            "dpuListRefreshInterval": 60,
            "metricWhiteList": []
        }
        ```

        <a name="zh-cn_topic_0000002511426332_table192202574407"></a>

        |参数|说明|
        |---|---|
        |hinicadm5CollectorInterval|全局指标（hinicadm5 counter）采集周期，单位为秒。取值范围为1~86400。默认值40。|
        |sysfsCollectorInterval|Interface级指标（sysfs文件接口）采集周期，单位为秒。取值范围为1~86400。默认值20。|
        |dpuListRefreshInterval|DPU设备列表刷新周期，单位为秒。取值范围为1~86400。默认值60。|
        |metricWhiteList|自定义指标白名单，支持前缀匹配（以`*`结尾）。为空时使用[默认白名单](../../../06_api/16_dpu_exporter.md#默认白名单)。|

    4. <a name="li18459954104719"></a>按"Esc"键，输入:wq!保存并退出。

5. 创建并编辑dpu-exporter.service文件。
    1. 执行以下命令，创建dpu-exporter.service文件。

        ```shell
        vi /home/ascend-dpu-exporter/dpu-exporter.service
        ```

    2. 参考如下内容，写入dpu-exporter.service文件中。

        <pre>
        [Unit]
        Description=Ascend dpu exporter
        Documentation=hiascend.com

        [Service]
        ExecStart=/bin/bash -c "/usr/local/bin/dpu-exporter -config=/etc/dpu-exporter/config.json -port=8080 >/dev/null 2>&1 &"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking
        User=hwMindX
        Group=hwMindX

        [Install]
        WantedBy=multi-user.target</pre>

        DPU Exporter默认侦听端口8080，可通过修改启动参数"-port"和"dpu-exporter.service"文件的"ExecStart"字段修改侦听端口。

    3. 按"Esc"键，输入:wq!保存并退出。

6. 创建并编辑dpu-exporter.timer文件。通过配置timer延时启动，可保证DPU Exporter启动时DPU卡已就位。
    1. 执行以下命令，创建dpu-exporter.timer文件。

        ```shell
         vi /home/ascend-dpu-exporter/dpu-exporter.timer
        ```

    2. 参考以下示例，并将其写入dpu-exporter.timer文件中。

        <pre>
        [Unit]
        Description=Timer for DPU Exporter Service

        [Timer]
        OnBootSec=60s            # 设置DPU Exporter延时启动时间，请根据实际情况调整
        Unit=dpu-exporter.service

        [Install]
        WantedBy=timers.target</pre>

    3. 按"Esc"键，输入:wq!保存并退出。

7. 依次执行以下命令，启用DPU Exporter服务。

    ```shell
    cd /home/ascend-dpu-exporter
    cp dpu-exporter /usr/local/bin
    mkdir -p /var/log/mindx-dl/dpu-exporter
    cp dpu-exporter.service /etc/systemd/system
    chattr +i /etc/systemd/system/dpu-exporter.service
    cp dpu-exporter.timer /etc/systemd/system
    chattr +i /etc/systemd/system/dpu-exporter.timer
    chmod 500 /usr/local/bin/dpu-exporter
    chown hwMindX:hwMindX /usr/local/bin/dpu-exporter
    chattr +i /usr/local/bin/dpu-exporter
    chown hwMindX:hwMindX /var/log/mindx-dl/dpu-exporter
    chmod 4755 /usr/sbin/hinicadm5
    systemctl enable dpu-exporter.timer
    systemctl start dpu-exporter
    systemctl start dpu-exporter.timer
    ```

    >[!NOTE]
    >首次部署时`systemctl start dpu-exporter`会立即启动服务，timer的`OnBootSec`延时仅在系统重启时生效。如需首次启动也通过timer延时，可仅执行`systemctl start dpu-exporter.timer`。

8. 验证服务是否启动成功。

    ```shell
    curl http://127.0.0.1:8080/metrics
    ```

    正常时应能看到以`dpu_`和`dpu_interface_`为前缀的指标。

## 参数说明<a name="section2042611570393"></a>

**表 2** DPU Exporter启动参数

<a name="table872410431915"></a>

|参数|类型|默认值|说明|
|--|--|--|--|
|-config|string|/etc/dpu-exporter/config.json|配置文件路径。配置文件为JSON格式，包含采集周期和指标白名单等配置项，详见[配置文件说明](#table192202574407)。|
|-port|int|8080|侦听端口，取值范围为1025~65535。|
|-cardType|string|huawei|DPU卡类型。目前仅支持<code>huawei</code>。|
|-logLevel|int|0|日志级别：<ul><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>|
|-maxAge|int|7|日志备份时间，取值范围为7~700，单位为天。|
|-maxBackups|int|30|转储后日志文件保留个数上限，取值范围为1~180，单位为个。|
|-logFile|string|/var/log/mindx-dl/dpu-exporter/dpu-exporter.log|日志文件路径，该参数当前不支持通过命令行修改，日志路径固定为默认值。<p>单个日志文件超过20MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：dpu-exporter-触发转储的时间.log，如：dpu-exporter-2023-10-07T03-38-24.402.log。</p>|
|-h或者-help|无|无|显示帮助信息。|

## 动态加载配置说明<a name="动态配置加载说明"></a>

DPU Exporter支持动态加载配置文件，无需重启组件即可使配置变更生效。

### 二进制部署场景

配置文件路径为`/etc/dpu-exporter/config.json`。

直接修改该文件即可，DPU Exporter会自动检测文件变更并重新加载配置。

### K8s HostPath挂载场景

- HostPath挂载的优点：
  - 配置修改后立即生效。
  - 每个节点可以独立配置，也可结合共享目录实现全局统一配置。
- HostPath挂载的缺点：配置变更不易追踪。

在部署YAML中配置HostPath挂载，将宿主机上的配置文件挂载到容器的`/etc/dpu-exporter`路径：

1. 在每个节点上准备配置文件。

    ```bash
    mkdir -p /etc/dpu-exporter
    cp config.json /etc/dpu-exporter/
    ```

2. 修改YAML中的volume挂载为HostPath方式。

    ```yaml
    volumeMounts:
      - name: config
        mountPath: /etc/dpu-exporter
        readOnly: true

    volumes:
      - name: config
        hostPath:
          path: /etc/dpu-exporter
          type: DirectoryOrCreate
    ```

修改宿主机上的配置文件即可，DPU Exporter会自动检测文件变更并重新加载配置。

### K8s ConfigMap挂载场景（默认）

- ConfigMap挂载的优点：
  - 统一管理所有节点的配置，支持一键更新所有节点配置。
  - 配置变更可追踪和版本控制。
- ConfigMap挂载的缺点：
  - 所有节点使用相同的配置，无法独立配置单个节点。
  - 配置生效有一定延迟（K8s更新ConfigMap到容器的时间）。

DPU Exporter的YAML默认使用ConfigMap方式挂载配置文件。ConfigMap在YAML中已定义，包含`config.json`键。

1. 更新ConfigMap中的配置。

    ```bash
    kubectl edit cm -n dpu-exporter dpu-exporter-config
    ```

    或直接重新部署YAML：

    ```bash
    kubectl apply -f dpu-exporter-v6.0.0.yaml
    ```

    >[!NOTICE]
    >
    > 必须直接挂载ConfigMap到目录，**不能使用`subPath`**：
    > - 使用`subPath`会导致ConfigMap更新后无法自动同步到容器内，需要重启才能生效。
    > - 修改ConfigMap后，容器内文件不能实时更新，需要等待一定时间（K8s机制，最长约10分钟）才能感知到文件变化。

更新ConfigMap后，K8s会自动更新容器中的配置文件，DPU Exporter会自动检测并重新加载配置。

### 配置变更验证

配置变更后，可以通过查看DPU Exporter的日志确认配置是否成功加载：

```shell
# 二进制部署
tail -100f /var/log/mindx-dl/dpu-exporter/dpu-exporter.log

# 镜像部署
kubectl logs -n dpu-exporter <pod-name> --tail=100
```

成功加载配置会打印类似如下日志：

```text
config hot-reload: config file modified, reloading...
collector chains rebuilt after config reload
```

## 白名单配置说明<a name="白名单配置说明"></a>

DPU Exporter通过白名单机制控制hinicadm5全局指标的采集范围，避免采集不必要的指标。

### 白名单优先级

- **自定义白名单优先**：如果用户配置了`metricWhiteList`（非空），则以自定义白名单为准。
- **默认白名单**：如果`metricWhiteList`为空或未配置，则使用内置默认白名单。

### 白名单匹配规则

- **精确匹配**：指标名与模式完全一致则匹配。
- **前缀匹配**：模式以`*`结尾时，匹配所有以该前缀开头的指标名。

### 配置示例

在config.json中配置自定义白名单：

```json
{
    "hinicadm5CollectorInterval": 40,
    "sysfsCollectorInterval": 20,
    "dpuListRefreshInterval": 60,
    "metricWhiteList": [
        "roce_rx_*",
        "roce_tx_*",
        "roce_cmdq_ctr_roce_cmd_2rst_qp"
    ]
}
```

上述配置表示：只采集以`roce_rx_`开头的指标、以`roce_tx_`开头的指标，以及精确匹配`roce_cmdq_ctr_roce_cmd_2rst_qp`的指标。
