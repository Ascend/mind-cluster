# Resilience Controller<a name="ZH-CN_TOPIC_0000002511426375"></a>

## （可选）导入证书和KubeConfig<a name="ZH-CN_TOPIC_0000002479226468"></a>

**使用前必读<a name="section18169249192720"></a>**

导入工具cert-importer在组件的软件包中。

- 使用之前请先查看[导入工具说明](#section890515124614)，根据实际情况选择对应的导入步骤。
- 导入KubeConfig文件参见[导入KubeConfig文件](#section1538945217341)。

**导入工具说明<a name="section890515124614"></a>**

- 导入文件的说明请参考[表1](#table66513321527)，详细命令参数请参考[表4](#table18529165716504)。

    **表 1**  组件导入文件说明

    <a name="table66513321527"></a>
    <table><thead align="left"><tr id="row866113218219"><th class="cellrowborder" valign="top" width="19.59195919591959%" id="mcps1.2.5.1.1"><p id="p19661432425"><a name="p19661432425"></a><a name="p19661432425"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.88168816881688%" id="mcps1.2.5.1.2"><p id="p5118134235115"><a name="p5118134235115"></a><a name="p5118134235115"></a>导入文件类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.95269526952695%" id="mcps1.2.5.1.3"><p id="p99612619162"><a name="p99612619162"></a><a name="p99612619162"></a>导入命令示例</p>
    </th>
    <th class="cellrowborder" valign="top" width="36.57365736573657%" id="mcps1.2.5.1.4"><p id="p176262101716"><a name="p176262101716"></a><a name="p176262101716"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row12463182714316"><td class="cellrowborder" valign="top" width="19.59195919591959%" headers="mcps1.2.5.1.1 "><p id="p72311217103014"><a name="p72311217103014"></a><a name="p72311217103014"></a><span id="ph14361567178"><a name="ph14361567178"></a><a name="ph14361567178"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="16.88168816881688%" headers="mcps1.2.5.1.2 "><p id="p33991858232"><a name="p33991858232"></a><a name="p33991858232"></a>连接<span id="ph4808918506"><a name="ph4808918506"></a><a name="ph4808918506"></a>K8s</span>的KubeConfig文件</p>
    <p id="p331133914167"><a name="p331133914167"></a><a name="p331133914167"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="26.95269526952695%" headers="mcps1.2.5.1.3 "><p id="p16682153041618"><a name="p16682153041618"></a><a name="p16682153041618"></a>./cert-importer -kubeConfig=<em id="i28511515200"><a name="i28511515200"></a><a name="i28511515200"></a>{kubeFile}</em>  -cpt=<em id="i11887152317202"><a name="i11887152317202"></a><a name="i11887152317202"></a>{component}</em></p>
    <p id="p115141742151614"><a name="p115141742151614"></a><a name="p115141742151614"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="36.57365736573657%" headers="mcps1.2.5.1.4 "><p id="p2833102831511"><a name="p2833102831511"></a><a name="p2833102831511"></a>由于<span id="ph88891493615"><a name="ph88891493615"></a><a name="ph88891493615"></a>K8s</span>自带的ServiceAccount的token文件会挂载到物理机上，有暴露风险，可通过外部导入加密KubeConfig文件替换ServiceAccount进行安全加固。</p>
    <p id="p18105124517162"><a name="p18105124517162"></a><a name="p18105124517162"></a></p>
    </td>
    </tr>
    </tbody>
    </table>

- 工具支持的操作如[表2](#table13221181211509)。

    **表 2**  操作说明

    <a name="table13221181211509"></a>
    <table><thead align="left"><tr id="row4222141214502"><th class="cellrowborder" valign="top" width="15.709999999999999%" id="mcps1.2.3.1.1"><p id="p6222131285015"><a name="p6222131285015"></a><a name="p6222131285015"></a>操作</p>
    </th>
    <th class="cellrowborder" valign="top" width="84.28999999999999%" id="mcps1.2.3.1.2"><p id="p1222181295014"><a name="p1222181295014"></a><a name="p1222181295014"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row022271217502"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p1822219129505"><a name="p1822219129505"></a><a name="p1822219129505"></a>新增</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p622220128501"><a name="p622220128501"></a><a name="p622220128501"></a>导入KubeConfig等文件。</p>
    </td>
    </tr>
    <tr id="row1622231295011"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p132221512105017"><a name="p132221512105017"></a><a name="p132221512105017"></a>更新</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p147469919538"><a name="p147469919538"></a><a name="p147469919538"></a>导入新的KubeConfig等文件，替换旧的文件。</p>
    <p id="p922217125500"><a name="p922217125500"></a><a name="p922217125500"></a>重新导入后，需要重启业务组件才生效。请提前规划证书的有效期，有效期要匹配产品生命周期，不能过长或者过短，避免业务组件重启导致业务中断。</p>
    </td>
    </tr>
    </tbody>
    </table>

- 默认情况下，导入成功后，工具会自动删除KubeConfig授权文件，用户可通过<b>-n</b>参数停用自动删除功能。如果不自动删除，用户应妥善保管相关配置文件，如果决定不再使用相关文件，请立即删除，防止意外泄露。
- 导入的文件会被重新加密并存入“/etc/mindx-dl”目录中，具体参考[表3](#table252713572507)。
- 如果从3.0.RC3及以后版本降级到3.0.RC3之前的旧版本，需在手动删除“/etc/mindx-dl/”目录下的文件后，重新使用旧版cert-importer工具导入。
- 导入工具加密需要系统有足够的熵池（random pool）。如果熵池不够，程序可能阻塞，可以安装haveged组件来进行补熵。

    安装命令可参考：

    - 类似CentOS操作系统执行**yum install haveged -y**命令进行安装，并执行**systemctl start haveged**命令启动haveged组件。
    - 类似Ubuntu操作系统执行**apt install haveged -y**命令进行安装，并执行**systemctl start haveged**命令启动haveged组件。

**导入KubeConfig文件<a name="section1538945217341"></a>**

1. 登录K8s管理节点。
2. 创建“/etc/kubernetes/mindxdl”文件夹，权限设置为750。

    ```shell
    rm -rf /etc/kubernetes/mindxdl
    mkdir /etc/kubernetes/mindxdl
    chmod 750 /etc/kubernetes/mindxdl
    ```

3. 参考[Kubernetes相关指导](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)自行创建名为resilience-controller-cfg.conf的KubeConfig文件，其中KubeConfig文件中的“user”字段为“resilience-controller”。将KubeConfig文件放到“/etc/kubernetes/mindxdl/”路径下。
4. 进入Resilience Controller安装包解压路径，将lib文件夹设置到当前窗口的环境变量LD\_LIBRARY\_PATH中，不需要持久化或继承给其他用户（证书导入工具需要配置自带的加密组件相关的so包路径）。
    1. 执行以下命令，将环境变量进行备份。

        ```shell
        export LD_LIBRARY_PATH_BAK=${LD_LIBRARY_PATH}
        ```

    2. 执行以下命令，将lib文件夹设置到当前环境变量LD\_LIBRARY\_PATH中。

        ```shell
        export LD_LIBRARY_PATH=`pwd`/lib/:${LD_LIBRARY_PATH}
        ```

5. 执行以下命令，为Resilience Controller组件导入KubeConfig文件。

    ```shell
    ./cert-importer -kubeConfig=/etc/kubernetes/mindxdl/resilience-controller-cfg.conf  -cpt=rc
    ```

    回显示例如下，请以实际回显为准，出现以下字段表示导入成功。

    ```ColdFusion
    encrypt kubeConfig successfully
    start to write data to disk
    [OP]import kubeConfig successfully
    change owner and set file mode successfully
    ```

    >[!NOTE] 
    >- 已经导入了KubeConfig配置文件，但是组件还是出现连接K8s异常的场景，可以参见[集群调度组件连接K8s异常](../../../faq.md#集群调度组件连接k8s异常)章节进行处理。
    >- 导入证书时，导入工具cert-importer会自动创建“/var/log/mindx-dl/cert-importer”目录，目录权限750，属主为root:root。

6. 执行以下命令，将备份的环境变量还原。

    ```shell
    export LD_LIBRARY_PATH=${LD_LIBRARY_PATH_BAK}
    ```

**表 3** 集群调度组件证书配置文件表

<a name="table252713572507"></a>
<table><thead align="left"><tr id="row4527257145015"><th class="cellrowborder" valign="top" width="17.5982401759824%" id="mcps1.2.5.1.1"><p id="p14528165725013"><a name="p14528165725013"></a><a name="p14528165725013"></a>组件</p>
</th>
<th class="cellrowborder" valign="top" width="19.24807519248075%" id="mcps1.2.5.1.2"><p id="p14528105765013"><a name="p14528105765013"></a><a name="p14528105765013"></a>证书等配置文件路径</p>
</th>
<th class="cellrowborder" valign="top" width="11.08889111088891%" id="mcps1.2.5.1.3"><p id="p105282572501"><a name="p105282572501"></a><a name="p105282572501"></a>目录及其文件属主</p>
</th>
<th class="cellrowborder" valign="top" width="52.064793520647946%" id="mcps1.2.5.1.4"><p id="p11528155755016"><a name="p11528155755016"></a><a name="p11528155755016"></a>配置文件说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row9528155785012"><td class="cellrowborder" valign="top" width="17.5982401759824%" headers="mcps1.2.5.1.1 "><p id="p1528155715501"><a name="p1528155715501"></a><a name="p1528155715501"></a><span id="ph1488142812262"><a name="ph1488142812262"></a><a name="ph1488142812262"></a>集群调度组件</span>证书相关根目录</p>
</td>
<td class="cellrowborder" valign="top" width="19.24807519248075%" headers="mcps1.2.5.1.2 "><p id="p7528357125019"><a name="p7528357125019"></a><a name="p7528357125019"></a>/etc/mindx-dl/</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="11.08889111088891%" headers="mcps1.2.5.1.3 "><p id="p1528457195011"><a name="p1528457195011"></a><a name="p1528457195011"></a>hwMindX:hwMindX</p>
<p id="p17618514195"><a name="p17618514195"></a><a name="p17618514195"></a></p>
<p id="p27716513196"><a name="p27716513196"></a><a name="p27716513196"></a></p>
<p id="p1532775483511"><a name="p1532775483511"></a><a name="p1532775483511"></a></p>
</td>
<td class="cellrowborder" valign="top" width="52.064793520647946%" headers="mcps1.2.5.1.4 "><p id="p9528857175013"><a name="p9528857175013"></a><a name="p9528857175013"></a>kmc_primary_store/master.ks：自动生成的主密钥，请勿删除。</p>
<p id="p1152811579509"><a name="p1152811579509"></a><a name="p1152811579509"></a>.config/backup.ks：自动生成的备份密钥，请勿删除。</p>
</td>
</tr>
<tr id="row207702393454"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p127701395458"><a name="p127701395458"></a><a name="p127701395458"></a><span id="ph1287272539"><a name="ph1287272539"></a><a name="ph1287272539"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1977073944518"><a name="p1977073944518"></a><a name="p1977073944518"></a>/etc/mindx-dl/resilience-controller/</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p176150132191"><a name="p176150132191"></a><a name="p176150132191"></a>.config/config6：导入的加密<span id="ph10615131313194"><a name="ph10615131313194"></a><a name="ph10615131313194"></a>K8s</span> KubeConfig文件，连接<span id="ph761518136190"><a name="ph761518136190"></a><a name="ph761518136190"></a>K8s</span>使用。</p>
<p id="p16156132195"><a name="p16156132195"></a><a name="p16156132195"></a>.config6：导入的加密<span id="ph761517138198"><a name="ph761517138198"></a><a name="ph761517138198"></a>K8s</span> KubeConfig文件备份。</p>
</td>
</tr>
</tbody>
</table>

**表 4**  导入工具参数说明

<a name="table18529165716504"></a>
<table><thead align="left"><tr id="row1852914572501"><th class="cellrowborder" valign="top" width="17.349999999999998%" id="mcps1.2.5.1.1"><p id="p5529175745012"><a name="p5529175745012"></a><a name="p5529175745012"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="19.41%" id="mcps1.2.5.1.2"><p id="p17529185775019"><a name="p17529185775019"></a><a name="p17529185775019"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="11.01%" id="mcps1.2.5.1.3"><p id="p1352935715507"><a name="p1352935715507"></a><a name="p1352935715507"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="52.23%" id="mcps1.2.5.1.4"><p id="p1552925711509"><a name="p1552925711509"></a><a name="p1552925711509"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row55021443133913"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p117491127105717"><a name="p117491127105717"></a><a name="p117491127105717"></a>-kubeConfig</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p18750132718575"><a name="p18750132718575"></a><a name="p18750132718575"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p975010276572"><a name="p975010276572"></a><a name="p975010276572"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9750162720574"><a name="p9750162720574"></a><a name="p9750162720574"></a>待导入的KubeConfig文件的路径。</p>
</td>
</tr>
<tr id="row45301657115017"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p8530165715011"><a name="p8530165715011"></a><a name="p8530165715011"></a>-cpt</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1353085745016"><a name="p1353085745016"></a><a name="p1353085745016"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p18558162664212"><a name="p18558162664212"></a><a name="p18558162664212"></a>rc</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9931932155418"><a name="p9931932155418"></a><a name="p9931932155418"></a>导入证书的组件名称为rc，表示<span id="ph131541756961"><a name="ph131541756961"></a><a name="ph131541756961"></a>Resilience Controller</span>。</p>
</td>
</tr>
<tr id="row953045718504"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p5530195785020"><a name="p5530195785020"></a><a name="p5530195785020"></a>-encryptAlgorithm</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p12530125719509"><a name="p12530125719509"></a><a name="p12530125719509"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p35312571506"><a name="p35312571506"></a><a name="p35312571506"></a>9</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p135316571501"><a name="p135316571501"></a><a name="p135316571501"></a>私钥口令加密算法：</p>
<a name="ul145317578507"></a><a name="ul145317578507"></a><ul id="ul145317578507"><li>8：AES128GCM</li><li>9：AES256GCM</li></ul>
<div class="note" id="note05311457165012"><a name="note05311457165012"></a><a name="note05311457165012"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p18531125718501"><a name="p18531125718501"></a><a name="p18531125718501"></a>无效参数值会被重置为默认值。</p>
</div></div>
</td>
</tr>
<tr id="row18531135717506"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1253175785015"><a name="p1253175785015"></a><a name="p1253175785015"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p75318575501"><a name="p75318575501"></a><a name="p75318575501"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p853175715507"><a name="p853175715507"></a><a name="p853175715507"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p35317578505"><a name="p35317578505"></a><a name="p35317578505"></a>查询<span id="ph19991165205214"><a name="ph19991165205214"></a><a name="ph19991165205214"></a>Resilience Controller</span>版本号。</p>
</td>
</tr>
<tr id="row2573635141612"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p138495616250"><a name="p138495616250"></a><a name="p138495616250"></a>-n</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p6384135614257"><a name="p6384135614257"></a><a name="p6384135614257"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p03848562252"><a name="p03848562252"></a><a name="p03848562252"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p2384145614255"><a name="p2384145614255"></a><a name="p2384145614255"></a>导入成功后是否删除<span id="ph418094814555"><a name="ph418094814555"></a><a name="ph418094814555"></a>KubeConfig</span>文件。</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true：导入成功后不删除<span id="ph39020528557"><a name="ph39020528557"></a><a name="ph39020528557"></a>KubeConfig</span>文件。</li><li>false：导入成功后删除<span id="ph7200135465511"><a name="ph7200135465511"></a><a name="ph7200135465511"></a>KubeConfig</span>文件。</li></ul>
</td>
</tr>
<tr id="row5485341194020"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p6232132695813"><a name="p6232132695813"></a><a name="p6232132695813"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p623242655813"><a name="p623242655813"></a><a name="p623242655813"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1232182685820"><a name="p1232182685820"></a><a name="p1232182685820"></a>/var/log/mindx-dl/cert-importer/cert-importer.log</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p102331826175814"><a name="p102331826175814"></a><a name="p102331826175814"></a>工具运行日志文件。转储后文件的命名格式为：cert-importer-触发转储的时间.log，如：cert-importer-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row8384164173412"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1138412411345"><a name="p1138412411345"></a><a name="p1138412411345"></a>-updateMk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1738494133414"><a name="p1738494133414"></a><a name="p1738494133414"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1238434133420"><a name="p1238434133420"></a><a name="p1238434133420"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p17211144255612"><a name="p17211144255612"></a><a name="p17211144255612"></a>是否更新KMC加密组件的主密钥。</p>
<a name="ul154871314165520"></a><a name="ul154871314165520"></a><ul id="ul154871314165520"><li>true：更新主密钥。</li><li>false：不更新主密钥。</li></ul>
</td>
</tr>
<tr id="row1397106345"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p53986014348"><a name="p53986014348"></a><a name="p53986014348"></a>-updateRk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p139811020345"><a name="p139811020345"></a><a name="p139811020345"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p139850143413"><a name="p139850143413"></a><a name="p139850143413"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p741952510563"><a name="p741952510563"></a><a name="p741952510563"></a>是否更新KMC加密组件的根密钥。</p>
<a name="ul14451957145511"></a><a name="ul14451957145511"></a><ul id="ul14451957145511"><li>true：更新根密钥。</li><li>false：不更新根密钥。</li></ul>
</td>
</tr>
<tr id="row050462052716"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p13504720112715"><a name="p13504720112715"></a><a name="p13504720112715"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1350422002713"><a name="p1350422002713"></a><a name="p1350422002713"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1650420209273"><a name="p1650420209273"></a><a name="p1650420209273"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p4505820152717"><a name="p4505820152717"></a><a name="p4505820152717"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

## 安装Resilience Controller<a name="ZH-CN_TOPIC_0000002479226460"></a>

- 使用**弹性训练**时，必须安装Resilience Controller。Resilience Controller连接K8s时，可以选择使用ServiceAccount或KubeConfig文件进行认证，两种方式差异可参考[使用ServiceAccount和KubeConfig差异](../../../appendix.md#使用serviceaccount和kubeconfig差异)。
- 不使用**弹性训练**的用户，可以不安装Resilience Controller，请直接跳过本章节。

**操作步骤<a name="section0531457718"></a>**

1. 以root用户登录K8s管理节点，并执行以下命令，查看Resilience Controller镜像和版本号是否正确。

    ```shell
    docker images | grep resilience-controller
    ```

    回显示例如下：

    ```ColdFusion
    resilience-controller                      v26.0.0             c532e9d0889c        About an hour ago         142MB
    ```

    - 是，执行[步骤2](#li10743192474541)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li10743192474541"></a>将Resilience Controller软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Resilience Controller的启动参数。启动参数的说明请参见[表1](#table195504370194)，也可执行<b>./resilience-controller -h</b>查看参数说明。
4. 在管理节点的YAML所在路径，执行以下命令，启动Resilience Controller。

    - 如果没有导入KubeConfig证书，执行如下命令。

        ```shell
        kubectl apply -f resilience-controller-v{version}.yaml
        ```

        启动示例如下：

        ```ColdFusion 
        serviceaccount/resilience-controller created
        clusterrole.rbac.authorization.k8s.io/pods-resilience-controller-role created
        clusterrolebinding.rbac.authorization.k8s.io/resilience-controller-rolebinding created
        deployment.apps/resilience-controller created
       ```       

    - 如果导入了KubeConfig证书，执行如下命令。

        ```shell
        kubectl apply -f resilience-controller-without-token-v{version}.yaml
        ```

        启动示例如下：

        ```ColdFusion
        deployment.apps/resilience-controller created
        ``` 

5. 执行以下命令，查看组件是否安装成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                                            READY    STATUS      RESTARTS   AGE
    ...
    resilience-controller-7667495b6b-hwmjw   1/1     Running   0         11s
    ...
    ```

>[!NOTE]
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1868556161717"></a>**

**表 1** Resilience Controller启动参数

<a name="table195504370194"></a>
<table><thead align="left"><tr id="row10550173721915"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p1855053711192"><a name="p1855053711192"></a><a name="p1855053711192"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p355063710197"><a name="p355063710197"></a><a name="p355063710197"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p055073781916"><a name="p055073781916"></a><a name="p055073781916"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3550237171920"><a name="p3550237171920"></a><a name="p3550237171920"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row3551143715196"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p65517376197"><a name="p65517376197"></a><a name="p65517376197"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p19551153781918"><a name="p19551153781918"></a><a name="p19551153781918"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15511378194"><a name="p15511378194"></a><a name="p15511378194"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18551173791915"><a name="p18551173791915"></a><a name="p18551173791915"></a>是否查询<span id="ph151418415511"><a name="ph151418415511"></a><a name="ph151418415511"></a>Resilience Controller</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row8551137161913"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1155183715199"><a name="p1155183715199"></a><a name="p1155183715199"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p105511137141920"><a name="p105511137141920"></a><a name="p105511137141920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p755113373192"><a name="p755113373192"></a><a name="p755113373192"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6551123716195"><a name="p6551123716195"></a><a name="p6551123716195"></a>日志级别：</p>
<a name="ul655113715194"></a><a name="ul655113715194"></a><ul id="ul655113715194"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row1455163771915"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p10551143710191"><a name="p10551143710191"></a><a name="p10551143710191"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p14551193781920"><a name="p14551193781920"></a><a name="p14551193781920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1655193715195"><a name="p1655193715195"></a><a name="p1655193715195"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p7551183716190"><a name="p7551183716190"></a><a name="p7551183716190"></a>日志备份时间限制，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row175527378195"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p455223751910"><a name="p455223751910"></a><a name="p455223751910"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p195521937131913"><a name="p195521937131913"></a><a name="p195521937131913"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15552137111920"><a name="p15552137111920"></a><a name="p15552137111920"></a>/var/log/mindx-dl/resilience-controller/run.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18552143713199"><a name="p18552143713199"></a><a name="p18552143713199"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：run-触发转储的时间.log，如run-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row1655213379191"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p11552163741920"><a name="p11552163741920"></a><a name="p11552163741920"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1552137171918"><a name="p1552137171918"></a><a name="p1552137171918"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p7552193711192"><a name="p7552193711192"></a><a name="p7552193711192"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p555233718199"><a name="p555233718199"></a><a name="p555233718199"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row33119022219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1532160192215"><a name="p1532160192215"></a><a name="p1532160192215"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p123213019227"><a name="p123213019227"></a><a name="p123213019227"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1832100102210"><a name="p1832100102210"></a><a name="p1832100102210"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1328016224"><a name="p1328016224"></a><a name="p1328016224"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>
