# 创建vNPU<a name="ZH-CN_TOPIC_0000002479226382"></a>

- 在物理机和虚拟机使用npu-smi工具创建vNPU的命令基本相同，所以本节命令可以适用于物理机和虚拟机，其中只有Atlas 推理系列产品支持在虚拟机创建vNPU。
- 当使用**静态虚拟化**创建vNPU并挂载到容器时，需要使用**npu-smi**命令创建vNPU，再参考[挂载vNPU](./06_mounting_vnpu.md)。
- 当使用**动态虚拟化**时，无需提前创建vNPU，请跳过本节，直接在容器拉起时按以下要求进行参数配置。
    - 使用Ascend Docker Runtime：参考[方式一：Ascend Docker Runtime挂载vNPU](./06_mounting_vnpu.md#方式一ascend-docker-runtime挂载vnpu)，通过ASCEND\_VISIBLE\_DEVICES和ASCEND\_VNPU\_SPECS参数从物理芯片上虚拟化出多个vNPU并挂载至容器。
    - 使用MindCluster集群调度组件（Ascend Device Plugin和Volcano）：参考[动态虚拟化](./06_mounting_vnpu.md#动态虚拟化)，运行任务时自动按照配置要求调用接口创建vNPU。

**创建vNPU方法<a name="section206799361399"></a>**

- 在物理机执行以下命令设置虚拟化模式（如果是在虚拟机内划分vNPU，不需要执行本命令），命令格式如下。

    **npu-smi set -t vnpu-mode -d** _mode_

    **表 1**  参数说明

    <a name="table11489191211336"></a>

    |类型|描述|
    |--|--|
    |mode|<p>虚拟化实例模式。取值为0或1：</p><ul><li>0：虚拟化实例容器模式</li><li>1：虚拟化实例虚拟机模式</li></ul>|

- 创建vNPU。命令格式如下：

    **npu-smi set -t create-vnpu -i** _id_ **-c** _chip\_id_ **-f** _vnpu\_config_  \[**-v** _vnpu\_id_\] \[**-g** _vgroup\_id_\]

    <a name="table1654283920393"></a>

    |类型|描述|
    |--|--|
    |id|设备ID。通过<b>npu-smi info -l</b>命令查出的NPU ID即为设备ID。|
    |chip_id|芯片ID。通过<b>npu-smi info -m</b>命令查出的Chip ID即为芯片ID。|
    |vnpu_config|虚拟化实例模板名称，详细请参见[虚拟化模板](./03_virtualization_templates.md)。|
    |vnpu_id|<p>指定需要创建的vNPU的ID。</p><ul><li>首次创建可以不指定该参数，由系统默认分配。若重启后业务需要使用重启前的vnpu_id，可以使用-v参数指定重启前的vnpu_id进行恢复。</li><li>取值范围：<ul><li>Atlas 推理系列产品<p>vnpu_id的取值范围为[phy_id \* 16 + 100, phy_id \* 16+107]。</p></li><li>Atlas 训练系列产品<p>vnpu_id的取值范围为[phy_id \* 16 + 100, phy_id \* 16+115]。</p></li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody">phy_id表示芯片物理ID，可通过执行<strong>ls /dev/davinci*</strong>命令获取芯片的物理ID。例如/dev/davinci0，表示芯片的物理ID为0。</div></div></li><li>vnpu_id传入4294967295时表示不指定虚拟设备号。</li><li>同一台服务器内不可重复创建相同vnpu_id的vNPU。</li></ul>|
    |vgroup_id|虚拟资源组vGroup的ID，取值范围为0~3。<p>vGroup是指虚拟化时NPU根据用户指定的虚拟化模板划分出虚拟资源组vGroup，每个vGroup包含若干AICore、AICPU、片上内存、DVPP资源。</p><p>仅<span>Atlas 推理系列产品</span>支持本参数。</p>|

    使用示例如下：

    - 在设备0中编号为0的芯片上根据模板vir02创建vNPU。

        ```shell
        npu-smi set -t create-vnpu -i 0 -c 0 -f vir02
                Status : OK         Message : Create vnpu success
        ```

    - 在设备0中编号为0的芯片上指定vnpu\_id为103创建vNPU设备，此vNPU的模板为vir02。

        ```shell
        npu-smi set -t create-vnpu -i 0 -c 0 -f vir02 -v 103
                Status : OK         Message : Create vnpu success
        ```

    - 在设备0中编号为0的芯片上指定vnpu\_id为100并指定vgroup\_id为1创建vNPU设备，此vNPU的模板为vir02。

        ```shell
        npu-smi set -t create-vnpu -i 0 -c 0 -f vir02 -v 100 -g 1
                Status : OK         Message : Create vnpu success
        ```

- 配置vNPU恢复状态。该参数用于设备重启时，设备能够保存vNPU配置信息，重启后，vNPU配置依然有效。

    **npu-smi set -t vnpu-cfg-recover -d** _mode_

    mode表示vNPU的配置恢复使能状态，“1”表示开启状态，“0”表示关闭状态，默认为使能状态。

    执行如下命令设置vNPU的配置恢复状态，以下命令表示将vNPU的配置恢复状态设置为使能状态。

    **npu-smi set -t vnpu-cfg-recover -d** _1_

    ```ColdFusion
           Status : OK
           Message : The VNPU config recover mode Enable is set successfully.
    ```

- 查询vNPU的配置恢复状态。

    以下命令表示查询当前环境中vNPU的配置恢复使能状态。

    **npu-smi info -t vnpu-cfg-recover**

    ```ColdFusion
    VNPU config recover mode : Enable
    ```

- 查询vNPU信息。命令格式：

    **npu-smi info -t info-vnpu -i** _id_ **-c** _chip\_id_

    <a name="table1585213289319"></a>

    |类型|描述|
    |--|--|
    |id|设备ID。通过<b>npu-smi info -l</b>命令查出的NPU ID即为设备ID。|
    |chip_id|芯片ID。通过<b>npu-smi info -m</b>命令查出的Chip ID即为芯片ID。|

    执行如下命令查询vNPU信息。以下命令表示查询设备0中编号为0的芯片的vNPU信息。

    **npu-smi info -t info-vnpu -i** _0_ **-c** _0_

    ![](../../../../figures/scheduling/1.png)

    >[!NOTE] 
    >Atlas 推理系列产品支持返回AICPU，Vgroup ID信息，Atlas 训练系列产品不支持返回AICPU，Vgroup ID信息。
