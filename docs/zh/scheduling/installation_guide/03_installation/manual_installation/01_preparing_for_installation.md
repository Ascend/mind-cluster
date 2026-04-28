# 安装前准备<a name="ZH-CN_TOPIC_0000002479386432"></a>

## 准备镜像<a name="ZH-CN_TOPIC_0000002479226488"></a>

用户可通过以下两种方式准备镜像，获取镜像后依次为安装的相应组件创建节点标签、创建用户、创建日志目录和创建命名空间。

- （推荐）[制作镜像](#section106851195114)。本章节以Ascend Operator为例，介绍了制作集群调度组件容器部署时所需镜像的操作步骤。软件包中的Dockerfile仅作为参考，用户可基于本示例制作定制化镜像。

- [从昇腾镜像仓库拉取镜像](#section133861705416)。用户可以从镜像仓库获取制作好的集群调度各组件的镜像。

>[!NOTE]
>
>- 拉取或者制作镜像完成后，请及时进行安全加固，如修复基础镜像的漏洞、安装第三方依赖导致的漏洞等。
>- 在K8s所使用的容器运行时中导入镜像。如K8s  1.24以上版本默认使用Containerd作为容器运行时，拉取或者制作完镜像后需要将镜像导入到Containerd中。
>- NPU Exporter和Ascend Device Plugin的运行用户为root，在对应的Dockerfile中配置了LD\_LIBRARY\_PATH环境变量，其中的值包含了驱动库的相关路径。组件运行时会使用到其中的文件，建议驱动安装时指定的运行用户为root，避免用户不一致带来的提权风险。

**制作镜像<a name="section106851195114"></a>**

1. 在[获取软件包](./00_obtaining_software_packages.md)章节，获取需要安装的集群调度组件软件包。
2. 将软件包解压后，上传到制作镜像服务器的任意目录。以Ascend Operator为例，放到“/home/ascend-operator”目录，目录结构如下。

    ```shell
    root@node:/home/ascend-operator# ll
    total 41388
    drwxr-xr-x 2 root root     4096 Aug 26 20:20 ./
    drwxr-xr-x 6 root root     4096 Aug 26 20:20 ../
    -r-x------ 1 root root 41992192 Aug 26 02:02 ascend-operator*
    -r-------- 1 root root   372291 Aug 26 02:02 ascend-operator-v{version}.yaml
    -r-------- 1 root root      482 Aug 26 02:02 Dockerfile
    ```

    >[!NOTE]
    >NPU Exporter和Ascend Device Plugin若以容器化的形式部署在Atlas 200I SoC A1 核心板上，需要进行如下操作。
    >1. 在制作镜像时检查宿主机HwHiAiUser、HwDmUser、HwBaseUser用户的UID和GID，并记录该GID和UID的取值。
    >2. 查看在Dockerfile-310P-1usoc中创建HwHiAiUser、HwDmUser、HwBaseUser用户时指定的GID和UID是否与宿主机的一致。如果一致则不做修改；如果不一致，请手动修改Dockerfile-310P-1usoc文件使其保持一致，同时需要保证每台宿主机上HwHiAiUser、HwDmUser、HwBaseUser用户的GID和UID的取值一致。

3. 检查制作集群调度组件镜像的节点是否存在如下基础镜像。

    - 执行**docker images | grep ubuntu**命令检查Ubuntu镜像，ARM架构和x86\_64架构镜像大小有差异。

        ```ColdFusion
        ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
        ```

    - 如果需要安装Volcano，则需要检查alpine镜像是否存在。执行**docker images | grep alpine**命令检查，回显示例如下，ARM架构和x86\_64架构镜像大小有差异。

        ```ColdFusion
        alpine            latest              a24bb4013296        2 years ago         5.57MB
        ```

    若上述基础镜像不存在，使用[表1](#table17241135718196)中相关命令拉取基础镜像（拉取镜像需要服务器能访问互联网）。

    **表 1**  获取基础镜像命令

    <a name="table17241135718196"></a>
    <table><thead align="left"><tr><th class="cellrowborder" valign="top" width="20%"><p>基础镜像</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%"><p>拉取镜像命令</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%"><p>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1"><p>ubuntu:22.04</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.2 "><pre class="screen">docker pull ubuntu:22.04</pre>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.3 "><p>拉取时自动识别系统架构。</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1"><p>alpine:latest</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.2 "><ul><li>x86_64架构<pre class="screen">docker pull alpine:latest</pre></li><li>ARM架构<pre class="screen">docker pull arm64v8/alpine:latest
   docker tag arm64v8/alpine:latest alpine:latest</pre></li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.3 "><p>-</p>
    </td>
    </tr>
    </tbody>
    </table>
    
4. 进入组件解压目录，执行**docker build**命令制作镜像，命令参考如下[表2](#table998719467243)。

    **表 2**  各组件镜像制作命令

    <a name="table998719467243"></a>
    <table><thead align="left"><tr id="row4988174618246"><th class="cellrowborder" valign="top" width="12.941294129412938%" id="mcps1.2.5.1.1"><p id="p14926203952810"><a name="p14926203952810"></a><a name="p14926203952810"></a>节点产品类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="13.081308130813083%" id="mcps1.2.5.1.2"><p id="p09883468245"><a name="p09883468245"></a><a name="p09883468245"></a>组件名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.76547654765477%" id="mcps1.2.5.1.3"><p id="p998884619247"><a name="p998884619247"></a><a name="p998884619247"></a>镜像制作命令</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.21192119211921%" id="mcps1.2.5.1.4"><p id="p438416952520"><a name="p438416952520"></a><a name="p438416952520"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row2098819467246"><td class="cellrowborder" valign="top" width="12.941294129412938%" headers="mcps1.2.5.1.1 "><p id="p179024214293"><a name="p179024214293"></a><a name="p179024214293"></a>其他产品</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="13.081308130813083%" headers="mcps1.2.5.1.2 "><p id="p34169197258"><a name="p34169197258"></a><a name="p34169197258"></a><span id="ph36246385212"><a name="ph36246385212"></a><a name="ph36246385212"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="54.76547654765477%" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen3237730141519"><a name="screen3237730141519"></a><a name="screen3237730141519"></a>docker build --no-cache -t ascend-k8sdeviceplugin:<em id="i02419301157"><a name="i02419301157"></a><a name="i02419301157"></a>{</em><em id="i133991029173612"><a name="i133991029173612"></a><a name="i133991029173612"></a>tag}</em> ./</pre>
    </td>
    <td class="cellrowborder" rowspan="9" valign="top" width="19.21192119211921%" headers="mcps1.2.5.1.4 "><p id="p10280193431010"><a name="p10280193431010"></a><a name="p10280193431010"></a><em id="i472612293915"><a name="i472612293915"></a><a name="i472612293915"></a>{tag}</em>需要参考软件包上的版本。如：软件包上版本为<span id="ph18653133316811"><a name="ph18653133316811"></a><a name="ph18653133316811"></a>26.0.0</span>，则<em id="i1572610273910"><a name="i1572610273910"></a><a name="i1572610273910"></a>{tag}</em>为v<span id="ph205239348813"><a name="ph205239348813"></a><a name="ph205239348813"></a>26.0.0</span>。</p>
    <div class="note" id="note1217913258443"><a name="note1217913258443"></a><a name="note1217913258443"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p11793259444"><a name="p11793259444"></a><a name="p11793259444"></a>请确保Dockerfile-310P-1usoc中HwDmUser和HwBaseUser的<span id="ph18833164913291"><a name="ph18833164913291"></a><a name="ph18833164913291"></a>GID</span>和<span id="ph5530185193011"><a name="ph5530185193011"></a><a name="ph5530185193011"></a>UID</span>与物理机上的保持一致。</p>
    </div></div>
    <p id="p7733142881719"><a name="p7733142881719"></a><a name="p7733142881719"></a></p>
    </td>
    </tr>
    <tr id="row11961911142910"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1519601142915"><a name="p1519601142915"></a><a name="p1519601142915"></a><span id="ph138789131469"><a name="ph138789131469"></a><a name="ph138789131469"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen11251535101518"><a name="screen11251535101518"></a><a name="screen11251535101518"></a>docker build --no-cache -t<strong id="b412563510158"><a name="b412563510158"></a><a name="b412563510158"></a> </strong>ascend-k8sdeviceplugin:<em id="i14896103963618"><a name="i14896103963618"></a><a name="i14896103963618"></a>{</em><em id="i108961395368"><a name="i108961395368"></a><a name="i108961395368"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row098844612415"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3927139182817"><a name="p3927139182817"></a><a name="p3927139182817"></a>其他产品</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.5.1.2 "><p id="p114161919102520"><a name="p114161919102520"></a><a name="p114161919102520"></a><span id="ph5113121424115"><a name="ph5113121424115"></a><a name="ph5113121424115"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen194843931520"><a name="screen194843931520"></a><a name="screen194843931520"></a>docker build --no-cache -t npu-exporter:<em id="i1233412449361"><a name="i1233412449361"></a><a name="i1233412449361"></a>{</em><em id="i16334174433615"><a name="i16334174433615"></a><a name="i16334174433615"></a>tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row435991410290"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6359161411292"><a name="p6359161411292"></a><a name="p6359161411292"></a><span id="ph1257419163460"><a name="ph1257419163460"></a><a name="ph1257419163460"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen18159134401518"><a name="screen18159134401518"></a><a name="screen18159134401518"></a>docker build --no-cache -t<strong id="b416024416154"><a name="b416024416154"></a><a name="b416024416154"></a> </strong>npu-exporter:<em id="i1316184923612"><a name="i1316184923612"></a><a name="i1316184923612"></a>{</em><em id="i21616493369"><a name="i21616493369"></a><a name="i21616493369"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row16602529173910"><td class="cellrowborder" rowspan="6" valign="top" headers="mcps1.2.5.1.1 "><p id="p119247391094"><a name="p119247391094"></a><a name="p119247391094"></a>其他产品</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4603162993920"><a name="p4603162993920"></a><a name="p4603162993920"></a><span id="ph2247144612408"><a name="ph2247144612408"></a><a name="ph2247144612408"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen118201953161519"><a name="screen118201953161519"></a><a name="screen118201953161519"></a>docker build --no-cache -t ascend-operator:<em id="i1582195311159"><a name="i1582195311159"></a><a name="i1582195311159"></a>{tag} </em>./</pre>
    </td>
    </tr>
    <tr id="row17988246152414"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1741731972511"><a name="p1741731972511"></a><a name="p1741731972511"></a><span id="ph16157133165316"><a name="ph16157133165316"></a><a name="ph16157133165316"></a>Infer Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen2020115813153"><a name="screen2020115813153"></a><a name="screen2020115813153"></a>docker build --no-cache -t infer-operator:<em id="i1078611616374"><a name="i1078611616374"></a><a name="i1078611616374"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row17988246152414"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1741731972511"><a name="p1741731972511"></a><a name="p1741731972511"></a><span id="ph16157133165316"><a name="ph16157133165316"></a><a name="ph16157133165316"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen2020115813153"><a name="screen2020115813153"></a><a name="screen2020115813153"></a>docker build --no-cache -t resilience-controller:<em id="i1078611616374"><a name="i1078611616374"></a><a name="i1078611616374"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row139888467245"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15417131916251"><a name="p15417131916251"></a><a name="p15417131916251"></a><span id="ph78731053479"><a name="ph78731053479"></a><a name="ph78731053479"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen27324211618"><a name="screen27324211618"></a><a name="screen27324211618"></a>docker build --no-cache -t noded:<em id="i693671211372"><a name="i693671211372"></a><a name="i693671211372"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row273319281179"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1973362871712"><a name="p1973362871712"></a><a name="p1973362871712"></a><span id="ph143563971716"><a name="ph143563971716"></a><a name="ph143563971716"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen134421047161717"><a name="screen134421047161717"></a><a name="screen134421047161717"></a>docker build --no-cache -t clusterd:<em id="i1344219474175"><a name="i1344219474175"></a><a name="i1344219474175"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row1498819461243"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p7417181910258"><a name="p7417181910258"></a><a name="p7417181910258"></a><span id="ph1841103815159"><a name="ph1841103815159"></a><a name="ph1841103815159"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7611881466"><a name="p7611881466"></a><a name="p7611881466"></a>进入<span id="ph11611128154615"><a name="ph11611128154615"></a><a name="ph11611128154615"></a>Volcano</span>组件解压目录，选择以下版本路径并进入。</p>
    <a name="ul1193395714453"></a><a name="ul1193395714453"></a><ul id="ul1193395714453"><li>v1.7.0版本执行以下命令。<pre class="screen" id="screen73221362140"><a name="screen73221362140"></a><a name="screen73221362140"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.7.0 ./ -f ./Dockerfile-controller</pre>
    </li><li>v1.9.0版本执行以下命令。<pre class="screen" id="screen20630163032915"><a name="screen20630163032915"></a><a name="screen20630163032915"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.9.0 ./ -f ./Dockerfile-scheduler
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./ -f ./Dockerfile-controller</pre>
    </li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p966311264620"><a name="p966311264620"></a><a name="p966311264620"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

    以Ascend Operator组件的镜像制作为例，执行<b>docker build --no-cache -t ascend-operator:v\{version\} .</b>命令进行制作，回显示例如下。**注意不要遗漏命令结尾的**“.”。

    ```ColdFusion
    DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
                Install the buildx component to build images with BuildKit:
                https://docs.docker.com/go/buildx/
    Sending build context to Docker daemon  42.37MB
    Step 1/5 : FROM ubuntu:22.04 as build
     ---> 1f37bb13f08a
    Step 2/5 : RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&     usermod root -s /usr/sbin/nologin
     ---> Running in d43f1927b1fd
    Removing intermediate container d43f1927b1fd
     ---> 9f1d64e06ee6
    Step 3/5 : COPY ./ascend-operator  /usr/local/bin/
     ---> 5022b58c516e
    Step 4/5 : RUN chown -R hwMindX:hwMindX /usr/local/bin/ascend-operator  &&    chmod 500 /usr/local/bin/ascend-operator &&    chmod 750 /home/hwMindX &&    echo 'umask 027' >> /etc/profile &&     echo 'source /etc/profile' >> /home/hwMindX/.bashrc
     ---> Running in a781bde3dc56
    Removing intermediate container a781bde3dc56
     ---> 3d7e2ee7a3bd
    Step 5/5 : USER hwMindX
     ---> Running in 338954be8d99
    Removing intermediate container 338954be8d99
     ---> 103f6a2b43a5
    Successfully built 103f6a2b43a5
    Successfully tagged ascend-operator:v{version}
    ```

5. 满足以下场景可以跳过本步骤。

    - 已将制作好的集群调度组件镜像上传到私有镜像仓库，各节点可以通过私有镜像仓库拉取集群调度组件的镜像。
    - 已在安装集群调度组件各节点制作好了组件对应的镜像。

    如不满足上述场景，则需要手动分发各组件镜像到各个节点。以NodeD组件为例，使用离线镜像包的方式，分发镜像到其他节点。

    1. 将制作完成的镜像保存成离线镜像。

        ```shell
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2. 将镜像拷贝到其他节点。

        ```shell
        scp noded-v{version}-linux-aarch64.tar root@{目标节点IP地址}:保存路径
        ```

    3. 以root用户登录各个节点载入离线镜像。

        ```shell
        docker load < noded-v{version}-linux-aarch64.tar
        ```

6. （可选）导入离线镜像到Containerd中。本步骤适用于容器运行时为Containerd场景，其他场景下可跳过。

    以NodeD组件为例，使用离线镜像包的方式，执行以下命令。

    ```shell
    ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
    ```

**从昇腾镜像仓库拉取镜像<a name="section133861705416"></a>**

1. 确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2. <a name="li1381232414410"></a>在左侧导航栏选择任务类型为“集群调度”，然后根据下表选择组件对应的镜像。拉取的镜像需要重命名后才能使用组件启动YAML进行部署，可参考[步骤3](#li14816124549)。

    **表 3**  镜像列表

    <a name="table981217243412"></a>
    <table><thead align="left"><tr id="row1781262416419"><th class="cellrowborder" valign="top" width="28.689999999999998%" id="mcps1.2.5.1.1"><p id="p168129241348"><a name="p168129241348"></a><a name="p168129241348"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="34.43%" id="mcps1.2.5.1.2"><p id="p581214248413"><a name="p581214248413"></a><a name="p581214248413"></a>镜像名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.21%" id="mcps1.2.5.1.3"><p id="p12812122410414"><a name="p12812122410414"></a><a name="p12812122410414"></a>镜像tag</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.67%" id="mcps1.2.5.1.4"><p id="p28136241144"><a name="p28136241144"></a><a name="p28136241144"></a>拉取镜像的节点</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row38132241945"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p138133241142"><a name="p138133241142"></a><a name="p138133241142"></a><span id="ph88139247418"><a name="ph88139247418"></a><a name="ph88139247418"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><a name="ul158133245418"></a><a name="ul158133245418"></a><ul id="ul158133245418"><li><a href="https://www.hiascend.com/developer/ascendhub/detail/54545fa4ff9f446e914bf44b85efdb61" target="_blank" rel="noopener noreferrer">volcanosh/vc-scheduler</a></li><li><a href="https://www.hiascend.com/developer/ascendhub/detail/16f17a3c95d54f9da710a9c51bfceaa3" target="_blank" rel="noopener noreferrer">volcanosh/vc-controller-manager</a></li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p38142241846"><a name="p38142241846"></a><a name="p38142241846"></a>根据需要选择镜像：</p>
    <p id="p1814102416419"><a name="p1814102416419"></a><a name="p1814102416419"></a>v1.7.0-v<span id="ph616117387810"><a name="ph616117387810"></a><a name="ph616117387810"></a>26.0.0</span></p>
    <p id="p9814824342"><a name="p9814824342"></a><a name="p9814824342"></a>v1.9.0-v<span id="ph57147381283"><a name="ph57147381283"></a><a name="ph57147381283"></a>26.0.0</span></p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p18131924748"><a name="p18131924748"></a><a name="p18131924748"></a>管理节点</p>
    <p id="p1081314241741"><a name="p1081314241741"></a><a name="p1081314241741"></a></p>
    </td>
    </tr>
    <tr id="row38143241742"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p128147241147"><a name="p128147241147"></a><a name="p128147241147"></a><span id="ph168144244410"><a name="ph168144244410"></a><a name="ph168144244410"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1381415241342"><a name="p1381415241342"></a><a name="p1381415241342"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/a066319600634cf6a1e522856a63a1c5" target="_blank" rel="noopener noreferrer">ascend-operator</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1881412416419"><a name="p1881412416419"></a><a name="p1881412416419"></a>v<span id="ph19259839285"><a name="ph19259839285"></a><a name="ph19259839285"></a>26.0.0</span></p>
    </td>
    </tr>
    <tr id="row1381419241342"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1814324740"><a name="p1814324740"></a><a name="p1814324740"></a><span id="ph88147247419"><a name="ph88147247419"></a><a name="ph88147247419"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p98151024149"><a name="p98151024149"></a><a name="p98151024149"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/b554929b470747448924bc786b5ab95d" target="_blank" rel="noopener noreferrer">clusterd</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1481592418419"><a name="p1481592418419"></a><a name="p1481592418419"></a>v<span id="ph9804039087"><a name="ph9804039087"></a><a name="ph9804039087"></a>26.0.0</span></p>
    </td>
    </tr>
    <tr id="row138151249410"><td class="cellrowborder" valign="top" width="28.689999999999998%" headers="mcps1.2.5.1.1 "><p id="p1881520248414"><a name="p1881520248414"></a><a name="p1881520248414"></a><span id="ph081511241449"><a name="ph081511241449"></a><a name="ph081511241449"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="34.43%" headers="mcps1.2.5.1.2 "><p id="p1681572413418"><a name="p1681572413418"></a><a name="p1681572413418"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/cc7e6c0a10834f1888d790174fba4bc5" target="_blank" rel="noopener noreferrer">noded</a></p>
    </td>
    <td class="cellrowborder" valign="top" width="17.21%" headers="mcps1.2.5.1.3 "><p id="p108159249411"><a name="p108159249411"></a><a name="p108159249411"></a>v<span id="ph19289104014814"><a name="ph19289104014814"></a><a name="ph19289104014814"></a>26.0.0</span></p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p128156248413"><a name="p128156248413"></a><a name="p128156248413"></a>计算节点</p>
    </td>
    </tr>
    <tr id="row08151024548"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p281518242412"><a name="p281518242412"></a><a name="p281518242412"></a><span id="ph481514241548"><a name="ph481514241548"></a><a name="ph481514241548"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p481512243413"><a name="p481512243413"></a><a name="p481512243413"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/1b1a8c3cc1ff4710bdb0222514a8a7a3" target="_blank" rel="noopener noreferrer">npu-exporter</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p081515241546"><a name="p081515241546"></a><a name="p081515241546"></a>v<span id="ph1878517407813"><a name="ph1878517407813"></a><a name="ph1878517407813"></a>26.0.0</span></p>
    </td>
    </tr>
    <tr id="row1781532410415"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p78163241644"><a name="p78163241644"></a><a name="p78163241644"></a><span id="ph148168241849"><a name="ph148168241849"></a><a name="ph148168241849"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1081612418413"><a name="p1081612418413"></a><a name="p1081612418413"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/a592da7bd2ab4dffa8864abd4eac5068" target="_blank" rel="noopener noreferrer">ascend-k8sdeviceplugin</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p19816132417413"><a name="p19816132417413"></a><a name="p19816132417413"></a>v<span id="ph210911425819"><a name="ph210911425819"></a><a name="ph210911425819"></a>26.0.0</span></p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。

3. <a name="li14816124549"></a>昇腾镜像仓库中拉取的集群调度镜像与组件启动YAML中的名字不一致，需要重命名拉取的镜像后才能启动。根据以下步骤将[步骤2](#li1381232414410)中获取的镜像重新命名，同时建议删除原始名字的镜像。具体操作如下。
    1. 执行以下命令，重命名镜像（用户需根据所使用的组件，选取对应命令执行）。

        ```shell       
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0 ascend-operator:v26.0.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v26.0.0 npu-exporter:v26.0.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0
        
        # 使用1.9.0版本的Volcano，需要将镜像tag修改为v1.9.0-v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0 volcanosh/vc-controller-manager:v1.7.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0 volcanosh/vc-scheduler:v1.7.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0 noded:v26.0.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0 clusterd:v26.0.0
        ```

    2. （可选）执行以下命令，删除原始名字镜像（用户需根据所使用的组件，选取对应命令执行）。

        ```shell
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0
        # 使用1.9.0版本的Volcano，需要将镜像tag修改为v1.9.0-v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0
        ```

4. （可选）导入离线镜像到Containerd中。本步骤适用于容器运行时为Containerd场景，其他场景下可跳过。

    以NodeD组件为例，使用离线镜像包的方式，执行以下步骤。

    1. 将制作完成的镜像保存成离线镜像。

        ```shell
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2. 将离线镜像导入Containerd中。

        ```shell
        ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
        ```

## 创建节点标签<a name="ZH-CN_TOPIC_0000002511426279"></a>

K8s集群中，如果将包含昇腾AI处理器的节点作为K8s的管理节点，此时该节点既是管理节点又是计算节点，除了需要管理节点对应的标签外，还需要根据节点的昇腾AI处理器类型，打上计算节点的相关标签。生产环境中，管理节点一般为通用服务器，不包含昇腾AI处理器。

**操作步骤<a name="section847765415564"></a>**

1. 在任意节点执行以下命令，查询节点名称。

    ```shell
    kubectl get node
    ```

    回显示例如下：

    ```ColdFusion
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2. 按照[表1](#table202738181704)的标签信息，为对应节点打标签，方便集群调度组件在各种不同形态的工作节点之间进行调度。为节点打标签的命令参考如下。

    ```shell
    kubectl label nodes 主机名称 标签
    ```

    以主机名称“ubuntu”，标签“masterselector=dls-master-node”为例，命令参考如下。

    ```shell
    kubectl label nodes ubuntu masterselector=dls-master-node
    ```

    回显示例如下，表示操作成功。

    ```ColdFusion
    node/ubuntu labeled
    ```

    >[!NOTE]
    >- [表1](#table202738181704)中各节点标签的详细说明请参见[K8s原生对象说明](../../../api/k8s.md)章节。
    >- 请按[表1](#table202738181704)，根据节点类型和产品类型，配置所列出的所有标签。
    >- 芯片型号的数值可通过**npu-smi info**命令查询，返回的“Name”字段对应信息为芯片型号，下文的\{_xxx_\}即取“910”字符作为芯片型号数值。

    **表 1**  节点对应的标签信息

    <a name="table202738181704"></a>
    <table><thead align="left"><tr id="row627331819017"><th class="cellrowborder" valign="top" width="31.840000000000003%" id="mcps1.2.4.1.1"><p id="p19273918201"><a name="p19273918201"></a><a name="p19273918201"></a>节点类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="25.96%" id="mcps1.2.4.1.2"><p id="p3273218803"><a name="p3273218803"></a><a name="p3273218803"></a>产品类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="42.199999999999996%" id="mcps1.2.4.1.3"><p id="p19273118301"><a name="p19273118301"></a><a name="p19273118301"></a>标签</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row227451815011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p142747189017"><a name="p142747189017"></a><a name="p142747189017"></a>管理节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p102741181908"><a name="p102741181908"></a><a name="p102741181908"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><p id="p1227417181004"><a name="p1227417181004"></a><a name="p1227417181004"></a>masterselector=dls-master-node</p>
    </td>
    </tr>
    <tr id="row127412189015"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p14274118905"><a name="p14274118905"></a><a name="p14274118905"></a>计算节点</p>
    <p id="p203704324914"><a name="p203704324914"></a><a name="p203704324914"></a></p>
    <p id="p4371534493"><a name="p4371534493"></a><a name="p4371534493"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p627418181808"><a name="p627418181808"></a><a name="p627418181808"></a><span id="ph42747181102"><a name="ph42747181102"></a><a name="ph42747181102"></a>Atlas 800 训练服务器（NPU满配）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul727421813014"></a><a name="ul727421813014"></a><ul id="ul727421813014"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row19274318806"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p742615141511"><a name="p742615141511"></a><a name="p742615141511"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p027411181309"><a name="p027411181309"></a><a name="p027411181309"></a><span id="ph127517181101"><a name="ph127517181101"></a><a name="ph127517181101"></a>Atlas 800 训练服务器（NPU半配）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul22751618203"></a><a name="ul22751618203"></a><ul id="ul22751618203"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=half</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row92751018202"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p554271313169"><a name="p554271313169"></a><a name="p554271313169"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p527551818016"><a name="p527551818016"></a><a name="p527551818016"></a><span id="ph1427511188015"><a name="ph1427511188015"></a><a name="ph1427511188015"></a>Atlas 800T A2 训练服务器</span>或<span id="ph102750181803"><a name="ph102750181803"></a><a name="ph102750181803"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul32752181202"></a><a name="ul32752181202"></a><ul id="ul32752181202"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph12761718301"><a name="ph12761718301"></a><a name="ph12761718301"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row8394133819129"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p1237115354918"><a name="p1237115354918"></a><a name="p1237115354918"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p2039613891219"><a name="p2039613891219"></a><a name="p2039613891219"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul3874134511121"></a><a name="ul3874134511121"></a><ul id="ul3874134511121"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-a3-16-super-pod</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p><span>Atlas 9000 A3 SuperPoD 集群算力系统</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul3874134511121"></a><a name="ul3874134511121"></a><ul id="ul3874134511121"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-a3-8-super-pod</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row69181319336"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p738423163315"><a name="p738423163315"></a><a name="p738423163315"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p1584884715522"><a name="p1584884715522"></a><a name="p1584884715522"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul537611425289"></a><a name="ul537611425289"></a><ul id="ul537611425289"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul13263154872811"></a><a name="ul13263154872811"></a><ul id="ul13263154872811"><li>host-arch=huawei-x86或host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul17911532280"></a><a name="ul17911532280"></a><ul id="ul17911532280"><li>accelerator-type=module-a3-16</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row271845218270"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p188095589274"><a name="p188095589274"></a><a name="p188095589274"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p164951627162819"><a name="p164951627162819"></a><a name="p164951627162819"></a><span id="ph19495127162814"><a name="ph19495127162814"></a><a name="ph19495127162814"></a>Atlas 800I A3 超节点服务器</span></p>
    <p id="p12463112181614"><a name="p12463112181614"></a><a name="p12463112181614"></a><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul16834964293"></a><a name="ul16834964293"></a><ul id="ul16834964293"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul128341660299"></a><a name="ul128341660299"></a><ul id="ul128341660299"><li>host-arch=huawei-x86或host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul168341764299"></a><a name="ul168341764299"></a><ul id="ul168341764299"><li>accelerator-type=module-a3-16</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row42763185011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p16530201015713"><a name="p16530201015713"></a><a name="p16530201015713"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p19276111815011"><a name="p19276111815011"></a><a name="p19276111815011"></a><span id="ph152766181106"><a name="ph152766181106"></a><a name="ph152766181106"></a>Atlas 800I A2 推理服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72766183018"></a><a name="ul72766183018"></a><ul id="ul72766183018"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph2027661812017"><a name="ph2027661812017"></a><a name="ph2027661812017"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1468510421395"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p868624283911"><a name="p868624283911"></a><a name="p868624283911"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p534220145119"><a name="p534220145119"></a><a name="p534220145119"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul19511133318489"></a><a name="ul19511133318489"></a><ul id="ul19511133318489"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph175351194911"><a name="ph175351194911"></a><a name="ph175351194911"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row13277101813019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p356115645715"><a name="p356115645715"></a><a name="p356115645715"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122778182014"><a name="p122778182014"></a><a name="p122778182014"></a><span id="ph3277518801"><a name="ph3277518801"></a><a name="ph3277518801"></a>Atlas 200T A2 Box16 异构子框</span></p>
    <p id="p1993115373112"><a name="p1993115373112"></a><a name="p1993115373112"></a><span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 异构子框</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul15277318601"></a><a name="ul15277318601"></a><ul id="ul15277318601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph52776181604"><a name="ph52776181604"></a><a name="ph52776181604"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-16</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1627716183019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p556216614577"><a name="p556216614577"></a><a name="p556216614577"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p32771718506"><a name="p32771718506"></a><a name="p32771718506"></a><span id="ph162771318306"><a name="ph162771318306"></a><a name="ph162771318306"></a>训练服务器（插<span id="ph4277131818016"><a name="ph4277131818016"></a><a name="ph4277131818016"></a>Atlas 300T 训练卡</span>）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72771181601"></a><a name="ul72771181601"></a><ul id="ul72771181601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=card</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row62791418607"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p45625617576"><a name="p45625617576"></a><a name="p45625617576"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122793182008"><a name="p122793182008"></a><a name="p122793182008"></a>推理服务器（插<span id="ph19279181811010"><a name="ph19279181811010"></a><a name="ph19279181811010"></a>Atlas 300I 推理卡</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul127919181101"></a><a name="ul127919181101"></a><ul id="ul127919181101"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row72822181005"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p165621264571"><a name="p165621264571"></a><a name="p165621264571"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p16282118603"><a name="p16282118603"></a><a name="p16282118603"></a><span id="ph182828181802"><a name="ph182828181802"></a><a name="ph182828181802"></a>Atlas 推理系列产品</span>（除<span id="ph828261816012"><a name="ph828261816012"></a><a name="ph828261816012"></a>Atlas 200I SoC A1 核心板</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul162825182010"></a><a name="ul162825182010"></a><ul id="ul162825182010"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row328212184011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p20562266579"><a name="p20562266579"></a><a name="p20562266579"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p228281818011"><a name="p228281818011"></a><a name="p228281818011"></a><span id="ph928241810010"><a name="ph928241810010"></a><a name="ph928241810010"></a><span id="ph122828181609"><a name="ph122828181609"></a><a name="ph122828181609"></a>Atlas 200I SoC A1 核心板</span></span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul202825181508"></a><a name="ul202825181508"></a><ul id="ul202825181508"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>servertype=soc</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p>Atlas 350 标卡</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3"><ul><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-npu</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

## 创建用户<a name="ZH-CN_TOPIC_0000002511346353"></a>

在对应组件安装的节点上执行以下命令创建用户。

- <a name="li1069651515405"></a>Ubuntu操作系统

    ```shell
    useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
    usermod -a -G HwHiAiUser hwMindX
    ```

- <a name="li19202165424015"></a>CentOS操作系统

    ```shell
    useradd -d /home/hwMindX -u 9000 -m -s /sbin/nologin hwMindX
    usermod -a -G HwHiAiUser hwMindX
    ```

>[!NOTE]
>
>- 其余操作系统创建用户：
>     - 基于Ubuntu操作系统开发的操作系统，参考[Ubuntu操作系统](#li1069651515405)。
>     - 基于CentOS操作系统开发的操作系统，参考[CentOS操作系统](#li19202165424015)。
>- HwHiAiUser是驱动或CANN软件包所需的软件运行用户。
>- 执行**getent passwd**命令，查看所有物理机（存储节点、管理节点、计算节点）和容器内，HwHiAiUser的UID和GID是否一致，且都为1000。如果被占用可能会导致服务不可用，可以参见[用户UID或GID被占用](../../../faq.md#用户uid或gid被占用)章节进行处理。

**表 1**  组件用户说明

<a name="table125971501113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0299839362_row86431704617"><th class="cellrowborder" valign="top" width="20.962096209620963%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0299839362_p464201754614"><a name="zh-cn_topic_0299839362_p464201754614"></a><a name="zh-cn_topic_0299839362_p464201754614"></a>组件</p>
</th>
<th class="cellrowborder" valign="top" width="34.13341334133413%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0299839362_p11647172468"><a name="zh-cn_topic_0299839362_p11647172468"></a><a name="zh-cn_topic_0299839362_p11647172468"></a>启动用户</p>
</th>
<th class="cellrowborder" valign="top" width="44.90449044904491%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0299839362_p56451734620"><a name="zh-cn_topic_0299839362_p56451734620"></a><a name="zh-cn_topic_0299839362_p56451734620"></a>是否使用特权容器</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0299839362_row3641172465"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p671453716107"><a name="p671453716107"></a><a name="p671453716107"></a><span id="ph14925450192719"><a name="ph14925450192719"></a><a name="ph14925450192719"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><a name="ul124012695512"></a><a name="ul124012695512"></a><ul id="ul124012695512"><li>二进制运行：hwMindX</li><li>容器运行：root</li></ul>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><a name="ul8401830195518"></a><a name="ul8401830195518"></a><ul id="ul8401830195518"><li>二进制运行：不涉及。</li><li>容器运行：需要使用特权容器，建议用户使用二进制运行。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row1064121764612"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p16641317134612"><a name="zh-cn_topic_0299839362_p16641317134612"></a><a name="zh-cn_topic_0299839362_p16641317134612"></a><span id="ph522114212719"><a name="ph522114212719"></a><a name="ph522114212719"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p53735269103"><a name="p53735269103"></a><a name="p53735269103"></a>root</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p29286561106"><a name="p29286561106"></a><a name="p29286561106"></a>需要使用特权容器。</p>
</td>
</tr>
<tr id="row10935147171519"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1935947181513"><a name="p1935947181513"></a><a name="p1935947181513"></a><span id="ph5551115391513"><a name="ph5551115391513"></a><a name="ph5551115391513"></a>NodeD</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row664817164615"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p0649177466"><a name="zh-cn_topic_0299839362_p0649177466"></a><a name="zh-cn_topic_0299839362_p0649177466"></a><span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span></p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p153424813128"><a name="p153424813128"></a><a name="p153424813128"></a>hwMindX</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p17327314131212"><a name="p17327314131212"></a><a name="p17327314131212"></a>不涉及。</p>
</td>
</tr>
<tr id="row24141825191817"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1941515259187"><a name="p1941515259187"></a><a name="p1941515259187"></a><span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span></p>
</td>
</tr>
<tr id="row29051413163917"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p390551333913"><a name="p390551333913"></a><a name="p390551333913"></a><span id="ph829115811272"><a name="ph829115811272"></a><a name="ph829115811272"></a>Resilience Controller</span></p>
</td>
</tr>
<tr id="row1674814434406"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p97491434407"><a name="p97491434407"></a><a name="p97491434407"></a><span id="ph1566531814589"><a name="ph1566531814589"></a><a name="ph1566531814589"></a>Infer Operator</span></p>
</td>
</tr>
<tr id="row1674814434406"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p97491434407"><a name="p97491434407"></a><a name="p97491434407"></a><span id="ph1566531814589"><a name="ph1566531814589"></a><a name="ph1566531814589"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row6784854202610"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p11621711181811"><a name="p11621711181811"></a><a name="p11621711181811"></a><span id="ph1790811553279"><a name="ph1790811553279"></a><a name="ph1790811553279"></a>Elastic Agent</span></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p161622011121819"><a name="p161622011121819"></a><a name="p161622011121819"></a>由用户自行决定，建议使用非root用户。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p1916271131815"><a name="p1916271131815"></a><a name="p1916271131815"></a>由用户自行决定，建议不使用特权容器。</p>
</td>
</tr>
<tr id="row315419369301"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p715593611302"><a name="p715593611302"></a><a name="p715593611302"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
</td>
</tr>
<tr id="row3502131311115"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p175021513201117"><a name="p175021513201117"></a><a name="p175021513201117"></a><span id="ph16988102112717"><a name="ph16988102112717"></a><a name="ph16988102112717"></a>Container Manager</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p1450212134110"><a name="p1450212134110"></a><a name="p1450212134110"></a>root</p>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p6502191318116"><a name="p6502191318116"></a><a name="p6502191318116"></a>不涉及。</p>
</td>
</tr>
</tbody>
</table>

## 创建日志目录<a name="ZH-CN_TOPIC_0000002511346417"></a>

在对应节点创建组件日志父目录和各组件的日志目录，并设置目录对应属主和权限。

**操作步骤<a name="section124928122416"></a>**

1. 执行以下命令，按照[表1 集群调度组件日志路径列表](#table957112617314)，在各节点创建组件日志父目录。

    ```shell
    mkdir -m 755 /var/log/mindx-dl
    chown root:root /var/log/mindx-dl
    ```

2. 根据所使用组件的具体情况，创建相应的日志目录。

    **表 1** 集群调度组件日志路径列表

    <a name="table957112617314"></a>
    <table><thead align="left"><tr id="row2057210616310"><th class="cellrowborder" valign="top" width="21.93%" id="mcps1.2.5.1.1"><p id="p10572761231"><a name="p10572761231"></a><a name="p10572761231"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="41.91%" id="mcps1.2.5.1.2"><p id="p11572156430"><a name="p11572156430"></a><a name="p11572156430"></a>创建日志目录命令</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.05%" id="mcps1.2.5.1.3"><p id="p25721364319"><a name="p25721364319"></a><a name="p25721364319"></a>日志路径创建节点</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.11%" id="mcps1.2.5.1.4"><p id="p16572661320"><a name="p16572661320"></a><a name="p16572661320"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row457296131"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p1572469315"><a name="p1572469315"></a><a name="p1572469315"></a><span id="ph9572196532"><a name="ph9572196532"></a><a name="ph9572196532"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1657216638"><a name="screen1657216638"></a><a name="screen1657216638"></a>mkdir -m 750 /var/log/mindx-dl/devicePlugin
   chown root:root /var/log/mindx-dl/devicePlugin</pre>
    </td>
    <td class="cellrowborder" rowspan="5" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p11572661536"><a name="p11572661536"></a><a name="p11572661536"></a>计算节点</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p557592110325"><a name="p557592110325"></a><a name="p557592110325"></a>-</p>
    </td>
    </tr>
    <tr id="row95721761536"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p125721269315"><a name="p125721269315"></a><a name="p125721269315"></a><span id="ph14572161034"><a name="ph14572161034"></a><a name="ph14572161034"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen457213611313"><a name="screen457213611313"></a><a name="screen457213611313"></a>mkdir -m 750 /var/log/mindx-dl/npu-exporter
   chown root:root /var/log/mindx-dl/npu-exporter</pre>
    </td>
    </tr>
    <tr id="row105739620318"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p195731868318"><a name="p195731868318"></a><a name="p195731868318"></a><span id="ph11573862310"><a name="ph11573862310"></a><a name="ph11573862310"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1957396735"><a name="screen1957396735"></a><a name="screen1957396735"></a>mkdir -m 750 /var/log/mindx-dl/noded
   chown root:root /var/log/mindx-dl/noded</pre>
    </td>
    </tr>
    <tr id="row55731961237"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15573961314"><a name="p15573961314"></a><a name="p15573961314"></a><span id="ph13573106431"><a name="ph13573106431"></a><a name="ph13573106431"></a>Elastic Agent</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen55735616314"><a name="screen55735616314"></a><a name="screen55735616314"></a>mkdir -m 750 /var/log/mindx-dl/elastic
   chown <em id="i15731661134"><a name="i15731661134"></a><a name="i15731661134"></a>由用户自行定义</em> /var/log/mindx-dl/elastic</pre>
    <div class="note" id="note3573061032"><a name="note3573061032"></a><a name="note3573061032"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p2057310617318"><a name="p2057310617318"></a><a name="p2057310617318"></a>将<span id="ph1472342453512"><a name="ph1472342453512"></a><a name="ph1472342453512"></a>Elastic Agent</span>日志目录挂载到容器内，详见<a href="../../../usage/resumable_training/06_configuring_the_job_yaml_file.md#任务yaml配置示例">任务YAML配置示例</a>章节中“修改训练脚本、代码的挂载路径”步骤。</p>
    </div></div>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul958614153510"></a><a name="ul958614153510"></a><ul id="ul958614153510"><li>目录属主由用户自定义。注意：安装<span id="ph67093892615"><a name="ph67093892615"></a><a name="ph67093892615"></a>Elastic Agent</span>的用户属组、调用<span id="ph1642075902418"><a name="ph1642075902418"></a><a name="ph1642075902418"></a>Elastic Agent</span>的运行用户属组、挂载宿主机的目录属组请保持一致。</li><li>用户可自定义<span id="ph1790811553279"><a name="ph1790811553279"></a><a name="ph1790811553279"></a>Elastic Agent</span>的运行日志的落盘路径，在该路径下，用户可查看<span id="ph1529820279122"><a name="ph1529820279122"></a><a name="ph1529820279122"></a>Elastic Agent</span>所有节点日志，无需逐一登录每个节点查看。</li></ul>
    </td>
    </tr>
    <tr id="row189638410329"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p7963164113217"><a name="p7963164113217"></a><a name="p7963164113217"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen929012103313"><a name="screen929012103313"></a><a name="screen929012103313"></a>mkdir  -m 750  <em id="i15660102313617"><a name="i15660102313617"></a><a name="i15660102313617"></a>训练脚本目录</em>/taskd_log
   chown <em id="i4956143053617"><a name="i4956143053617"></a><a name="i4956143053617"></a>由用户自行定义</em> <em id="i6187123720366"><a name="i6187123720366"></a><a name="i6187123720366"></a>训练脚本目录</em>/taskd_log </pre>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul9461980353"></a><a name="ul9461980353"></a><ul id="ul9461980353"><li>目录属主由用户自定义。</li><li><span id="ph1524182517352"><a name="ph1524182517352"></a><a name="ph1524182517352"></a>TaskD</span>在运行过程中可以自动创建对应日志目录，日志目录前缀一般为任务YAML中执行<strong id="b5881131073711"><a name="b5881131073711"></a><a name="b5881131073711"></a>bash命令</strong>或拉起训练时所在目录。</li></ul>
    </td>
    </tr>
    <tr id="row65749616319"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p8574136838"><a name="p8574136838"></a><a name="p8574136838"></a><span id="ph13574365316"><a name="ph13574365316"></a><a name="ph13574365316"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen05746613313"><a name="screen05746613313"></a><a name="screen05746613313"></a>mkdir -m 750 /var/log/mindx-dl/ascend-operator
   chown hwMindX:hwMindX /var/log/mindx-dl/ascend-operator</pre>
    </td>
    <td class="cellrowborder" rowspan="6" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p65611868135"><a name="p65611868135"></a><a name="p65611868135"></a>管理节点</p>
    </td>
    <td class="cellrowborder" rowspan="6" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p11355115061313"><a name="p11355115061313"></a><a name="p11355115061313"></a>-</p>
    </td>
    </tr>
    <tr id="row45741461130"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p18574466314"><a name="p18574466314"></a><a name="p18574466314"></a><span id="ph13574176736"><a name="ph13574176736"></a><a name="ph13574176736"></a>Infer Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1574064313"><a name="screen1574064313"></a><a name="screen1574064313"></a>mkdir -m 750 /var/log/mindx-dl/infer-operator
   chown hwMindX:hwMindX /var/log/mindx-dl/infer-operator</pre>
    </td>
    </tr>
    <tr id="row45741461130"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p18574466314"><a name="p18574466314"></a><a name="p18574466314"></a><span id="ph13574176736"><a name="ph13574176736"></a><a name="ph13574176736"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1574064313"><a name="screen1574064313"></a><a name="screen1574064313"></a>mkdir -m 750 /var/log/mindx-dl/resilience-controller
   chown hwMindX:hwMindX /var/log/mindx-dl/resilience-controller</pre>
    </td>
    </tr>
    <tr id="row68981954111810"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p28991454191811"><a name="p28991454191811"></a><a name="p28991454191811"></a><span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen161652618196"><a name="screen161652618196"></a><a name="screen161652618196"></a>mkdir -m 750 /var/log/mindx-dl/clusterd
   chown hwMindX:hwMindX /var/log/mindx-dl/clusterd</pre>
    </td>
    </tr>
    <tr id="row957413616315"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.5.1.1 "><p id="p1657414618311"><a name="p1657414618311"></a><a name="p1657414618311"></a><span id="ph185741164311"><a name="ph185741164311"></a><a name="ph185741164311"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen145741661036"><a name="screen145741661036"></a><a name="screen145741661036"></a>mkdir -m 750 /var/log/mindx-dl/volcano-controller
   chown hwMindX:hwMindX /var/log/mindx-dl/volcano-controller</pre>
    </td>
    </tr>
    <tr id="row18574568314"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><pre class="screen" id="screen1257416635"><a name="screen1257416635"></a><a name="screen1257416635"></a>mkdir -m 750 /var/log/mindx-dl/volcano-scheduler
   chown hwMindX:hwMindX /var/log/mindx-dl/volcano-scheduler</pre>
    </td>
    </tr>
    <tr id="row14307175681213"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p1030717560124"><a name="p1030717560124"></a><a name="p1030717560124"></a><span id="ph172417011305"><a name="ph172417011305"></a><a name="ph172417011305"></a>Container Manager</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen44681417291"><a name="screen44681417291"></a><a name="screen44681417291"></a>mkdir -m 750 /var/log/mindx-dl/container-manager
   chown root:root /var/log/mindx-dl/container-manager</pre>
    </td>
    <td class="cellrowborder" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p53074565125"><a name="p53074565125"></a><a name="p53074565125"></a>需要使用容器恢复特性的节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p1518124119135"><a name="p1518124119135"></a><a name="p1518124119135"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

## 创建命名空间<a name="ZH-CN_TOPIC_0000002479226384"></a>

- 集群调度的NodeD、Resilience Controller、ClusterD、Infer Operator和Ascend Operator组件会运行在K8s的mindx-dl命名空间下，请在K8s的管理节点执行如下命令，创建对应的命名空间。

    ```shell
    kubectl create ns mindx-dl
    ```

- MindCluster上报超节点信息、pingmesh配置信息、公共故障信息需手动创建名为cluster-system命名空间。请在K8s的管理节点执行如下命令。

    ```shell
    kubectl create ns cluster-system
    ```

- NPU Exporter的命名空间为npu-exporter；Volcano的命名空间为volcano-system；Ascend Device Plugin的命名空间为kube-system，上述组件的命名空间由系统创建，用户无需再次创建。
