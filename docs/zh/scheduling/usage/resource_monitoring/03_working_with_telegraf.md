# 通过Telegraf使用<a name="ZH-CN_TOPIC_0000002479227042"></a>

本章节指导用户安装部署Telegraf软件，并通过Telegraf查看资源监测的相关数据信息，数据信息的相关说明可参见[Telegraf数据信息说明](../../api/npu_exporter.md#telegraf数据信息说明)章节。

**二进制集成Telegraf<a name="section31082142614"></a>**

>[!NOTE] 
>除了二进制集成外，集群调度支持通过修改NPU Exporter开源代码，集成Telegraf源码。

1. （可选）如果没有创建NPU Exporter的日志目录，需要依次执行以下命令，创建日志目录。

    ```shell
    mkdir -m 750 /var/log/mindx-dl/npu-exporter
    chown hwMindX:hwMindX /var/log/mindx-dl/npu-exporter
    ```

2. 从[昇腾社区](https://www.hiascend.com/zh/developer/download/community/result?module=dl+cann)获取NPU Exporter软件包，并从中解压出NPU Exporter二进制文件npu-exporter，并上传至环境任意路径（如“/home/npu\_plugin”）。
3. 执行以下命令，创建npu\_plugin.conf文件。

    ```shell
    vi npu_plugin.conf
    ```

    在文件中添加NPU Exporter二进制文件路径，示例如下。

    <pre>
    [[inputs.execd]]
      command = ["/home/npu_plugin/npu-exporter", "-platform=Telegraf", "-poll_interval=10s", "-hccsBWProfilingTime=200"] 
      signal = "none"  
    [[outputs.file]] 
      files=["stdout"]</pre>

    command字段的输入参数说明如[表1](#table5347115241118)所示。

    **表 1**  参数说明

    <a name="table5347115241118"></a>

    |参数名|类型|默认值|取值说明|是否必选|
    |--|--|--|--|--|
    |-platform|string|Prometheus|指定对接平台，取值如下：<ul><li>Prometheus：对接Prometheus</li><li>Telegraf：对接Telegraf</li></ul>|是|
    |-poll_interval|duration(int)|1s|Telegraf数据上报的间隔时间，此参数在对接Telegraf平台时才起作用，即需要指定-platform=Telegraf时才生效，否则该参数不生效。|否|
    |-hccsBWProfilingTime|int|200|HCCS链路带宽采样时长，取值范围[1，1000]，单位为ms。|否|

4. （可选）如果没有安装Telegraf，需执行以下步骤安装Telegraf。
    - **离线安装（推荐）**
        1. 进入[Telegraf下载页面](https://github.com/influxdata/telegraf/releases)。
        2. 选择需要安装的版本，完成下载，如：telegraf-1.34.3\_linux\_arm64.tar.gz。
        3. 将上述安装包上传到服务器的任意路径下。
        4. 在软件包所在目录执行如下命令进行解压。示例如下。

            ```shell
            tar -zxvf telegraf-1.34.3_linux_arm64.tar.gz
            ```

        5. 进入解压目录，在./usr/bin路径下找到Telegraf二进制文件，将该文件拷贝到任意路径下（如“/home/npu\_plugin”）。

    - **在线安装**
        1. 进入[Telegraf下载页面](https://www.influxdata.com/downloads/)。
        2. 在下拉框选择操作系统及Telegraf版本。

            **图 1**  下载Telegraf<a name="fig131640329479"></a>  
            ![](../../../figures/scheduling/下载Telegraf.png "下载Telegraf")

        3. 拷贝弹框中的安装命令到待安装设备上，执行命令，完成安装。

5. 执行以下命令，运行Telegraf。
    - 如使用离线安装，请执行以下命令运行Telegraf。

        ```shell
        ./telegraf --config npu_plugin.conf
        ```

    - 如使用在线安装，请执行以下命令运行Telegraf。

        ```shell
        telegraf --config npu_plugin.conf
        ```

        Telegraf运行成功后，回显示例如下，从npu_chip_link_speed开始之后的信息即为监测的昇腾AI处理器的数据信息。

        ```ColdFusion
        2023-09-15T10:11:31Z I! Loading config file: ../npu_plugin.conf
        2023-09-15T10:11:31Z I! Starting Telegraf 1.26.0
        2023-09-15T10:11:31Z I! Available plugins: 236 inputs, 9 aggregators, 27 processors, 22 parsers, 57 outputs, 2 secret-stores2023-09-15T10:11:31Z I! Loaded inputs: execd
        2023-09-15T10:11:31Z I! Loaded aggregators: 
        2023-09-15T10:11:31Z I! Loaded processors: 
        2023-09-15T10:11:31Z I! Loaded secretstores: 
        2023-09-15T10:11:31Z I! Loaded outputs: file
        2023-09-15T10:11:31Z I! Tags enabled: host=xxx
        2023-09-15T10:11:31Z I! [agent] Config: Interval:10s, Quiet:false, Hostname:"xxx", Flush Interval:10s
        2023-09-15T10:11:31Z I! [inputs.execd] Starting process: /xxx/npu-exporter [-platform=Telegraf -poll_interval=1m]
        Ascend910-0,host=xxx npu_chip_link_speed=104857600000i,npu_chip_roce_rx_cnp_pkt_num=0i,npu_chip_roce_unexpected_ack_num=0i,npu_chip_optical_vcc=3245.1,npu_chip_optical_rx_power_1=0.8585,npu_chip_info_hbm_used_memory=0i,npu_chip_mac_rx_pause_num=0i,npu_chip_roce_tx_all_pkt_num=0i,npu_chip_roce_tx_cnp_pkt_num=0i,npu_chip_info_temperature=46,npu_chip_mac_rx_bad_pkt_num=0i,npu_chip_roce_tx_err_pkt_num=0i,npu_chip_optical_rx_power_3=0.8466,npu_chip_optical_rx_power_0=0.7933,npu_chip_info_network_status=0i,npu_chip_mac_rx_pfc_pkt_num=0i,npu_chip_mac_tx_bad_pkt_num=0i,npu_chip_roce_rx_all_pkt_num=0i,npu_chip_mac_rx_bad_oct_num=0i,npu_chip_optical_tx_power_1=0.9162,npu_chip_info_utilization=0,npu_chip_info_power=73.9000015258789,npu_chip_info_link_status=1i,npu_chip_info_bandwidth_rx=0,npu_chip_mac_tx_pfc_pkt_num=0i,npu_chip_roce_rx_err_pkt_num=0i,npu_chip_roce_verification_err_num=0i,npu_chip_optical_state=1i,npu_chip_info_bandwidth_tx=0,npu_chip_mac_tx_bad_oct_num=0i,npu_chip_roce_out_of_order_num=0i,npu_chip_roce_qp_status_err_num=0i,npu_chip_optical_rx_power_2=0.855,npu_chip_optical_tx_power_0=0.9095,npu_chip_info_hbm_utilization=0,npu_chip_link_up_num=2i,npu_chip_info_health_status=1i,npu_chip_mac_tx_pause_num=0i,npu_chip_roce_new_pkt_rty_num=0i,npu_chip_optical_temp=53,npu_chip_optical_tx_power_2=1.0342,npu_chip_optical_tx_power_3=0.9715 1694772754612200641
        ```
