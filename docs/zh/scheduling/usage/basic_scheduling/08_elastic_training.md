# 弹性训练<a name="ZH-CN_TOPIC_0000002479227142"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002479227148"></a>

>[!NOTE] 
>本章节描述的是基于Resilience Controller组件的弹性训练，该组件已经日落，相关资料将于2026年9月30日的版本删除。最新的弹性训练能力请参见[弹性训练](../resumable_training/01_solutions_principles.md#弹性训练)。

当出现硬件故障，且无备用设备时，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数，重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。

**前提条件<a name="section722033433815"></a>**

- 确保环境中有配置相应的存储方案，比如使用NFS（Network File System），用户可以参见[安装NFS](../../common_operations.md#安装nfs)进行操作。

    NFS需要用户根据使用情况进行目录隔离，NFS的随机读写性能必须能够在15分钟内保存完整的CKPT文件，建议用户使用专业的存储服务器，NFS具体性能要求给出如下参考。

    ![](../../../figures/scheduling/6-2-2-1-折线图.png)

- 在命令行场景下使用弹性训练特性，需要确保已经安装如下组件。
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - Volcano（弹性训练特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    - Ascend Operator
    - NodeD
    - Resilience Controller
    - ClusterD

- 若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。

**使用方式<a name="section1215781619816"></a>**

弹性训练特性的使用方式如下：

- [通过命令行使用](#ZH-CN_TOPIC_0000002511427031)：安装集群调度组件，通过命令行使用弹性训练特性。
- [集成后使用](#ZH-CN_TOPIC_0000002511347077)：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section252320491398"></a>**

- 资源监测可以和训练场景下的所有特性一起使用。
- 集群中同时跑多个训练任务，每个任务使用的特性可以不同。
- 集群调度组件管理的训练节点出现故障（安装昇腾AI处理器并启用NodeD的节点网络故障或者芯片故障）后，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。
- 重调度功能由Kubernetes（简称K8s）配合Volcano或者其他调度器实现。
- 更多说明详见[表1](#table1337017499206)。

    **表 1**  使用说明

    <a name="table1337017499206"></a>
    <table><thead align="left"><tr id="row1537112499205"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="p1737115497204"><a name="p1737115497204"></a><a name="p1737115497204"></a>场景</p>
    </th>
    <th class="cellrowborder" valign="top" width="80%" id="mcps1.2.3.1.2"><p id="p73711249152017"><a name="p73711249152017"></a><a name="p73711249152017"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row12371949152010"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p5371204918208"><a name="p5371204918208"></a><a name="p5371204918208"></a>环境要求</p>
    <p id="p53711949192013"><a name="p53711949192013"></a><a name="p53711949192013"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p317010582217"><a name="p317010582217"></a><a name="p317010582217"></a>需要保证<span id="ph1093220553219"><a name="ph1093220553219"></a><a name="ph1093220553219"></a>K8s</span>集群中各节点时间一致，避免程序误判。</p>
    </td>
    </tr>
    <tr id="row1937117496204"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p132798152215"><a name="p132798152215"></a><a name="p132798152215"></a>用于检测NPU芯片间连通性的IP地址推荐配置为路由器的IP地址。</p>
    </td>
    </tr>
    <tr id="row12371154922014"><td class="cellrowborder" rowspan="3" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p83713498203"><a name="p83713498203"></a><a name="p83713498203"></a>故障处理</p>
    <p id="p16371114912201"><a name="p16371114912201"></a><a name="p16371114912201"></a></p>
    <p id="p20371194932016"><a name="p20371194932016"></a><a name="p20371194932016"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p471123192217"><a name="p471123192217"></a><a name="p471123192217"></a>使用单机多卡进行训练，当出现故障时，优先按照原任务规格进行恢复，且任务规格遵循8、4、2、1卡的恢复策略。</p>
    </td>
    </tr>
    <tr id="row43711949102018"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1611164119224"><a name="p1611164119224"></a><a name="p1611164119224"></a>若<span id="ph23311639112215"><a name="ph23311639112215"></a><a name="ph23311639112215"></a>Resilience Controller</span>在重新调度任务的过程中，该任务出现新的故障，将不再进行处理。</p>
    </td>
    </tr>
    <tr id="row53711449152015"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p45358632316"><a name="p45358632316"></a><a name="p45358632316"></a>若在集群资源有限的场景中，当多个任务同时故障触发重调度，可能会出现由于资源不足而导致任务处于Pending状态。</p>
    </td>
    </tr>
    <tr id="row046671715231"><td class="cellrowborder" rowspan="3" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p1471143519233"><a name="p1471143519233"></a><a name="p1471143519233"></a>特性说明</p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p87431448172312"><a name="p87431448172312"></a><a name="p87431448172312"></a>本特性不适用于虚拟化实例场景。</p>
    </td>
    </tr>
    <tr id="row1526201918238"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p57511456152312"><a name="p57511456152312"></a><a name="p57511456152312"></a>本特性目前支持服务器和芯片间数据并行和混合并行的分布式vcjob类型的训练任务。</p>
    </td>
    </tr>
    <tr id="row16951221172311"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1993291632412"><a name="p1993291632412"></a><a name="p1993291632412"></a>本特性仅支持设备故障和服务器网络故障检测，说明如下：</p>
    <a name="ul175182082413"></a><a name="ul175182082413"></a><ul id="ul175182082413"><li>设备故障支持<span id="ph1914015620494"><a name="ph1914015620494"></a><a name="ph1914015620494"></a>《<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540096" target="_blank" rel="noopener noreferrer">Atlas 中心训练服务器 25.5.0 健康管理故障定义</a>》</span>中DCMI接口上报的<span class="parmvalue" id="parmvalue11232151011244"><a name="parmvalue11232151011244"></a><a name="parmvalue11232151011244"></a>“重执行业务”</span>、<span class="parmvalue" id="parmvalue112321310182410"><a name="parmvalue112321310182410"></a><a name="parmvalue112321310182410"></a>“热复位芯片”</span>和<span class="parmvalue" id="parmvalue1423217104248"><a name="parmvalue1423217104248"></a><a name="parmvalue1423217104248"></a>“隔离芯片”</span>类型的错误。</li><li>设备网络探测工具hccn_tool检测到的设备网络故障；服务器网络故障依赖于<span id="ph1523271015245"><a name="ph1523271015245"></a><a name="ph1523271015245"></a>NodeD</span>组件的节点状态上报机制，<span id="ph1123281019246"><a name="ph1123281019246"></a><a name="ph1123281019246"></a>NodeD</span>未正确安装或者节点间网络不通都会影响该故障检测功能。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

**支持的产品形态<a name="section10503153618487"></a>**

支持Atlas 800 训练服务器产品使用弹性训练。

**使用流程<a name="section9435132545416"></a>**

通过命令行使用弹性训练特性流程可以参见[图1](#fig1445992135513)。

**图 1**  使用流程<a name="fig1445992135513"></a>  
![](../../../figures/scheduling/使用流程-6.png "使用流程-6")

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427031"></a>

### （可选）配置组件<a name="ZH-CN_TOPIC_0000002479227154"></a>

如果用户在安装Ascend Device Plugin和NodeD时，已经配置了弹性训练相关功能，则可以跳过本章节；若没有配置，则需要对组件[MindCluster Ascend Device Plugin](#zh-cn_topic_0000001609393673_section22911654123018)和[MindCluster NodeD](#section4599195414500)进行相关配置才能正常使用本特性。

**配置Ascend Device Plugin<a name="zh-cn_topic_0000001609393673_section22911654123018"></a>**

在重调度策略开启的情况下，Ascend Device Plugin的异常也会触发故障重调度。

1. 修改Ascend Device Plugin组件的启动YAML，修改如下所示加粗部分。

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
                     <strong>-volcanoType=true                    # 重调度场景下必须使用Volcano。</strong>
                     <strong>-autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品</strong>
                     -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒。
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

    如在Atlas 训练系列产品启动该组件，示例如下。

    ```shell
    kubectl apply -f device-plugin-volcano-v26.0.0.yaml
    ```

**配置NodeD<a name="section4599195414500"></a>**

用户可以通过手动修改NodeD的启动YAML来配置节点状态上报间隔。

1. 执行以下命令，编辑NodeD组件的启动YAML文件。

    ```shell
    vi noded-v{version}.yaml
    ```

2. 在YAML文件的“args”行修改“-**reportInterval**”参数，如下所示：

    ```Yaml
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/home/hwMindX/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 -reportInterval=5" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: false
                capabilities:
                  drop: [ "ALL" ]
                runAsUser: 9000
                runAsGroup: 9000
              volumeMounts:
                - name: log-noded
    ...
    ```

### 制作镜像<a name="ZH-CN_TOPIC_0000002511427037"></a>

弹性训练需要训练基础镜像，用户需要根据所使用的训练框架参见[制作镜像](../../common_operations.md#制作镜像)章节进行制作。

>[!NOTE]
>MindSpore框架的[盘古模型](#ZH-CN_TOPIC_0000002479387110)，还需要参考本章继续制作适配盘古模型的镜像。

**前提条件<a name="zh-cn_topic_0272789326_section193545302315"></a>**

请按照[表1](#zh-cn_topic_0272789326_table13971125465512)所示，获取对应操作系统的软件包与打包镜像所需Dockerfile文件与脚本文件，断点续训软件包名称中\{version\}表示版本号。

**表 1**  所需软件

<a name="zh-cn_topic_0272789326_table13971125465512"></a>

|软件包|是否必选|说明|获取方法|
|--|--|--|--|
|mindformers-<em>{version}</em>-py3-none-any.whl|是|MindSpore Transformers套件，构建大模型训练、微调、评估、推理、部署的全流程开发套件。MindSpore的master版本请使用r0.3分支代码版本。|[获取链接](https://gitcode.com/mindspore/mindformers/tree/master)|
|Dockerfile|是|制作镜像需要。|用户根据业务自行准备。|

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 
>本章节以Ubuntu操作系统为例。

**操作步骤<a name="section173381914413"></a>**

1. 以**root**用户登录服务器。
2. 将准备的软件包MindFormers源码上传到服务器任意目录（如“/home/test”）。
3. 执行以下步骤准备Dockerfile文件。
    1. 进入软件包所在目录，执行以下命令创建Dockerfile文件（文件名示例“Dockerfile”）。

        ```shell
        vi Dockerfile
        ```

    2. 请参见[Dockerfile](#zh-cn_topic_0272789326_li104026527188)编写示例，将内容写入Dockerfile文件后执行:wq命令保存内容。

4. 进入软件包所在目录，执行以下命令，构建容器镜像，**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t  [OPTIONS] 镜像名_系统架构:镜像tag .
    ```

    例如：

    ```shell
    docker build -t test_train_arm64:v1.0 .
    ```

    命令解释如[表2](#zh-cn_topic_0272789326_zh-cn_topic_0256378845_table47051919193111)所示。

    **表 2**  命令参数说明

    <a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_table47051919193111"></a>

    |参数|说明|
    |--|--|
    |-t|指定镜像名称。|
    |<em>OPTIONS</em>|“--disable-content-trust”选项：忽略校验，默认开启。出于安全考虑，这里推荐设置关闭。|
    |<em>镜像名</em><em>_系统架构:</em><em>镜像tag</em>|镜像名称与标签，请用户根据实际情况写入。|

    当出现“Successfully built xxx”表示镜像构建成功。

5. 构建完成后，执行以下命令查看镜像信息。

    ```shell
    docker images
    ```

    回显示例如下。

    ```ColdFusion
    REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64          v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

**编写示例<a name="zh-cn_topic_0272789326_section3523631151714"></a>**

使用过程中请根据实际情况修改软件包版本及架构。

1. <a name="zh-cn_topic_0272789326_li104026527188"></a>Dockerfile编写示例。

    - Ubuntu  ARM系统Dockerfile示例。

        ```Dockerfile
        FROM xxx # 基础训练镜像 
        ARG MINDFORMERS_PKG=mindformers-{version}-py3-none-any.whl
        
        WORKDIR /tmp 
        COPY . ./ 
         
        ENV http_proxy xxx
        ENV https_proxy xxx
        
        # 配置Python pip源 
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf
         
        # 安装MindFormers
        RUN pip install $MINDFORMERS_PKG
        
         
        ENV http_proxy "" 
        ENV https_proxy "" 
        
        ```

    - Ubuntu  x86\_64系统Dockerfile示例。

        ```Dockerfile
        FROM xxx # 基础训练镜像 
        ARG MINDFORMERS_PKG=mindformers-{version}-py3-none-any.whl
        
        WORKDIR /tmp 
        COPY . ./ 
         
        ENV http_proxy xxx
        ENV https_proxy xxx
        
        # 配置Python pip源 
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf
        
        # 安装MindFormers
        RUN pip install $MINDFORMERS_PKG
        
         
        ENV http_proxy "" 
        ENV https_proxy "" 
        
        ```

    为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。

### 脚本适配<a name="ZH-CN_TOPIC_0000002479387132"></a>

本章节提供了故障恢复脚本适配示例。用户请根据实际情况选择对应的脚本适配示例。

- ResNet50模型适配
    - [基于PyTorch的故障恢复](#section72859254718)
    - [基于MindSpore的故障恢复](#section127532091511)
    - [基于TensorFlow的故障恢复](#section2352206112211)

- Pangu\_alpha模型适配（MindSpore框架）

    [基于Pangu\_alpha的故障恢复示例](#section1844516123710)

>[!NOTE]  
>下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。

**PyTorch的故障恢复示例<a name="section72859254718"></a>**

1. <a name="li14102111234717"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2. 自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```ColdFusion
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```ColdFusion
        11G
        ```

4. 将[步骤1](#li14102111234717)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch”目录上传至环境，如“/data/atlas\_dls/public/code/”目录。
5. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/pytorch/resnet50”目录中的train\_start.sh、utils.sh和rank\_table.sh文件，在训练代码中创建“scripts”目录，在管理节点构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/#
    scripts/
    ├── rank_table.sh
    ├── utils.sh
    └── train_start.sh
    ```

6. 在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch”路径下修改main.py代码，修改以下加粗内容，改动内容涉及模型保存和加载的逻辑调整。

    <pre codetype="Python">
    import argparse
    <strong>import glob</strong>
    import os
    ...
        if args.resume:
            <strong>candidate_ckpt_path = ""</strong>
            <strong>for p in glob.glob(f"./rank*"):</strong>
                <strong>best_ckpt_path = os.path.join(p, "model_best.pth.tar")</strong>
                <strong>if os.path.exists(best_ckpt_path):</strong>
                    <strong>candidate_ckpt_path = best_ckpt_path</strong>
                    <strong>break</strong>
            <strong>if candidate_ckpt_path:</strong>
                <strong>print("[gpu id:", args.gpu, "]", "=> loading checkpoint '{}'".format(candidate_ckpt_path))</strong>
                <strong># Map model to be loaded to specified single npu.</strong>
                <strong>loc = 'npu:{}'.format(args.gpu)</strong>
                <strong>checkpoint = torch.load(candidate_ckpt_path, map_location=loc)</strong>
                <strong>print(f"load checkpoint to : {loc}")</strong>
                <strong>args.start_epoch = checkpoint['epoch']</strong>
                <strong>best_acc1 = checkpoint['best_acc1']</strong>
                <strong>model.load_state_dict(checkpoint['state_dict'])</strong>
                <strong>optimizer.load_state_dict(checkpoint['optimizer'])</strong>
                <strong>print("[gpu id:", args.gpu, "]", "=> loaded checkpoint '{}' (epoch {})".format(candidate_ckpt_path, checkpoint['epoch']))</strong>
            <strong>else:</strong>
                <strong>print("no valid ckpt found to resume.")</strong>
    ...
            if not args.multiprocessing_distributed or (args.multiprocessing_distributed and args.rank % ngpus_per_node == 0):
                <strong>save_path = f"./rank_{args.rank}"</strong>
                <strong>if not os.path.exists(save_path):</strong>
                    <strong>os.makedirs(save_path, exist_ok=True)</strong>
                save_checkpoint({
                    'epoch': epoch + 1,
                    'arch': args.arch,
                    'state_dict': model.state_dict(),
                    'best_acc1': best_acc1,
                    'optimizer': optimizer.state_dict(),
                <strong>}, is_best, save_path=save_path)</strong>
    ...
    ...
    # 修改原有save_checkpoint函数
    <strong>def save_checkpoint(state, is_best, filename='checkpoint.pth.tar', save_path="./"):</strong>
        <strong>if is_best:</strong>
            <strong>target_path = os.path.join(save_path, 'model_best.pth.tar')</strong>
            <strong>torch.save(state, target_path)</strong>
            <strong>print(f"save ckpt to {target_path} done. Best epoch for now is :{state['epoch']}")</strong></pre>

**MindSpore的故障恢复示例<a name="section127532091511"></a>**

1. 下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支代码，将“models/official/cv/ResNet”目录重命名为“resnet”并作为训练代码。
2. 执行以下命令，在管理节点创建代码目录，并上传训练代码到该目录。

    ```shell
    mkdir /data/atlas_dls/code
    ```

3. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/resnet50”目录中的“train\_start.sh”和“main.sh”文件，结合训练代码中“resnet/scripts”目录，在管理节点构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/resnet/scripts/#
    scripts/
    ├── main.sh
     ...
    ├── run_distribute_train.sh
    ├── run_distribute_train_gpu.sh
    └── train_start.sh
    ```

4. 修改“/data/atlas\_dls/public/code/resnet/scripts”目录下的“train\_start.sh”文件。

    1. 将“dataset\_path”修改为容器内实际的数据集目录。
    2. “config\_yaml\_path”修改为容器内实际的配置文件路径。

    ```shell
    根据实际情况进行修改，全局配置参数：数据集路径，配置参数文件路径；其他模型适配，请根据实际情况增删参数。
    dataset_path=/job/data/imagenet/train
    config_yaml_path=/job/code/resnet/resnet50_imagenet2012_config.yaml
    ```

    train\_start.sh脚本通过调用main.sh脚本启动训练任务。在适配其他模型时，请根据其训练启动脚本（本示例为train.py）的使用指导，调整main.sh脚本中的环境变量配置、启动脚本路径、启动脚本参数。

    ```shell
    # main.sh: 针对本示例（ResNet50模型），用户不需要再修改此脚本；其他模型适配，请根据实际情况，增、删或修改环境变量配置，然后修改训练启动脚本路径和对应的参数，即main.sh脚本中Python命令调用的部分。
    # 本例中，单机单卡的Python命令如下：
    python ${ROOT_PATH}/../train.py --data_path=${DATA_PATH} --config_path=${CONFIG_PATH} 
    # 本例中，单机多卡和分布式的命令如下：
    python ${ROOT_PATH}/../train.py --run_distribute=True --device_num=${RANK_SIZE} --data_path=${DATA_PATH} --config_path=${CONFIG_PATH} 
    ```

5. 修改“/data/atlas\_dls/public/code/resnet/config/”目录的配置文件“resnet50\_imagenet2012\_config.yaml”。模型保存和加载设置，图编译保存和加载设置。

    ```Yaml
    ...
    run_distribute: False
    enable_profiling: False
    data_path: "/cache/data"
    output_dir: "/job/code/output" # 修改checkpoint保存路径，请用户根据实际情况进行修改
    load_path: "/cache/checkpoint_path/"
    device_target: "Ascend"
    checkpoint_path: "./checkpoint/"
    checkpoint_file_path: ""
    ...
    net_name: "resnet50"
    dataset: "imagenet2012"
    device_num: 1
    pre_trained: "/job/code/output/resnet50/imagenet2012/ckpt" # 容器内预训练模型加载路径（支持目录和文件），支持在指定路径下对.ckpt文件进行模糊查找，将搜寻最新的.ckpt文件进行加载，请用户参考训练YAML根据实际情况进行修改
    run_eval: False
    eval_dataset_path: ""
    parameter_server: False
    filter_weight: False
    save_best_ckpt: True
    eval_start_epoch: 40
    ...
    network_dataset: "resnet50_imagenet2012"
    
    
    # 再训练选项 
    save_graphs: False  # 是否开启图编译结果保存
    save_graphs_path: "./graphs" # 图编译结果保存路径
    has_trained_epoch: 0 # 模型预训练的epoch，默认是0
    has_trained_step: 0 # 模型预训练的step，默认是0
    ---
    # 每项配置的帮助说明
    enable_modelarts: "Whether training on modelarts, default: False"
    ...
    batch_size: "Batch size for training and evaluation"
    epoch_size: "Total training epochs."
    checkpoint_path: "The location of the checkpoint file."
    checkpoint_file_path: "The location of the checkpoint file."
    save_graphs: "Whether save graphs during training, default: False."
    save_graphs_path: "Path to save graphs."
    ```

6. resnet代码的启动脚本为train.py，检查train.py中是否存在保存Checkpoint的代码，示例代码如下。

    - 如果存在，则跳过本步骤。
    - 如果不存在，则补充以下保存Checkpoint的代码样例，其中所用参数需要用户在配置文件中定义和设置。其他模型适配，请参考如下片段，根据启动脚本具体内容，添加保存Checkpoint的代码。如有需要，请参考[MindSpore官网](https://www.mindspore.cn/)教程进行修改。

    <pre codetype="Python">
    ...
        # 模型保存代码
        <strong>if config.save_checkpoint:</strong>
            ckpt_append_info = [{"epoch_num": 0, "step_num": 0}]
            config_ck = CheckpointConfig(save_checkpoint_steps=config.save_checkpoint_epochs * step_size,
                                         keep_checkpoint_max=config.keep_checkpoint_max,
                                         append_info=ckpt_append_info)
            <strong>ckpt_cb = ModelCheckpoint(prefix=config.net_name, directory=config.save_ckpt_dir+"_"+str(config.rank_id), config=config_ck)</strong>
            cb += [ckpt_cb]
    ...</pre>

7. resnet代码的启动脚本为train.py，检查train.py中是否存在加载Checkpoint的代码，如果存在，则执行配置完成，进行下一章节操作；否则执行[步骤8](#li1621315181018)。
8. <a name="li1621315181018"></a>在train.py中补充加载Checkpoint的代码。以下为Checkpoint加载样例，其中所用参数需要用户在配置文件中定义和设置。其他模型适配，请参考如下片段，根据启动脚本具体内容，添加加载Checkpoint的代码。如有需要，请参考[MindSpore官网](https://www.mindspore.cn/)教程进行修改。
    1. 修改“src/utils.py”，添加读取epoch代码，加载CKPT后，训练日志中将从CKPT保存时刻所处的epoch开始打印。

        <pre codetype="Python">
        ...
        def init_weight(net, cfg):
            """init_weight"""
            if cfg.pre_trained:
                if not os.path.isfile(cfg.pre_trained):
                    cfg.logger.warning("There is not ckpt file: %s", cfg.pre_trained)
                else:
                    param_dict = ms.load_checkpoint(cfg.pre_trained)
                    if cfg.filter_weight:
                        filter_list = [x.name for x in net.end_point.get_parameters()]
                        filter_checkpoint_parameter_by_list(param_dict, filter_list)
                    ms.load_param_into_net(net, param_dict)
                    <strong>cfg.start_epoch = int(param_dict.get('epoch_num', ms.Tensor(0, ms.int32)).asnumpy().item())</strong>
                    cfg.logger.info("Pre trained ckpt mode: %s loading", cfg.pre_trained)
        ...</pre>

    2. 修改train.py，替换原有的init\_weight函数，使用\_try\_to\_init\_weight尝试加载CKPT文件，避免出现加载到不完整的CKPT，导致训练报错的问题。

        ```Python
        import glob
        ...
        # 找寻pre_trained目录下最新的*.ckpt文件
        def _find_latest_ckpt():
            ckpt_files = glob.glob(config.pre_trained+"*/*.ckpt")
            if ckpt_files:
                ckpt_files.sort(key=os.path.getmtime, reverse=True)
            return ckpt_files
        
        # 尝试加载CKPT文件，尝试次数为INIT_WEIGHT_MAX_ATTEMPTS次
        def _try_to_init_weight(net, config):
            if os.path.isfile(config.pre_trained):
                latest_ckpt = [config.pre_trained]
            else:
                latest_ckpt = _find_latest_ckpt()
        
            if not latest_ckpt:
                config.logger.warning("There is not ckpt file: %s", config.pre_trained)
                return
        
            init_weight_attempts = 0
            INIT_WEIGHT_MAX_ATTEMPTS = 5
            while(latest_ckpt and init_weight_attempts < INIT_WEIGHT_MAX_ATTEMPTS): 
                try:
                    config.pre_trained = latest_ckpt[0]
                    init_weight(net, config)
                    break
                except Exception:
                    config.logger.warning("Pre trained ckpt %s format is incorrect, try to load the last most recent ckpt", config.pre_trained)
                    if latest_ckpt[1:]:
                        latest_ckpt = latest_ckpt[1:]
                        init_weight_attempts+=1
                        continue
                    else:
                        config.logger.error("no more ckpt to load", config.pre_trained)
                        raise ValueError("ckpt format is incorrect, no more ckpt to load, load ckpt failed.")
        
        ...
        @moxing_wrapper()
        def train_net():
            """train net"""
            target = config.device_target
            set_parameter()
            set_output_dir(config)
            config.logger = get_logger(config.log_dir, config.rank_id, config.parameter_server)
            dataset = create_dataset(dataset_path=config.data_path, do_train=True,
                                     batch_size=config.batch_size, train_image_size=config.train_image_size,
                                     eval_image_size=config.eval_image_size, target=target,
                                     distribute=config.run_distribute)
            step_size = dataset.get_dataset_size()
            net = resnet(class_num=config.class_num)
            if config.parameter_server:
                net.set_param_ps()
            # 替换原有的init_weight函数，使用_try_to_init_weight尝试加载CKPT文件，避免加载到不完整的CKPT，导致训练报错
            _try_to_init_weight(net, config)
        
            if config.resume_ckpt:
                resume_param = ms.load_checkpoint(config.resume_ckpt,
                                                  choice_func=lambda x: not x.startswith(('learning_rate', 'global_step')))
                config.start_epoch = int(resume_param.get('epoch_num', ms.Tensor(0, ms.int32)).asnumpy().item())
            lr = ms.Tensor(init_lr(step_size=step_size))
        ...
        ```

**TensorFlow的故障恢复示例<a name="section2352206112211"></a>**

1. <a name="li360413424258"></a>下载[TensorFlow代码仓](https://gitee.com/ascend/ModelZoo-TensorFlow/tree/master/TensorFlow2/built-in/cv/image_classification/ResNet50_ID0360_for_TensorFlow2.X)中master分支中的“ResNet50\_ID0360\_for\_TensorFlow2.X”作为训练代码，请根据该模型代码TensorFlow版本选择训练镜像中的TensorFlow版本包。
2. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet\_TF”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# pwd
        /data/atlas_dls/public/dataset/resnet50/imagenet_TF
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# du -sh
        42G
        ```

3. 在本地解压[步骤1](#li360413424258)中下载的训练代码，将“ModelZoo-TensorFlow-master/TensorFlow2/built-in/cv/image\_classification/”下的“ResNet50\_ID0360\_for\_TensorFlow2.X”目录重命名为“ResNet50\_for\_TensorFlow\_2.6\_code/”目录。
4. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/basic-training/ranktable”目录中的train\_start.sh、utils.sh和rank\_table.sh文件，在训练代码中创建“scripts”目录，在管理节点构造如下的目录结构。

    ```text
    /data/atlas_dls/public/code/ResNet50_for_TensorFlow_2.6_code/
    ├──  scripts
    │   ├──  train_start.sh
    │   ├──  utils.sh
    │   ├──  rank_table.sh
    │    ...
    ```

5. 修改训练代码。补充加载CKPT文件时的日志打印。修改"tensorflow/tf2\_common/training/controller.py"。

    <pre codetype="Python">
    class Controller(object):
      """Class that facilitates training and evaluation of models."""
      def __init__(
        ...
        # Restore Model if needed.
        if self.checkpoint_manager is not None:
          model_restored = self._restore_model()
          <strong>logging.info("loading checkpoint %s", model_restored)</strong>
          if not model_restored and self.checkpoint_manager.checkpoint_interval:
            # If the model is not restored from a checkpoint, save an initial
            # checkpoint.
            ckpt_path = self.checkpoint_manager.save(
                checkpoint_number=self.global_step)
            logging.info("Saved checkpoints in %s", ckpt_path)
        # Create and initialize the interval triggers.
        self.eval_trigger = utils.IntervalTrigger(self.eval_interval,
                                                  self.eval_offset)</pre>

**Pangu\_alpha模型适配示例<a name="section1844516123710"></a>**

1. 下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/nlp/Pangu_alpha)中master分支代码，将“models/official/nlp/Pangu\_alpha”目录重命名为“pangu\_alpha”并作为训练代码，使用该版本模型脚本需保证在镜像中安装的MindSpore版本不低于2.0.0，并且安装mindformers组件。
2. 执行以下命令，在管理节点创建代码目录。

    ```shell
    mkdir /data/atlas_dls/code
    ```

3. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu\_alpha”目录中的“train\_start.sh”和“main.sh”文件，结合训练代码中“pangu\_alpha/scripts”目录，在管理节点构造如下的目录结构。对于盘古百亿模型，使用“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu\_alpha\_13B”目录中的对应文件。

    ```text
    root@ubuntu:/data/atlas_dls/code/pangu_alpha/scripts/# 
    scripts/
    ├── main.sh
    ├── run_cluster_export.sh
    ├── run_distribute_eval_gpu.sh
    ├── run_distribute_eval.sh
     ...
    ├── run_distribute_train.sh
    ├── run_standalone_eval.sh
    ├── run_standalone_export.sh
    ├── run_standalone_predict.sh
    └── train_start.sh
    ```

4. 修改“/data/atlas\_dls/code/pangu\_alpha/scripts”目录下的“train\_start.sh”文件，将“dataset”修改为容器内实际的数据集目录。

    ```shell
    ...
    # 训练数据集路径，根据实际情况修改
    # 安全提示，涉及对路径和输入参数的校验
    dataset="/job/data/train_data"
    
    # 设置训练环境变量
    set_env
    
    # 单节点训练场景
    if [[ "$server_count" == "1" ]]; then
        server_id=0
        if [ ${device_count} -lt 8 ]; then
            echo "Less than 8 card training is not supported for pangu alpha model." | tee log
        fi
        if [ ${device_count} -eq 8 ]; then
            bash main.sh ${device_count} ${server_count} ${RANK_TABLE_FILE} ${server_id} ${dataset}
        fi
    
    # 分布式训练场景
    else
        server_id=$(get_server_id)
        if [ $? -eq 1 ];then
            echo "get server id failed."
            exit 1
        fi
        echo "server id is: "${server_id}
        bash main.sh ${device_count} ${server_count} ${RANK_TABLE_FILE} ${server_id} ${dataset}
    
    ```

5. 百亿及以下模型可跳过该步骤。训练千亿模型时，期望恢复时间小于5min，需要进行额外脚本适配。下文以[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/nlp/Pangu_alpha)中pangu\_alpha的master分支为例（**已完成弹性训练任务配置和脚本适配**）。
    1. 修改“src/pangu\_alpha\_config.py”文件，主要涉及三个参数的更改：args\_opt.num\_layers、args\_opt.stage\_num、args\_opt.micro\_size。

        ```Python
        def set_parse_200B(args_opt):
            """
                Set config for 200B mode
            """
            args_opt.embedding_size = 16384
            args_opt.num_layers = 32                 # 模型层次
            args_opt.num_heads = 128
            if args_opt.per_batch_size == 0:
                args_opt.per_batch_size = 1
            args_opt.word_emb_dp = 0
            if args_opt.run_type == "train":
                args_opt.start_lr = 6e-5
                args_opt.end_lr = 6e-6
               args_opt.stage_num = 8               # 流水线阶段的数量
               args_opt.micro_size = 16             # 流水线并行模式下的微批次大小，其取值应大于args_opt.stage_num
                args_opt.op_level_model_parallel_num = 16
                if args_opt.optimizer_shard = 1:
                    args_opt.op_level_model_parallel_num = 8
            elif args_opt.run_type == "predict":
                args_opt.stage_num = 4
                args_opt.micro_size = 1
                args_opt.op_level_model_parallel_num = 16
                if args_opt.optimizer_shard == 1:
                    args_opt.op_level_model_parallel_num = 8
        ```

    2. 此外，需要指定或者直接修改“src/utils.py”中的“micro\_batch\_interleaved”参数为“1”（请参考train.py脚本的“run\_train\_pipeline”函数中“stage\_device\_num”、“data\_parallel\_num”、“batch\_size”、“micro\_batch\_interleaved”之间的计算关系。最终结果需要满足“PanguAlphaConfig”的“batch\_size”值是“TransformerOpParallelConfig”的“data\_parallel”的倍数）。

6. pangu代码的启动脚本为train.py，检查train.py中是否存在保存Checkpoint的代码，代码示例如下。

    - 如果存在，则跳过本步骤。
    - 如果不存在，则补充以下保存Checkpoint的代码样例，其中所用参数可参照[步骤9](#li13178638874)在配置文件“src/utils.py”中定义和设置。

    ```Python
    ...
    
        # 保存Checkpoint的代码调用
        add_checkpoint_callback_policy(args_opt, callback, rank)
    ...
    # 保存Checkpoint代码定义
    def add_checkpoint_callback_policy(args_param, callback, rank_id):
        r"""
        Add checkpoint policy to callback.
        """
        # 安全提示，涉及对路径和输入参数的校验
        if args_param.save_checkpoint:
            # checkpoint保存epoch_num和step_num info信息
            ckpt_append_info = [{"epoch_num": args_param.has_trained_epoches, "step_num": args_param.has_trained_steps}]
            ckpt_config = CheckpointConfig(save_checkpoint_steps=args_param.save_checkpoint_steps,
                                           keep_checkpoint_max=args_param.keep_checkpoint_max,
                                           integrated_save=False,
                                           append_info=ckpt_append_info
                                           )
    
    
            ckpoint_cb = ModelCheckpoint(prefix=args_param.ckpt_name_prefix + str(rank_id),
                                         directory=os.path.join(args_param.save_checkpoint_path, f"rank_{rank_id}"),
                                         config=ckpt_config)
    
    
            callback.append(ckpoint_cb)
    ...
    ```

7. pangu代码的启动脚本为train.py，检查train.py中是否存在加载Checkpoint的代码，如果存在，则执行[步骤10](#li6181138370)；否则执行[步骤8](#li12175938673)。
8. <a name="li12175938673"></a>在train.py中补充加载checkpoint的代码。以下为Checkpoint加载样例，存在部分加载Checkpoint的代码，需要添加弹性训练特性相关Checkpoint加载代码，其中所用参数可参照[步骤9](#li13178638874)在配置文件“src/utils.py”中定义和设置。

    ```Python
    ...
    # 如果运行的模型没有开启pipeline并行，则修改在以下函数
    def set_parallel_context(args_opt):
    # 如果运行的模型开启pipeline并行，则修改在以下函数
    # 安全提示，涉及对路径和输入参数的校验
    def set_pipeline_parallel_context(args_opt):
    # 在mindspore.set_auto_parallel_context前添加以下代码，请参考[MindSpore文档分布式并行接口说明](https://www.mindspore.cn/tutorials/experts/zh-CN/r2.0/index.html)对set_auto_parallel_context参数的使用说明
            
             
            # 弹性训练中增加内容
            if not os.path.exists(args_opt.strategy_load_ckpt_path):
                args_opt.strategy_load_ckpt_path = ""
    
            # 弹性训练中增加内容，strategy_ckpt_save_file_path参数可以根据容器内路径指定
            strategy_ckpt_save_file_path = '/job/data/code/fault_torlence/pangu_alpha/strategy.ckpt' 
            if args_opt.strategy_load_ckpt_path == strategy_ckpt_save_file_path:
                 strategy_ckpt_save_file_path = '/job/data/code/fault_torlence/pangu_alpha/strategy_new.ckpt'
     
            # 将strategy_ckpt_save_file='strategy.ckpt'修改成strategy_ckpt_save_file=strategy_ckpt_save_file_path，如果set_auto_parallel_context里没有指定strategy_ckpt_save_file参数，则需要手动添加strategy_ckpt_save_file=strategy_ckpt_save_file_path，如下粗体所示
            mindspore.set_auto_parallel_context(
                parallel_mode=args_opt.parallel_mode, gradients_mean=False, search_mode=args_opt.search_mode,
                full_batch=bool(args_opt.full_batch), loss_repeated_mean=True,
                device_num=device_num, enable_parallel_optimizer=bool(args_opt.optimizer_shard),
                pipeline_stages=args_opt.stage_num, enable_alltoall=bool(args_opt.enable_alltoall),
                strategy_ckpt_save_file=strategy_ckpt_save_file_path)
           
    ...
    ...
    # checkpoint加载代码定义
    # 安全提示，涉及对路径和输入参数的校验
    def restore_checkpoint(args_param, sink_size, dataset, model, network, epoch):
        r"""
        Load checkpoint process.
        """
        print("======start single checkpoint", flush=True)
        ckpt_name = args_param.ckpt_name_prefix
        # 为了文档简洁易读, 此处省略了命令行参数save_checkpoint_path和ckpt_name的校验, 请用户自行添加相关校验
        ckpt_pattern = os.path.join(args_param.save_checkpoint_path, "rank_{}".format(D.get_rank()),
                                    f"{ckpt_name}*.ckpt")
        ckpt_all_files = glob.glob(ckpt_pattern)
        if not ckpt_all_files:
            print(f"There is no ckpt file in {args_param.save_checkpoint_path}, "
                  f"current ckpt_files found is {ckpt_files} "
                  f"with pattern {ckpt_pattern}, so skip the loading.")
            return
        ckpt_exp_pattern = os.path.join(
            args_param.save_checkpoint_path,
            "rank_{}".format(D.get_rank()),
            f"{ckpt_name}*_breakpoint.ckpt",
        )
        ckpt_exp_files = glob.glob(ckpt_exp_pattern)
        ckpt_files = []
        for file in ckpt_all_files:
            if file not in ckpt_exp_files:
                ckpt_files.append(file)
    
        if not ckpt_files:
            print(
                f"There is no ckpt file in {args_param.save_checkpoint_path}, "
                f"current ckpt_files found is {ckpt_files} "
                f"with pattern {ckpt_pattern}, so skip the loading."
            )
            return
        ckpt_files.sort(key=os.path.getmtime, reverse=True)
        time_stamp = datetime.datetime.now()
        print(f"time stamp {time_stamp.strftime('%Y.%m.%d-%H:%M:%S')} pre trained ckpt model {ckpt_files} loading",
              flush=True)
        # 加载checkpoint最新文件
        print(f'Start to load from {ckpt_files[0]}')
        param_dict = load_checkpoint(ckpt_files[0])
        if param_dict.get("epoch_num") and param_dict.get("step_num"):
            args_param.has_trained_epoches = int(param_dict["epoch_num"].data.asnumpy())
            args_param.has_trained_steps = int(param_dict["step_num"].data.asnumpy())
        model.build(train_dataset=dataset, sink_size=sink_size, epoch=epoch)
        load_param_into_net(network, param_dict)
    ...
    ```

9. <a name="li13178638874"></a>修改“src/utils.py”文件中的参数。

    ```Python
    ...
        opt.add_argument("--vocab_size",
                          type=int,
                          default=50304, # 根据训练数据集进行修改，此处已修改为样例数据集的取值
                          help="vocabulary size, default is 40000.")
    ...
        opt.add_argument("--data_column_name",
                         type=str,
                         default="text", # 根据数据集定义的字段进行修改，此处已修改为样例数据集的取值
                         help="Column name of datasets")
    ...
        parser.add_argument("--strategy_load_ckpt_path",
                            type=str,
                            default="/job/data/code/fault_torlence/pangu_alpha/strategy/strategy.ckpt", # 弹性训练中，根据用户习惯指定容器内路径，且路径不会被训练覆盖。
                            help="The training prallel strategy for the model.")
        parser.add_argument("--tokenizer_path",
                            type=str,
                            default="./tokenizer_path",
                            help="The path where stores vocab and vocab model file")
    ...
    def add_retrain_params(opt):
        """
        Add parameters about retrain.
        """
        opt.add_argument("--pre_trained",
                         type=str,
                         default="/job/data/code/fault_torlence/pangu_alpha/8p", # 指定预训练模型路径，
                         help="Pretrained checkpoint path.")
        opt.add_argument("--save_checkpoint_path",  
                         type=str,
                         default="/job/data/code/fault_torlence/pangu_alpha/8p",   # 指定模型保存路径
                         help="Save checkpoint path.")
        opt.add_argument("--keep_checkpoint_max", # 指定模型保存策略：最大数量
                         type=int,
                         default=1,
                         help="Max checkpoint save number.")
        opt.add_argument("--save_checkpoint_steps", # 指定模型保存策略：保存间隔
                         type=int,
                         default=20,
                         help="Save checkpoint step number.")
        opt.add_argument("--save_checkpoint", # 指定当次训练是否保存模型
                         type=ast.literal_eval,
                         default=True,
                         help="Whether save checkpoint in local disk.")
        opt.add_argument("--ckpt_name_prefix", # 指定模型保存策略：文件名前缀
                         type=str,
                         default="pangu",
                         help="Saving checkpoint name prefix.")
    ...
    ```

10. <a name="li6181138370"></a>在“/data/atlas\_dls/code/pangu\_alpha”目录下构建空文件“group\_info\_env”。

    ```text
    root@ubuntu:/data/atlas_dls/code/pangu_alpha/# 
    pangu_alpha/
    ├── README.md
    ├── README_CN.md
    ├── group_info_env
     ...
    ├── scripts
    ├── serving_increment
    ├── src
    ├── tasks.py
    └── train.py
     ```

11. 修改train.py文件中的“group\_info\_env”路径。

    ```Python
    ...
        # env variable prepare
        group_info_file = os.getenv("GROUP_INFO_FILE")
        if group_info_file:
            with open(os.path.expanduser("/job/code/group_info_env"), "a") as outfile:
                outfile.write(f"export GROUP_INFO_FILE_REFLECT={group_info_file}\n")
    ...
    ```

### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479227132"></a>

#### 选择YAML示例<a name="ZH-CN_TOPIC_0000002479387110"></a>

集群调度并未专门为用户提供弹性训练的YAML示例，用户可以获取断点续训的YAML并进行修改即可使用。

**表 1**  获取YAML

<a name="table194871213135113"></a>
<table><thead align="left"><tr id="row15488613165116"><th class="cellrowborder" valign="top" width="12.98259651930386%" id="mcps1.2.6.1.1"><p id="p1694212055114"><a name="p1694212055114"></a><a name="p1694212055114"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.2"><p id="p894252065116"><a name="p894252065116"></a><a name="p894252065116"></a>模型</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.3"><p id="p1694216209514"><a name="p1694216209514"></a><a name="p1694216209514"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.4"><p id="p8942122025113"><a name="p8942122025113"></a><a name="p8942122025113"></a>获取链接</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.5"><p id="p7942172025115"><a name="p7942172025115"></a><a name="p7942172025115"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row10488181375112"><td class="cellrowborder" rowspan="4" valign="top" width="12.98259651930386%" headers="mcps1.2.6.1.1 "><p id="p9942132045117"><a name="p9942132045117"></a><a name="p9942132045117"></a>Volcano Job</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.2 "><p id="p1394262015120"><a name="p1394262015120"></a><a name="p1394262015120"></a>ResNet50</p>
</td>
<td class="cellrowborder" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.3 "><p id="p10942152013510"><a name="p10942152013510"></a><a name="p10942152013510"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.4 "><p id="p9942102010518"><a name="p9942102010518"></a><a name="p9942102010518"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/910" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.5 "><p id="p1694222015119"><a name="p1694222015119"></a><a name="p1694222015119"></a>示例默认为单机8卡任务</p>
<p id="p1161014614466"><a name="p1161014614466"></a><a name="p1161014614466"></a></p>
</td>
</tr>
<tr id="row20488131310512"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p92173245117"><a name="p92173245117"></a><a name="p92173245117"></a>a800_pytorch_vcjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p191773377514"><a name="p191773377514"></a><a name="p191773377514"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/train/resumable-training/fault-rescheduling/withRanktable/pytorch/resnet50/yamls/910/a800_pytorch_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row348851319516"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p32203210515"><a name="p32203210515"></a><a name="p32203210515"></a>a800_vcjob.yaml（MindSpore架构）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5177173765117"><a name="p5177173765117"></a><a name="p5177173765117"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/resnet50/yamls/a800_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1561116674613"><a name="p1561116674613"></a><a name="p1561116674613"></a>示例默认为单机单卡任务</p>
</td>
</tr>
<tr id="row16489613125118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p548911315117"><a name="p548911315117"></a><a name="p548911315117"></a>盘古</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19293215114"><a name="p19293215114"></a><a name="p19293215114"></a>a800_vcjob.yaml（MindSpore架构）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p317719373516"><a name="p317719373516"></a><a name="p317719373516"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu_alpha/yamls/a800_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p156115614615"><a name="p156115614615"></a><a name="p156115614615"></a>示例默认为2*8卡任务</p>
</td>
</tr>
</tbody>
</table>

#### YAML参数说明<a name="ZH-CN_TOPIC_0000002479387134"></a>

本章节提供使用弹性训练配置YAML的操作示例。在具体操作前，用户需要了解相关YAML示例的参数说明，再进行操作。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="27.21%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.230000000000004%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="row12812154533917"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p14813194563914"><a name="p14813194563914"></a><a name="p14813194563914"></a>maxRetry</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p6813134514395"><a name="p6813134514395"></a><a name="p6813134514395"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p10813124553917"><a name="p10813124553917"></a><a name="p10813124553917"></a><span id="ph41465155410"><a name="ph41465155410"></a><a name="ph41465155410"></a>Pod</span>删除重启次数，弹性训练需关闭<span id="ph1357741225415"><a name="ph1357741225415"></a><a name="ph1357741225415"></a>Pod</span>重启，需要设置为0。</p>
</td>
</tr>
<tr id="row917012162413"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p871672217415"><a name="p871672217415"></a><a name="p871672217415"></a>minReplicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p14170516144111"><a name="p14170516144111"></a><a name="p14170516144111"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p317081614417"><a name="p317081614417"></a><a name="p317081614417"></a>最小副本数，需要设置为任务需要的最小节点的数量。</p>
</td>
</tr>
<tr id="row20143651155812"><td class="cellrowborder" rowspan="5" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p19309122810151"><a name="p19309122810151"></a><a name="p19309122810151"></a>fault-scheduling</p>
<p id="p1430895155915"><a name="p1430895155915"></a><a name="p1430895155915"></a></p>
<p id="p7308145135911"><a name="p7308145135911"></a><a name="p7308145135911"></a></p>
<p id="p111494487596"><a name="p111494487596"></a><a name="p111494487596"></a></p>
<p id="p1190084975913"><a name="p1190084975913"></a><a name="p1190084975913"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p1058012817249"><a name="p1058012817249"></a><a name="p1058012817249"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1432806105912"><a name="p1432806105912"></a><a name="p1432806105912"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="ph0511135011612"><a name="ph0511135011612"></a><a name="ph0511135011612"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="ph2149141202811"><a name="ph2149141202811"></a><a name="ph2149141202811"></a>Pod</span>。</p>
</td>
</tr>
<tr id="row6949124320594"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p2032819617590"><a name="zh-cn_topic_0000001570873348_p2032819617590"></a><a name="zh-cn_topic_0000001570873348_p2032819617590"></a>force</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" headers="mcps1.2.4.1.2 "><p id="p4784925400"><a name="p4784925400"></a><a name="p4784925400"></a>暂不支持。</p>
<p id="p85061256401"><a name="p85061256401"></a><a name="p85061256401"></a>当前仅支持grace模式。</p>
</td>
</tr>
<tr id="row69001145205918"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p153287615911"><a name="zh-cn_topic_0000001570873348_p153287615911"></a><a name="zh-cn_topic_0000001570873348_p153287615911"></a>off</p>
</td>
</tr>
<tr id="row16149114820595"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p8667174914916"><a name="zh-cn_topic_0000001570873348_p8667174914916"></a><a name="zh-cn_topic_0000001570873348_p8667174914916"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="row2900144905917"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p4602135219491"><a name="zh-cn_topic_0000001570873348_p4602135219491"></a><a name="zh-cn_topic_0000001570873348_p4602135219491"></a>其他值</p>
</td>
</tr>
<tr id="row128861384219"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p11288121310421"><a name="p11288121310421"></a><a name="p11288121310421"></a>elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7288191354217"><a name="p7288191354217"></a><a name="p7288191354217"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1628816134422"><a name="p1628816134422"></a><a name="p1628816134422"></a>开启弹性训练。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在准备训练镜像章节制作或者获取的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：huawei-arm</p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：huawei-x86</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<p id="p533665619531"><a name="p533665619531"></a><a name="p533665619531"></a><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p155013815543"><a name="p155013815543"></a><a name="p155013815543"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，取值如下：</p>
<div class="p" id="p106711919185315"><a name="p106711919185315"></a><a name="p106711919185315"></a><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</div>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p530216266485"><a name="zh-cn_topic_0000001609074269_p530216266485"></a><a name="zh-cn_topic_0000001609074269_p530216266485"></a>请求的NPU数量，请根据实际修改，请求整卡时不能再请求vNPU。</p>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 ">
    <p id="p1294216211553"><a name="p1294216211553"></a><a name="p1294216211553"></a>Atlas 800 训练服务器（NPU满配）取值为：ascend-910</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="row15462632114"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/AscendXXX']</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
</td>
</tr>
<tr id="row1149173454010"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><div class="note" id="note1031313718402"><a name="note1031313718402"></a><div class="notebody"><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE]  
>新任务副本数范围为\[minReplicas, replicas\]，具体数值由当前集群中的可用节点数确定，多节点分布式训练时有效。

#### 配置YAML<a name="ZH-CN_TOPIC_0000002479227138"></a>

**操作步骤<a name="section6131855154814"></a>**

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    使用**弹性训练**特性，参考本配置。以a800\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

    ```Yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
    ...
      labels:
        ring-controller.atlas: ascend-910   # 标识任务使用的芯片的产品类型
    ...
    ---
    apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API
    kind: Job                               # 目前只支持Job类型
    metadata:
      name: mindx-dls-test                  # 任务名，可自定义
      labels:
        ring-controller.atlas: ascend-910    # 标识任务使用的芯片的产品类型
        fault-scheduling: "grace"        # 开启故障重调度
        elastic-scheduling: "on"          # 开启弹性训练，需添加""号
      annotations:
        minReplicas: "1"                 # 最小副本数
    ...
    spec:
      minAvailable: 1                  # 设置为1
    ...
      maxRetry: 0              #设置为0
    ...
      - name: "default-test"
          template:
            metadata:
    ...
            spec:
    ...
              env:
    ...
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
    ...
                resources:  
                  requests:
                    huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8
                  limits:
                    huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
    ...
                nodeSelector:
                  host-arch: huawei-arm       # 可选值，根据实际情况填写
    ...
    ```

2. 使用弹性训练功能，需要扩展内存，请按注释添加参数。此外还要使用“maxRetry”机制，示例如下。

    ```Yaml
    ...
              volumeMounts:                             #弹性训练扩容
              - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

3. 若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m           
                  memory: 100Gi       
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

4. 修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

5. 如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
    - **TensorFlow命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ tensorflow/resnet_ctl_imagenet_main.py --data_dir=/job/data/imagenet_TF --distribution_strategy=one_device --use_tf_while_loop=true --epochs_between_evals=1 --skip_eval --enable_checkpoint_and_export;"
        ...
        ```

    - **PyTorch命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024 --resume=true;"
        ...
        ```

    - 使用**MindSpore架构**的模型，包括ResNet50模型和Pangu\_alpha模型需要跳过此步骤。

6. YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

    ```Yaml
    ...
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ...
            volumes:
    ...
            - name: code
              nfs:
                server: 127.0.0.1        # NFS服务器IP地址
                path: "xxxxxx"           # 配置训练脚本路径
            - name: data
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 配置训练集路径
            - name: output
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```

### 下发任务<a name="ZH-CN_TOPIC_0000002511427035"></a>

**操作步骤<a name="section12502215114011"></a>**

本章节以MindSpore框架的ResNet50模型为例，下发训练任务。

1. 登录管理节点，进入YAML文件所在路径。
2. 在管理节点执行以下命令，使用YAML下发训练任务。

    ```shell
    kubectl apply -f XXX.yaml
    ```

    例如：

    ```shell
    kubectl apply -f a800_vcjob.yaml
    ```

    回显如下：

    ```ColdFusion
    configmap/rings-config-mindx-dls-test created
    job.batch.volcano.sh/mindx-dls-test created
    ```

    >[!NOTE] 
    >如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002479227140"></a>

训练任务下发成功后，训练任务就可正常运行。可通过如下内容查看训练任务运行情况。

**查看所有训练任务<a name="section181299581348"></a>**

查看当前节点上运行的所有训练任务，操作步骤如下。

1. 登录管理节点。
2. 执行以下命令，查看训练任务运行情况。

    ```shell
    kubectl get pods -A -o wide
    ```

    回显示例如下

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          92s     192.168.70.118    ubuntu-155   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Running   0          92s     192.168.185.213   ubuntu-177   <none>           <none>
    ```

### 查看结果<a name="ZH-CN_TOPIC_0000002479387114"></a>

#### 构造故障<a name="ZH-CN_TOPIC_0000002511347079"></a>

用户可以参考本章节构造故障。

**（可选）构造NPU芯片故障<a name="section182989331585"></a>**

通过断开NPU网络链路模拟的参数面网络故障。NPU网络故障不影响单机训练任务。用户在断开链路后需手动恢复，否则该故障会一直存在。

1. 登录计算节点。
2. 执行以下命令，构造NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -link -s down
    ```

    >[!NOTE]  
    >device\_id为NPU的ID，可以通过npu-smi info命令查看NPU的ID。

3. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障构造成功。

    ```ColdFusion
    net health status: Fault
    ```

4. 执行以下命令，恢复NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -cfg recovery
    ```

5. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障已经恢复。

    ```ColdFusion
    net health status: Success
    ```

#### 查看运行结果<a name="ZH-CN_TOPIC_0000002511427041"></a>

当节点发生故障时，Volcano会将该训练任务删除，Resilience Controller根据可用资源修改任务资源需求，Volcano调度到剩余可用资源上继续运行。

**弹性训练情况<a name="section55191324318"></a>**

1. 登录管理节点，执行以下命令查看训练任务运行情况。

    ```shell
    ~# kubectl get pods -A -o wide
    ```

    以全部资源为2节点16卡，下发2节点16卡任务为例，回显示例如下。该回显表示训练任务正常执行时的任务运行情况。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          47s     192.168.70.82   Node-1   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Running   0          47s     192.168.39.9    Node-2     <none>           <none>
    ……
    ```

2. 当Node-1发生NPU网络故障时，Volcano删除任务。执行以下命令查看训练任务终止情况。

    ```shell
     kubectl get pods -A -o wide
    ```

    回显示例如下，表示训练任务被删除。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              0/1     Terminating   0          6m59s     192.168.70.82   Node-1   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Terminating   0          6m59s     192.168.39.9    Node-2     <none>           <none>
    ……
    ```

3. 等待一段时间，执行以下命令查看训练任务弹性伸缩情况。

    ```shell
     kubectl get pods -A -o wide
    ```

    回显示例如下，表示训练任务根据当前可用节点数将2节点16卡任务伸缩为1节点8卡任务。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          107s    192.168.70.86   Node-2   <none>           <none>
    ……
    ```

**查看单个Pod运行情况<a name="section89223312467"></a>**

执行以下命令，查看单个Pod的训练任务运行情况。

```shell
kubectl logs mindx-dls-test-default-test-0 -n vcjob -f
```

- 回显示例如下表示发生故障时，使用最近保存的第39步的Checkpoint文件恢复，实现训练任务第40个epoch开始继续训练。

    ```ColdFusion
    ...
    2023-06-09 22:17:33,441:INFO:--> pre_trained: /job/code/mindspore/output/resnet50/imagenet2012/ckpt_0/resnet50-39_48.ckpt
    2023-06-09 22:17:33,441:INFO:--> run_eval: False
    2023-06-09 22:17:33,441:INFO:--> eval_dataset_path: 
    2023-06-09 22:17:33,441:INFO:--> parameter_server: False
    2023-06-09 22:17:33,441:INFO:--> filter_weight: False
    2023-06-09 22:17:33,441:INFO:--> save_best_ckpt: True
    2023-06-09 22:17:33,441:INFO:--> eval_start_epoch: 40
    2023-06-09 22:17:33,441:INFO:--> eval_interval: 1
    2023-06-09 22:17:33,441:INFO:--> enable_cache: False
    2023-06-09 22:17:33,441:INFO:--> cache_session_id: 
    2023-06-09 22:17:33,441:INFO:--> mode_name: GRAPH
    2023-06-09 22:17:33,441:INFO:--> boost_mode: O0
    2023-06-09 22:17:33,441:INFO:--> conv_init: XavierUniform
    2023-06-09 22:17:33,441:INFO:--> dense_init: TruncatedNormal
    2023-06-09 22:17:33,442:INFO:--> all_reduce_fusion_config: [85, 160]
    2023-06-09 22:17:33,442:INFO:--> train_image_size: 224
    2023-06-09 22:17:33,442:INFO:--> eval_image_size: 224
    2023-06-09 22:17:33,442:INFO:--> device_id: 0
    2023-06-09 22:17:33,442:INFO:--> width: 224
    2023-06-09 22:17:33,442:INFO:--> height: 224
    2023-06-09 22:17:33,442:INFO:--> file_name: resnet50
    2023-06-09 22:17:33,442:INFO:--> file_format: MINDIR
    2023-06-09 22:17:33,442:INFO:--> ckpt_file: 
    2023-06-09 22:17:33,442:INFO:--> network_dataset: resnet50_imagenet2012
    2023-06-09 22:17:33,442:INFO:--> save_graphs: False
    2023-06-09 22:17:33,442:INFO:--> save_graphs_path: ./graphs
    2023-06-09 22:17:33,442:INFO:--> has_trained_epoch: 0
    2023-06-09 22:17:33,442:INFO:--> has_trained_step: 0
    2023-06-09 22:17:33,442:INFO:--> result_path: 
    2023-06-09 22:17:33,442:INFO:--> label_path: 
    2023-06-09 22:17:33,442:INFO:--> config_path: /job/code/mindspore/config/resnet50_imagenet2012_config.yaml
    2023-06-09 22:17:33,442:INFO:--> rank_id: 0
    2023-06-09 22:17:33,442:INFO:--> save_ckpt_dir: /job/code/mindspore/output/resnet50/imagenet2012/ckpt
    2023-06-09 22:17:33,442:INFO:--> log_dir: /job/code/mindspore/output/resnet50/imagenet2012/log
    2023-06-09 22:17:33,442:INFO:--> logger: <LOGGER resnet (NOTSET)>
    2023-06-09 22:17:33,442:INFO:
    [WARNING] DEVICE(312,fffd6e363470,python):2023-06-09-22:17:33.999.925 [mindspore/ccsrc/plugin/device/ascend/hal/hardware/ge_graph_executor.cc:128] RunGEInitGraph] Can not find init_subgraph.kernel_graph_0 subgraph, don't need data init subgraph in INFER mode.
    [WARNING] DEVICE(312,fffd6e363470,python):2023-06-09-22:17:43.733.157 [mindspore/ccsrc/plugin/device/ascend/hal/hardware/ge_graph_executor.cc:128] RunGEInitGraph] Can not find init_subgraph.kernel_graph_1 sub graph, don't need data init subgraph in INFER mode.
    ....2023-06-09 22:18:45,025:INFO:epoch: [40/90] loss: 3.465011, epoch time: 71.582 s, per step time: 1491.285 ms
    2023-06-09 22:18:49,453:INFO:epoch: [41/90] loss: 3.396700, epoch time: 4.428 s, per step time: 92.245 ms
    .2023-06-09 22:19:02,685:INFO:epoch: [42/90] loss: 3.297215, epoch time: 13.232 s, per step time: 275.659 ms
    2023-06-09 22:19:07,323:INFO:epoch: [43/90] loss: 3.289656, epoch time: 4.638 s, per step time: 96.622 ms
    2023-06-09 22:19:11,746:INFO:epoch: [44/90] loss: 3.266534, epoch time: 4.423 s, per step time: 92.139 ms
    2023-06-09 22:19:16,913:INFO:epoch: [45/90] loss: 3.180886, epoch time: 5.167 s, per step time: 107.650 ms
    2023-06-09 22:19:21,377:INFO:epoch: [46/90] loss: 2.895963, epoch time: 4.464 s, per step time: 92.997 ms
    2023-06-09 22:19:25,798:INFO:epoch: [47/90] loss: 2.815258, epoch time: 4.420 s, per step time: 92.090 ms
    2023-06-09 22:19:31,122:INFO:epoch: [48/90] loss: 2.826911, epoch time: 5.324 s, per step time: 110.918 ms
    2023-06-09 22:19:35,591:INFO:epoch: [49/90] loss: 2.712467, epoch time: 4.469 s, per step time: 93.098 ms
    ...
    ```

### 删除任务<a name="ZH-CN_TOPIC_0000002511347063"></a>

1. 登录管理节点，进入YAML文件所在路径。
2. 在管理节点执行以下命令，使用YAML删除训练任务。

    ```shell
    kubectl delete -f XXX.yaml
    ```

    例如：

    ```shell
    kubectl delete -f a800_vcjob.yaml
    ```

    回显如下：

    ```ColdFusion
    configmap/rings-config-mindx-dls-test deleted
    job.batch.volcano.sh/mindx-dls-test deleted
    ```

## 集成后使用<a name="ZH-CN_TOPIC_0000002511347077"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1. 根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2. 根据K8s官方提供的API库，对任务进行创建、查询、删除等操作。
3. 创建、查询或删除操作任务时，用户需要将[示例YAML](#准备任务yaml)的内容转换成K8s官方API中定义的对象，通过官方库中提供的API发送给K8s的API Server或者将YAML内容转换为JSON格式直接发送给K8s的API Server。
