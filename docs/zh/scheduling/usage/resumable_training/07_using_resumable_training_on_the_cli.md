# 通过命令行使用<a name="ZH-CN_TOPIC_0000002479386546"></a>

## （可选）配置组件<a name="ZH-CN_TOPIC_0000002511346449"></a>

如果用户在安装Ascend Device Plugin和NodeD时，已经配置了断点续训相关功能，则可以跳过本章节；若没有配置，则需要对[Ascend Device Plugin](#section14208511958)和[NodeD](#section162092113510)进行相关配置。

**配置Ascend Device Plugin<a name="section14208511958"></a>**

只支持以容器化方式启动Ascend Device Plugin。

1. 根据所使用的故障处理模式，修改Ascend Device Plugin组件的启动YAML，修改如下所示加粗部分。
    1. 重调度模式

        >[!NOTE] 
        >在重调度模式下，Ascend Device Plugin的异常也会触发故障重调度。

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin  
                         -useAscendDocker=true 
                         <strong>-volcanoType=true                    # 重调度场景下必须使用Volcano</strong>
                         <strong>-autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品</strong>
                         <strong>-listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒</strong>
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

    2. （可选）优雅容错模式：在重调度配置的基础上，新增“-hotReset”字段。

        >[!NOTE] 
        >- 优雅容错功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。
        >- “-hotReset”字段取值为1对应的功能已经日落。

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin  
                         -useAscendDocker=true 
                         -volcanoType=true                    # 重调度场景下必须使用Volcano
                         -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                         <strong>-hotReset=1 # 开启优雅容错模式，系统会尝试自动复位故障芯片</strong>
                         -listWatchPeriod=5                   # 健康状态检查周期，范围[3,1800]；单位为秒
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

2. 在K8s管理节点执行以下命令，启动Ascend Device Plugin。

    ```shell
    kubectl apply -f device-plugin-xxx-v{version}.yaml
    ```

    如在Atlas 训练系列产品环境下启动该组件，示例如下。

    ```shell
    kubectl apply -f device-plugin-volcano-v{version}.yaml
    ```

**配置NodeD<a name="section162092113510"></a>**

配置节点状态发送间隔时间。用户可以通过手动修改NodeD的启动YAML，配置上报节点状态的间隔时间。

1. 进入组件解压目录，执行以下命令，打开NodeD组件的启动YAML文件。

    ```shell
    vi noded-v{version}.yaml
    ```

2. 在YAML文件的“args”行修改“-**reportInterval**”参数，如下所示：

    <pre codetype="yaml">
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 <strong>-reportInterval=5</strong>" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...</pre>

## 制作镜像<a name="ZH-CN_TOPIC_0000002511426469"></a>

### 制作MindSpeed-LLM训练镜像（PyTorch框架）<a name="ZH-CN_TOPIC_0000002479386504"></a>

[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)作为昇腾大模型训练框架，旨在为昇腾芯片提供端到端的大语言模型训练方案，包含分布式预训练、分布式指令微调、分布式偏好对齐以及对应的开发工具链。[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)包括了仓库拉取、环境搭建与大模型训练等章节，制作MindSpeed-LLM训练框架镜像可以结合本章节和[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)。

断点续训可以基于基础训练镜像制作，基础训练镜像的制作可参考[使用Dockerfile构建容器镜像（PyTorch）](../../common_operations.md#使用dockerfile构建容器镜像pytorch)章节进行操作。

本章节结合基础训练镜像的制作步骤，展示基于Ubuntu 20.04来构建训练镜像。

>[!NOTE] 
>以下示例使用MindSpeed-LLM  2.3.0版本。

**准备软件包<a name="zh-cn_topic_0000002039339945_section18254161612586"></a>**

请按照[表1](#zh-cn_topic_0000002039339945_table1172542119019)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。软件包名称中{version}表示版本号、{arch}表示架构、{chip_type}表示芯片类型。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002039339945_table1172542119019"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339945_row157251121508"><th class="cellrowborder" valign="top" width="24.55%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002039339945_p1441653254"><a name="zh-cn_topic_0000002039339945_p1441653254"></a><a name="zh-cn_topic_0000002039339945_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="25.45%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002039339945_p2052053751"><a name="zh-cn_topic_0000002039339945_p2052053751"></a><a name="zh-cn_topic_0000002039339945_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002039339945_p657531455"><a name="zh-cn_topic_0000002039339945_p657531455"></a><a name="zh-cn_topic_0000002039339945_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002039339945_p1859531759"><a name="zh-cn_topic_0000002039339945_p1859531759"></a><a name="zh-cn_topic_0000002039339945_p1859531759"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339945_row16726192116014"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1754534515"><a name="zh-cn_topic_0000002039339945_p1754534515"></a><a name="zh-cn_topic_0000002039339945_p1754534515"></a>taskd-<em id="i2511021165615"><a name="i2511021165615"></a><a name="i2511021165615"></a>{version}</em>-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p4321152352612"><a name="p4321152352612"></a><a name="p4321152352612"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p205155310512"><a name="zh-cn_topic_0000002039339945_p205155310512"></a><a name="zh-cn_topic_0000002039339945_p205155310512"></a>集群调度组件断点续训whl包。</p>
<div class="note" id="zh-cn_topic_0000002039339945_note494818501423"><a name="zh-cn_topic_0000002039339945_note494818501423"></a><a name="zh-cn_topic_0000002039339945_note494818501423"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p159489506423"><a name="zh-cn_topic_0000002039339945_p159489506423"></a><a name="zh-cn_topic_0000002039339945_p159489506423"></a>安装<span id="ph1670711477256"><a name="ph1670711477256"></a><a name="ph1670711477256"></a>TaskD</span>组件前需确保<span id="zh-cn_topic_0000002039339945_ph998914174412"><a name="zh-cn_topic_0000002039339945_ph998914174412"></a><a name="zh-cn_topic_0000002039339945_ph998914174412"></a>PyTorch</span>框架已正确安装，当前支持的<span id="zh-cn_topic_0000002039339945_ph2908133144419"><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a>PyTorch</span>版本为：2.1.0、2.3.0、2.4.0、2.5.0、2.6.0、2.7.1。TaskD运行依赖PyTorch框架，请选择无已知安全漏洞的PyTorch版本或从官方社区获取已修复安全问题的对应版本。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p19595310517"><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1386820525510"><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p12512532515"><a name="zh-cn_topic_0000002039339945_p12512532515"></a><a name="zh-cn_topic_0000002039339945_p12512532515"></a>用户通过获取链接得到的是<span id="ph480901420289"><a name="ph480901420289"></a><a name="ph480901420289"></a>TaskD</span>压缩包Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a><a name="i1328312515383"></a>{arch}</em>.zip，需要通过解压后，获得相应的whl软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row572619211108"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1863531756"><a name="zh-cn_topic_0000002039339945_p1863531756"></a><a name="zh-cn_topic_0000002039339945_p1863531756"></a>mindio_ttp-<em id="zh-cn_topic_0000002039339945_i15340181201416"><a name="zh-cn_topic_0000002039339945_i15340181201416"></a><a name="zh-cn_topic_0000002039339945_i15340181201416"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002039339945_i19614531957"><a name="zh-cn_topic_0000002039339945_i19614531957"></a><a name="zh-cn_topic_0000002039339945_i19614531957"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p96553958"><a name="zh-cn_topic_0000002039339945_p96553958"></a><a name="zh-cn_topic_0000002039339945_p96553958"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p66253359"><a name="zh-cn_topic_0000002039339945_p66253359"></a><a name="zh-cn_topic_0000002039339945_p66253359"></a><span id="zh-cn_topic_0000002039339945_ph845710020145"><a name="zh-cn_topic_0000002039339945_ph845710020145"></a><a name="zh-cn_topic_0000002039339945_ph845710020145"></a>MindIO TFT</span>安装包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1862053354"><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row672652117018"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1468532510"><a name="zh-cn_topic_0000002039339945_p1468532510"></a><a name="zh-cn_topic_0000002039339945_p1468532510"></a>apex-0.1+ascend-cp3x-cp3x-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1869531256"><a name="zh-cn_topic_0000002039339945_p1869531256"></a><a name="zh-cn_topic_0000002039339945_p1869531256"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p156353454"><a name="zh-cn_topic_0000002039339945_p156353454"></a><a name="zh-cn_topic_0000002039339945_p156353454"></a>MindSpeed-LLM依赖</p>
<p id="zh-cn_topic_0000002039339945_p1861353651"><a name="zh-cn_topic_0000002039339945_p1861353651"></a><a name="zh-cn_topic_0000002039339945_p1861353651"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p166105316515"><a name="zh-cn_topic_0000002039339945_p166105316515"></a><a name="zh-cn_topic_0000002039339945_p166105316515"></a>混合精度训练是在训练时混合使用单精度（float32）与半精度(float16)数据类型，将两者结合在一起，并使用相同的超参数实现了与float32几乎相同的精度。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a>请参见<span id="zh-cn_topic_0000002039339945_ph156792413596"><a name="zh-cn_topic_0000002039339945_ph156792413596"></a><a name="zh-cn_topic_0000002039339945_ph156792413596"></a>《Ascend Extension for PyTorch 软件安装指南》中的“<a href="https://www.hiascend.com/document/detail/zh/Pytorch/730/configandinstg/instg/docs/installing_apex.md">安装APEX模块</a>”章节</span>，根据实际情况编译APEX软件包。</p>
<p id="zh-cn_topic_0000002039339945_p1761531257"><a name="zh-cn_topic_0000002039339945_p1761531257"></a><a name="zh-cn_topic_0000002039339945_p1761531257"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row197268213011"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1361953254"><a name="zh-cn_topic_0000002039339945_p1361953254"></a><a name="zh-cn_topic_0000002039339945_p1361953254"></a>torch_npu-2.7.1.<em id="zh-cn_topic_0000002039339945_i16204112111321"><a name="zh-cn_topic_0000002039339945_i16204112111321"></a><a name="zh-cn_topic_0000002039339945_i16204112111321"></a>{version}</em>-cp3x-cp3x-manylinux_2_28_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p56185314510"><a name="zh-cn_topic_0000002039339945_p56185314510"></a><a name="zh-cn_topic_0000002039339945_p56185314510"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p186653757"><a name="zh-cn_topic_0000002039339945_p186653757"></a><a name="zh-cn_topic_0000002039339945_p186653757"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p15720535517"><a name="zh-cn_topic_0000002039339945_p15720535517"></a><a name="zh-cn_topic_0000002039339945_p15720535517"></a>Ascend Extension for PyTorch插件是基于昇腾的深度学习适配框架，使昇腾NPU可以支持PyTorch框架，为PyTorch框架的使用者提供昇腾AI处理器的超强算力。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p10718533510"><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a href="https://www.hiascend.com/document/detail/zh/Pytorch/720/configandinstg/instg/insg_0004.html" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1165115165020"><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p167047813263"><a name="zh-cn_topic_0000002039339945_p167047813263"></a><a name="zh-cn_topic_0000002039339945_p167047813263"></a>如果使用MindSpeed-LLM代码仓上的<span id="zh-cn_topic_0000002039339945_ph1987542822613"><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a>PyTorch</span>模型，需要使用<span id="zh-cn_topic_0000002039339945_ph1412723132619"><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a>Ascend Extension for PyTorch</span> 2.6.0及以上版本。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row1412215399516"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><ul id="zh-cn_topic_0000002039339945_ul104867135415"><li><span id="zh-cn_topic_0000002039339945_ph0853174512272"><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a>x86_64</span>架构：torch-2.7.1+cpu.cxx11.abi-cp3x-cp3x-linux_x86_64.whl</li><li><span id="zh-cn_topic_0000002039339945_ph7852164518272"><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a>ARM</span>架构：torch-2.7.1+cpu-cp3x-cp3x-manylinux_2_28_aarch64.whl</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p871953357"><a name="zh-cn_topic_0000002039339945_p871953357"></a><a name="zh-cn_topic_0000002039339945_p871953357"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p17715531453"><a name="zh-cn_topic_0000002039339945_p17715531453"></a><a name="zh-cn_topic_0000002039339945_p17715531453"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a>官方<span id="zh-cn_topic_0000002039339945_ph19355165113512"><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a>PyTorch</span>包。</p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p99745421447"><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a href="https://download.pytorch.org/whl/torch/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<p id="zh-cn_topic_0000002039339945_p483943610920"><a name="zh-cn_topic_0000002039339945_p483943610920"></a><a name="zh-cn_topic_0000002039339945_p483943610920"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row151232039750"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p4714531658"><a name="zh-cn_topic_0000002039339945_p4714531658"></a><a name="zh-cn_topic_0000002039339945_p4714531658"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1774531250"><a name="zh-cn_topic_0000002039339945_p1774531250"></a><a name="zh-cn_topic_0000002039339945_p1774531250"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5720531514"><a name="zh-cn_topic_0000002039339945_p5720531514"></a><a name="zh-cn_topic_0000002039339945_p5720531514"></a>CANN算子包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p87153755"><a name="zh-cn_topic_0000002039339945_p87153755"></a><a name="zh-cn_topic_0000002039339945_p87153755"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note13775154104217"><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p2075812144313"><a name="zh-cn_topic_0000002039339945_p2075812144313"></a><a name="zh-cn_topic_0000002039339945_p2075812144313"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="row1173819266428"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="p7738142674217"><a name="p7738142674217"></a><a name="p7738142674217"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p12738626184214"><a name="p12738626184214"></a><a name="p12738626184214"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13738226114218"><a name="p13738226114218"></a><a name="p13738226114218"></a>CANN Toolkit开发套件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p19271154916428"><a name="p19271154916428"></a><a name="p19271154916428"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="note3272104913427"><a name="note3272104913427"></a><a name="note3272104913427"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p2272174920421"><a name="p2272174920421"></a><a name="p2272174920421"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row121231639952"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1048218391768"><a name="zh-cn_topic_0000002039339945_p1048218391768"></a><a name="zh-cn_topic_0000002039339945_p1048218391768"></a>MindSpeed</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p3482939367"><a name="zh-cn_topic_0000002039339945_p3482939367"></a><a name="zh-cn_topic_0000002039339945_p3482939367"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p048215393610"><a name="zh-cn_topic_0000002039339945_p048215393610"></a><a name="zh-cn_topic_0000002039339945_p048215393610"></a>MindSpeed是针对昇腾设备的大模型加速库。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p7482193913619"><a name="zh-cn_topic_0000002039339945_p7482193913619"></a><a name="zh-cn_topic_0000002039339945_p7482193913619"></a>git clone https://gitcode.com/Ascend/MindSpeed.git</p>
<p id="zh-cn_topic_0000002039339945_p9482139663"><a name="zh-cn_topic_0000002039339945_p9482139663"></a><a name="zh-cn_topic_0000002039339945_p9482139663"></a>cd MindSpeed</p>
<p id="zh-cn_topic_0000002039339945_p1948213912618"><a name="zh-cn_topic_0000002039339945_p1948213912618"></a><a name="zh-cn_topic_0000002039339945_p1948213912618"></a>git checkout 2.3.0_core_r0.12.1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row144125121466"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p848215391768"><a name="zh-cn_topic_0000002039339945_p848215391768"></a><a name="zh-cn_topic_0000002039339945_p848215391768"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1548319391465"><a name="zh-cn_topic_0000002039339945_p1548319391465"></a><a name="zh-cn_topic_0000002039339945_p1548319391465"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p16483239968"><a name="zh-cn_topic_0000002039339945_p16483239968"></a><a name="zh-cn_topic_0000002039339945_p16483239968"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p14483939561"><a name="zh-cn_topic_0000002039339945_p14483939561"></a><a name="zh-cn_topic_0000002039339945_p14483939561"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p748314394617"><a name="zh-cn_topic_0000002039339945_p748314394617"></a><a name="zh-cn_topic_0000002039339945_p748314394617"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row17301171913614"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p148310396616"><a name="zh-cn_topic_0000002039339945_p148310396616"></a><a name="zh-cn_topic_0000002039339945_p148310396616"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p348313399610"><a name="zh-cn_topic_0000002039339945_p348313399610"></a><a name="zh-cn_topic_0000002039339945_p348313399610"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p1348313391961"><a name="zh-cn_topic_0000002039339945_p1348313391961"></a><a name="zh-cn_topic_0000002039339945_p1348313391961"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5483239564"><a name="zh-cn_topic_0000002039339945_p5483239564"></a><a name="zh-cn_topic_0000002039339945_p5483239564"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p2483339861"><a name="zh-cn_topic_0000002039339945_p2483339861"></a><a name="zh-cn_topic_0000002039339945_p2483339861"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row93022191368"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15485173911615"><a name="zh-cn_topic_0000002039339945_p15485173911615"></a><a name="zh-cn_topic_0000002039339945_p15485173911615"></a>Dllogger代码仓</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p248514391563"><a name="zh-cn_topic_0000002039339945_p248514391563"></a><a name="zh-cn_topic_0000002039339945_p248514391563"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164850391565"><a name="zh-cn_topic_0000002039339945_p164850391565"></a><a name="zh-cn_topic_0000002039339945_p164850391565"></a>PyTorch日志工具。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1548619391164"><a name="zh-cn_topic_0000002039339945_p1548619391164"></a><a name="zh-cn_topic_0000002039339945_p1548619391164"></a>git clone https://github.com/NVIDIA/dllogger.git</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row33025197610"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p0486133919618"><a name="zh-cn_topic_0000002039339945_p0486133919618"></a><a name="zh-cn_topic_0000002039339945_p0486133919618"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1148619393610"><a name="zh-cn_topic_0000002039339945_p1148619393610"></a><a name="zh-cn_topic_0000002039339945_p1148619393610"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164861439461"><a name="zh-cn_topic_0000002039339945_p164861439461"></a><a name="zh-cn_topic_0000002039339945_p164861439461"></a>用于安装pip模块。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p34865396611"><a name="zh-cn_topic_0000002039339945_p34865396611"></a><a name="zh-cn_topic_0000002039339945_p34865396611"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row03021191165"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15486163919616"><a name="zh-cn_topic_0000002039339945_p15486163919616"></a><a name="zh-cn_topic_0000002039339945_p15486163919616"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1448693915619"><a name="zh-cn_topic_0000002039339945_p1448693915619"></a><a name="zh-cn_topic_0000002039339945_p1448693915619"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p74862391165"><a name="zh-cn_topic_0000002039339945_p74862391165"></a><a name="zh-cn_topic_0000002039339945_p74862391165"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p9571628104613"><a name="zh-cn_topic_0000002039339945_p9571628104613"></a><a name="zh-cn_topic_0000002039339945_p9571628104613"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 
>本章节以单台Atlas 800T A2 训练服务器、Ubuntu 20.04 Arm、配套Python  3.10为例来介绍训练镜像的制作，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000002039339945_section20489630477"></a>**

1. 参照[表1](#zh-cn_topic_0000002039339945_table1172542119019)，在宿主机上完成软件包的准备工作。
2. 编写如下Dockerfile。

    <pre>
    FROM ubuntu:20.04 
    WORKDIR /root 
    COPY . . 
      
    ARG PYTORCH_WHL=torch-2.7.1+cpu-cp310-cp310-manylinux_2_28_aarch64.whl 
    ARG PYTORCH_NPU_WHL=torch_npu-2.7.1.{version}-cp310-cp310-manylinux_2_28_aarch64.whl 
    ARG APEX_WHL=apex-0.1+ascend-cp310-cp310-linux_aarch64.whl 
    ARG HOST_ASCEND_BASE=/usr/local/Ascend 
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann 
    # 示例使用的CANN版本为8.5.0,使用过程中请根据实际情况修改
    ARG TOOLKIT=Ascend-cann-toolkit_8.5.0_linux-aarch64.run    
    ARG OPS=Ascend-cann-910b-ops_8.5.0_linux-aarch64.run 
    ARG TASKD_WHL=taskd-7.3.0-py3-none-linux_aarch64.whl   
    ARG MINDIO_TTP_WHL=mindio_ttp-1.0.0-py3-none-linux_aarch64.whl 
    ARG MINDSPEED=MindSpeed 
    ARG DLLOGGER=dllogger 
      
    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf 
      
    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\ 
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list 
      
    ARG DEBIAN_FRONTEND=noninteractive 
      
    # 系统包 
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
    # 建立Python软链
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
    # 系统
    RUN umask 0022 && apt update && \
            apt-get install -y --no-install-recommends \
            gcc g++ make cmake vim \
            zlib1g zlib1g-dev \
            openssl libsqlite3-dev libssl-dev \
            libffi-dev unzip pciutils \
            net-tools libblas-dev \
            gfortran libblas3 libopenblas-dev \
            curl unzip liblapack3 liblapack-dev \
            libhdf5-dev libxml2 patch
    # 时区
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
    # 配置pip源
    RUN mkdir -p ~/.pip \
    && echo '[global] \n\
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf
    # pip3.10
    RUN cd /tmp && \
        apt-get download python3-distutils && \
        dpkg-deb -x python3-distutils_*.deb / && \
        rm python3-distutils_*.deb && \
        cd - && \
        python get-pip.py && \
        rm get-pip.py
    RUN umask 0022 && \
        pip install sympy==1.4 && \
        pip install cffi && \
        pip install pathlib2 && \
        pip install grpcio && \
        pip install grpcio-tools && \
        pip install torchvision==0.22.1 && \
        pip install transformers==4.51.0 && \
        pip install absl-py && \
        pip install datasets && \
        pip install tokenizers==0.20.1 && \
        pip install pyOpenSSL
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
    # 安装torch、torch_npu、apex包
    RUN umask 0022 && pip install $PYTORCH_WHL && \
        pip install $PYTORCH_NPU_WHL && \
        pip install $APEX_WHL
      
    # Ascend包 
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录 
    RUN umask 0022 &&  \ 
        cp ascend_install.info /etc/ && \ 
        mkdir -p /usr/local/Ascend/driver/ && \ 
        cp version.info /usr/local/Ascend/driver/ && \ 
        chmod +x $TOOLKIT && \ 
        chmod +x $OPS 
      
    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet 
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc 
    RUN umask 0022 && ./$OPS --install --quiet 
      
    # 只为了安装toolkit包，所以需要清理，容器启动时通过Ascend Docker Runtime挂载进来 
    RUN rm -f version.info && rm -f ascend_install.info \ 
        rm -rf /usr/local/Ascend/driver/ 
      
    RUN umask 0022 && cd $MINDSPEED && \ 
        pip install -r requirements.txt && \ 
        pip install -e . && \ 
        echo "export PYTHONPATH=/root/MindSpeed:\$PYTHONPATH" >> ~/.bashrc 
      
    RUN umask 0022 && cd $DLLOGGER && \ 
        python setup.py build && \ 
        python setup.py install 
      
    # 导入环境变量 
    ENV HCCL_WHITELIST_DISABLE=1 
      
    # 创建/lib64/ld-linux-aarch64.so.1 
    RUN umask 0022 && \ 
        if [ ! -d "/lib64" ]; \ 
        then \ 
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \ 
        fi 
      
    # MindCluster断点续训适配脚本。
    RUN umask 0022 && \ 
        pip install $TASKD_WHL && \ 
        pip install $MINDIO_TTP_WHL 
      
    # 可选，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令。
    RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py  
    
    # 增加安装任务调度依赖库 
    RUN pip install apscheduler 
      
    RUN rm -rf tmp && \ 
        rm -f $PYTORCH_WHL && \ 
        rm -f $PYTORCH_NPU_WHL && \ 
        rm -f $APEX_WHL && \ 
        rm -f $TOOLKIT && \ 
        rm -f $OPS && \ 
        rm -f $TASKD_WHL && \ 
        rm -f $MINDIO_TTP_WHL && \ 
        rm -rf $DLLOGGER && \ 
        rm -rf Dockerfile 
    ## 最后打包成镜像mindspeed-dl:v1</pre>

    >[!NOTE] 
    >Python 3.10若无法通过PPA直接安装成功，或者deadsnakes PPA不提供Python 3.10版本的镜像源，则可下载源码手动编译安装。

3. 构建镜像。执行以下命令生成镜像。为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t mindspeed-dl:v1 .
    ```

### 制作MindFormers训练镜像（MindSpore框架）<a name="ZH-CN_TOPIC_0000002511426451"></a>

[MindSpore Transformers套件](https://gitcode.com/mindspore/mindformers)（以下简称MindFormers）的目标是构建一个大模型训练、微调、评估、推理、部署的全流程开发套件，提供业内主流的Transformer类预训练模型和SOTA下游任务应用，涵盖丰富的并行特性。期望帮助用户轻松地实现大模型训练和创新研发。

[MindSpore Transformers文档](https://www.mindspore.cn/mindformers/docs/zh-CN/r1.3.0/start/overview.html)的快速入门包括了安装与快速启动章节，可以在镜像制作时参考。

训练镜像可以基于基础训练镜像，结合MindFormers文档自行制作，基础训练镜像的制作可参考[使用Dockerfile构建容器镜像（MindSpore）](../../common_operations.md#使用dockerfile构建容器镜像mindspore)章节进行操作。

本章节结合基础训练镜像的制作步骤，展示基于Ubuntu 20.04来构建训练镜像。

**准备软件包<a name="zh-cn_topic_0000002003180012_section181941327124212"></a>**

请按照[表1](#zh-cn_topic_0000002003180012_table223643812168)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。软件包名称中{version}表示版本号、{arch}表示架构、{chip_type}表示芯片类型。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002003180012_table223643812168"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003180012_row6236938171619"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002003180012_p3390131317171"><a name="zh-cn_topic_0000002003180012_p3390131317171"></a><a name="zh-cn_topic_0000002003180012_p3390131317171"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002003180012_p173901213151712"><a name="zh-cn_topic_0000002003180012_p173901213151712"></a><a name="zh-cn_topic_0000002003180012_p173901213151712"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002003180012_p239018134178"><a name="zh-cn_topic_0000002003180012_p239018134178"></a><a name="zh-cn_topic_0000002003180012_p239018134178"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002003180012_p1539051321714"><a name="zh-cn_topic_0000002003180012_p1539051321714"></a><a name="zh-cn_topic_0000002003180012_p1539051321714"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003180012_row13237173817161"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p6390191319171"><a name="zh-cn_topic_0000002003180012_p6390191319171"></a><a name="zh-cn_topic_0000002003180012_p6390191319171"></a>MindFormers代码仓</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p13390113131712"><a name="zh-cn_topic_0000002003180012_p13390113131712"></a><a name="zh-cn_topic_0000002003180012_p13390113131712"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p153901913131711"><a name="zh-cn_topic_0000002003180012_p153901913131711"></a><a name="zh-cn_topic_0000002003180012_p153901913131711"></a>构建一个大模型训练、微调、评估、推理、部署的全流程开发套件，提供业内主流的Transformer类预训练模型和SOTA下游任务应用，涵盖丰富的并行特性<span id="ph19351335211"><a name="ph19351335211"></a><a name="ph19351335211"></a>。</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p3390131316172"><a name="zh-cn_topic_0000002003180012_p3390131316172"></a><a name="zh-cn_topic_0000002003180012_p3390131316172"></a>git clone https://gitcode.com/mindspore/mindformers.git</p>
<p id="zh-cn_topic_0000002003180012_p5390101317175"><a name="zh-cn_topic_0000002003180012_p5390101317175"></a><a name="zh-cn_topic_0000002003180012_p5390101317175"></a>cd mindformers</p>
<p id="zh-cn_topic_0000002003180012_p9390151318171"><a name="zh-cn_topic_0000002003180012_p9390151318171"></a><a name="zh-cn_topic_0000002003180012_p9390151318171"></a>git checkout 15ff59dd55b84b4dfc7de03f7f20f6e2be3669ec</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row14237113817167"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p133901013201717"><a name="zh-cn_topic_0000002003180012_p133901013201717"></a><a name="zh-cn_topic_0000002003180012_p133901013201717"></a>requirements.txt文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p10390813171719"><a name="zh-cn_topic_0000002003180012_p10390813171719"></a><a name="zh-cn_topic_0000002003180012_p10390813171719"></a>否</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p439011371714"><a name="zh-cn_topic_0000002003180012_p439011371714"></a><a name="zh-cn_topic_0000002003180012_p439011371714"></a>由于通过pip安装MindSpore时，可能出现依赖的组件安装报错，故可以先安装依赖。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6390121315177"><a name="zh-cn_topic_0000002003180012_p6390121315177"></a><a name="zh-cn_topic_0000002003180012_p6390121315177"></a>wget https://gitcode.com/mindspore/mindspore/raw/r2.4.1/requirements.txt</p>
<div class="note" id="zh-cn_topic_0000002003180012_note14449193224617"><a name="zh-cn_topic_0000002003180012_note14449193224617"></a><a name="zh-cn_topic_0000002003180012_note14449193224617"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p15449133274617"><a name="zh-cn_topic_0000002003180012_p15449133274617"></a><a name="zh-cn_topic_0000002003180012_p15449133274617"></a>MindSpore软件包与<span id="zh-cn_topic_0000002003180012_ph327965117217"><a name="zh-cn_topic_0000002003180012_ph327965117217"></a><a name="zh-cn_topic_0000002003180012_ph327965117217"></a>Atlas 训练系列产品</span>需配套使用，请参见MindSpore<a href="https://www.mindspore.cn/install" target="_blank" rel="noopener noreferrer">安装指南</a>查看对应关系。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row2023743821619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p123901513131714"><a name="zh-cn_topic_0000002003180012_p123901513131714"></a><a name="zh-cn_topic_0000002003180012_p123901513131714"></a>mindspore-<em id="zh-cn_topic_0000002003180012_i42701940155017"><a name="zh-cn_topic_0000002003180012_i42701940155017"></a><a name="zh-cn_topic_0000002003180012_i42701940155017"></a>{version}</em>-cp3x-cp3x-linux_aarch64.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p73901313101718"><a name="zh-cn_topic_0000002003180012_p73901313101718"></a><a name="zh-cn_topic_0000002003180012_p73901313101718"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p1839181315178"><a name="zh-cn_topic_0000002003180012_p1839181315178"></a><a name="zh-cn_topic_0000002003180012_p1839181315178"></a>MindSpore whl包<span id="ph441575419329"><a name="ph441575419329"></a><a name="ph441575419329"></a>。</span></p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，请根据实际情况选择对应软件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6391181310177"><a name="zh-cn_topic_0000002003180012_p6391181310177"></a><a name="zh-cn_topic_0000002003180012_p6391181310177"></a><a href="https://www.mindspore.cn/install/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row32371838111619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p13917139175"><a name="zh-cn_topic_0000002003180012_p13917139175"></a><a name="zh-cn_topic_0000002003180012_p13917139175"></a>mindio_ttp-<em id="zh-cn_topic_0000002003180012_i14277191551111"><a name="zh-cn_topic_0000002003180012_i14277191551111"></a><a name="zh-cn_topic_0000002003180012_i14277191551111"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002003180012_i16391113201710"><a name="zh-cn_topic_0000002003180012_i16391113201710"></a><a name="zh-cn_topic_0000002003180012_i16391113201710"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p183915136176"><a name="zh-cn_topic_0000002003180012_p183915136176"></a><a name="zh-cn_topic_0000002003180012_p183915136176"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p13906311171017"><a name="zh-cn_topic_0000002003180012_p13906311171017"></a><a name="zh-cn_topic_0000002003180012_p13906311171017"></a><span id="zh-cn_topic_0000002003180012_ph845710020145"><a name="zh-cn_topic_0000002003180012_ph845710020145"></a><a name="zh-cn_topic_0000002003180012_ph845710020145"></a>MindIO TFT</span>安装包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p11392111316172"><a name="zh-cn_topic_0000002003180012_p11392111316172"></a><a name="zh-cn_topic_0000002003180012_p11392111316172"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row1423815380168"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p0915027134813"><a name="zh-cn_topic_0000002003180012_p0915027134813"></a><a name="zh-cn_topic_0000002003180012_p0915027134813"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1139314132170"><a name="zh-cn_topic_0000002003180012_p1139314132170"></a><a name="zh-cn_topic_0000002003180012_p1139314132170"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p193931713181710"><a name="zh-cn_topic_0000002003180012_p193931713181710"></a><a name="zh-cn_topic_0000002003180012_p193931713181710"></a>CANN算子包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p139312131171"><a name="zh-cn_topic_0000002003180012_p139312131171"></a><a name="zh-cn_topic_0000002003180012_p139312131171"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note13501612171513"><a name="zh-cn_topic_0000002003180012_note13501612171513"></a><a name="zh-cn_topic_0000002003180012_note13501612171513"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p1519161921516"><a name="zh-cn_topic_0000002003180012_p1519161921516"></a><a name="zh-cn_topic_0000002003180012_p1519161921516"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row8238173810165"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p1439381351714"><a name="zh-cn_topic_0000002003180012_p1439381351714"></a><a name="zh-cn_topic_0000002003180012_p1439381351714"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1239319131176"><a name="zh-cn_topic_0000002003180012_p1239319131176"></a><a name="zh-cn_topic_0000002003180012_p1239319131176"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p5393121371719"><a name="zh-cn_topic_0000002003180012_p5393121371719"></a><a name="zh-cn_topic_0000002003180012_p5393121371719"></a>CANN Toolkit开发套件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p239319132175"><a name="zh-cn_topic_0000002003180012_p239319132175"></a><a name="zh-cn_topic_0000002003180012_p239319132175"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note1733918441613"><a name="zh-cn_topic_0000002003180012_note1733918441613"></a><a name="zh-cn_topic_0000002003180012_note1733918441613"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p533924121616"><a name="zh-cn_topic_0000002003180012_p533924121616"></a><a name="zh-cn_topic_0000002003180012_p533924121616"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row4825411181413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p1882511116142"><a name="zh-cn_topic_0000002003180012_p1882511116142"></a><a name="zh-cn_topic_0000002003180012_p1882511116142"></a>taskd-{version}-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p10825811101413"><a name="zh-cn_topic_0000002003180012_p10825811101413"></a><a name="zh-cn_topic_0000002003180012_p10825811101413"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p4825711171415"><a name="zh-cn_topic_0000002003180012_p4825711171415"></a><a name="zh-cn_topic_0000002003180012_p4825711171415"></a>集群调度组件断点续训whl包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p18169645192413"><a name="zh-cn_topic_0000002003180012_p18169645192413"></a><a name="zh-cn_topic_0000002003180012_p18169645192413"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note079418496154"><a name="zh-cn_topic_0000002003180012_note079418496154"></a><a name="zh-cn_topic_0000002003180012_note079418496154"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002003180012_ul79293962319"></a><a name="zh-cn_topic_0000002003180012_ul79293962319"></a><ul id="zh-cn_topic_0000002003180012_ul79293962319"><li>MindSpore场景下使用优雅容错、Pod级别重调度、进程级别重调度、进程级在线恢复，必须安装该whl包。</li><li>用户通过获取链接得到的是<span id="zh-cn_topic_0000002003180012_ph11742444163719"><a name="zh-cn_topic_0000002003180012_ph11742444163719"></a><a name="zh-cn_topic_0000002003180012_ph11742444163719"></a>TaskD</span>压缩包Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a><a name="i1328312515383"></a>{arch}</em>.zip，需要通过解压后，获得相应的whl软件包。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row15183115071614"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p83941613131719"><a name="zh-cn_topic_0000002003180012_p83941613131719"></a><a name="zh-cn_topic_0000002003180012_p83941613131719"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p4394913121712"><a name="zh-cn_topic_0000002003180012_p4394913121712"></a><a name="zh-cn_topic_0000002003180012_p4394913121712"></a>是</p>
<p id="zh-cn_topic_0000002003180012_p0394413131713"><a name="zh-cn_topic_0000002003180012_p0394413131713"></a><a name="zh-cn_topic_0000002003180012_p0394413131713"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p73942134170"><a name="zh-cn_topic_0000002003180012_p73942134170"></a><a name="zh-cn_topic_0000002003180012_p73942134170"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p239415135179"><a name="zh-cn_topic_0000002003180012_p239415135179"></a><a name="zh-cn_topic_0000002003180012_p239415135179"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row218375021618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p5394141315176"><a name="zh-cn_topic_0000002003180012_p5394141315176"></a><a name="zh-cn_topic_0000002003180012_p5394141315176"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1039401316175"><a name="zh-cn_topic_0000002003180012_p1039401316175"></a><a name="zh-cn_topic_0000002003180012_p1039401316175"></a>是</p>
<p id="zh-cn_topic_0000002003180012_p14394171331717"><a name="zh-cn_topic_0000002003180012_p14394171331717"></a><a name="zh-cn_topic_0000002003180012_p14394171331717"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p03941613151715"><a name="zh-cn_topic_0000002003180012_p03941613151715"></a><a name="zh-cn_topic_0000002003180012_p03941613151715"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p11394141321712"><a name="zh-cn_topic_0000002003180012_p11394141321712"></a><a name="zh-cn_topic_0000002003180012_p11394141321712"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row61841150171618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p6394713201715"><a name="zh-cn_topic_0000002003180012_p6394713201715"></a><a name="zh-cn_topic_0000002003180012_p6394713201715"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p14394121310177"><a name="zh-cn_topic_0000002003180012_p14394121310177"></a><a name="zh-cn_topic_0000002003180012_p14394121310177"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p63959135172"><a name="zh-cn_topic_0000002003180012_p63959135172"></a><a name="zh-cn_topic_0000002003180012_p63959135172"></a>用于安装pip模块</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6395913121711"><a name="zh-cn_topic_0000002003180012_p6395913121711"></a><a name="zh-cn_topic_0000002003180012_p6395913121711"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row618410501169"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p123971013121710"><a name="zh-cn_topic_0000002003180012_p123971013121710"></a><a name="zh-cn_topic_0000002003180012_p123971013121710"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p18397713121711"><a name="zh-cn_topic_0000002003180012_p18397713121711"></a><a name="zh-cn_topic_0000002003180012_p18397713121711"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p639719136172"><a name="zh-cn_topic_0000002003180012_p639719136172"></a><a name="zh-cn_topic_0000002003180012_p639719136172"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p639714133170"><a name="zh-cn_topic_0000002003180012_p639714133170"></a><a name="zh-cn_topic_0000002003180012_p639714133170"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 
>本章节以单台Atlas 800T A2 训练服务器、Ubuntu 20.04、配套Python 3.10为例来介绍制作镜像的详细过程，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000002003180012_section614453171018"></a>**

1. 在宿主机上完成软件包的准备工作。
2. 构建如下的Dockerfile。

    <pre>
    FROM ubuntu:20.04
     
    WORKDIR /root
     
    COPY . .
     
    ARG HOST_ASCEND_BASE=/usr/local/Ascend
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann
    # 示例使用的CANN版本为8.5.0,使用过程中请根据实际情况修改
    ARG TOOLKIT=Ascend-cann-toolkit_8.5.0_linux-aarch64.run    
    ARG OPS=Ascend-cann-910b-ops_8.5.0_linux-aarch64.run
    ARG MINDIO_TTP_WHL=mindio_ttp-1.0.0-py3-none-linux_aarch64.whl
    ARG MINDFORMERS=mindformers
    ARG MINDSPORE_REQUIREMENTS=requirements.txt
    ARG MINDSPORE_WHL=mindspore-2.5.0-cp310-cp310-linux_aarch64.whl
    ARG TASKD_WHL=taskd-7.0.RC1-py3-none-linux_aarch64.whl    
     
    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf
     
    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list
     
    ARG DEBIAN_FRONTEND=noninteractive
     
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
     
    # 建立Python软链接
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
     
    # 系统包
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
            gcc g++ make cmake vim \
            zlib1g zlib1g-dev \
            openssl libsqlite3-dev libssl-dev \
            libffi-dev unzip pciutils \
            net-tools libblas-dev \
            gfortran libblas3 libopenblas-dev \
            curl unzip liblapack3 liblapack-dev \
            libhdf5-dev libxml2 patch
     
    # 时区
    # RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
     
    # 配置pip源
    RUN mkdir -p ~/.pip \
    && echo '[global] \n\
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf
     
    # pip3.10
    RUN cd /tmp && \
        apt-get download python3-distutils && \
        dpkg-deb -x python3-distutils_*.deb / && \
        rm python3-distutils_*.deb && \
        cd - && \
        python get-pip.py && \
    rm get-pip.py
     
    RUN umask 0022 && \
        pip install sympy==1.4 && \
        pip install cffi && \
        pip install pathlib2 && \
        pip install grpcio && \
        pip install grpcio-tools && \
        pip install absl-py && \
        pip install datasets && \
        pip install tokenizers==0.20.1 && \
        pip install pyOpenSSL
     
    # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
     
    # Ascend包
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录
    RUN umask 0022 &&  \
        cp ascend_install.info /etc/ && \
        mkdir -p /usr/local/Ascend/driver/ && \
        cp version.info /usr/local/Ascend/driver/ && \
        chmod +x $TOOLKIT && \
        chmod +x $OPS
     
    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc
    RUN umask 0022 && ./$OPS --install --quiet
     
    # 只为了安装toolkit包，所以需要清理，容器启动时通过ascend docker挂载进来
    RUN rm -f version.info && \
        rm -rf /usr/local/Ascend/driver/
     
    # 安装mindspore
    RUN umask 0022 && pip uninstall te topi hccl -y && \
             pip install sympy && \
             pip install /usr/local/Ascend/cann/lib64/hccl-*-py3-none-any.whl
    RUN umask 0022 && \
        pip install -r $MINDSPORE_REQUIREMENTS && \
        pip install $MINDSPORE_WHL
     
    # 安装mindformers
    RUN umask 0022 && cd $MINDFORMERS && \
        pip install -r requirements.txt
     
    # MindCluster无损失断点续训适配脚本
    RUN umask 0022 && \
        pip install $MINDIO_TTP_WHL --target=$(pip show mindspore | awk '/Location:/ {print $2}') && \
        pip install $TASKD_WHL
     
    # 环境变量
    ENV HCCL_WHITELIST_DISABLE=1
     
    # 创建/lib64/ld-linux-aarch64.so.1
    RUN umask 0022 && \
        if [ ! -d "/lib64" ]; \
        then \
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
        fi
     
    # 增加安装任务调度依赖库
    RUN pip install apscheduler
      
    RUN rm -rf tmp && \
        rm -f $TOOLKIT && \
        rm -f $OPS && \
        rm -f $MINDIO_TTP_WHL && \
        rm -f $MINDSPORE_REQUIREMENTS && \
        rm -f $MINDSPORE_WHL
    ## 最后打包成镜像mindformers-dl:v1</pre>

3. 构建镜像。执行以下命令生成镜像。为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t mindformers-dl:v1 .
    ```

### 制作强化学习后训练镜像（Verl框架）<a name="ZH-CN_TOPIC_0000002511426439"></a>

[Verl](https://verl.readthedocs.io/en/latest/index.html)是一款专为大语言模型（LLM）后训练阶段设计的灵活、高效且具备生产就绪能力的强化学习训练框架。

**构建镜像**

详细请参见[Verl官网文档-构建镜像](https://github.com/verl-project/verl/blob/main/docs/ascend_tutorial/quick_start/dockerfile_build_guidance.rst)。vLLM和Megatron分别作为推理和训练后端。

**安装软件**

详细请参见[Verl官网文档-安装软件](https://github.com/verl-project/verl/blob/main/docs/ascend_tutorial/quick_start/ascend_quick_start.rst)。

>[!NOTE] 
>若需使用Pod重调度功能，建议MindSpeed版本不早于commit id为6390a8ee2f0e59ae237753cce51289a3fe490905的版本。

**（可选）安装jemalloc**

制作镜像时，可以选择安装jemalloc，以优化内存管理。源码获取链接：[jemalloc](https://github.com/jemalloc/jemalloc/releases)

1. 执行以下命令进行安装。

   ```shell
   tar -xvf jemalloc-{version}.tar.bz2
   cd jemalloc-{version}
   ./configure --prefix=/usr/local
   make
   make install
   ```

2. 安装完成后设置环境变量。以安装路径“/usr/local/lib/libjemalloc.so.2”为例。

   ```shell
   export LD_PRELOAD=/usr/local/lib/libjemalloc.so.2
   ```

## 脚本适配<a name="ZH-CN_TOPIC_0000002511426481"></a>

### 流程说明<a name="ZH-CN_TOPIC_0000002511346469"></a>

模型脚本需要适配CKPT之后才可以使用断点续训功能，脚本适配大致流程和逻辑如[图1](#fig88341718121515)所示。

**图 1**  脚本适配流程<a name="fig88341718121515"></a>  
![](../../../figures/scheduling/脚本适配流程.png "脚本适配流程")

### 适配示例<a name="ZH-CN_TOPIC_0000002511346445"></a>

本章节将指导用户step by step地完成断点续训的适配步骤。

- [PyTorch场景适配示例（基于MindSpeed-LLM）](#zh-cn_topic_0000002003180016_section412442472511)
- [MindSpore场景适配示例（基于MindFormers）](#zh-cn_topic_0000002003180016_section718243883518)
- [强化学习后训练场景适配示例（基于Verl）](#section1335017512276)

>[!NOTE]
> 
>- 为保证优雅容错与进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>- 断点续训展示的组件代码为开源代码，其中涉及到相关安全说明请参见[安全说明](../../appendix.md#安全说明)。
>- 下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
>- 模型的参数配置，根据模型仓的模型配置以实际情况来写。若修改不当，可能会引发不可预知的问题。
>- 若训练过程中出现“Failed to bind the IP port. Reason: The IP address and port have been bound already”报错，可以按照如下进行配置，详情请参见《CANN 环境变量参考》中的“[HCCL_HOST_SOCKET_PORT_RANGE](https://www.hiascend.com/document/detail/zh/canncommercial/850/maintenref/envvar/envref_07_0143.html)”章节。
>
>   ```shell
>   export HCCL_HOST_SOCKET_PORT_RANGE="60000-60050"
>   export HCCL_NPU_SOCKET_PORT_RANGE="61000-61050"
>   ```
>
>- 若使用TaskD组件且训练容器使用Host网络，则先通过`sysctl net.ipv4.ip_local_reserved_ports`查询当前预留端口配置后，通过`sysctl -w net.ipv4.ip_local_reserved_ports="xxx,9601,9602"`新增预留端口9601、9602（其中xxx指的是前面查出来已配置的端口，若无则省略）。

**PyTorch场景适配示例（基于MindSpeed-LLM）<a name="zh-cn_topic_0000002003180016_section412442472511"></a>**

训练代码与数据集准备，可以参考[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/docs/pytorch/solutions/pretrain/pretrain.md)。下面以两台Atlas 800T A2 训练服务器为例，说明具体操作步骤。

1. 拉取训练代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/MindSpeed-LLM.git
    git clone https://github.com/NVIDIA/Megatron-LM.git
    cd MindSpeed-LLM
    git checkout 2.3.0
    cd ..
    cd Megatron-LM 
    git checkout core_v0.12.1
    cp -r megatron ../MindSpeed-LLM #此处目的是将Megatron-LM项目下的Megatron目录复制到MindSpeed-LLM项目下
    ## 重命名MindSpeed-LLM为QWEN3_for_PyTorch_2.7_code
    cd ..
    mv MindSpeed-LLM QWEN3_for_PyTorch_2.7_code
    ```

2. 获取模型权重。

    请用户自行从[Qwen3](https://huggingface.co/Qwen/Qwen3-8B/tree/main)下载模型权重放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-8b-hf”。

3. 获取数据集。

    请用户自行从[Alpaca](https://huggingface.co/datasets/tatsu-lab/alpaca/blob/main/data/train-00000-of-00001-a09b74b3ef9c3b56.parquet)下载数据集（以Alpaca数据集为例）放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-alpaca”。

4. 处理数据集。
    1. 启动容器。

        ```shell
        docker run -it -v /data/atlas_dls/public/:/data/atlas_dls/public/ -e ASCEND_VISIBLE_DEVICES=0-7 mindspeed-dl:v1 bash
        ```

    2. 在容器中执行如下操作。

        ```shell
        export TORCH_DEVICE_BACKEND_AUTOLOAD=0
        source /usr/local/Ascend/cann/set_env.sh
        cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code
        # 可选，如下为安装MindSpeed加速库操作，可在任意目录下执行。若制作镜像时已安装，则跳过该操作
        git clone https://gitcode.com/ascend/MindSpeed.git 
        cd MindSpeed 
        git checkout 2.3.0_core_r0.12.1
        pip install -r requirements.txt 
        pip install -e . 
        export PYTHONPATH=/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/MindSpeed:$PYTHONPATH
        cd ..
        ```

    3. 处理数据集。

        Qwen3要求使用Transformers\>=4.51.0，因此Python需使用3.9及以上版本且需要安装4.51.0及以上的Transformers。

        ```Python
        python preprocess_data.py \
            --input /data/atlas_dls/public/dataset/qwen3-alpaca/train-00000-of-00001-a09b74b3ef9c3b56.parquet \ # 数据集文件路径
            --tokenizer-name-or-path /data/atlas_dls/public/dataset/qwen3-8b-hf \ # 开源模型权重文件目录
            --tokenizer-type PretrainedFromHF \
            --handler-name GeneralPretrainHandler \
            --output-prefix /data/atlas_dls/public/dataset/qwen3-alpaca/alpaca \ # 会生成alpaca_text_document.bin和.idx文件
            --json-keys text \
            --workers 4 \
            --log-interval 1000
        ```

        >[!NOTE] 
        >若出现报错：/usr/local/lib/python3.10/dist-packages/sklearn/utils/../../scikit\_learn.libs/libgomp-947d5fa1.so.1.0.0: cannot allocate memory in static TLS block，可执行以下命令预加载libgomp库。
        >
        >```shell
        >export LD_PRELOAD="/usr/local/lib/python3.10/dist-packages/scikit_learn.libs/libgomp-947d5fa1.so.1.0.0"
        >```

5. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3”目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

6. 获取[训练任务YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml)。该YAML中已经配置了Pod级别重调度、进程级别重调度、进程级在线恢复、弹性训练等。根据实际情况配置挂载卷的服务器IP地址、各种重调度级别等。

    进程级别重调度、进程级在线恢复、弹性训练等训练进程级别的恢复与优雅容错不可同时存在。优雅容错的配置步骤请参见[优雅容错模式](#可选配置组件)。

7. 配置训练启动脚本train\_start.sh和训练任务YAML，请根据实际情况进行修改。
    1. 修改启动脚本基础参数。

        ```shell
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/ttplogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/trainlogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/demo/
        # 日志保存路径，可根据实际情况修改
        export ASCEND_PROCESS_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP       # 设置plog保存路径，其中$MINDX_TASK_ID为Ascend Operator注入的任务UID环境变量，$XDL_IP为任务YAML中写入的环境变量status.hostIP
        export TTP_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/ttplogs/ttplog$XDL_IP-$RANK    # 设置TTP日志保存路径，其中$RANK为Ascend Operator为PyTorch框架注入的环境变量
        export TRAIN_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/trainlogs/$XDL_IP-$RANK      # 设置训练日志保存路径
        export GLOO_SOCKET_IFNAME=enp189s0f0               # 物理机上可以通信的网口，根据主节点高速网卡实际情况进行配置，如任务YAML中配置hostNetwork为false，则设置为eth0
        export HCCL_SOCKET_IFNAME=enp189s0f0               # 如任务YAML中配置hostNetwork为false，则设置为eth0
         
        CKPT_SAVE_DIR="/job/code/output/ckpt" # 训练完成后的权重保存路径
        DATA_PATH="/job/data/alpaca_text_document" # 数据集路径，填入数据预处理时保存的数据路径
        TOKENIZER_PATH="/job/data/qwen3-8b-hf" # 词表路径，填入下载的开源权重词表路径
        CKPT_LOAD_DIR="/job/code/output/ckpt" # 权重加载路径
        ```

    2. 使用TaskD完成进程级别重调度、进程级在线恢复、进程级别原地恢复或弹性训练，还需拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
             
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
             
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码，拉起TaskD  Manager。

            在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi
                

            torchrun $DISTRIBUTED_ARGS ...
            ```

        3. 修改训练任务YAML，新增容器端口，在所有的Pod下增加TaskD通信使用的端口9601（如已有则跳过）。

            ```Yaml
            ...
                    spec:
            ...
                      containers:
            ...
                        ports:                          
                         - containerPort: 9601              
                           name: taskd-port 
            ...
            ```

**MindSpore场景适配示例（基于MindFormers）<a name="zh-cn_topic_0000002003180016_section718243883518"></a>**

训练代码与数据集准备，可以参考[MindFormers文档](https://gitcode.com/mindspore/mindformers/tree/master/configs/qwen3)。下面以两台Atlas 900 A3 SuperPoD 超节点为例，说明具体操作步骤。

1. 准备代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/mindspore/mindformers.git
    cd mindformers
    git checkout 15ff59dd55b84b4dfc7de03f7f20f6e2be3669ec
    # 将mindformers重命名为QWEN3_for_MS_code
    cd ..
    mv mindformers QWEN3_for_MS_code
    ```

2. 准备数据集。

    请用户自行从[DagsHub](https://dagshub.com/DagsHub/WIkiText-103/src/main/dataset/tokens/wiki.train.tokens)下载数据集并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset”。

3. 转换数据集。
    1. 下载数据集转换脚本。

        从[数据集转换](https://gitee.com/mindspore/mindformers/issues/ICOKGY)下载数据集转换脚本并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/gen\_wiki\_json.py”。

    2. 下载tokenizer文件。

        从[Qwen3-32B](https://huggingface.co/Qwen/Qwen3-32B/tree/main)下载tokenizer文件并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/Qwen3-32B-tokenizer”。

    3. 转换数据集。
        1. 启动容器并挂载所需文件。

            ```shell
            docker run -it -v /data/atlas_dls/public/code/:/data/atlas_dls/public/code/ mindformers-dl:v1 bash
            ```

        2. 执行转换脚本，将wiki.train.tokens转换为jsonl格式。

            ```shell
            # 执行该脚本需要的Python环境，请提前准备Python环境
            cd /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset
            python gen_wiki_json.py --input wiki.train.tokens  --output wiki.jsonl 
            ```

        3. 将jsonl格式数据转为bin格式数据。

            ```shell
            # 执行时若报错ModuleNotFoundError: No module named 'xxx'，请自行安装依赖
            cd /data/atlas_dls/public/code/QWEN3_for_MS_code
            python toolkit/data_preprocess/megatron/preprocess_indexed_dataset.py \
              --input /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset/wiki.jsonl \
              --output-prefix /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset/wiki103-megatron \
              --tokenizer-type HuggingFaceTokenizer \
              --tokenizer-dir /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset/Qwen3-32B-tokenizer # 其他规格的模型可以调整为对应的tokenizer路径
            ```

            运行完成后，“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset”目录下会生成“wiki103-megatron\_text\_document.bin”和“wiki103-megatron\_text\_document.idx”文件。填写数据集路径时，需要使用“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/wiki103-megatron\_text\_document”，不需要带后缀名。

4. 获取[训练任务YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/yamls/ms_multinodes_acjob_superpod.yaml)和[训练启动脚本](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/msrun_launcher.sh)，并进行修改。
    1. 若训练任务YAML中“hostNetwork”参数值为“false”，则需要将启动脚本中“GLOO\_SOCKET\_IFNAME”的值设置为“eth0”。示例如下：

        ```shell
        export GLOO_SOCKET_IFNAME=eth0  #eth0是容器内可以通信的网口
        export HCCL_SOCKET_IFNAME=eth0
        ```

        然后根据实际情况修改启动脚本中的其他参数。

    2. 根据实际情况修改任务YAML中挂载卷的服务器IP地址等配置。
    3. 使用TaskD完成进程级别重调度、进程级在线恢复、进程级别原地恢复、借轨通信任务暂停与回切或在线压测，还需拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码拉起TaskD  Manager。在以下代码中，前两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

            ```Python
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi
            msrun ...
            ```

        3. 修改训练任务YAML，新增容器端口，在所有的Pod下增加TaskD通信使用的端口9601（如已有则跳过）。

            ```Yaml
            ...
                    spec:
            ...
                      containers:
            ...
                        ports:                          
                         - containerPort: 9601              
                           name: taskd-port 
            ...
            ```

5. 修改参数模型配置文件。
    1. 打开代码目录下“configs/qwen3/pretrain\_qwen3\_32b\_4k.yaml”文件。

        ```shell
        vi configs/qwen3/pretrain_qwen3_32b_4k.yaml
        ```

    2. 按“i”进入编辑模式，修改参数模型配置文件。
        1. 修改如下加粗配置，包括数据集路径、分布式并行参数、模型参数等。以下模型参数仅供参考，如有需要请自行修改。

            <pre codetype="yaml">
            train_dataset: &train_dataset
              data_loader:
                type: BlendedMegatronDatasetDataLoader
                datasets_type: "GPTDataset"
                sizes:
                  - 8000  # Number of samples in the training set
                  - 0     # Number of samples in the test set (currently unsupported)
                  - 0     # Number of samples in the evaluation set (currently unsupported)
                config:
                  seed: 1234  # Random seed for data sampling
                  split: "1, 0, 0"  # Proportions for training, test, and evaluation sets (test/eval currently unsupported)
                  seq_length: 4096  # Sequence length of the dataset
                  eod_mask_loss: False  # Whether to calculate loss at the end-of-document (EOD)
                  reset_position_ids: False  # Whether to reset position_ids at EOD
                  create_attention_mask: True  # Whether to include attention_mask in the dataset
                  reset_attention_mask: False  # Whether to reset attention_mask at EOD, creating a stepped attention_mask
                  create_compressed_eod_mask: False  # Whether to include a compressed attention_mask
                  eod_pad_length: 128  # Length of the compressed attention_mask
                  eod: 1  # Token ID for EOD in the dataset
                  pad: -1  # Token ID for padding in the dataset
                  data_path:  # Sampling proportion and path for the Megatron dataset
                    - '1'
                    <strong>- "/job/data/wiki103-megatron_text_document" # 数据集路径</strong>
            ……
            # Parallel configuration
            parallel_config:
              <strong>data_parallel: &dp 4  # Number of data parallel. If using the high availability feature, it must be an even number.</strong>
              <strong>model_parallel: 8  # Number of model parallel</strong>
              <strong>pipeline_stage: 1  # Number of pipeline parallel</strong>
              <strong>micro_batch_num: 1  # Pipeline parallel microbatch size</strong>
              use_seq_parallel: False  # Whether to enable sequence parallelism
              gradient_aggregation_group: 1  # Size of the gradient communication operator fusion group
            # When model_parallel > 1, setting micro_batch_interleave_num to 2 may accelerate the training process.
            micro_batch_interleave_num: 1
            ……
            model:
              model_config:
                # Configurations from Hugging Face
                <strong>vocab_size: 75968            # 此处改小了模型参数仅供测试，如有需要请自行调整</strong>
                <strong>hidden_size: 2560           # 此处改小了模型参数仅供测试，如有需要请自行调整</strong>
                <strong>intermediate_size: 12800   # 此处改小了模型参数仅供测试，如有需要请自行调整</strong>
                <strong>num_hidden_layers: 32      # 此处改小了模型参数仅供测试，如有需要请自行调整</strong>
                <strong>num_attention_heads: 32    # 此处改小了模型参数仅供测试，如有需要请自行调整</strong>
                num_key_value_heads: 8
                head_dim: 128
                hidden_act: 'swiglu'
                max_position_embeddings: 4096
                seq_length: 4096
                initializer_range: 0.02
                rms_norm_eps: 1.e-6
                use_cache: True
                tie_word_embeddings: False
                rope_theta: 1000000.
                attention_bias: False
                use_flash_attention: True
                add_bias_linear: False
                eos_token_id: 151645
                pad_token_id: 151643
                bos_token_id: 151643
                attention_dropout: 0.0
                # Configurations from MindFormers
                hidden_dropout: 0.0
                input_sliced_sig: True
                untie_embeddings_and_output_weights: True
                position_embedding_type: "rope"
                qk_layernorm: True
                use_contiguous_weight_layout_attention: False
                qkv_concat: True
                <strong>offset: [0]</strong>
                params_dtype: "float32"
                compute_dtype: "bfloat16"
                layernorm_compute_dtype: "float32"
                softmax_compute_dtype: "float32"
                rotary_dtype: "float32"
                residual_dtype: "float32"
                model_type: "qwen3"
                architectures: ["Qwen3ForCausalLM"]</pre>

        2. （可选）使用临终CKPT的场景，在保存CKPT后通过Pod级别重调度加载CKPT，需修改如下配置字段。

            首次拉起必须保证“load\_checkpoint”参数值的目录下存在正常可用的CKPT或该目录为空，否则可能导致训练无法正常拉起。

            ```Yaml
            resume_training: True 
            src_strategy_path_or_dir: './output/strategy'
            load_checkpoint: './output/checkpoint'
            ```

    3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

**强化学习后训练场景适配示例（基于Verl）<a name="section1335017512276"></a>**

MindCluster仅支持Verl框架Job级别和Pod级别重调度，其中Pod级别重调度仅支持GRPO算法。Verl的训练任务被Ray集群所管理，为适配MindCluster的Ascend 
Job任务部署，每个Worker节点上部署一个Pod，Pod内承载该Ray集群上的所有进程。Ray集群的head节点为Master Pod，head节点启动Ray集群，其他Worker Pod启动后加入Ray集群，然后由head节点提交任务。

下面以两台Atlas 900 A3 SuperPoD 超节点为例，说明具体操作步骤。

1. 准备Verl代码。

   ```shell
   git clone https://github.com/volcengine/verl.git
   cd verl
   git checkout b97ebfd5062223337ae065c2250f8ab5c0e08e5e
   rm -rf recipe
   git clone https://github.com/verl-project/verl-recipe.git
   cd verl-recipe
   git checkout 474494acafc6482a7e16be2c82e957bd8ca11a3f
   cd ..
   mv verl-recipe recipe
   ```

2. 获取Qwen3-32B模型。获取链接：[Qwen/Qwen3-32B](https://huggingface.co/Qwen/Qwen3-32B/tree/main)

3. 获取及转换gsm8k数据集。获取链接：[openai/gsm8k](https://huggingface.co/datasets/openai/gsm8k/tree/main/main)

   1. 准备数据集目录。

      在某目录（如“/data/dataset/”）下创建gsm8k/main目录。将下载的数据集train-00000-of-00001.parquet和test-00000-of-00001.parquet文件放入“/data/dataset/gsm8k/main”目录下。

   2. 启动容器。

      下面以Verl代码和数据集文件均在“/data”路径为例，具体挂载路径请根据实际修改。

      ```shell
      docker run -it \
      -v /data:/data \
      -v /usr/local/Ascend/driver:/usr/local/Ascend/driver \
      verl:v1 /bin/bash
      ```

   3. 进入代码目录，修改gsm8k.py脚本。

      ```shell
      cd /data/code/verl
      vi examples/data_preprocess/gsm8k.py
      ```

      找到以下代码段：

      ```Python
      if local_dataset_path is not None:
          dataset = datasets.load_dataset(local_dataset_path, "main")
      else:
          dataset = datasets.load_dataset(data_source, "main")
      ```

      修改为：

      ```Python
      if local_dataset_path is not None:
          dataset = datasets.load_dataset(local_dataset_path)
      else:
          dataset = datasets.load_dataset(data_source, "main")
      ```

   4. 执行预处理。

      ```shell
       python3 examples/data_preprocess/gsm8k.py \
       --local_save_dir /data/datasets/gsm8k \
       --local_dataset_path /data/datasets/gsm8k
       ```

4. 准备训练脚本和配置文件。

   1. 获取训练脚本和配置文件。

      从[示例仓库](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/grpo)获取示例训练脚本start_grpo.sh和run_grpo_qwen3_32b_a3b_megatron.sh放置到verl根目录下，获取配置文件runtime_env.yaml放置到“verl/recipe/fault_recover/config”目录下。

   2. 修改训练脚本。

      - 修改start_grpo.sh关键配置：

         ```shell
         # 网卡配置（当hostNetwork=true时根据实际网卡信息修改）
         export HCCL_SOCKET_IFNAME=eth0
         export TP_SOCKET_IFNAME=eth0
         export GLOO_SOCKET_IFNAME=eth0
            
         # 依赖路径（根据实际路径配置）
         export PYTHONPATH=$PYTHONPATH:/data/code/Megatron-LM
        
         # 日志目录（根据实际路径配置）
         export path_log_dir=/data/logs/$MINDX_TASK_ID/trainlog
         export ASCEND_PROCESS_LOG_PATH=/data/logs/$MINDX_TASK_ID/plog
        
         # 内存优化库（根据实际路径配置）
         export LD_PRELOAD=/usr/local/lib/libjemalloc.so.2:$LD_PRELOAD
         ```
      
      - 修改run_grpo_qwen3_32b_a3b_megatron.sh配置：
      
         ```shell
           MODEL_PATH=/data/models/Qwen3-32B                 # 模型路径，根据实际情况修改
           CKPTS_DIR="/data/ckpt/Qwen3-32B-save/"            # 保存checkpoint路径，根据实际情况修改，推荐使用共享存储
           TRAIN_FILE="/data/datasets/gsm8k/train.parquet"   # 数据集路径，根据实际情况修改
           TEST_FILE="/data/datasets/gsm8k/test.parquet"     # 数据集路径，根据实际情况修改
        ```

5. 准备任务YAML，下发任务。

    获取[verl-grpo.yaml](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/grpo)（示例默认配置了Pod级别和Job级别重调度，可根据实际情况自行配置），执行如下启动命令：

    ```shell
    kubectl apply -f verl-grpo.yaml
    ```

    启动后，日志中可能出现类似如下错误信息，此为正常现象，因为Head节点通常未挂载NPU卡。

    ```ColdFusion
    [ERROR] RUNTIME(38734,python3): ... [driver.cc:64]38734 GetDeviceCount:Call drvGetDevNum, drvRetCode=7.
    [ERROR] ASCENDCL(38734,python3): ... aclrtGetDeviceCountImpl:get device count failed, runtime result = 507899.
    [ERROR] APP(38734,python3): ... "[PTA]:"get device count of NPU failed""
    ```

## 准备任务YAML<a name="ZH-CN_TOPIC_0000002511426415"></a>

集群调度组件为用户提供YAML示例，用户需要根据使用的功能、模型类型和任务类型等，并根据使用的故障处理模式，选择相应的YAML示例并根据需求进行相应修改后才可使用。

**表 1**  训练任务YAML示例

<a name="table350244433714"></a>
<table><thead align="left"><tr id="row135031644183710"><th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.1"><p id="p8503244173715"><a name="p8503244173715"></a><a name="p8503244173715"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="16.173234646929384%" id="mcps1.2.8.1.2"><p id="p145038448375"><a name="p145038448375"></a><a name="p145038448375"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="8.521704340868173%" id="mcps1.2.8.1.3"><p id="p919210345266"><a name="p919210345266"></a><a name="p919210345266"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="13.672734546909378%" id="mcps1.2.8.1.4"><p id="p5503544193713"><a name="p5503544193713"></a><a name="p5503544193713"></a>模型</p>
</th>
<th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.5"><p id="p19672186404"><a name="p19672186404"></a><a name="p19672186404"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="15.433086617323463%" id="mcps1.2.8.1.6"><p id="p1096741894013"><a name="p1096741894013"></a><a name="p1096741894013"></a>获取链接</p>
</th>
<th class="cellrowborder" valign="top" width="15.413082616523303%" id="mcps1.2.8.1.7"><p id="p2967518174012"><a name="p2967518174012"></a><a name="p2967518174012"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row4503174412371"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p09365292408"><a name="p09365292408"></a><a name="p09365292408"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul129364297402"></a><a name="ul129364297402"></a><ul id="ul129364297402"><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p319343422611"><a name="p319343422611"></a><a name="p319343422611"></a><span id="ph310231710274"><a name="ph310231710274"></a><a name="ph310231710274"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p493616294406"><a name="p493616294406"></a><a name="p493616294406"></a><span id="ph22631282914"><a name="ph22631282914"></a><a name="ph22631282914"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p893610293406"><a name="p893610293406"></a><a name="p893610293406"></a>pytorch_multinodes_acjob_910b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1987716427402"><a name="p1987716427402"></a><a name="p1987716427402"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml" target="_blank" rel="noopener noreferrer">pytorch_multinodes_acjob_910b.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p8936152964011"><a name="p8936152964011"></a><a name="p8936152964011"></a>示例默认使用2*8卡任务</p>
</td>
</tr>
<tr id="row91607510384"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p89371529174019"><a name="p89371529174019"></a><a name="p89371529174019"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul393742934014"></a><a name="ul393742934014"></a><ul id="ul393742934014"><li><span id="ph139426426441"><a name="ph139426426441"></a><a name="ph139426426441"></a>Atlas 800T A2 训练服务器</span></li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p1319333422617"><a name="p1319333422617"></a><a name="p1319333422617"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p1893752924017"><a name="p1893752924017"></a><a name="p1893752924017"></a><span id="ph234505228"><a name="ph234505228"></a><a name="ph234505228"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p1493742904013"><a name="p1493742904013"></a><a name="p1493742904013"></a><span id="ph153229411739"><a name="ph153229411739"></a><a name="ph153229411739"></a>ms_multinodes_acjob_superpod.yaml</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1637217494110"><a name="p1637217494110"></a><a name="p1637217494110"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/yamls/ms_multinodes_acjob_superpod.yaml" target="_blank" rel="noopener noreferrer">ms_multinodes_acjob_superpod.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p79373296408"><a name="p79373296408"></a><a name="p79373296408"></a>示例默认使用2*16卡任务</p>
</td>
</tr>
<tr id="row4955626202119"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p95874374213"><a name="p95874374213"></a><a name="p95874374213"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><p id="p553054612118"><a name="p553054612118"></a><a name="p553054612118"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p1419311343262"><a name="p1419311343262"></a><a name="p1419311343262"></a>Verl</p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p558710375211"><a name="p558710375211"></a><a name="p558710375211"></a>Qwen3-30B</p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p85871337112114"><a name="p85871337112114"></a><a name="p85871337112114"></a>verl-resche.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1833456152315"><a name="p1833456152315"></a><a name="p1833456152315"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/verl-resche.yaml" target="_blank" rel="noopener noreferrer">verl-resche.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p4587737102119"><a name="p4587737102119"></a><a name="p4587737102119"></a>示例默认使用2*16卡任务</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 
>当前断点续训并未提供Atlas 900 A3 SuperPoD 超节点产品的示例YAML，用户可以在示例YAML中的labels下新增annotations字段即可。示例如下：
>
>```Yaml
>...
>  labels: 
>...
>  annotations:
>    sp-block: "32"   # 逻辑超节点芯片数量，sp-block字段的详细说明，可以参见YAML参数说明。
>...
>```

## 下发任务<a name="ZH-CN_TOPIC_0000002479226548"></a>

示例YAML中，任务部署在default命名空间下。本章节以Pytorch框架为例，下发训练任务。

1. 登录管理节点，进入YAML文件所在路径。
2. 在管理节点执行以下命令，使用YAML下发训练任务。

    ```shell
    kubectl apply -f XXX.yaml
    ```

    例如：

    ```shell
    kubectl apply -f pytorch_multinodes_acjob_910b.yaml
    ```

    回显如下：

    ```ColdFusion
    configmap/reset-config-default-test-pytorch created
    ascendjob.mindxdl.gitee.com/default-test-pytorch created
    ```

## 查看任务进程<a name="ZH-CN_TOPIC_0000002511426461"></a>

训练任务下发成功后，训练任务就可正常运行。可通过如下内容查看训练任务运行情况。

**查看所有训练任务<a name="section16792164211375"></a>**

查看当前节点上运行的所有训练任务，操作步骤如下。

1. 登录管理节点，进入YAML文件所在路径。
2. 执行以下命令，查看训练任务运行情况。

    ```shell
    kubectl get pods -A -o wide
    ```

    回显示例如下。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node1          <none>           <none>
    default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
    ……
    ```

**查看单个Pod的训练任务<a name="zh-cn_topic_0000001621551937_section1141119143319"></a>**

查看其中一个Pod上运行的训练任务，操作步骤如下。

执行以下命令，查看训练任务运行情况。

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显示例如下，出现loss即表示任务正常运行。

![](../../../figures/scheduling/unnaming-(7).png)

**查看是否存在CKPT文件<a name="section979416428371"></a>**

故障恢复功能是通过参考CKPT文件实现的，用户需要查看存储节点上是否存在CKPT文件。

用户可以等待训练任务运行时间超过用户设置的保存CKPT文件的时间后，查看设置的保存CKPT文件的路径下是否存在周期性CKPT文件，操作步骤如下。

1. 登录存储节点，执行以下命令，进入CKPT文件路径。

    ```shell
    cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/output/ckpt
    ```

2. 执行以下命令，查看当前目录是否存在周期性CKPT文件。

    ```shell
    ll ./
    ```

    回显示例如下，说明存在周期性CKPT文件。

    ```ColdFusion
    total 8
    drwx-xr-x-  18 root root   8192 Jun 22 18:39 iter_0000100
    -rw-r--r--  1 root root    2    Jun 22 18:39 latest_checkpointed_iteration.txt
    ```

3. （可选）如果使用临终遗言，可以在保存CKPT的路径下，执行以下命令，查看当前目录是否存在临终CKPT文件。

    ```shell
    ll ./
    ```

    回显示例如下，说明存在临终CKPT文件。

    ```ColdFusion
    total 8
    drwx-xr-x-  18 root root   8192 Jun 22 15:39 iter_0000009
    -rw-r--r--  1 root root    2    Jun 22 15:39 latest_checkpointed_iteration.txt
    ```

## 查看训练结果<a name="ZH-CN_TOPIC_0000002479386554"></a>

### （可选）构造故障<a name="ZH-CN_TOPIC_0000002511426449"></a>

本章节将指导用户构造简单的故障，包括节点故障、参数面网络故障和业务面故障。

>[!NOTE] 
>构造芯片故障存在安全风险，如需构造请联系华为技术支持工程师处理。

**构造节点故障<a name="section173881558133914"></a>**

通过重启训练节点，模拟节点下电导致节点状态丢失。该故障在节点重启完成后可自动恢复。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，重启该训练节点，模拟节点状态丢失故障。

    ```shell
    reboot
    ```

3. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

4. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

**构造参数面网络故障<a name="section22113033919"></a>**

通过断开NPU网络链路模拟参数面网络故障。NPU网络故障不影响单机训练任务。用户在断开链路后需手动恢复，否则该故障会一直存在。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，构造NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -link -s down
    ```

    >[!NOTE] 
    >device\_id为NPU的ID，可以通过<b>npu-smi info</b>命令查看NPU的ID。

3. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障构造成功。

    ```ColdFusion
    net health status: Fault
    ```

4. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

6. 执行以下命令，恢复NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -cfg recovery
    ```

7. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障已经恢复。

    ```ColdFusion
    net health status: Success
    ```

**构造业务面故障<a name="section9891038124213"></a>**

通过删除训练进程，模拟业务面故障。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，使用训练启动脚本，查询训练进程信息。

    ```shell
    ps -ef | grep python| grep 训练启动脚本.py
    ```

3. 执行以下命令，手动删除PID最小的训练进程。

    ```shell
    kill -9 pid
    ```

4. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-10-16 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

### 重调度模式<a name="ZH-CN_TOPIC_0000002479386534"></a>

**重调度情况<a name="section87441013105513"></a>**

>[!NOTE] 
>当节点发生故障时，Volcano会将该训练任务调度到其他满足条件的节点上继续运行。

登录管理节点，执行以下命令查看训练任务运行情况。

```shell
kubectl get pods -A -o wide
```

故障前，若训练任务调度到了node1和node2上面，当node1节点上发生故障，此时Volcano组件会将node1和node2上训练任务重调度到node2和node3节点上，重调度后回显示例如下。

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node3          <none>           <none>
……
```

**查看其中一个Pod运行情况<a name="section28985295314"></a>**

执行以下命令，查看单个Pod的训练任务运行情况。

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显如下表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

```ColdFusion
2025-09-08 11:34:00.400331 warn 1900637 [77840][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.401841 warn 1900631 [28432][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.402489 warn 1900639 [10928][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.426989 warn 1900627 [98608][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.429141 warn 1900634 [24592][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
(min, max) time across ranks (ms):
    load-checkpoint ................................: (32107.12, 32108.53)
(min, max) time across ranks (ms):
    model-and-optimizer-setup ......................: (32528.79, 32544.35)
    train/valid/test-data-iterators-setup ..........: (72.68, 656.79)
[rank16]:[W908 11:34:01.252908110 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank24]:[W908 11:34:01.254614170 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank17]:[W908 11:34:01.421349990 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank20]:[W908 11:34:01.431165020 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank19]:[W908 11:34:01.431240250 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank30]:[W908 11:34:01.431707980 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
...
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
 [2025-09-08 11:37:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 6932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
 [2025-09-08 11:37:03] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 | 
...
```

**查看任务重调度记录<a name="section97707231547"></a>**

执行如下命令查看任务重调度记录。

```shell
kubectl describe cm -n mindx-dl job-reschedule-reason
```

回显示例如下。

```ColdFusion
Name:         job-reschedule-reason
Namespace:    mindx-dl
Labels:       <none>
Annotations:  <none>
Data
====
recent-reschedule-records:
----
{"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b":{"JobID":"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0908 11:36:10","RescheduleTimeStamp":1759683370,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"default-test-pytorch-worker-0","NodeName":"node2","NodeRankIndex":"1"}]}]}}
Events:  <none>
```

### 优雅容错模式<a name="ZH-CN_TOPIC_0000002511346479"></a>

本章节指导用户查看使用故障处理的优雅容错模式的训练信息。当芯片发生故障时，进程退出后进行优雅容错处理，恢复后重新拉起进程。

**日志说明<a name="section83075820188"></a>**

重新拉起的训练进程的训练日志在“_训练脚本路径_/newlog”中，具体说明如下。

- QWEN3（PyTorch）训练日志：“/data/atlas\_dls/public/code/QWEN3\_for\_PyTorch\_2.7\_code/alllogs”。
- QWEN3（MindSpore）训练日志：“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/alllogs”。

**操作步骤<a name="section25042117188"></a>**

1. 登录管理节点，执行以下命令查看芯片情况。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程占用片上内存，正常训练中。

    ![](../../../figures/scheduling/1-13.png)

2. 故障发生后，执行以下命令查看芯片信息。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已退出，释放片上内存。

    ![](../../../figures/scheduling/2.png)

3. 故障恢复后，执行以下命令查看芯片信息。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已重新拉起占用片上内存，正常训练中。

    ![](../../../figures/scheduling/3.png)

## 删除任务<a name="ZH-CN_TOPIC_0000002479386566"></a>

**操作步骤<a name="section324819211118"></a>**

在下发任务的YAML目录执行以下命令，删除对应的训练任务。

```shell
kubectl delete -f XXX.yaml
```

示例如下：

```shell
kubectl delete -f pytorch_multinodes_acjob_910b.yaml
```

回显示例如下：

```ColdFusion
configmap "reset-config-default-test-pytorch" deleted
ascendjob.mindxdl.gitee.com "default-test-pytorch" deleted
```

## 运行维护<a name="ZH-CN_TOPIC_0000002479386520"></a>

**前提条件<a name="section18751194535314"></a>**

此功能只适用于特定场景下，用户需要使用重调度功能，且Ascend Device Plugin的启动YAML中已设置autoStowing参数为false。

**操作方法<a name="section8557331115714"></a>**

- 用户可以使用以下命令，将健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-Recover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-Recover**”标签，该标签中的芯片会重新放入资源池中供程序调度。

    >[!NOTE] 
    >该命令仅作清除Recover标签信息使用，请不要用于添加标签。

- 用户可以使用以下命令，将参数面网络健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-NetworkRecover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-NetworkRecover**”标签，同时也会清除“**huawei.com/Ascend910-NetworkUnhealthy**”中对应的芯片。

    >[!NOTE] 
    >该命令仅作清除NetworkRecover标签信息使用，请不要用于添加标签。
