# Installation Preparation<a name="ZH-CN_TOPIC_0000002479386432"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:13:07.840Z pushedAt=2026-06-09T06:22:06.814Z -->

## Preparing an Image<a name="ZH-CN_TOPIC_0000002479226488"></a>

You can prepare images in the following two ways. After obtaining the images, create node labels, users, log directories, and namespaces for the corresponding components to be installed in sequence.

- (Recommended) [Create Images](#section106851195114): Taking Ascend Operator as an example, this section describes the steps for creating images required for container deployment of cluster scheduling components. The Dockerfile in the software package is for reference only, and you can create customized images based on this example.

- [Pull Images from Ascend Image Repository](#section133861705416): You can obtain pre-built images of cluster scheduling components from the image repository.

>[!NOTE]
>
>- After pulling or creating images, perform security hardening in a timely manner, such as fixing vulnerabilities in the base image and those introduced by third-party dependencies.
>- Import the images into the container runtime used by K8s. For example, K8s versions 1.24 and later use Containerd as the default container runtime. After pulling or creating images, you need to import them into Containerd.
>- The running user for NPU Exporter and Ascend Device Plugin is `root`. The `LD_LIBRARY_PATH` environment variable is configured in their corresponding Dockerfiles, and its value includes paths related to driver libraries. The components will use files in these paths at runtime. It is recommended that the running user specified during driver installation be `root` to avoid privilege escalation risks caused by user inconsistency.

**Create Images<a name="section106851195114"></a>**

1. In the [Obtaining Software Packages](./00_obtaining_software_packages.md) chapter, obtain the software packages of the cluster scheduling components that need to be installed.
2. After decompressing the software packages, upload them to any directory on the image creation server. Taking Ascend Operator as an example, place them in the `/home/ascend-operator` directory. The directory structure is as follows.

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
    >If NPU Exporter and Ascend Device Plugin are deployed as images on the Atlas 200I SoC A1 core board, perform the following operations.
    >1. When creating the image, check the UID and GID of the `HwHiAiUser`, `HwDmUser`, and `HwBaseUser` users on the host, and record the values of these GIDs and UIDs.
    >2. Check whether the GID and UID specified when creating the `HwHiAiUser`, `HwDmUser`, and `HwBaseUser` users in `Dockerfile-310P-1usoc` are consistent with those on the host. If they are consistent, no modification is required; if they are inconsistent, manually modify the `Dockerfile-310P-1usoc` file to make them consistent. At the same time, ensure that the GID and UID values of the `HwHiAiUser`, `HwDmUser`, and `HwBaseUser` users are consistent on each host.

3. Check whether the following base images exist on the node used for creating cluster scheduling component images.

    - Run the `docker images | grep ubuntu` command to check the Ubuntu image. The image sizes differ between ARM architecture and x86_64 architecture.

        ```ColdFusion
        ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
        ```

    - If you need to install Volcano, check whether the alpine image exists. Run the `docker images | grep alpine` command to check. The sample output is as follows. The image sizes differ between ARM and x86_64 architectures.

        ```ColdFusion
        alpine            latest              a24bb4013296        2 years ago         5.57MB
        ```

    If the preceding base images do not exist, use the relevant commands in [Table 1](#table17241135718196) to pull the base images (pulling images requires the server to have internet access).

    **Table 1**  Commands for obtaining base images

    <a name="table17241135718196"></a>
    <table><thead align="left"><tr><th class="cellrowborder" valign="top" width="20%"><p>Base Image</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%"><p>Image Pull Command</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%"><p>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1"><p>ubuntu:22.04</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.2 "><pre class="screen">docker pull ubuntu:22.04</pre>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.3 "><p>The system architecture is automatically identified during the pull.</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1"><p>alpine:latest</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.2 "><ul><li>x86_64 architecture<pre class="screen">docker pull alpine:latest</pre></li><li>ARM architecture<pre class="screen">docker pull arm64v8/alpine:latest
   docker tag arm64v8/alpine:latest alpine:latest</pre></li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.3 "><p>-</p>
    </td>
    </tr>
    </tbody>
    </table>

4. Go to the component extraction directory and run the `docker build` command to create the image. For command details, see [Table 2](#table998719467243).

    **Table 2** Image creation commands for each component

    <a name="table998719467243"></a>
    <table><thead align="left"><tr id="row4988174618246"><th class="cellrowborder" valign="top" width="12.941294129412938%" id="mcps1.2.5.1.1"><p id="p14926203952810"><a name="p14926203952810"></a><a name="p14926203952810"></a>Node/Product Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="13.081308130813083%" id="mcps1.2.5.1.2"><p id="p09883468245"><a name="p09883468245"></a><a name="p09883468245"></a>Component</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.76547654765477%" id="mcps1.2.5.1.3"><p id="p998884619247"><a name="p998884619247"></a><a name="p998884619247"></a>Image Creation Command</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.21192119211921%" id="mcps1.2.5.1.4"><p id="p438416952520"><a name="p438416952520"></a><a name="p438416952520"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row2098819467246"><td class="cellrowborder" valign="top" width="12.941294129412938%" headers="mcps1.2.5.1.1 "><p id="p179024214293"><a name="p179024214293"></a><a name="p179024214293"></a>Other Products</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="13.081308130813083%" headers="mcps1.2.5.1.2 "><p id="p34169197258"><a name="p34169197258"></a><a name="p34169197258"></a><span id="ph36246385212"><a name="ph36246385212"></a><a name="ph36246385212"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="54.76547654765477%" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen3237730141519"><a name="screen3237730141519"></a><a name="screen3237730141519"></a>docker build --no-cache -t ascend-k8sdeviceplugin:<em id="i02419301157"><a name="i02419301157"></a><a name="i02419301157"></a>{</em><em id="i133991029173612"><a name="i133991029173612"></a><a name="i133991029173612"></a>tag}</em> ./</pre>
    </td>
    <td class="cellrowborder" rowspan="9" valign="top" width="19.21192119211921%" headers="mcps1.2.5.1.4 "><p id="p10280193431010"><a name="p10280193431010"></a><a name="p10280193431010"></a><em id="i472612293915"><a name="i472612293915"></a><a name="i472612293915"></a>{tag}</em> must reference the version on the software package. For example, if the version on the software package is <span id="ph18653133316811"><a name="ph18653133316811"></a><a name="ph18653133316811"></a>26.0.0</span>, then <em id="i1572610273910"><a name="i1572610273910"></a><a name="i1572610273910"></a>{tag}</em> is v<span id="ph205239348813"><a name="ph205239348813"></a><a name="ph205239348813"></a>26.0.0</span>.</p>
    <div class="note" id="note1217913258443"><a name="note1217913258443"></a><a name="note1217913258443"></a><span class="notetitle">[!NOTE] Description</span><div class="notebody"><p id="p11793259444"><a name="p11793259444"></a><a name="p11793259444"></a>Ensure that the <span id="ph18833164913291"><a name="ph18833164913291"></a><a name="ph18833164913291"></a>GID</span> and <span id="ph5530185193011"><a name="ph5530185193011"></a><a name="ph5530185193011"></a>UID</span> of HwDmUser and HwBaseUser in Dockerfile-310P-1usoc are consistent with those on the physical machine.</p>
    </div></div>
    <p id="p7733142881719"><a name="p7733142881719"></a><a name="p7733142881719"></a></p>
    </td>
    </tr>
    <tr id="row11961911142910"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1519601142915"><a name="p1519601142915"></a><a name="p1519601142915"></a><span id="ph138789131469"><a name="ph138789131469"></a><a name="ph138789131469"></a>Atlas 200I SoC A1 Core Board</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen11251535101518"><a name="screen11251535101518"></a><a name="screen11251535101518"></a>docker build --no-cache -t<strong id="b412563510158"><a name="b412563510158"></a><a name="b412563510158"></a> </strong>ascend-k8sdeviceplugin:<em id="i14896103963618"><a name="i14896103963618"></a><a name="i14896103963618"></a>{</em><em id="i108961395368"><a name="i108961395368"></a><a name="i108961395368"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row098844612415"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3927139182817"><a name="p3927139182817"></a><a name="p3927139182817"></a>Other Products</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.5.1.2 "><p id="p114161919102520"><a name="p114161919102520"></a><a name="p114161919102520"></a><span id="ph5113121424115"><a name="ph5113121424115"></a><a name="ph5113121424115"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen194843931520"><a name="screen194843931520"></a><a name="screen194843931520"></a>docker build --no-cache -t npu-exporter:<em id="i1233412449361"><a name="i1233412449361"></a><a name="i1233412449361"></a>{</em><em id="i16334174433615"><a name="i16334174433615"></a><a name="i16334174433615"></a>tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row435991410290"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6359161411292"><a name="p6359161411292"></a><a name="p6359161411292"></a><span id="ph1257419163460"><a name="ph1257419163460"></a><a name="ph1257419163460"></a>Atlas 200I SoC A1 Core Board</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen18159134401518"><a name="screen18159134401518"></a><a name="screen18159134401518"></a>docker build --no-cache -t<strong id="b416024416154"><a name="b416024416154"></a><a name="b416024416154"></a> </strong>npu-exporter:<em id="i1316184923612"><a name="i1316184923612"></a><a name="i1316184923612"></a>{</em><em id="i21616493369"><a name="i21616493369"></a><a name="i21616493369"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row16602529173910"><td class="cellrowborder" rowspan="6" valign="top" headers="mcps1.2.5.1.1 "><p id="p119247391094"><a name="p119247391094"></a><a name="p119247391094"></a>Other Products</p>
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
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7611881466"><a name="p7611881466"></a><a name="p7611881466"></a>Go to the decompressed directory of the <span id="ph11611128154615"><a name="ph11611128154615"></a><a name="ph11611128154615"></a>Volcano</span> component, select the following version path and enter it.</p>
    <a name="ul1193395714453"></a><a name="ul1193395714453"></a><ul id="ul1193395714453"><li>For v1.7.0, run the following commands.<pre class="screen" id="screen73221362140"><a name="screen73221362140"></a><a name="screen73221362140"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.7.0 ./ -f ./Dockerfile-controller</pre>
    </li><li>For v1.9.0, run the following commands.<pre class="screen" id="screen20630163032915"><a name="screen20630163032915"></a><a name="screen20630163032915"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.9.0 ./ -f ./Dockerfile-scheduler
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./ -f ./Dockerfile-controller</pre>
    </li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p966311264620"><a name="p966311264620"></a><a name="p966311264620"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

Taking the image creation of Ascend Operator as an example, run the `docker build --no-cache -t ascend-operator:v{version} .` command to create the image. The sample output is as follows. **Note: Do not omit the "." at the end of the command.**

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

5. You can skip this step in the following scenarios.

    - The created cluster scheduling component images have been uploaded to a private image repository, and each node can pull the cluster scheduling component images from the private image repository.
    - The corresponding component images have been created on each node where the cluster scheduling components are installed.

    In other scenarios, you need to manually distribute each component image to each node. Taking NodeD as an example, use the offline image package to distribute the image to other nodes.

    1. Save the created image as an offline image.

        ```shell
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2. Copy the image to other nodes.

        ```shell
        scp noded-v{version}-linux-aarch64.tar root@{Target node IP}:Save path
        ```

    3. Log in to each node as the `root` user and load the offline image.

        ```shell
        docker load < noded-v{version}-linux-aarch64.tar
        ```

6. (Optional) Import the offline image into Containerd. This step applies to scenarios where the container runtime is Containerd. It can be skipped in other scenarios.

    Taking NodeD as an example, execute the following command.

    ```shell
    ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
    ```

**Pull Images from the Ascend Image Repository<a name="section133861705416"></a>**

1. After ensuring the server can access the internet, visit the [Ascend Image Repository](https://www.hiascend.com/developer/ascendhub).
2. <a name="li1381232414410"></a>In the left navigation pane, select the task type as "Cluster Scheduling", and then select the corresponding component images according to the following table. The pulled images need to be renamed before they can be deployed using the component startup YAML. For details, see [Step 3](#li14816124549).

    **Table 3**  Image list

    <a name="table981217243412"></a>
    <table><thead align="left"><tr id="row1781262416419"><th class="cellrowborder" valign="top" width="28.689999999999998%" id="mcps1.2.5.1.1"><p id="p168129241348"><a name="p168129241348"></a><a name="p168129241348"></a>Component</p>
    </th>
    <th class="cellrowborder" valign="top" width="34.43%" id="mcps1.2.5.1.2"><p id="p581214248413"><a name="p581214248413"></a><a name="p581214248413"></a>Image Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.21%" id="mcps1.2.5.1.3"><p id="p12812122410414"><a name="p12812122410414"></a><a name="p12812122410414"></a>Image Tag</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.67%" id="mcps1.2.5.1.4"><p id="p28136241144"><a name="p28136241144"></a><a name="p28136241144"></a>Node to Pull Image</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row38132241945"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p138133241142"><a name="p138133241142"></a><a name="p138133241142"></a><span id="ph88139247418"><a name="ph88139247418"></a><a name="ph88139247418"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><a name="ul158133245418"></a><a name="ul158133245418"></a><ul id="ul158133245418"><li><a href="https://www.hiascend.com/developer/ascendhub/detail/54545fa4ff9f446e914bf44b85efdb61" target="_blank" rel="noopener noreferrer">volcanosh/vc-scheduler</a></li><li><a href="https://www.hiascend.com/developer/ascendhub/detail/16f17a3c95d54f9da710a9c51bfceaa3" target="_blank" rel="noopener noreferrer">volcanosh/vc-controller-manager</a></li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p38142241846"><a name="p38142241846"></a><a name="p38142241846"></a>Select the image as needed:</p>
    <p id="p1814102416419"><a name="p1814102416419"></a><a name="p1814102416419"></a>v1.7.0-v<span id="ph616117387810"><a name="ph616117387810"></a><a name="ph616117387810"></a>26.0.0</span></p>
    <p id="p9814824342"><a name="p9814824342"></a><a name="p9814824342"></a>v1.9.0-v<span id="ph57147381283"><a name="ph57147381283"></a><a name="ph57147381283"></a>26.0.0</span></p>
    </td>
    <td class="cellrowborder" rowspan="4" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p18131924748"><a name="p18131924748"></a><a name="p18131924748"></a>Management node</p>
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
    <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>Infer Operator</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p><a href="https://www.hiascend.com/developer/ascendhub/detail/13f3dee71712420d8b583b9275c04899" target="_blank" rel="noopener noreferrer">infer-operator</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>v26.0.0</p>
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
    <td class="cellrowborder" rowspan="3" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p128156248413"><a name="p128156248413"></a><a name="p128156248413"></a>Compute node</p>
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
    >If you do not have download permissions, apply for permissions as prompted on the page. After submitting the application, wait for the administrator's approval. Once approved, you can download the image.

3. <a name="li14816124549"></a>The cluster scheduling images pulled from the Ascend image repository have different names from those in the component startup YAML files. You need to rename the pulled images before starting the components. Follow the steps below to rename the images obtained in <a href="#li1381232414410">Step 2</a>. It is also recommended to delete the images with the original names. The specific operations are as follows:
    1. Rename images (users need to select the corresponding command based on the component used).

        ```shell
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0 ascend-operator:v26.0.0

        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.0.0 infer-operator:v26.0.0

        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v26.0.0 npu-exporter:v26.0.0

        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0

        # When using Volcano version 1.9.0, you need to change the image tag to v1.9.0-v26.0.0.
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0 volcanosh/vc-controller-manager:v1.7.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0 volcanosh/vc-scheduler:v1.7.0

        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0 noded:v26.0.0

        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0 clusterd:v26.0.0
        ```

    2. (Optional) Delete images with the original names (users need to select the corresponding command based on the component used).

        ```shell
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0
        # When using Volcano v1.9.0, change the image tag to v1.9.0-v26.0.0.
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0
        ```

4. (Optional) Import the offline image into Containerd. This step applies to scenarios where the container runtime is Containerd. Skip this step for other scenarios.

    Taking the NodeD component as an example, use the offline image package and perform the following steps.

    1. Save the created image as an offline image.

        ```shell
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2. Import the offline image into Containerd.

        ```shell
        ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
        ```

## Creating Node Labels<a name="ZH-CN_TOPIC_0000002511426279"></a>

In a K8s cluster, if a node containing an Ascend AI Processor is used as the K8s management node, this node serves as both a management node and a compute node. In addition to the labels required for the management node, you also need to apply the relevant labels for the compute node based on the type of Ascend AI Processor on the node. In a production environment, the management node is typically a general-purpose server and does not contain an Ascend AI Processor.

**Procedure<a name="section847765415564"></a>**

1. Run the following command on any node to query the node name.

    ```shell
    kubectl get node
    ```

    Sample output:

    ```ColdFusion
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2. Label the corresponding nodes based on the label information in [Table 1](#table202738181704) to facilitate the cluster scheduling components in scheduling across various types of worker nodes.

    ```shell
    kubectl label nodes Hostname Label
    ```

    Taking the hostname `ubuntu` and the label `masterselector=dls-master-node` as an example, the command is as follows.

    ```shell
    kubectl label nodes ubuntu masterselector=dls-master-node
    ```

    The sample output is as follows, indicating a successful operation.

    ```ColdFusion
    node/ubuntu labeled
    ```

    >[!NOTE]
    >
    >- For a detailed description of each node label in [Table 1](#table202738181704), see the [K8s Native Object Description](../../../api/k8s.md) section.
    >- Configure all the labels listed in [Table 1](#table202738181704) based on the node type and product type.
    >- {_xxx_} uses `910` as an example chip model.

    **Table 1** Label information corresponding to nodes

    <a name="table202738181704"></a>
    <table><thead align="left"><tr id="row627331819017"><th class="cellrowborder" valign="top" width="31.840000000000003%" id="mcps1.2.4.1.1"><p id="p19273918201"><a name="p19273918201"></a><a name="p19273918201"></a>Node Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="25.96%" id="mcps1.2.4.1.2"><p id="p3273218803"><a name="p3273218803"></a><a name="p3273218803"></a>Product Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="42.199999999999996%" id="mcps1.2.4.1.3"><p id="p19273118301"><a name="p19273118301"></a><a name="p19273118301"></a>Label</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row227451815011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p142747189017"><a name="p142747189017"></a><a name="p142747189017"></a>Management node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p102741181908"><a name="p102741181908"></a><a name="p102741181908"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><p id="p1227417181004"><a name="p1227417181004"></a><a name="p1227417181004"></a>masterselector=dls-master-node</p>
    </td>
    </tr>
    <tr id="row127412189015"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p14274118905"><a name="p14274118905"></a><a name="p14274118905"></a>Compute node</p>
    <p id="p203704324914"><a name="p203704324914"></a><a name="p203704324914"></a></p>
    <p id="p4371534493"><a name="p4371534493"></a><a name="p4371534493"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p627418181808"><a name="p627418181808"></a><a name="p627418181808"></a><span id="ph42747181102"><a name="ph42747181102"></a><a name="ph42747181102"></a>Atlas 800 training server (full NPU configuration)</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul727421813014"></a><a name="ul727421813014"></a><ul id="ul727421813014"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row19274318806"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p742615141511"><a name="p742615141511"></a><a name="p742615141511"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p027411181309"><a name="p027411181309"></a><a name="p027411181309"></a><span id="ph127517181101"><a name="ph127517181101"></a><a name="ph127517181101"></a>Atlas 800 training server (half NPU configuration)</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul22751618203"></a><a name="ul22751618203"></a><ul id="ul22751618203"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=half</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row92751018202"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p554271313169"><a name="p554271313169"></a><a name="p554271313169"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p527551818016"><a name="p527551818016"></a><a name="p527551818016"></a><span id="ph1427511188015"><a name="ph1427511188015"></a><a name="ph1427511188015"></a>Atlas 800T A2 training server</span> or <span id="ph102750181803"><a name="ph102750181803"></a><a name="ph102750181803"></a>Atlas 900 A2 PoD cluster basic unit</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul32752181202"></a><a name="ul32752181202"></a><ul id="ul32752181202"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph12761718301"><a name="ph12761718301"></a><a name="ph12761718301"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row8394133819129"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p1237115354918"><a name="p1237115354918"></a><a name="p1237115354918"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p2039613891219"><a name="p2039613891219"></a><a name="p2039613891219"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul3874134511121"></a><a name="ul3874134511121"></a><ul id="ul3874134511121"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-a3-16-super-pod</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p><span>Atlas 9000 A3 SuperPoD cluster computing system</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul3874134511121"></a><a name="ul3874134511121"></a><ul id="ul3874134511121"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-a3-8-super-pod</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row69181319336"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p738423163315"><a name="p738423163315"></a><a name="p738423163315"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p1584884715522"><a name="p1584884715522"></a><a name="p1584884715522"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 SuperPoD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul537611425289"></a><a name="ul537611425289"></a><ul id="ul537611425289"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul13263154872811"></a><a name="ul13263154872811"></a><ul id="ul13263154872811"><li>host-arch=huawei-x86 or host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul17911532280"></a><a name="ul17911532280"></a><ul id="ul17911532280"><li>accelerator-type=module-a3-16</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row271845218270"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p188095589274"><a name="p188095589274"></a><a name="p188095589274"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p164951627162819"><a name="p164951627162819"></a><a name="p164951627162819"></a><span id="ph19495127162814"><a name="ph19495127162814"></a><a name="ph19495127162814"></a>Atlas 800I A3 SuperPoD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul16834964293"></a><a name="ul16834964293"></a><ul id="ul16834964293"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul128341660299"></a><a name="ul128341660299"></a><ul id="ul128341660299"><li>host-arch=huawei-x86 or host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul168341764299"></a><a name="ul168341764299"></a><ul id="ul168341764299"><li>accelerator-type=module-a3-16</li><li>server-usage=infer</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p><span>Atlas 800T A3 SuperPoD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><ul><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul><ul><li>host-arch=huawei-x86 or host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul><ul><li>accelerator-type=module-a3-16</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row42763185011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p16530201015713"><a name="p16530201015713"></a><a name="p16530201015713"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p19276111815011"><a name="p19276111815011"></a><a name="p19276111815011"></a><span id="ph152766181106"><a name="ph152766181106"></a><a name="ph152766181106"></a>Atlas 800I A2 inference server</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72766183018"></a><a name="ul72766183018"></a><ul id="ul72766183018"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph2027661812017"><a name="ph2027661812017"></a><a name="ph2027661812017"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1468510421395"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p868624283911"><a name="p868624283911"></a><a name="p868624283911"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p534220145119"><a name="p534220145119"></a><a name="p534220145119"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box heterogeneous subrack</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul19511133318489"></a><a name="ul19511133318489"></a><ul id="ul19511133318489"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph175351194911"><a name="ph175351194911"></a><a name="ph175351194911"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row13277101813019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p356115645715"><a name="p356115645715"></a><a name="p356115645715"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122778182014"><a name="p122778182014"></a><a name="p122778182014"></a><span id="ph3277518801"><a name="ph3277518801"></a><a name="ph3277518801"></a>Atlas 200T A2 Box16 heterogeneous subrack</span></p>
    <p id="p1993115373112"><a name="p1993115373112"></a><a name="p1993115373112"></a><span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 heterogeneous subrack</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul15277318601"></a><a name="ul15277318601"></a><ul id="ul15277318601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph52776181604"><a name="ph52776181604"></a><a name="ph52776181604"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-16</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1627716183019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p556216614577"><a name="p556216614577"></a><a name="p556216614577"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p32771718506"><a name="p32771718506"></a><a name="p32771718506"></a><span id="ph162771318306"><a name="ph162771318306"></a><a name="ph162771318306"></a>Training server (with <span id="ph4277131818016"><a name="ph4277131818016"></a><a name="ph4277131818016"></a>Atlas 300T training card</span>)</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72771181601"></a><a name="ul72771181601"></a><ul id="ul72771181601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=card</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row62791418607"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p45625617576"><a name="p45625617576"></a><a name="p45625617576"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122793182008"><a name="p122793182008"></a><a name="p122793182008"></a>Inference server (with <span id="ph19279181811010"><a name="ph19279181811010"></a><a name="ph19279181811010"></a>Atlas 300I inference card</span>)</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul127919181101"></a><a name="ul127919181101"></a><ul id="ul127919181101"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row72822181005"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p165621264571"><a name="p165621264571"></a><a name="p165621264571"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p16282118603"><a name="p16282118603"></a><a name="p16282118603"></a><span id="ph182828181802"><a name="ph182828181802"></a><a name="ph182828181802"></a>Atlas inference series products</span> (except <span id="ph828261816012"><a name="ph828261816012"></a><a name="ph828261816012"></a>Atlas 200I SoC A1 core board</span>)</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul162825182010"></a><a name="ul162825182010"></a><ul id="ul162825182010"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row328212184011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p20562266579"><a name="p20562266579"></a><a name="p20562266579"></a>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p228281818011"><a name="p228281818011"></a><a name="p228281818011"></a><span id="ph928241810010"><a name="ph928241810010"></a><a name="ph928241810010"></a><span id="ph122828181609"><a name="ph122828181609"></a><a name="ph122828181609"></a>Atlas 200I SoC A1 core board</span></span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul202825181508"></a><a name="ul202825181508"></a><ul id="ul202825181508"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>servertype=soc</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p>Compute node</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p>Atlas 350 standard card</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3"><ul><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm or host-arch=huawei-x86</li><li>accelerator=huawei-npu</li><li>(Optional) nodeDEnable=on</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

## Creating a User<a name="ZH-CN_TOPIC_0000002511346353"></a>

Run the following commands on the node where the corresponding component is installed to create a user account:

- <a name="li1069651515405"></a>Ubuntu operating system

  ```shell
  useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
  usermod -a -G HwHiAiUser hwMindX
  ```

- <a name="li19202165424015"></a>CentOS operating system

  ```shell
  useradd -d /home/hwMindX -u 9000 -m -s /sbin/nologin hwMindX
  usermod -a -G HwHiAiUser hwMindX
  ```

>[!NOTE]
>
>- Creating users on other operating systems:
>     - [Ubuntu](#li1069651515405).
>     - [CentOS](#li19202165424015).
>- `HwHiAiUser` is the software runtime user required by the driver or CANN package.
>- Run the `getent passwd` command to check whether the UID and GID of `HwHiAiUser` are consistent across all physical machines (storage nodes, management nodes, and compute nodes) and inside containers, and that both are 1000. If they are occupied, the service may become unavailable. For details, see [User UID or GID Occupied](https://gitcode.com/Ascend/mind-cluster/issues/337).

**Table 1**  Component user description

<a name="table125971501113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0299839362_row86431704617"><th class="cellrowborder" valign="top" width="20.962096209620963%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0299839362_p464201754614"><a name="zh-cn_topic_0299839362_p464201754614"></a><a name="zh-cn_topic_0299839362_p464201754614"></a>Component</p>
</th>
<th class="cellrowborder" valign="top" width="34.13341334133413%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0299839362_p11647172468"><a name="zh-cn_topic_0299839362_p11647172468"></a><a name="zh-cn_topic_0299839362_p11647172468"></a>Startup User</p>
</th>
<th class="cellrowborder" valign="top" width="44.90449044904491%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0299839362_p56451734620"><a name="zh-cn_topic_0299839362_p56451734620"></a><a name="zh-cn_topic_0299839362_p56451734620"></a>Use Privileged Container</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0299839362_row3641172465"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p671453716107"><a name="p671453716107"></a><a name="p671453716107"></a><span id="ph14925450192719"><a name="ph14925450192719"></a><a name="ph14925450192719"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><a name="ul124012695512"></a><a name="ul124012695512"></a><ul id="ul124012695512"><li>Binary run: hwMindX</li><li>Container run: root</li></ul>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><a name="ul8401830195518"></a><a name="ul8401830195518"></a><ul id="ul8401830195518"><li>Binary run: Not involved.</li><li>Container run: Requires a privileged container. It is recommended to use binary run.</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row1064121764612"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p16641317134612"><a name="zh-cn_topic_0299839362_p16641317134612"></a><a name="zh-cn_topic_0299839362_p16641317134612"></a><span id="ph522114212719"><a name="ph522114212719"></a><a name="ph522114212719"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p53735269103"><a name="p53735269103"></a><a name="p53735269103"></a>root</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p29286561106"><a name="p29286561106"></a><a name="p29286561106"></a>Requires a privileged container.</p>
</td>
</tr>
<tr id="row10935147171519"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1935947181513"><a name="p1935947181513"></a><a name="p1935947181513"></a><span id="ph5551115391513"><a name="ph5551115391513"></a><a name="ph5551115391513"></a>NodeD</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row664817164615"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p0649177466"><a name="zh-cn_topic_0299839362_p0649177466"></a><a name="zh-cn_topic_0299839362_p0649177466"></a><span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span></p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p153424813128"><a name="p153424813128"></a><a name="p153424813128"></a>hwMindX</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p17327314131212"><a name="p17327314131212"></a><a name="p17327314131212"></a>Not involved.</p>
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
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p161622011121819"><a name="p161622011121819"></a><a name="p161622011121819"></a>Decided by the user. It is recommended to use a non-root user.</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p1916271131815"><a name="p1916271131815"></a><a name="p1916271131815"></a>Decided by the user. It is recommended not to use a privileged container.</p>
</td>
</tr>
<tr id="row315419369301"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p715593611302"><a name="p715593611302"></a><a name="p715593611302"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
</td>
</tr>
<tr id="row3502131311115"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p175021513201117"><a name="p175021513201117"></a><a name="p175021513201117"></a><span id="ph16988102112717"><a name="ph16988102112717"></a><a name="ph16988102112717"></a>Container Manager</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p1450212134110"><a name="p1450212134110"></a><a name="p1450212134110"></a>root</p>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p6502191318116"><a name="p6502191318116"></a><a name="p6502191318116"></a>Not involved.</p>
</td>
</tr>
</tbody>
</table>

## Creating Log Directories<a name="ZH-CN_TOPIC_0000002511346417"></a>

Create the parent log directory for components and the log directories for each component on the corresponding nodes, and set the owner and permissions for the directories.

**Procedure<a name="section124928122416"></a>**

1. Run the following command to create the parent log directory for components on each node according to [Table 1 Cluster scheduling component log path list](#table957112617314).

    ```shell
    mkdir -m 755 /var/log/mindx-dl
    chown root:root /var/log/mindx-dl
    ```

2. Create the corresponding log directories based on the specific components used.

    **Table 1** Cluster scheduling component log path list

    <a name="table957112617314"></a>
    <table><thead align="left"><tr id="row2057210616310"><th class="cellrowborder" valign="top" width="21.93%" id="mcps1.2.5.1.1"><p id="p10572761231"><a name="p10572761231"></a><a name="p10572761231"></a>Component</p>
    </th>
    <th class="cellrowborder" valign="top" width="41.91%" id="mcps1.2.5.1.2"><p id="p11572156430"><a name="p11572156430"></a><a name="p11572156430"></a>Command for Creating Log Directory</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.05%" id="mcps1.2.5.1.3"><p id="p25721364319"><a name="p25721364319"></a><a name="p25721364319"></a>Node for Log Path Creation</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.11%" id="mcps1.2.5.1.4"><p id="p16572661320"><a name="p16572661320"></a><a name="p16572661320"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row457296131"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p1572469315"><a name="p1572469315"></a><a name="p1572469315"></a><span id="ph9572196532"><a name="ph9572196532"></a><a name="ph9572196532"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1657216638"><a name="screen1657216638"></a><a name="screen1657216638"></a>mkdir -m 750 /var/log/mindx-dl/devicePlugin
   chown root:root /var/log/mindx-dl/devicePlugin</pre>
    </td>
    <td class="cellrowborder" rowspan="5" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p11572661536"><a name="p11572661536"></a><a name="p11572661536"></a>Compute node</p>
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
   chown <em id="i15731661134"><a name="i15731661134"></a><a name="i15731661134"></a>User-defined</em> /var/log/mindx-dl/elastic</pre>
    <div class="note" id="note3573061032"><a name="note3573061032"></a><a name="note3573061032"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p2057310617318"><a name="p2057310617318"></a><a name="p2057310617318"></a>Mount Elastic Agent log directory to the container. For details, see "Modify the mount paths for the training script and code" step in <a href="../../../usage/resumable_training/06_configuring_the_job_yaml_file.md#job-yaml-configuration-examples">Job YAML Configuration Examples</a>.</p>
    </div></div>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul958614153510"></a><a name="ul958614153510"></a><ul id="ul958614153510"><li>The directory owner is user-defined. Note: Ensure that the user group for installing <span id="ph67093892615"><a name="ph67093892615"></a><a name="ph67093892615"></a>Elastic Agent</span>, the user group for running <span id="ph1642075902418"><a name="ph1642075902418"></a><a name="ph1642075902418"></a>Elastic Agent</span>, and the group for the mounted host machine directory are consistent.</li><li>Users can customize the disk path for <span id="ph1790811553279"><a name="ph1790811553279"></a><a name="ph1790811553279"></a>Elastic Agent</span> runtime logs. Under this path, users can view logs of all <span id="ph1529820279122"><a name="ph1529820279122"></a><a name="ph1529820279122"></a>Elastic Agent</span> nodes without logging in to each node individually.</li></ul>
    </td>
    </tr>
    <tr id="row189638410329"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p7963164113217"><a name="p7963164113217"></a><a name="p7963164113217"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen929012103313"><a name="screen929012103313"></a><a name="screen929012103313"></a>mkdir  -m 750  <em id="i15660102313617"><a name="i15660102313617"></a><a name="i15660102313617"></a>Training script directory</em>/taskd_log
   chown <em id="i4956143053617"><a name="i4956143053617"></a><a name="i4956143053617"></a>User-defined</em> <em id="i6187123720366"><a name="i6187123720366"></a><a name="i6187123720366"></a>Training script directory</em>/taskd_log </pre>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul9461980353"></a><a name="ul9461980353"></a><ul id="ul9461980353"><li>The directory owner is user-defined.</li><li><span id="ph1524182517352"><a name="ph1524182517352"></a><a name="ph1524182517352"></a>TaskD</span> can automatically create the corresponding log directory during runtime. The log directory prefix is generally the directory where the <strong id="b5881131073711"><a name="b5881131073711"></a><a name="b5881131073711"></a>bash command</strong> is executed in the job YAML or where the training is launched.</li></ul>
    </td>
    </tr>
    <tr id="row65749616319"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p8574136838"><a name="p8574136838"></a><a name="p8574136838"></a><span id="ph13574365316"><a name="ph13574365316"></a><a name="ph13574365316"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen05746613313"><a name="screen05746613313"></a><a name="screen05746613313"></a>mkdir -m 750 /var/log/mindx-dl/ascend-operator
   chown hwMindX:hwMindX /var/log/mindx-dl/ascend-operator</pre>
    </td>
    <td class="cellrowborder" rowspan="6" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p65611868135"><a name="p65611868135"></a><a name="p65611868135"></a>Management node</p>
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
    <td class="cellrowborder" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p53074565125"><a name="p53074565125"></a><a name="p53074565125"></a>Nodes that require the container recovery feature</p>
    </td>
    <td class="cellrowborder" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p1518124119135"><a name="p1518124119135"></a><a name="p1518124119135"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

## Creating a Namespace<a name="ZH-CN_TOPIC_0000002479226384"></a>

- The NodeD, Resilience Controller, ClusterD, Infer Operator, and Ascend Operator components for cluster scheduling run in the `mindx-dl` namespace of K8s. Run the following command on the management node of K8s to create the corresponding namespace.

    ```shell
    kubectl create ns mindx-dl
    ```

- To report SuperPoD information, pingmesh configuration, and public fault information, MindCluster requires the manual creation of a namespace named `cluster-system`. Run the following command on the management node of K8s.

    ```shell
    kubectl create ns cluster-system
    ```

- The namespace for NPU Exporter is `npu-exporter`; the namespace for Volcano is `volcano-system`; the namespace for Ascend Device Plugin is `kube-system`. The namespaces for the preceding components are created by the system, and you do not need to create them again.
