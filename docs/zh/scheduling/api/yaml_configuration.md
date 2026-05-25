# 任务YAML配置说明<a name="yaml_configuration"></a>

## acjob任务yaml参数说明<a name="acjob"></a>

在acjob训练任务中，可使用的YAML参数说明如下表所示。

**表 1** acjob任务关键字段说明

|字段路径|类型|格式|描述|
|--|--|--|--|
|apiVersion|字符串 (string)|-|定义对象表示的版本化资源模式。服务器会转换为最新内部值，拒绝不识别的版本。更多信息请参见[Types](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds)。|
|kind|字符串 (string)|-|表示此对象对应的REST资源类型。值通过端点推断，不可更新，采用驼峰命名。更多信息请参见[Resources](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources)。|
|metadata|对象 (object)|-|Kubernetes元数据（如命名空间、标签等）。更多信息请参见[Metadata](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata)。|
|metadata.labels.app|字符串 (string)|-|<p>表明MindIE Motor任务在Ascend Job中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。</p><ul id="ul139591420161415"><li>acjob的任务YAML同时包含jobID和app这2个字段时，<span id="zh-cn_topic_0000001951418201_ph1566531814589"><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a>Ascend Operator</span>组件会自动传入环境变量MINDX_TASK_ID、APP_TYPE、MINDX_SERVER_IP及MINDX_SERVER_DOMAIN，并将其标识为MindIE推理任务。</li><li>关于以上环境变量的详细说明请参见<a href="./environment_variable_description.md#ascend-operator环境变量说明">Ascend Operator注入的训练环境变量</a>。</li><li>该参数仅支持在<span id="ph1493312176292"><a name="ph1493312176292"></a><a name="ph1493312176292"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph1893331752914"><a name="ph1893331752914"></a><a name="ph1893331752914"></a>Atlas 800I A2 推理服务器</span>上使用。</li></ul>
| metadata.labels.mind-cluster/scaling-rule: scaling-rule  | 字符串 (string) |  | 标记扩缩容规则对应的ConfigMap名称。 仅支持MindIE Motor推理任务在Atlas 800I A3 超节点服务器和Atlas 800I A2 推理服务器上使用本参数。  |
|metadata.labels.mind-cluster/group-name: group0  |字符串 (string) | | 标记扩缩容规则中对应的group名称。  仅支持MindIE Motor推理任务在Atlas 800I A3 超节点服务器和Atlas 800I A2 推理服务器上使用本参数。  |
|metadata.labels.framework|字符串 (string)|-|AI框架类型，取值为pytorch或mindspore。|
|metadata.labels.jobID|字符串 (string)|-|当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置。该参数仅支持在Atlas 800I A3 超节点服务器和Atlas 800I A2 推理服务器上使用。|
|metadata.labels.pod-rescheduling  | ||<p>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p><ul><li>on：开启Pod级别重调度 </li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul><div class="note" id="zh-cn_topic_0000002039339953_note1430334413223"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><ul id="zh-cn_topic_0000002039339953_ul461013147314"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li></ul></div></div>
|metadata.labels.process-recover-enable|字符串 (string)|-|<p>Ascend Operator会根据用户配置的recover-strategy自动给任务打上process-recover-enable=on标签，无需用户手动指定。</p><ul><li>on：开启进程级别重调度及进程级在线恢复。<p>进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过Job级别重调度恢复训练。</p></li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>|
|metadata.annotations.recover-strategy|字符串 (string)|-|<p>任务可用恢复策略。recover-strategy配置在任务YAML的annotations下，取值为6种策略的随意组合，策略之间由逗号分割。</p><ul><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>|
|metadata.labels.subHealthyStrategy|字符串 (string)|-|<p>节点状态为亚健康（SubHealthy）的节点的处理策略。</p><ul><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li><li>默认取值为ignore。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用graceExit策略时，需保证任务开启了临终CKPT保存功能。</li><li>hotSwitch策略的使用约束请参见<a href="../usage/resumable_training/01_solutions_principles.md#亚健康热切">使用约束</a>。</li></ul></div></div>|
|metadata.labels.fault-scheduling|字符串 (string)|-|<ul><li>grace：配置任务采用优雅删除模式，并在过程中先优雅删除原Pod，15分钟后若还未成功，使用强制删除原Pod。进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</li><li>force：配置任务采用强制删除模式，在过程中强制删除原Pod。</li><li>off：该任务不使用断点续训特性，K8s的maxRetry仍然生效。</li><li>无（无fault-scheduling字段）：该任务不使用断点续训特性，K8s的maxRetry仍然生效。</li><li>其他值：该任务不使用断点续训特性，K8s的maxRetry仍然生效。</li></ul>|
|metadata.labels.fault-retry-times|整数 (integer)|int32|<p>处理业务面故障，必须配置业务面可无条件重试的次数。</p><ul><li>0 &lt; fault-retry-times：处理业务面故障，必须配置业务面可无条件重试的次数。</li><li>无（无fault-retry-times）或0：该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅Atlas 800T A2 训练服务器和Atlas 900 A2 PoD 集群基础单元支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul></div></div>|
|metadata.labels.ring-controller.atlas|字符串 (string)|-|用于区分任务使用的芯片的类型。<ul><li>Atlas A2 训练系列产品、A200T A3 Box8 超节点服务器、Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器取值为：ascend-{xxx}b</li><li>Atlas 800 训练服务器，服务器（插Atlas 300T 训练卡）取值为：ascend-910</li><li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD取值为：ascend-npu</li></ul>|
|metadata.labels.podgroup-sched-enable|字符串 (string)|-|<p>仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。</p><ul><li>取值配置为字符串"true"时，表示开启批量调度功能。</li><li>取值配置为其他字符串时，表示批量调度功能不生效，使用普通调度。</li></ul><p>若不配置该参数，表示批量调度功能不生效，使用普通调度。</p><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>该参数只支持使用Volcano调度器的整卡调度特性。</li><li>仅支持在Atlas 900 A3 SuperPoD 超节点和Atlas 800T A3 超节点服务器中使用本参数。</li></ul></div>
|metadata.labels.tor-affinity|字符串 (string)|-|<p>默认值为null，表示不使用交换机亲和性调度。</p><ul><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度</li></ul><span class="notetitle">[!NOTE] 说明</span><div class="notebody">用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</div><p>用户需要根据任务类型进行配置。</p><ul id="ul59831535122714"><li>交换机亲和性调度1.0版本支持<span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span>和<span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_3"><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>|
|metadata.annotations['sp-block']|字符串 (string)|-|<p>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。若用户未指定该字段，调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</p><ul><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul><p>了解详细说明请参见<a href="../usage/basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>仅支持在Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器、Atlas 800I A3 超节点服务器中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="https://gitcode.com/Ascend/mind-cluster/issues/377">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul></div>|
|metadata.annotations.huawei.com/schedule_policy|字符串 (string)|-|配置任务需要调度的AI芯片布局形态。Volcano会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。目前支持<a href="#schedule_policy">huawei.com/schedule_policy配置说明</a>中的配置。|
|huawei.com/affinity-config|字符串 (string)|-|<p>level1=x,level2=y,...</p><p>其中x,y...为对应的网络层级子任务大小。</p><p>配置任务的多级调度的亲和性层级。</p><p>要求满足格式为leveli=ni样式的字符串的拼接，中间使用英文逗号分隔。其中，i为网络层级序号，ni为该网络层级子任务的副本数量。例如，对于总副本数量为8的任务“level1=2,level2=4”，表示任务Pod中每2个Pod分配到有相同level1标签的节点上，每4个Pod分配到有相同level2标签的节点上。</p><p>网络层级配置需要满足以下要求：<ul><li>任务层级大于1层时，层级n的值必须是n-1的整数倍。</li><li>任务总副本数量必须是所有层级的整数倍。</li><li>任务层级配置必须从level1开始，从小到大连续的。</li></ul></p>|
|spec|对象 (object)|-|AscendJob期望状态的规格描述。必填字段：replicaSpecs。|
|spec.template.metadata.annotations.huawei.com/recover_policy_path|字符串 (string)|-|任务重调度策略。当取值为 pod 则只支持Pod级重调度，不升级为Job级别。当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask|
|spec.template.metadata.annotations.huawei.com/schedule_minAvailable|整数|-|默认值为任务总副本数。Ascend Operator启用“gang”调度生效，且调度器为Volcano时，任务运行总副本数。|
|metadata.annotations.wait-reschedule-timeout|整数 (integer)|int32|<p>30~270</p><p>进程级别重调度处理时等待故障节点重调度的超时时间，单位为秒，默认值为270。</p>|
|spec.replicaSpecs|对象 (object)|-|ReplicaType到ReplicaSpec的映射，指定MS集群配置。示例：{ "Scheduler": ReplicaSpec, "Worker": ReplicaSpec }。|
|spec.replicaSpecs.[ReplicaType]|对象 (object)|-|副本的描述。|
|spec.replicaSpecs.[ReplicaType].replicas|整数 (integer)|int32|副本数量，表示给定模板所需的副本数。默认为1。|
|spec.replicaSpecs.[ReplicaType].restartPolicy|字符串 (string)|-|。<p>容器重启策略，默认为Never。当配置业务面故障无条件重试时，容器重启策略取值必须为"Never"。</p><ul><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>vcjob类型的训练任务不支持ExitCode。</p></div></div>|
|spec.replicaSpecs.[ReplicaType].template|对象 (object)|-|Kubernetes Pod模板，更多信息请参见[Kubernetes Pod模板](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-template-v1/)。|
|spec.replicaSpecs.[ReplicaType].template.spec.hostNetwork|字符串 (string)|-|<ul><li>true：使用HostIP创建Pod。此种情况下，需要在YAML中同步配置环境变量HCCL_IF_IP为status.hostIP。当集群规模较大（节点数量>1000时），推荐使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。未传入此参数或此参数的值为false时，不需要配置上述环境变量。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody">当采用HostIp方式创建Pod，依然存在创建Pod速度慢且Pod之间通信速度慢的问题。此时推荐采用挂载RankTable文件的方式，通过解析RankTable文件获得Pod的hostIP，并将其注入到对应框架任务的环境变量中（如ms框架注入到环境变量MS_SCHED_HOST中），实现建链。</div>|
|spec.replicaSpecs.[ReplicaType].template.spec.nodeSelector.host-arch|字符串 (string)|-|<p>需要运行训练任务的节点架构，请根据实际修改。分布式任务中，请确保运行训练任务的节点架构相同。</p><p>Atlas 200I SoC A1 核心板节点仅支持huawei-arm。</p><ul><li>ARM环境：huawei-arm</li><li>x86_64环境：huawei-x86</li></ul>|
|spec.replicaSpecs.[ReplicaType].template.spec.nodeSelector.accelerator-type|字符串 (string)|-|<p>根据所使用芯片类型不同，取值如下：</p><ul><li>Atlas 800 训练服务器（NPU满配）：module</li><li>Atlas 800 训练服务器（NPU半配）：half</li><li>服务器（插Atlas 300T 训练卡）：card</li><li>Atlas 800T A2 训练服务器、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件、Atlas 800I A3 超节点服务器和Atlas 900 A2 PoD 集群基础单元：module-{xxx}b-8</li><li>Atlas 200T A2 Box16 异构子框：module-{xxx}b-16</li><li>A200T A3 Box8 超节点服务器：module-a3-16</li><li>（可选）Atlas 800 训练服务器（NPU满配）可以省略该标签。</li><li>Atlas 900 A3 SuperPoD 超节点：module-a3-16-super-pod</li><li>（可选）Atlas 350 标卡：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li></ul><p>根据需要运行训练任务的节点类型，选取不同的值。如果节点是Atlas 800 训练服务器（NPU满配），可以省略该标签。对于Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD，若使用pingmesh功能则此标签为必选。</p><span class="notetitle">[!NOTE] 说明</span><div class="notebody">下文的{<em>xxx</em>}即取“910”字符作为芯片型号数值。</div>|
|spec.replicaSpecs.[ReplicaType].template.spec.containers.name|字符串 (string)|-|容器名称，当前必须为ascend。|
|spec.replicaSpecs.[ReplicaType].template.spec.containers.image|字符串 (string)|-|训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。|
|spec.replicaSpecs.[ReplicaType].template.spec.containers.ports|对象 (object)|-|分布式训练集合通信端口。“name”取值只能为“ascendjob-port”，“containerPort”用户可根据实际情况设置，若未进行设置则采用默认端口2222。|
|<ul><li>spec.replicaSpecs.[ReplicaType].template.spec.containers.resources.requests</li><li>spec.replicaSpecs.[ReplicaType].template.spec.containers.resources.limits</li></ul>|对象 (object)|-|<p>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。limits需要和requests的芯片名称和数量需保持一致。</p><p><strong>整卡调度：</strong></p><ul><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD ：<ul><li>配置为 huawei.com/npu: <em>x</em></li></ul></li><li>推理服务器（插Atlas 300I 推理卡）：<ul><li>配置为 huawei.com/Ascend310: <em>x</em></li></ul></li><li>Atlas 推理系列产品非混插模式：<ul><li>配置为 huawei.com/Ascend310P: <em>x</em></li></ul></li><li>Atlas 推理系列产品混插模式：<ul><li>配置为 huawei.com/Ascend310P-V: <em>x</em></li><li>配置为 huawei.com/Ascend310P-VPro: <em>x</em></li><li>配置为 huawei.com/Ascend310P-IPro: <em>x</em></li></ul></li><li>其他产品配置为 huawei.com/Ascend910: <em>x</em></li></ul><p>根据所使用芯片类型不同，x取值如下：</p><ul><li>Atlas 800 训练服务器（NPU满配）：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul></li><li>Atlas 800 训练服务器（NPU半配）：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul></li><li>服务器（插Atlas 300T 训练卡）：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul></li><li>Atlas 800T A2 训练服务器和Atlas 900 A2 PoD 集群基础单元：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul></li><li>Atlas 200T A2 Box16 异构子框：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul></li><li>Atlas 900 A3 SuperPoD 超节点、A200T A3 Box8 超节点服务器、Atlas 800T A3 超节点服务器：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对Atlas 900 A3 SuperPoD 超节点的逻辑超节点亲和任务：16</li></ul></li><li>Atlas 350 标卡（无互联节点内8卡）：<ul><li>单机：1、2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul></li><li>Atlas 350 标卡（无互联节点内16卡）：<ul><li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li><li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li></ul></li><li>Atlas 350 标卡（4P mesh 8卡）：<ul><li>单机（满足亲和性）：1、2、3、4、8</li><li>单机（不保证亲和性）：5、6、7</li><li>分布式（满足亲和性）：1、2、3、4、8</li><li>分布式（不保证亲和性）：5、6、7</li></ul></li><li>Atlas 350 标卡（4P mesh 16卡）：<ul><li>单机（满足亲和性）：1、2、3、4、8、12、16</li><li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li><li>分布式（满足亲和性）：1、2、3、4、8、12、16</li><li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li></ul></li><li>Atlas 850 系列硬件产品（普通集群）：<ul><li>单机：1、2、4、8</li><li>分布式：1、2、4、8</li></ul></li><li>Atlas 850 系列硬件产品（超节点集群）：<ul><li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul></li><li>Atlas 950 SuperPoD 集群：<ul><li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul></li></ul><p><strong>静态vNPU调度：</strong></p><p>huawei.com/Ascend910-Y: 1</p><p>取值为1。只能使用一个NPU下的vNPU。</p><p>如huawei.com/Ascend910-6c.1cpu.16g: 1</p>|
|spec.replicaSpecs.{Master\|Scheduler\|Worker}.template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath|字符串 (string)|-|<p>取值为metadata.annotations['huawei.com/AscendXXX']，其中XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p><p>Ascend Docker Runtime会获取该参数值，用于给容器挂载相应类型的NPU。</p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>该参数只支持使用Volcano调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</li></ul></div></div>|
|spec.replicaSpecs.{Master\|Scheduler\|Worker}.template.spec.terminationGracePeriodSeconds|整数 (integer)|<p>0 &lt; terminationGracePeriodSeconds &lt; <strong id="zh-cn_topic_0000002039339953_b09468052417"><a name="zh-cn_topic_0000002039339953_b09468052417"></a><a name="zh-cn_topic_0000002039339953_b09468052417"></a>grace-over-time</strong>参数取值</p>|<p>容器收到SIGTERM到被K8s强制停止经历的时间，该时间需要大于0且小于volcano-v<em>{version}</em>.yaml文件中"<strong>grace-over-time</strong>"参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参见K8s官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/">容器生命周期回调</a>。</p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p></div></div>|
|spec.runPolicy|对象 (object)|-|封装分布式训练作业的运行时策略（如资源清理、活动时间）。|
|spec.runPolicy.backoffLimit|整数 (integer)|int32|作业失败前允许的重试次数（可选）。<ul><li>0 &lt; backoffLimit：任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与backoffLimit取值相同时，任务将不再进行重调度。</li><li>无（无backoffLimit）或backoffLimit ≤ 0：不限制总重调度次数。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>同时配置了backoffLimit和fault-retry-times参数时，当已经重调度次数与backoffLimit或fault-retry-times取值有一个相同时，将不再进行重调度。</p><p>若不配置backoffLimit，但是配置了fault-retry-times参数，则使用fault-retry-times的重调度次数。</p></div></div>|
|spec.runPolicy.activeDeadlineSeconds|整数 (integer)|int64|作业保持活动的最长时间（秒），值必须为正整数。当前无意义，后续版本将会删除。|
|spec.runPolicy.cleanPodPolicy|字符串 (string)|-|作业完成后清理Pod的策略。默认值为Running。当前无意义，后续版本将会删除。|
|spec.runPolicy.ttlSecondsAfterFinished|整数 (integer)|int32|作业完成后的TTL（生存时间）。默认为无限，实际删除可能延迟。当前无意义，后续版本将会删除。|
|spec.runPolicy.schedulingPolicy|对象 (object)|-|调度策略（如gang-scheduling）。|
|spec.runPolicy.schedulingPolicy.minAvailable|整数 (integer)|int32|最小可用资源数，默认值为任务总副本数。Ascend Operator启用"gang"调度生效，且调度器为Volcano时，任务运行总副本数。|
|spec.runPolicy.schedulingPolicy.minResources|对象 (object)|-|按资源名称分配的最小资源集合（支持整数或字符串格式）。|
|spec.runPolicy.schedulingPolicy.priorityClass|字符串 (string)|-|优先级类名称。|
|spec.runPolicy.schedulingPolicy.queue|字符串 (string)|-|调度队列名称。默认值为“default”，用户需根据自身情况填写。Ascend Operator启用“gang”调度生效，且调度器为Volcano时，任务所属队列。|
|spec.schedulerName|字符串 (string)|-|Ascend Operator启用"gang"调度时所选择的调度器。默认值为"volcano"，用户需根据自身情况填写。|
|spec.successPolicy|字符串 (string)|-|标记AscendJob成功的标准，当前无意义，仅当所有Pod成功时，才会判定任务成功。后续版本将会删除。|
|status|对象 (object)|-|AscendJob的最新观察状态（只读）。必填字段：conditions、replicaStatuses。|
|status.completionTime|字符串 (string)|date-time|作业完成时间（RFC3339格式，UTC）。|
|status.conditions|数组 (array)|-|当前作业条件数组。|
|status.conditions[type]|字符串 (string)|-|作业条件的类型（如 "Complete"）。|
|status.conditions[status]|字符串 (string)|-|条件状态：True、False、Unknown。|
|status.conditions[lastTransitionTime]|字符串 (string)|date-time|条件状态转换的时间。|
|status.conditions[lastUpdateTime]|字符串 (string)|date-time|条件更新后的最终时间。|
|status.conditions[message]|字符串 (string)|-|条件的详细描述。|
|status.conditions[reason]|字符串 (string)|-|条件转换的原因。|
|status.lastReconcileTime|字符串 (string)|date-time|作业最后一次调和的时间（RFC3339格式，UTC）。|
|status.replicaStatuses|对象 (object)|-|副本类型到副本状态的映射。|
|status.replicaStatuses.[ReplicaType].active|整数 (integer)|int32|正在运行的Pod数量。|
|status.replicaStatuses.[ReplicaType].failed|整数 (integer)|int32|已失败的Pod数量。|
|status.replicaStatuses.[ReplicaType].succeeded|整数 (integer)|int32|已成功的Pod数量。|
|status.replicaStatuses.[ReplicaType].labelSelector|对象 (object)|-|Pod标签选择器（定义如何筛选Pod）。|
|status.replicaStatuses.[ReplicaType].labelSelector.matchExpressions|数组 (array)|-|标签匹配规则（支持In、NotIn、Exists、DoesNotExist等操作符）。|
|status.replicaStatuses.[ReplicaType].labelSelector.matchLabels|对象 (object)|-|标签匹配的键值对（等价于matchExpressions条件）。|
|status.startTime|字符串 (string)|date-time|作业开始时间（RFC3339格式，UTC）。|
|metadata.annotations['huawei.com/AscendXXX']|字符串 (string)||XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。Ascend Docker Runtime会获取该参数值，用于给容器挂载相应类型的NPU。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>该参数只支持使用Volcano调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</li></ul></div></div>
|huawei.com/Ascend910|数字||请求的NPU数量，请根据实际修改。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</li></ul></div></div><span>Atlas 800 训练服务器（NPU满配）</span>：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul><span>Atlas 800训练服务器（NPU半配）</span>：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul><span>服务器（插Atlas 300T训练卡</span>）：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul><span>Atlas 800T A2训练服务器和Atlas 900 A2 PoD集群基础单元</span>：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul><span>Atlas 200T A2 Box16 异构子框和Atlas 200I A2 Box16 异构子框</span>：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul><span>Atlas 900 A3 SuperPoD 超节点、A200T A3 Box8 超节点服务器、Atlas 800T A3 超节点服务器</span>：<ul><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对Atlas 900 A3 SuperPoD 超节点的逻辑超节点亲和任务：16</li></ul></div><span>Atlas 350 标卡（无互联节点内8卡）</span>：<ul><li>单机：1、2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul><span>Atlas 350 标卡（无互联节点内16卡）</span>：<ul><li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li><li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li></ul><span>Atlas 350 标卡（4P mesh 8卡）</span>：<ul><li>单机（满足亲和性）：1、2、3、4、8</li><li>单机（不保证亲和性）：5、6、7</li><li>分布式（满足亲和性）：1、2、3、4、8</li><li>分布式（不保证亲和性）：5、6、7</li></ul><span>Atlas 350 标卡（4P mesh 16卡）</span>：<ul><li>单机（满足亲和性）：1、2、3、4、8、12、16</li><li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li><li>分布式（满足亲和性）：1、2、3、4、8、12、16</li><li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li></ul><span>Atlas 850 系列硬件产品（普通集群）</span>：<ul><li>单机：1、2、4、8</li><li>分布式：1、2、4、8</li></ul><span>Atlas 850 系列硬件产品（超节点集群）</span>：<ul><li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul><span>Atlas 950 SuperPoD</span>：<ul><li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul>
|super-pod-affinity|字符串 (string)||<p>仅支持在Atlas 900 A3 SuperPoD 超节点中使用本参数。超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p><ul><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
|<ul><li>customJobKey</li><li>custom-job-id</li></ul> |||<p>支持通过customJobKey或custom-job-id设置作业唯一标识符，方便用户根据该标识符过滤作业相关的告警、ISSUE等关键信息。在资源AscendJob的metadata.labels标签中设置。</p><ul><li>customJobKey：用户自定义标签，以二级跳转的方式设置作业唯一标识符，如：<p>customJobKey: tid</p><p>tid: "123456"</p></li><li>custom-job-id：用户自定义标签，直接设置作业唯一标识符，如：<p>custom-job-id："123456"</p></li></ul>
|huawei.com/scheduler.softShareDev.aicoreQuota|字符串 (string)||[1, 100]，请求的AICore百分比。|
|huawei.com/scheduler.softShareDev.hbmQuota|字符串 (string)||<p>[1, maxHBM]，请求的高带宽内存量，单位为MB。</p><p>maxHBM为通过<b>npu-smi info</b>命令查询出的HBM-Usage(MB)中HBM的值。</p>|
|huawei.com/scheduler.softShareDev.policy|字符串 (string)||<p>软切分策略，取值有：</p><ul><li>fixed-share</li><li>elastic</li><li>best-effort</li></ul>|
|podAffinity|字符串 (string)||<p>表示逻辑超节点会往具有更多亲和性Pod的物理超节点调度。</p><p>仅支持MindIE Motor推理任务Atlas 800I A3 超节点服务器上使用本参数。</p>|
|sp-fit|字符串 (string)|| <p>超节点调度策略。仅支持MindIE Motor推理任务Atlas 800I A3 超节点服务器上使用本参数。</p><ul><li>idlest：逻辑超节点会往更空闲的物理超节点调度。</li><li>非idlest：逻辑超节点会优先占满物理超节点。</li></ul>
|metadata.labels['duo']|字符串 (string)||<p>仅支持推理服务器（插Atlas 300I Duo 推理卡）的参数。</p><ul id="ul145791915920"><li>true：使用<span id="ph19427048143715"><a name="ph19427048143715"></a><a name="ph19427048143715"></a>Atlas300I Duo 推理卡</span>。</li><li>false：不使用<span id="ph1069395411377"><a name="ph1069395411377"></a><a name="ph1069395411377"></a>Atlas300I Duo 推理卡</span>。</li></ul>
|metadata.labels['npu-310-strategy']|字符串 (string)||<p>仅支持推理服务器（插Atlas 300I Duo 推理卡）的参数。</p><ul><li>card：按推理卡调度，request请求的昇腾AI处理器个数不超过2，使用同一张Atlas 300I Duo 推理卡上的昇腾AI处理器。</li><li>chip：按昇腾AI处理器调度，请求的昇腾AI处理器个数不超过单个节点的最大值。</li></ul>
|metadata.labels['distributed']|字符串 (string)||<p>仅支持推理服务器（插Atlas 300I Duo 推理卡）的参数。</p><p>是否使用分布式推理。</p><ul><li>true：使用分布式推理。使用chip模式时，必须将任务调度到整张Atlas 300I Duo 推理卡。若任务需要的昇腾AI处理器数量为单数时，使用单个昇腾AI处理器的部分，将优先调度到剩余昇腾AI处理器数量为1的Atlas 300I Duo 推理卡上。</li><li>false：使用非分布式推理。使用chip模式时，请求的昇腾AI处理器个数不超过单个节点的最大值。</li></ul><div class="note" id="note595619820324"><a name="note595619820324"></a><a name="note595619820324"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul1857014516418"></a><a name="ul1857014516418"></a><ul id="ul1857014516418"><li>无论是否为分布式推理，card模式的调度策略不变。</li><li>当distributed为true时，只支持单机多卡；当distributed为false时，只支持多机多卡。</li><li>当distributed为true时，不支持Deployment任务。</li></ul>    </div></div>

## vcjob任务yaml参数说明<a name="vcjob"></a>

在vcjob任务中，可使用的YAML参数说明如下表所示。

**表 2** vcjob任务关键字段说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>spec.minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>spec.tasks[].replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1458223119296"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p7582183112296"><a name="zh-cn_topic_0000001951418201_p7582183112296"></a><a name="zh-cn_topic_0000001951418201_p7582183112296"></a>spec.maxRetry</p>
<p id="zh-cn_topic_0000001951418201_p1758196165112"><a name="zh-cn_topic_0000001951418201_p1758196165112"></a><a name="zh-cn_topic_0000001951418201_p1758196165112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1026835111"><a name="zh-cn_topic_0000001951418201_p1026835111"></a><a name="zh-cn_topic_0000001951418201_p1026835111"></a>0&lt; maxRetry</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p2216813515"><a name="zh-cn_topic_0000001951418201_p2216813515"></a><a name="zh-cn_topic_0000001951418201_p2216813515"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与maxRetry取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note394611531302"><a name="zh-cn_topic_0000001951418201_note394611531302"></a><a name="zh-cn_topic_0000001951418201_note394611531302"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p0947553193013"><a name="zh-cn_topic_0000001951418201_p0947553193013"></a><a name="zh-cn_topic_0000001951418201_p0947553193013"></a>同时配置了maxRetry和fault-retry-times参数时，当已经重调度次数与maxRetry或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11581962517"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13882123719515"><a name="zh-cn_topic_0000001951418201_p13882123719515"></a><a name="zh-cn_topic_0000001951418201_p13882123719515"></a>无（无maxRetry）或maxRetry等于0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1637895110"><a name="zh-cn_topic_0000001951418201_p1637895110"></a><a name="zh-cn_topic_0000001951418201_p1637895110"></a>不配置maxRetry或配置maxRetry取值为0时，系统默认进行3次重调度。</p>
</td>
</tr>
<tr id="row917012162413"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p871672217415"><a name="p871672217415"></a><a name="p871672217415"></a>minReplicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p14170516144111"><a name="p14170516144111"></a><a name="p14170516144111"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p317081614417"><a name="p317081614417"></a><a name="p317081614417"></a>最小副本数，需要设置为任务需要的最小节点的数量。</p>
</td>
</tr>
<tr>
<td rowspan="2">metadata.labels.fault-scheduling</td>
<td>grace</td>
<td>配置任务采用优雅删除模式，并在过程中先优雅删除原Pod，15分钟后若还未成功，使用强制删除原Pod。进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</td>
</tr>
<tr>
<td>force</td>
<td>配置任务采用强制删除模式，在过程中强制删除原Pod。</td>
</tr>
<tr id="row128861384219"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p11288121310421"><a name="p11288121310421"></a><a name="p11288121310421"></a>metadata.labels.elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7288191354217"><a name="p7288191354217"></a><a name="p7288191354217"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1628816134422"><a name="p1628816134422"></a><a name="p1628816134422"></a>开启弹性训练。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>spec.tasks[0].template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）spec.tasks[0].template.spec.nodeSelector.host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：<span id="ph4569134274515"><a name="ph4569134274515"></a><a name="ph4569134274515"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：<span id="ph7394135434515"><a name="ph7394135434515"></a><a name="ph7394135434515"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。Atlas 200I SoC A1 核心板节点仅支持huawei-arm。</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod：只支持Pod级重调度，不升级为Job级别。（当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask）</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>目前支持<a href="#schedule_policy">huawei.com/schedule_policy配置说明</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
</td>
</tr>
<tr><td>servertype</td><td><ul><li>npu-{aicore核数}</li><li>soc</li><li>Ascend910-{aicore核数}</li><li>Ascend310P-{aicore核数}</li></ul></td><td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>服务器类型。</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>soc：调度到<span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 核心板</span>节点上，必须要加上此配置，并参考<span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>“infer-310p-1usoc.yaml”</span>文件进行目录挂载。</li><li>其他类型节点不需要此参数。</li></ul>
    </td></tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['sp-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>指定逻辑超节点芯片数量。</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>了解详细说明请参见<a href="../usage/basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器、Atlas 800I A3 超节点服务器中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="https://gitcode.com/Ascend/mind-cluster/issues/377">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul id="ul961424647"><li>交换机亲和性调度1.0版本支持<span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a>Atlas 训练系列产品</span>和<span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span>和<span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>spec.tasks[0].template.spec.nodeSelector.accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a><ul id="ul14200073713"><li><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001609074269_ph1284164912438"><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph4528511506"><a name="zh-cn_topic_0000001609074269_ph4528511506"></a><a name="zh-cn_topic_0000001609074269_ph4528511506"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph486033685311"><a name="ph486033685311"></a><a name="ph486033685311"></a>Atlas 800T A2 训练服务器</span>和<span id="ph1296712308221"><a name="ph1296712308221"></a><a name="ph1296712308221"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-8</li><li><span id="ph1114211211203"><a name="ph1114211211203"></a><a name="ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph5811017182112"><a name="ph5811017182112"></a><a name="ph5811017182112"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b-16</li><li><span id="ph115277505269"><a name="ph115277505269"></a><a name="ph115277505269"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li>（可选）Atlas 350 标卡：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1954213851616"><a name="p1954213851616"></a><a name="p1954213851616"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note19666163011214"><a name="note19666163011214"></a><a name="note19666163011214"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1105153313533"><a name="p1105153313533"></a><a name="p1105153313533"></a><span id="ph710573305319"><a name="ph710573305319"></a><a name="ph710573305319"></a>下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209_1"><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>spec.tasks[0].template.spec.containers[0].resources.requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>整卡调度：</strong></p>
<ul><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD ：<ul><li>配置为huawei.com/npu: <em>x</em></li></ul></li><li>推理服务器（插Atlas 300I 推理卡）：<ul><li>配置为huawei.com/Ascend310: <em>x</em></li></ul></li><li>Atlas 推理系列产品非混插模式：<ul><li>配置为huawei.com/Ascend310P: <em>x</em></li></ul></li><li>Atlas 推理系列产品混插模式：<ul><li>配置为huawei.com/Ascend310P-V: <em>x</em></li><li>配置为huawei.com/Ascend310P-VPro: <em>x</em></li><li>配置为huawei.com/Ascend310P-IPro: <em>x</em></li></ul></li><li>其他产品配置为huawei.com/Ascend910: <em>x</em></li></ul>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T 训练卡</span>）：<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>单机单芯片：1</li><li>单机多芯片：2</li><li>分布式：2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 训练服务器</span>和<span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>单机多芯片：2、4、6、8、10、12、14、16</li><li>分布式：16</li></ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内8卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8</li>
        <li>分布式：1、2、3、4、5、6、7、8</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内16卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
        <li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 8卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8</li>
        <li>单机（不保证亲和性）：5、6、7</li>
        <li>分布式（满足亲和性）：1、2、3、4、8</li>
        <li>分布式（不保证亲和性）：5、6、7</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 16卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
        <li>分布式（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（普通集群）</span>：
    <ul>
        <li>单机：1、2、4、8</li>
        <li>分布式：1、2、4、8</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（超节点集群）</span>：
    <ul>
        <li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
</ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>静态vNPU调度：</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>如huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<ul id="ul10782193418818"><li>仅<span id="ph1038285416813"><a name="ph1038285416813"></a><a name="ph1038285416813"></a>Atlas 推理系列产品</span>非混插模式支持静态vNPU调度。</li><li>推理服务器（插<span id="ph1990710374611"><a name="ph1990710374611"></a><a name="ph1990710374611"></a>Atlas 300I 推理卡</span>）和<span id="ph629210161695"><a name="ph629210161695"></a><a name="ph629210161695"></a>Atlas 推理系列产品</span>混插模式不支持静态vNPU调度。</li><li><strong id="b179331118122318"><a name="b179331118122318"></a><a name="b179331118122318"></a><em id="i14933131862318"><a name="i14933131862318"></a><a name="i14933131862318"></a>Y</em></strong>取值可参考<a href="../usage/virtual_instance/virtual_instance_with_hdk/static_vnpu_scheduling/02_mounting_vnpu_static.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与vNPU类型关系表的对应产品的“vNPU类型”列。<p id="p208621211164518"><a name="p208621211164518"></a><a name="p208621211164518"></a>以vNPU类型<em id="i412654718449"><a name="i412654718449"></a><a name="i412654718449"></a>Ascend310P-4c.3cpu</em>为例，<strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong>取值为4c.3cpu，不包括前面的Ascend310P。</p>
    </li></ul>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>spec.tasks[0].template.spec.containers[0].resources.limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">[!NOTE] 说明</span>
    <div class="notebody">
        <ul>
            <li>
                <p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>该参数只支持使用<span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>调度器的整卡调度特性。使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
            </li>
            <li><p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</p></li>
        </ul>
</div>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a>spec.tasks[0].template.spec.containers[0].resources.{requests|limits}['huawei.com/Ascend910']</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li><span id="ph157984201135"><a name="ph157984201135"></a><a name="ph157984201135"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph745323894316"><a name="zh-cn_topic_0000001951418201_ph745323894316"></a><a name="zh-cn_topic_0000001951418201_ph745323894316"></a>Atlas 900 A2 PoD 集群基础单元</span><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_ph419517625020"><a name="zh-cn_topic_0000001951418201_ph419517625020"></a><a name="zh-cn_topic_0000001951418201_ph419517625020"></a>Atlas 200T A2 Box16 异构子框</span><span id="ph1891953184717"><a name="ph1891953184717"></a><a name="ph1891953184717"></a>和</span><span id="ph1149713543472"><a name="ph1149713543472"></a><a name="ph1149713543472"></a>Atlas 200I A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000001951418201_ul191955617509"></a><a name="zh-cn_topic_0000001951418201_ul191955617509"></a><ul id="zh-cn_topic_0000001951418201_ul191955617509"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内8卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8</li>
        <li>分布式：1、2、3、4、5、6、7、8</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内16卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
        <li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 8卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8</li>
        <li>单机（不保证亲和性）：5、6、7</li>
        <li>分布式（满足亲和性）：1、2、3、4、8</li>
        <li>分布式（不保证亲和性）：5、6、7</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 16卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
        <li>分布式（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（普通集群）</span>：
    <ul>
        <li>单机：1、2、4、8</li>
        <li>分布式：1、2、4、8</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（超节点集群）</span>：
    <ul>
        <li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
</ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a>请求的NPU数量，请根据实际修改，请求整卡时不能再同时请求vNPU。</p>
<div class="note" id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><span class="notetitle">[!NOTE] 说明</span>
    <div class="notebody"><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a>
        <ul id="zh-cn_topic_0000001951418201_ul54321224184319">
            <li>
                <strong id="zh-cn_topic_0000001951418201_b16213840172320"><a name="zh-cn_topic_0000001951418201_b16213840172320"></a><a name="zh-cn_topic_0000001951418201_b16213840172320"></a>优雅容错模式</strong>支持<span id="zh-cn_topic_0000001951418201_ph158146714142"><a name="zh-cn_topic_0000001951418201_ph158146714142"></a><a name="zh-cn_topic_0000001951418201_ph158146714142"></a>Atlas 800 训练服务器</span>，且资源请求数量只能为4N、8N，N为训练节点数。
            </li>
            <li>
                <strong id="zh-cn_topic_0000001951418201_b1091614581433"><a name="zh-cn_topic_0000001951418201_b1091614581433"></a><a name="zh-cn_topic_0000001951418201_b1091614581433"></a>优雅容错模式</strong>支持<span id="ph184881417142314"><a name="ph184881417142314"></a><a name="ph184881417142314"></a>Atlas 800T A2 训练服务器</span>或<span id="zh-cn_topic_0000001951418201_ph9246916444"><a name="zh-cn_topic_0000001951418201_ph9246916444"></a><a name="zh-cn_topic_0000001951418201_ph9246916444"></a>Atlas 900 A2 PoD 集群基础单元</span>，且资源请求数量只能为8N，N为训练节点数。
            </li>
            <li>
                <p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需将参数名称修改为huawei.com/npu。</p>
            </li>
        </ul>
    </div>
</div>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>{metadata, spec.tasks[0].template.metadata}.labels['ring-controller.atlas']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>推理服务器（插<span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I 推理卡</span>）：ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a>Atlas 推理系列产品</span>：ascend-310P</li><li>Atlas 800 训练服务器，服务器（插<span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 训练系列产品</term></span>、<span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li><li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD取值为：ascend-npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>此处的{<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</td>
</tr>
<tr id="rowcustomjobkey2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey2"><a name="pcustomjobkey2"></a><a name="pcustomjobkey2"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue2"><a name="pcustomjobkeyvalue2"></a><a name="pcustomjobkeyvalue2"></a>用户自定义标签，以二级跳转的方式设置作业唯一标识符，如：<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc2"><a name="pcustomjobkeydesc2"></a><a name="pcustomjobkeydesc2"></a>支持通过customJobKey或custom-job-id设置作业唯一标识符，方便用户根据该标识符过滤作业相关的告警、ISSUE等关键信息。<br> <ul><li>vcjob任务在资源Job的metadata.labels标签中设置。<br></li> <li>deploy任务在资源Deployment的spec.template.metadata.labels标签中设置。</li></ul></p>
</td>
</tr>
<tr id="rowcustomjobid2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid2"><a name="pcustomjobid2"></a><a name="pcustomjobid2"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue2"><a name="pcustomjobidvalue2"></a><a name="pcustomjobidvalue2"></a>用户自定义标签，直接设置作业唯一标识符，如：<br> custom-job-id："123456"</p>
</td>
</tr>
<tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>spec.tasks[0].template.metadata.labels['vnpu-level']</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>低配，默认值，选择最低配置的虚拟化实例模板。</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>性能优先。</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>在集群资源充足的情况下，将选择尽量高配的虚拟化实例模板；在整个集群资源已使用过多的情况下，如大部分物理NPU都已使用，每个物理NPU只剩下小部分AICore，不足以满足高配虚拟化实例模板时，将使用相同AICore数量下较低配置的其他模板。具体选择请参考<a href="../usage/virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md">虚拟化模板</a>章节。</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>spec.tasks[0].template.metadata.labels['vnpu-dvpp']</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>该<span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span>使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>该<span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span>不使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>默认值，不关注是否使用DVPP。</p>
    </td>
    </tr>
<tr id="zh-cn_topic_0000001951418201_row4635558201210"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1499116019135"><a name="zh-cn_topic_0000001951418201_p1499116019135"></a><a name="zh-cn_topic_0000001951418201_p1499116019135"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p599118017133"><a name="zh-cn_topic_0000001951418201_p599118017133"></a><a name="zh-cn_topic_0000001951418201_p599118017133"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000001951418201_ul139911803137"></a><a name="zh-cn_topic_0000001951418201_ul139911803137"></a><ul id="zh-cn_topic_0000001951418201_ul139911803137"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul169911906135"></a><a name="zh-cn_topic_0000001951418201_ul169911906135"></a>recover-strategy配置在任务YAML annotations下，取值为5种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10152132415157"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10821192541514"><a name="zh-cn_topic_0000001951418201_p10821192541514"></a><a name="zh-cn_topic_0000001951418201_p10821192541514"></a>metadata.labels.pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><ul id="zh-cn_topic_0000001951418201_ul5821162501510"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p78221125201514"><a name="zh-cn_topic_0000001951418201_p78221125201514"></a><a name="zh-cn_topic_0000001951418201_p78221125201514"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note5822925151516"><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><ul id="zh-cn_topic_0000001951418201_ul17822112517158"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row576132216324"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1772202423212"><a name="zh-cn_topic_0000001951418201_p1772202423212"></a><a name="zh-cn_topic_0000001951418201_p1772202423212"></a>metadata.labels.subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><ul id="zh-cn_topic_0000001951418201_ul972624133214"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1973102463218"><a name="zh-cn_topic_0000001951418201_p1973102463218"></a><a name="zh-cn_topic_0000001951418201_p1973102463218"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note173271703519"><a name="zh-cn_topic_0000001951418201_note173271703519"></a><a name="zh-cn_topic_0000001951418201_note173271703519"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p163271901355"><a name="zh-cn_topic_0000001951418201_p163271901355"></a><a name="zh-cn_topic_0000001951418201_p163271901355"></a>使用graceExit策略时，需保证任务开启了临终CKPT保存功能。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1314311835012"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p123205151739"><a name="zh-cn_topic_0000001951418201_p123205151739"></a><a name="zh-cn_topic_0000001951418201_p123205151739"></a>metadata.labels.fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p196969196112"><a name="zh-cn_topic_0000001951418201_p196969196112"></a><a name="zh-cn_topic_0000001951418201_p196969196112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1192310597344"><a name="zh-cn_topic_0000001951418201_p1192310597344"></a><a name="zh-cn_topic_0000001951418201_p1192310597344"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p109232597342"><a name="zh-cn_topic_0000001951418201_p109232597342"></a><a name="zh-cn_topic_0000001951418201_p109232597342"></a>处理业务面故障，必须配置业务面可无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15571815115017"><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><ul id="zh-cn_topic_0000001951418201_ul15238182410364"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph1377171612516"><a name="ph1377171612516"></a><a name="ph1377171612516"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph14104952376"><a name="zh-cn_topic_0000001951418201_ph14104952376"></a><a name="zh-cn_topic_0000001951418201_ph14104952376"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260912190502"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2966613113520"><a name="zh-cn_topic_0000001951418201_p2966613113520"></a><a name="zh-cn_topic_0000001951418201_p2966613113520"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p2096618130353"><a name="zh-cn_topic_0000001951418201_p2096618130353"></a><a name="zh-cn_topic_0000001951418201_p2096618130353"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row128551542131510"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10285161985210"><a name="zh-cn_topic_0000001951418201_p10285161985210"></a><a name="zh-cn_topic_0000001951418201_p10285161985210"></a>spec.policies</p>
<p id="zh-cn_topic_0000001951418201_p490916512164"><a name="zh-cn_topic_0000001951418201_p490916512164"></a><a name="zh-cn_topic_0000001951418201_p490916512164"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p056810252162"><a name="zh-cn_topic_0000001951418201_p056810252162"></a><a name="zh-cn_topic_0000001951418201_p056810252162"></a>event，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><ul id="zh-cn_topic_0000001951418201_ul1781384818238"><li>PodFailed：Pod失败</li><li>PodEvicted：Pod被驱逐</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p180717598243"><a name="zh-cn_topic_0000001951418201_p180717598243"></a><a name="zh-cn_topic_0000001951418201_p180717598243"></a>Pod状态。与action字段搭配使用，表示当Pod处于某种状态时，<span id="zh-cn_topic_0000001951418201_ph525518226126"><a name="zh-cn_topic_0000001951418201_ph525518226126"></a><a name="zh-cn_topic_0000001951418201_ph525518226126"></a>Volcano</span>的处理策略。默认值为PodEvicted。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1390814541612"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1590911581611"><a name="zh-cn_topic_0000001951418201_p1590911581611"></a><a name="zh-cn_topic_0000001951418201_p1590911581611"></a>action，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><ul id="zh-cn_topic_0000001951418201_ul17824133752420"><li>RestartJob：重新启动训练任务。</li><li>Ignore：<span id="zh-cn_topic_0000001951418201_ph141051824104819"><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a>忽略。开源Volcano</span>不做任何处理，由<span id="zh-cn_topic_0000001951418201_ph631119334409"><a name="zh-cn_topic_0000001951418201_ph631119334409"></a><a name="zh-cn_topic_0000001951418201_ph631119334409"></a>Ascend-volcano-plugin</span>插件进行处理。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p20698124111427"><a name="zh-cn_topic_0000001951418201_p20698124111427"></a><a name="zh-cn_topic_0000001951418201_p20698124111427"></a><span id="zh-cn_topic_0000001951418201_ph10699341154214"><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a>Volcano</span>对处于某种状态的Pod的处理策略。默认值为RestartJob。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note128691230174312"><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><ul id="zh-cn_topic_0000001951418201_ul13166894452"><li>开启Pod级别重调度需要删除policies及其子参数event和action。</li><li>使用业务面故障无条件重试时（或同时使用Pod级别重调度和业务面故障无条件重试），需要将event配置为PodFailed；action配置为Ignore。</li><li>如果不使用集群调度组件的<span id="zh-cn_topic_0000001951418201_ph8224175173014"><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a>Volcano</span>或者开源<span id="zh-cn_topic_0000001951418201_ph83286473313"><a name="zh-cn_topic_0000001951418201_ph83286473313"></a><a name="zh-cn_topic_0000001951418201_ph83286473313"></a>Volcano</span>没有集成<span id="zh-cn_topic_0000001951418201_ph617718254597"><a name="zh-cn_topic_0000001951418201_ph617718254597"></a><a name="zh-cn_topic_0000001951418201_ph617718254597"></a>Ascend-volcano-plugin</span>插件，需要参考<a href="https://gitcode.com/Ascend/mind-cluster/issues/362">使用Volcano和Ascend Operator组件场景下，业务面故障的任务所有Pod的Status全部变为Failed，任务无法触发无条件重试重调度</a>修改开源Volcano代码。</li><li>开源Volcano还提供了policies的其他取值，不建议用户修改为其他取值，否则可能影响断点续训功能的正常使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11217021145014"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1929464718814"><a name="zh-cn_topic_0000001951418201_p1929464718814"></a><a name="zh-cn_topic_0000001951418201_p1929464718814"></a>spec.tasks[0].template.spec.restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><ul id="zh-cn_topic_0000001951418201_ul193373071216"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。<div class="note" id="zh-cn_topic_0000001951418201_note278954373014"><a name="zh-cn_topic_0000001951418201_note278954373014"></a><a name="zh-cn_topic_0000001951418201_note278954373014"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p14789194311309"><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><a name="zh-cn_topic_0000001951418201_p14789194311309"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1129434710811"><a name="zh-cn_topic_0000001951418201_p1129434710811"></a><a name="zh-cn_topic_0000001951418201_p1129434710811"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue182751614652"><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a>“Never”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1116371844811"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a>spec.tasks[0].template.spec.terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a>0 &lt; terminationGracePeriodSeconds &lt;<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a> grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a>容器收到SIGTERM到被<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a>K8s</span>强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000001951418201_i1645121221719"><a name="zh-cn_topic_0000001951418201_i1645121221719"></a><a name="zh-cn_topic_0000001951418201_i1645121221719"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a>K8s</span>官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note17641176363"><a name="zh-cn_topic_0000001951418201_note17641176363"></a><a name="zh-cn_topic_0000001951418201_note17641176363"></a><div class="notebody"><p id="zh-cn_topic_0000001951418201_p97641517103616"><a name="zh-cn_topic_0000001951418201_p97641517103616"></a><a name="zh-cn_topic_0000001951418201_p97641517103616"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

## deploy任务yaml参数说明<a name="deploy"></a>

在deploy任务中，可使用的YAML参数说明如下表所示。

**表 3** deploy任务关键字段说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>spec.replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>spec.template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr>
<td rowspan="2">spec.template.metadata.labels['fault-scheduling']</td>
<td>grace</td>
<td>配置任务采用优雅删除模式，并在过程中先优雅删除原Pod，15分钟后若还未成功，使用强制删除原Pod。</td>
</tr>
<tr>
<td>force</td>
<td>配置任务采用强制删除模式，在过程中强制删除原Pod。</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）spec.template.{metadata.labels, spec.nodeSelector}['host-arch']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：<span id="ph4569134274515"><a name="ph4569134274515"></a><a name="ph4569134274515"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：<span id="ph7394135434515"><a name="ph7394135434515"></a><a name="ph7394135434515"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。Atlas 200I SoC A1 核心板节点仅支持huawei-arm。</p>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>{metadata, spec.template.metadata}.labels['ring-controller.atlas']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>推理服务器（插<span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I 推理卡</span>）：ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a>Atlas 推理系列产品</span>：ascend-310P</li><li>Atlas 800 训练服务器，服务器（插<span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 训练系列产品</term></span>、<span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li><li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD取值为：ascend-npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>此处的{<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>spec.template.metadata.labels['vnpu-level']</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>低配，默认值，选择最低配置的虚拟化实例模板。</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>性能优先。</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>在集群资源充足的情况下，将选择尽量高配的虚拟化实例模板；在整个集群资源已使用过多的情况下，如大部分物理NPU都已使用，每个物理NPU只剩下小部分AI Core，不足以满足高配虚拟化实例模板时，将使用相同AI Core数量下较低配置的其他模板。具体选择请参考<a href="../usage/virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md">虚拟化模板</a>章节。</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>spec.template.metadata.labels['vnpu-dvpp']</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>该<span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span>使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>该<span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span>不使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>默认值，不关注是否使用DVPP。</p>
    </td>
    </tr>
<tr id="row33457612918"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p137845934610"><a name="p137845934610"></a><a name="p137845934610"></a>spec.template.metadata.labels['npu-310-strategy']</p>
</td>
<td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul1967514291118"></a><a name="ul1967514291118"></a><p>仅支持推理服务器（插Atlas 300I 推理卡）的参数</p><ul id="ul1967514291118"><li>card：按推理卡调度，request请求的<span id="ph1978781173013"><a name="ph1978781173013"></a><a name="ph1978781173013"></a>昇腾AI处理器</span>个数不超过4，使用同一张<span id="ph933971152917"><a name="ph933971152917"></a><a name="ph933971152917"></a>Atlas 300I 推理卡</span>上的<span id="ph77331623132919"><a name="ph77331623132919"></a><a name="ph77331623132919"></a>昇腾AI处理器</span>。</li><li>chip：按<span id="ph14705121219305"><a name="ph14705121219305"></a><a name="ph14705121219305"></a>昇腾AI处理器</span>调度，请求的芯片个数不超过单个节点的最大值。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p2799246194816"><a name="p2799246194816"></a><a name="p2799246194816"></a>-</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod：只支持Pod级重调度，不升级为Job级别。（当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask）</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>需要写到pod的annotation或label里面。</p><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>需要写到pod的annotation或label里面。</p><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>目前支持<a href="#schedule_policy">huawei.com/schedule_policy配置说明</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>需要写到pod的annotation或label里面。</p><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
</td>
</tr>
<tr><td>spec.template.spec.nodeSelector['servertype']</td>
<td><ul><li>npu-{aicore核数}</li><li>soc</li><li>Ascend910-{aicore核数}</li><li>Ascend310P-{aicore核数}</li></ul></td>
<td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>服务器类型。</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>soc：调度到<span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 核心板</span>节点上，必须要加上此配置，并参考<span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>“infer-310p-1usoc.yaml”</span>文件进行目录挂载。</li><li>其他类型节点不需要此参数。</li></ul>
    </td></tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>spec.template.spec.nodeSelector['accelerator-type']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a><ul id="ul14200073713"><li><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001609074269_ph1284164912438"><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph4528511506"><a name="zh-cn_topic_0000001609074269_ph4528511506"></a><a name="zh-cn_topic_0000001609074269_ph4528511506"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph486033685311"><a name="ph486033685311"></a><a name="ph486033685311"></a>Atlas 800T A2 训练服务器</span>和<span id="ph1296712308221"><a name="ph1296712308221"></a><a name="ph1296712308221"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-8</li><li><span id="ph1114211211203"><a name="ph1114211211203"></a><a name="ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph5811017182112"><a name="ph5811017182112"></a><a name="ph5811017182112"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b-16</li><li><span id="ph115277505269"><a name="ph115277505269"></a><a name="ph115277505269"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li>（可选）Atlas 350 标卡：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1954213851616"><a name="p1954213851616"></a><a name="p1954213851616"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note19666163011214"><a name="note19666163011214"></a><a name="note19666163011214"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1105153313533"><a name="p1105153313533"></a><a name="p1105153313533"></a><span id="ph710573305319"><a name="ph710573305319"></a><a name="ph710573305319"></a>下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209_1"><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['sp-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>指定逻辑超节点芯片数量。</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>了解详细说明请参见<a href="../usage/basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器、Atlas 800I A3 超节点服务器中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="https://gitcode.com/Ascend/mind-cluster/issues/377">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">[!NOTE] 说明</span>
    <div class="notebody">
        <ul>
            <li>
                <p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>该参数只支持使用<span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>调度器的整卡调度特性。使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
            </li>
            <li><p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</p></li>
        </ul>
</div>
</div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul id="ul961424647"><li>交换机亲和性调度1.0版本支持<span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a>Atlas 训练系列产品</span>和<span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span>和<span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>spec.template.spec.containers[0].resources.requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>整卡调度：</strong></p>
<ul><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD ：<ul><li>配置为huawei.com/npu: <em>x</em></li></ul></li><li>推理服务器（插Atlas 300I 推理卡）：<ul><li>配置为huawei.com/Ascend310: <em>x</em></li></ul></li><li>Atlas 推理系列产品非混插模式：<ul><li>配置为huawei.com/Ascend310P: <em>x</em></li></ul></li><li>Atlas 推理系列产品混插模式：<ul><li>配置为huawei.com/Ascend310P-V: <em>x</em></li><li>配置为huawei.com/Ascend310P-VPro: <em>x</em></li><li>配置为huawei.com/Ascend310P-IPro: <em>x</em></li></ul></li><li>其他产品配置为huawei.com/Ascend910: <em>x</em></li></ul>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T 训练卡</span>）：<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>单机单芯片：1</li><li>单机多芯片：2</li><li>分布式：2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 训练服务器</span>和<span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>单机多芯片：2、4、6、8、10、12、14、16</li><li>分布式：16</li></ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内8卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8</li>
        <li>分布式：1、2、3、4、5、6、7、8</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内16卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
        <li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 8卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8</li>
        <li>单机（不保证亲和性）：5、6、7</li>
        <li>分布式（满足亲和性）：1、2、3、4、8</li>
        <li>分布式（不保证亲和性）：5、6、7</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 16卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
        <li>分布式（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（普通集群）</span>：
    <ul>
        <li>单机：1、2、4、8</li>
        <li>分布式：1、2、4、8</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（超节点集群）</span>：
    <ul>
        <li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
</ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>静态vNPU调度：</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>如huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<ul id="ul10782193418818"><li>仅<span id="ph1038285416813"><a name="ph1038285416813"></a><a name="ph1038285416813"></a>Atlas 推理系列产品</span>非混插模式支持静态vNPU调度。</li><li>推理服务器（插<span id="ph1990710374611"><a name="ph1990710374611"></a><a name="ph1990710374611"></a>Atlas 300I 推理卡</span>）和<span id="ph629210161695"><a name="ph629210161695"></a><a name="ph629210161695"></a>Atlas 推理系列产品</span>混插模式不支持静态vNPU调度。</li><li><strong id="b179331118122318"><a name="b179331118122318"></a><a name="b179331118122318"></a><em id="i14933131862318"><a name="i14933131862318"></a><a name="i14933131862318"></a>Y</em></strong>取值可参考<a href="../usage/virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的对应产品的“vNPU类型”列。<p id="p208621211164518"><a name="p208621211164518"></a><a name="p208621211164518"></a>以vNPU类型<em id="i412654718449"><a name="i412654718449"></a><a name="i412654718449"></a>Ascend310P-4c.3cpu</em>为例，<strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong>取值为4c.3cpu，不包括前面的Ascend310P。</p>
    </li></ul>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>spec.template.spec.containers[0].resources.limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</td>
</tr>
<tr id="rowcustomjobkey2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey2"><a name="pcustomjobkey2"></a><a name="pcustomjobkey2"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue2"><a name="pcustomjobkeyvalue2"></a><a name="pcustomjobkeyvalue2"></a>用户自定义标签，以二级跳转的方式设置作业唯一标识符，如：<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc2"><a name="pcustomjobkeydesc2"></a><a name="pcustomjobkeydesc2"></a>支持通过customJobKey或custom-job-id设置作业唯一标识符，方便用户根据该标识符过滤作业相关的告警、ISSUE等关键信息。<br> <ul><li>vcjob任务在资源Job的metadata.labels标签中设置。<br></li> <li>deploy任务在资源Deployment的spec.template.metadata.labels标签中设置。</li></ul></p>
</td>
</tr>
<tr id="rowcustomjobid2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid2"><a name="pcustomjobid2"></a><a name="pcustomjobid2"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue2"><a name="pcustomjobidvalue2"></a><a name="pcustomjobidvalue2"></a>用户自定义标签，直接设置作业唯一标识符，如：<br> custom-job-id："123456"</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row4635558201210"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1499116019135"><a name="zh-cn_topic_0000001951418201_p1499116019135"></a><a name="zh-cn_topic_0000001951418201_p1499116019135"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p599118017133"><a name="zh-cn_topic_0000001951418201_p599118017133"></a><a name="zh-cn_topic_0000001951418201_p599118017133"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000001951418201_ul139911803137"></a><a name="zh-cn_topic_0000001951418201_ul139911803137"></a><ul id="zh-cn_topic_0000001951418201_ul139911803137"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul169911906135"></a><a name="zh-cn_topic_0000001951418201_ul169911906135"></a>recover-strategy配置在任务YAML annotations下，取值为5种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10152132415157"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10821192541514"><a name="zh-cn_topic_0000001951418201_p10821192541514"></a><a name="zh-cn_topic_0000001951418201_p10821192541514"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><ul id="zh-cn_topic_0000001951418201_ul5821162501510"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p78221125201514"><a name="zh-cn_topic_0000001951418201_p78221125201514"></a><a name="zh-cn_topic_0000001951418201_p78221125201514"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note5822925151516"><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><ul id="zh-cn_topic_0000001951418201_ul17822112517158"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row576132216324"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1772202423212"><a name="zh-cn_topic_0000001951418201_p1772202423212"></a><a name="zh-cn_topic_0000001951418201_p1772202423212"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><ul id="zh-cn_topic_0000001951418201_ul972624133214"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1973102463218"><a name="zh-cn_topic_0000001951418201_p1973102463218"></a><a name="zh-cn_topic_0000001951418201_p1973102463218"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note173271703519"><a name="zh-cn_topic_0000001951418201_note173271703519"></a><a name="zh-cn_topic_0000001951418201_note173271703519"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p163271901355"><a name="zh-cn_topic_0000001951418201_p163271901355"></a><a name="zh-cn_topic_0000001951418201_p163271901355"></a>使用graceExit策略时，需保证任务开启了临终CKPT保存功能。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1314311835012"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p123205151739"><a name="zh-cn_topic_0000001951418201_p123205151739"></a><a name="zh-cn_topic_0000001951418201_p123205151739"></a>fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p196969196112"><a name="zh-cn_topic_0000001951418201_p196969196112"></a><a name="zh-cn_topic_0000001951418201_p196969196112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1192310597344"><a name="zh-cn_topic_0000001951418201_p1192310597344"></a><a name="zh-cn_topic_0000001951418201_p1192310597344"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p109232597342"><a name="zh-cn_topic_0000001951418201_p109232597342"></a><a name="zh-cn_topic_0000001951418201_p109232597342"></a>处理业务面故障，必须配置业务面可无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15571815115017"><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><ul id="zh-cn_topic_0000001951418201_ul15238182410364"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph1377171612516"><a name="ph1377171612516"></a><a name="ph1377171612516"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph14104952376"><a name="zh-cn_topic_0000001951418201_ph14104952376"></a><a name="zh-cn_topic_0000001951418201_ph14104952376"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260912190502"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2966613113520"><a name="zh-cn_topic_0000001951418201_p2966613113520"></a><a name="zh-cn_topic_0000001951418201_p2966613113520"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p2096618130353"><a name="zh-cn_topic_0000001951418201_p2096618130353"></a><a name="zh-cn_topic_0000001951418201_p2096618130353"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11217021145014"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1929464718814"><a name="zh-cn_topic_0000001951418201_p1929464718814"></a><a name="zh-cn_topic_0000001951418201_p1929464718814"></a>restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><ul id="zh-cn_topic_0000001951418201_ul193373071216"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。<div class="note" id="zh-cn_topic_0000001951418201_note278954373014"><a name="zh-cn_topic_0000001951418201_note278954373014"></a><a name="zh-cn_topic_0000001951418201_note278954373014"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p14789194311309"><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><a name="zh-cn_topic_0000001951418201_p14789194311309"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1129434710811"><a name="zh-cn_topic_0000001951418201_p1129434710811"></a><a name="zh-cn_topic_0000001951418201_p1129434710811"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue182751614652"><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a>“Never”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1116371844811"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a>terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a>0 &lt; terminationGracePeriodSeconds &lt;<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a> grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a>容器收到SIGTERM到被<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a>K8s</span>强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000001951418201_i1645121221719"><a name="zh-cn_topic_0000001951418201_i1645121221719"></a><a name="zh-cn_topic_0000001951418201_i1645121221719"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a>K8s</span>官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note17641176363"><a name="zh-cn_topic_0000001951418201_note17641176363"></a><a name="zh-cn_topic_0000001951418201_note17641176363"></a><div class="notebody"><p id="zh-cn_topic_0000001951418201_p97641517103616"><a name="zh-cn_topic_0000001951418201_p97641517103616"></a><a name="zh-cn_topic_0000001951418201_p97641517103616"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
</tbody>
</table>


## 其他参数说明

### huawei.com/schedule_policy配置说明<a name="schedule_policy"></a>

**表 4**  huawei.com/schedule\_policy配置说明

|配置|说明|
|--|--|
|chip4-node8|1个节点8张芯片，每4个芯片形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的整模块场景/Atlas 350 标卡共8张卡，每4张卡通过UB扣板连接。|
|chip1-node2|1个节点2张芯片。例如，Atlas 300T 训练卡的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。|
|chip4-node4|1个节点4张芯片，形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的半配场景。|
|chip8-node8|1个节点8张卡，8张卡都在1个互联环上。例如，Atlas 800T A2 训练服务器 /Atlas 850 系列硬件产品。|
|chip8-node16|1个节点16张卡，每8张卡在1个互联环上。例如，Atlas 200T A2 Box16 异构子框。|
|chip2-node8|1个节点8张卡，每2张卡在1个互联环上。|
|chip2-node16|1个节点16张卡，每2张卡在1个互联环上。例如，Atlas 800T A3 超节点服务器。|
|chip2-node8-sp|1个节点8张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 9000 A3 SuperPoD 集群算力系统。|
|chip2-node16-sp|1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 900 A3 SuperPoD 超节点。|
|chip4-node16|1个节点16张卡，每4张卡都在1个互联环上。例如，Atlas 350 标卡共16张卡，每4张卡通过UB扣板连接。|
|chip1-node8|1个节点8张卡，每张卡之间无互联。例如，Atlas 350 标卡共8张卡，每张卡之间无互联。|
|chip1-node16|1个节点16张卡，每张卡之间无互联。例如，Atlas 350 标卡共16张卡，每张卡之间无互联。|
|chip8-node8-sp|1个节点8张卡，8张卡都在1个互联环上，多个服务器形成超节点。例如，Atlas 850 系列硬件产品（超节点服务器）。|
|chip8-node8-ra64-sp|1个节点8张卡，8张卡都在1个互联环上，64个节点组成一个计算框，多个框形成超节点。例如，Atlas 950 SuperPoD。|
|chip1-softShareDev|软切分虚拟化专用调度策略。|
|multilevel|多级调度场景使用，多级调度的详细使用方法请参见[多级调度](../usage/basic_scheduling/05_multi_level_scheduling.md)。|
