# Resilience Controller<a name="ZH-CN_TOPIC_0000002511426375"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:15:26.064Z pushedAt=2026-06-09T06:22:06.900Z -->

## (Optional) Importing Certificates and KubeConfig<a name="ZH-CN_TOPIC_0000002479226468"></a>

**Read Before Use<a name="section18169249192720"></a>**

The import tool `cert-importer` is included in the component's package.

- Before use, see [Import Tool Description](#section890515124614) and select the appropriate import procedure based on your actual situation.
- For importing the KubeConfig file, see [Importing the KubeConfig File](#section1538945217341a).

**Import Tool Description<a name="section890515124614"></a>**

- For file import instructions, see [Table 1](#table66513321527). For detailed command parameters, see [Table 4](#table18529165716504).

    **Table 1** File importing

    <a name="table66513321527"></a>
    <table><thead align="left"><tr id="row866113218219"><th class="cellrowborder" valign="top" width="19.59195919591959%" id="mcps1.2.5.1.1"><p id="p19661432425"><a name="p19661432425"></a><a name="p19661432425"></a>Component</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.88168816881688%" id="mcps1.2.5.1.2"><p id="p5118134235115"><a name="p5118134235115"></a><a name="p5118134235115"></a>Imported File Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.95269526952695%" id="mcps1.2.5.1.3"><p id="p99612619162"><a name="p99612619162"></a><a name="p99612619162"></a>Import Command Example</p>
    </th>
    <th class="cellrowborder" valign="top" width="36.57365736573657%" id="mcps1.2.5.1.4"><p id="p176262101716"><a name="p176262101716"></a><a name="p176262101716"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row12463182714316"><td class="cellrowborder" valign="top" width="19.59195919591959%" headers="mcps1.2.5.1.1 "><p id="p72311217103014"><a name="p72311217103014"></a><a name="p72311217103014"></a><span id="ph14361567178"><a name="ph14361567178"></a><a name="ph14361567178"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="16.88168816881688%" headers="mcps1.2.5.1.2 "><p id="p33991858232"><a name="p33991858232"></a><a name="p33991858232"></a>KubeConfig file for connecting to <span id="ph4808918506"><a name="ph4808918506"></a><a name="ph4808918506"></a>K8s</span></p>
    <p id="p331133914167"><a name="p331133914167"></a><a name="p331133914167"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="26.95269526952695%" headers="mcps1.2.5.1.3 "><p id="p16682153041618"><a name="p16682153041618"></a><a name="p16682153041618"></a>./cert-importer -kubeConfig=<em id="i28511515200"><a name="i28511515200"></a><a name="i28511515200"></a>{kubeFile}</em>  -cpt=<em id="i11887152317202"><a name="i11887152317202"></a><a name="i11887152317202"></a>{component}</em></p>
    <p id="p115141742151614"><a name="p115141742151614"></a><a name="p115141742151614"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="36.57365736573657%" headers="mcps1.2.5.1.4 "><p id="p2833102831511"><a name="p2833102831511"></a><a name="p2833102831511"></a>The token file of the ServiceAccount bundled with <span id="ph88891493615"><a name="ph88891493615"></a><a name="ph88891493615"></a>K8s</span> is mounted to the physical machine, posing an exposure risk. You can import an encrypted KubeConfig file externally to replace the ServiceAccount for security hardening.</p>
    <p id="p18105124517162"><a name="p18105124517162"></a><a name="p18105124517162"></a></p>
    </td>
    </tr>
    </tbody>
    </table>

- For operations supported by the tool, see [Table 2](#table13221181211509).

    **Table 2** Operations

    <a name="table13221181211509"></a>
    <table><thead align="left"><tr id="row4222141214502"><th class="cellrowborder" valign="top" width="15.709999999999999%" id="mcps1.2.3.1.1"><p id="p6222131285015"><a name="p6222131285015"></a><a name="p6222131285015"></a>Operation</p>
    </th>
    <th class="cellrowborder" valign="top" width="84.28999999999999%" id="mcps1.2.3.1.2"><p id="p1222181295014"><a name="p1222181295014"></a><a name="p1222181295014"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row022271217502"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p1822219129505"><a name="p1822219129505"></a><a name="p1822219129505"></a>Add</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p622220128501"><a name="p622220128501"></a><a name="p622220128501"></a>Import files such as KubeConfig.</p>
    </td>
    </tr>
    <tr id="row1622231295011"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p132221512105017"><a name="p132221512105017"></a><a name="p132221512105017"></a>Update</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p147469919538"><a name="p147469919538"></a><a name="p147469919538"></a>Import new files such as KubeConfig to replace the old ones.</p>
    <p id="p922217125500"><a name="p922217125500"></a><a name="p922217125500"></a>After re-importing, you need to restart the service components for the changes to take effect. Plan the certificate validity period in advance. The validity period must match the product lifecycle and cannot be too long or too short, to avoid service interruption caused by component restarts.</p>
    </td>
    </tr>
    </tbody>
    </table>

- By default, after the import succeeds, the tool automatically deletes the KubeConfig authorization file. You can disable the automatic deletion function using the `-n` parameter. If automatic deletion is not enabled, you should properly keep the relevant configuration files. If you decide to no longer use the relevant files, delete them immediately to prevent accidental leakage.
- The imported files will be re-encrypted and stored in the `/etc/mindx-dl` directory. For details, see [Table 3](#table252713572507).
- If you downgrade from version 3.0.RC3 or later to a version earlier than 3.0.RC3, you must manually delete the files in the `/etc/mindx-dl/` directory and then re-import them using the old ``cert-importer`` tool.
- The import tool encryption requires the system to have a sufficient entropy pool. If the entropy pool is insufficient, the program may block. You can install the haveged component to supplement entropy.

    For installation commands, refer to the following:

    - For CentOS-like operating systems, run the `yum install haveged -y` command to install, and run the `systemctl start haveged` command to start the haveged component.
    - For Ubuntu-like operating systems, run the `apt install haveged -y` command to install, and run the `systemctl start haveged` command to start the haveged component.

**Importing the KubeConfig File<a name="section1538945217341a"></a>**

1. Log in to the K8s management node.
2. Create the `/etc/kubernetes/mindxdl` folder with permissions set to `750`.

    ```shell
    rm -rf /etc/kubernetes/mindxdl
    mkdir /etc/kubernetes/mindxdl
    chmod 750 /etc/kubernetes/mindxdl
    ```

3. Refer to [Kubernetes-related guidance](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) to create a KubeConfig file named `resilience-controller-cfg.conf`, where the `user` field in the KubeConfig file is `resilience-controller`. Place the KubeConfig file in the `/etc/kubernetes/mindxdl/` path.
4. Navigate to the Resilience Controller installation package extraction path, and set the lib folder to the environment variable `LD_LIBRARY_PATH` of the current window. This does not need to be persisted or inherited by other users (the certificate import Tool requires the configuration of the so package path related to its built-in encryption components).
    1. Run the following command to back up the environment variables.

        ```shell
        export LD_LIBRARY_PATH_BAK=${LD_LIBRARY_PATH}
        ```

    2. Run the following command to set the lib folder to the current environment variable `LD_LIBRARY_PATH`.

        ```shell
        export LD_LIBRARY_PATH=`pwd`/lib/:${LD_LIBRARY_PATH}
        ```

5. Run the following command to import the KubeConfig file for the Resilience Controller component.

    ```shell
    ./cert-importer -kubeConfig=/etc/kubernetes/mindxdl/resilience-controller-cfg.conf  -cpt=rc
    ```

    If information similar to the following is displayed, the import succeeded.

    ```ColdFusion
    encrypt kubeConfig successfully
    start to write data to disk
    [OP]import kubeConfig successfully
    change owner and set file mode successfully
    ```

    >[!NOTE]
    >- If the KubeConfig configuration file has been imported but the component still fails to connect to K8s, see [Cluster Scheduling Component Fails to Connect to K8s](https://gitcode.com/Ascend/mind-cluster/issues/344) for troubleshooting.
    >- When importing a certificate, `cert-importer` automatically creates the `/var/log/mindx-dl/cert-importer` directory with permissions `750` and owner `root:root`.

6. Run the following command to restore the backed-up environment variables.

    ```shell
    export LD_LIBRARY_PATH=${LD_LIBRARY_PATH_BAK}
    ```

**Table 3** Cluster scheduling component certificate configuration files

<a name="table252713572507"></a>
<table><thead align="left"><tr id="row4527257145015"><th class="cellrowborder" valign="top" width="17.5982401759824%" id="mcps1.2.5.1.1"><p id="p14528165725013"><a name="p14528165725013"></a><a name="p14528165725013"></a>Component</p>
</th>
<th class="cellrowborder" valign="top" width="19.24807519248075%" id="mcps1.2.5.1.2"><p id="p14528105765013"><a name="p14528105765013"></a><a name="p14528105765013"></a>File Path</p>
</th>
<th class="cellrowborder" valign="top" width="11.08889111088891%" id="mcps1.2.5.1.3"><p id="p105282572501"><a name="p105282572501"></a><a name="p105282572501"></a>Directory and File Owner</p>
</th>
<th class="cellrowborder" valign="top" width="52.064793520647946%" id="mcps1.2.5.1.4"><p id="p11528155755016"><a name="p11528155755016"></a><a name="p11528155755016"></a>Configuration File Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row9528155785012"><td class="cellrowborder" valign="top" width="17.5982401759824%" headers="mcps1.2.5.1.1 "><p id="p1528155715501"><a name="p1528155715501"></a><a name="p1528155715501"></a><span id="ph1488142812262"><a name="ph1488142812262"></a><a name="ph1488142812262"></a>Cluster scheduling component-related</span> root directory</p>
</td>
<td class="cellrowborder" valign="top" width="19.24807519248075%" headers="mcps1.2.5.1.2 "><p id="p7528357125019"><a name="p7528357125019"></a><a name="p7528357125019"></a>/etc/mindx-dl/</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="11.08889111088891%" headers="mcps1.2.5.1.3 "><p id="p1528457195011"><a name="p1528457195011"></a><a name="p1528457195011"></a>hwMindX:hwMindX</p>
<p id="p17618514195"><a name="p17618514195"></a><a name="p17618514195"></a></p>
<p id="p27716513196"><a name="p27716513196"></a><a name="p27716513196"></a></p>
<p id="p1532775483511"><a name="p1532775483511"></a><a name="p1532775483511"></a></p>
</td>
<td class="cellrowborder" valign="top" width="52.064793520647946%" headers="mcps1.2.5.1.4 "><p id="p9528857175013"><a name="p9528857175013"></a><a name="p9528857175013"></a>kmc_primary_store/master.ks: Automatically generated master key. Do not delete.</p>
<p id="p1152811579509"><a name="p1152811579509"></a><a name="p1152811579509"></a>.config/backup.ks: Automatically generated backup key. Do not delete.</p>
</td>
</tr>
<tr id="row207702393454"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p127701395458"><a name="p127701395458"></a><a name="p127701395458"></a><span id="ph1287272539"><a name="ph1287272539"></a><a name="ph1287272539"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1977073944518"><a name="p1977073944518"></a><a name="p1977073944518"></a>/etc/mindx-dl/resilience-controller/</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p176150132191"><a name="p176150132191"></a><a name="p176150132191"></a>.config/config6: Imported encrypted <span id="ph10615131313194"><a name="ph10615131313194"></a><a name="ph10615131313194"></a>K8s</span> KubeConfig file, used for connecting to <span id="ph761518136190"><a name="ph761518136190"></a><a name="ph761518136190"></a>K8s</span>.</p>
<p id="p16156132195"><a name="p16156132195"></a><a name="p16156132195"></a>.config6: Backup of the imported encrypted <span id="ph761517138198"><a name="ph761517138198"></a><a name="ph761517138198"></a>K8s</span> KubeConfig file.</p>
</td>
</tr>
</tbody>
</table>

**Table 4** Parameters of cert-importer

<a name="table18529165716504"></a>
<table><thead align="left"><tr id="row1852914572501"><th class="cellrowborder" valign="top" width="17.349999999999998%" id="mcps1.2.5.1.1"><p id="p5529175745012"><a name="p5529175745012"></a><a name="p5529175745012"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="19.41%" id="mcps1.2.5.1.2"><p id="p17529185775019"><a name="p17529185775019"></a><a name="p17529185775019"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="11.01%" id="mcps1.2.5.1.3"><p id="p1352935715507"><a name="p1352935715507"></a><a name="p1352935715507"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="52.23%" id="mcps1.2.5.1.4"><p id="p1552925711509"><a name="p1552925711509"></a><a name="p1552925711509"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row55021443133913"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p117491127105717"><a name="p117491127105717"></a><a name="p117491127105717"></a>-kubeConfig</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p18750132718575"><a name="p18750132718575"></a><a name="p18750132718575"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p975010276572"><a name="p975010276572"></a><a name="p975010276572"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9750162720574"><a name="p9750162720574"></a><a name="p9750162720574"></a>Path of the KubeConfig file to be imported.</p>
</td>
</tr>
<tr id="row45301657115017"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p8530165715011"><a name="p8530165715011"></a><a name="p8530165715011"></a>-cpt</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1353085745016"><a name="p1353085745016"></a><a name="p1353085745016"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p18558162664212"><a name="p18558162664212"></a><a name="p18558162664212"></a>rc</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9931932155418"><a name="p9931932155418"></a><a name="p9931932155418"></a>The component name for importing the certificate is rc, which stands for <span id="ph131541756961"><a name="ph131541756961"></a><a name="ph131541756961"></a>Resilience Controller</span>.</p>
</td>
</tr>
<tr id="row953045718504"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p5530195785020"><a name="p5530195785020"></a><a name="p5530195785020"></a>-encryptAlgorithm</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p12530125719509"><a name="p12530125719509"></a><a name="p12530125719509"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p35312571506"><a name="p35312571506"></a><a name="p35312571506"></a>9</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p135316571501"><a name="p135316571501"></a><a name="p135316571501"></a>Private key passphrase encryption algorithm:</p>
<a name="ul145317578507"></a><a name="ul145317578507"></a><ul id="ul145317578507"><li>8: AES128GCM</li><li>9: AES256GCM</li></ul>
<div class="note" id="note05311457165012"><a name="note05311457165012"></a><a name="note05311457165012"></a><span class="notetitle">Note</span><div class="notebody"><p id="p18531125718501"><a name="p18531125718501"></a><a name="p18531125718501"></a>Invalid parameter values will be reset to the default value.</p>
</div></div>
</td>
</tr>
<tr id="row18531135717506"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1253175785015"><a name="p1253175785015"></a><a name="p1253175785015"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p75318575501"><a name="p75318575501"></a><a name="p75318575501"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p853175715507"><a name="p853175715507"></a><a name="p853175715507"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p35317578505"><a name="p35317578505"></a><a name="p35317578505"></a>Queries the <span id="ph19991165205214"><a name="ph19991165205214"></a><a name="ph19991165205214"></a>Resilience Controller</span> version number.</p>
</td>
</tr>
<tr id="row2573635141612"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p138495616250"><a name="p138495616250"></a><a name="p138495616250"></a>-n</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p6384135614257"><a name="p6384135614257"></a><a name="p6384135614257"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p03848562252"><a name="p03848562252"></a><a name="p03848562252"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p2384145614255"><a name="p2384145614255"></a><a name="p2384145614255"></a>Whether to delete the <span id="ph418094814555"><a name="ph418094814555"></a><a name="ph418094814555"></a>KubeConfig</span> file after a successful import.</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true: Do not delete the <span id="ph39020528557"><a name="ph39020528557"></a><a name="ph39020528557"></a>KubeConfig</span> file after a successful import.</li><li>false: Delete the <span id="ph7200135465511"><a name="ph7200135465511"></a><a name="ph7200135465511"></a>KubeConfig</span> file after a successful import.</li></ul>
</td>
</tr>
<tr id="row5485341194020"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p6232132695813"><a name="p6232132695813"></a><a name="p6232132695813"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p623242655813"><a name="p623242655813"></a><a name="p623242655813"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1232182685820"><a name="p1232182685820"></a><a name="p1232182685820"></a>/var/log/mindx-dl/cert-importer/cert-importer.log</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p102331826175814"><a name="p102331826175814"></a><a name="p102331826175814"></a>Tool runtime log file. The naming format for the dumped file is: cert-importer-<trigger dump time/>.log, for example: cert-importer-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row8384164173412"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1138412411345"><a name="p1138412411345"></a><a name="p1138412411345"></a>-updateMk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1738494133414"><a name="p1738494133414"></a><a name="p1738494133414"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1238434133420"><a name="p1238434133420"></a><a name="p1238434133420"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p17211144255612"><a name="p17211144255612"></a><a name="p17211144255612"></a>Whether to update the master key of the KMC encryption component.</p>
<a name="ul154871314165520"></a><a name="ul154871314165520"></a><ul id="ul154871314165520"><li>true: Update the master key.</li><li>false: Do not update the master key.</li></ul>
</td>
</tr>
<tr id="row1397106345"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p53986014348"><a name="p53986014348"></a><a name="p53986014348"></a>-updateRk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p139811020345"><a name="p139811020345"></a><a name="p139811020345"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p139850143413"><a name="p139850143413"></a><a name="p139850143413"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p741952510563"><a name="p741952510563"></a><a name="p741952510563"></a>Whether to update the root key of the KMC encryption component.</p>
<a name="ul14451957145511"></a><a name="ul14451957145511"></a><ul id="ul14451957145511"><li>true: Update the root key.</li><li>false: Do not update the root key.</li></ul>
</td>
</tr>
<tr id="row050462052716"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p13504720112715"><a name="p13504720112715"></a><a name="p13504720112715"></a>-h or -help</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1350422002713"><a name="p1350422002713"></a><a name="p1350422002713"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1650420209273"><a name="p1650420209273"></a><a name="p1650420209273"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p4505820152717"><a name="p4505820152717"></a><a name="p4505820152717"></a>Displays help information.</p>
</td>
</tr>
</tbody>
</table>

## Installing Resilience Controller<a name="ZH-CN_TOPIC_0000002479226460"></a>

- To use elastic training, Resilience Controller must be installed. When Resilience Controller connects to K8s, you can choose to authenticate using a ServiceAccount or a KubeConfig file. For differences between the two methods, see [Differences Between ServiceAccount and KubeConfig](../../../appendix.md#differences-between-using-serviceaccount-and-kubeconfig).
- Users who do not use elastic training can skip this chapter.

**Procedure<a name="section0531457718"></a>**

1. Log in to the K8s management node as the `root` user and run the following command to check whether the Resilience Controller image and version number are correct.

    ```shell
    docker images | grep resilience-controller
    ```

    The Response Example is as follows:

    ```ColdFusion
    resilience-controller                      v26.0.0             c532e9d0889c        About an hour ago         142MB
    ```

    - If correct, proceed to [Step 2](#li10743192474541).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

2. <a name="li10743192474541"></a>Copy the YAML file from the extracted Resilience Controller package directory to any directory on the Kubernetes management node.
3. If you do not need to modify the component startup parameters, you can skip this step. Otherwise, modify the Resilience Controller startup parameters in the YAML file based on your actual situation. For details on the startup parameters, see [Table 1](#table195504370194), or run `./resilience-controller -h` to view the parameter descriptions.
4. In the directory where the YAML file is located on the management node, run the following command to start Resilience Controller.

    - If the KubeConfig certificate has not been imported, run the following command.

        ```shell
        kubectl apply -f resilience-controller-v{version}.yaml
        ```

        Example:

        ```ColdFusion
        serviceaccount/resilience-controller created
        clusterrole.rbac.authorization.k8s.io/pods-resilience-controller-role created
        clusterrolebinding.rbac.authorization.k8s.io/resilience-controller-rolebinding created
        deployment.apps/resilience-controller created
       ```

    - If the KubeConfig certificate has been imported, run the following command.

        ```shell
        kubectl apply -f resilience-controller-without-token-v{version}.yaml
        ```

        Example:

        ```ColdFusion
        deployment.apps/resilience-controller created
        ```

5. Run the following command to check whether the component is installed successfully.

    ```shell
    kubectl get pod -n mindx-dl
    ```

The response example is as follows. **Running** indicates that the component startup is successful.

    ```ColdFusion
    NAME                                            READY    STATUS      RESTARTS   AGE
    ...
    resilience-controller-7667495b6b-hwmjw   1/1     Running   0         11s
    ...
    ```

>[!NOTE]
>
>- If the pod status of a component is not `Running` after installation, see [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
>- If the pod status of a component is `ContainerCreating` after installation, see [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343).
>- If the component fails to start, see [Cluster Scheduling Component Startup Failure, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
>- If the component starts successfully but the corresponding pod cannot be found, see [Component Startup YAML Executed Successfully but Corresponding Pod Not Found](https://gitcode.com/Ascend/mind-cluster/issues/345).

**Parameter Description<a name="section1868556161717"></a>**

**Table 1** Resilience Controller startup parameters

<a name="table195504370194"></a>
<table><thead align="left"><tr id="row10550173721915"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p1855053711192"><a name="p1855053711192"></a><a name="p1855053711192"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p355063710197"><a name="p355063710197"></a><a name="p355063710197"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p055073781916"><a name="p055073781916"></a><a name="p055073781916"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3550237171920"><a name="p3550237171920"></a><a name="p3550237171920"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row3551143715196"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p65517376197"><a name="p65517376197"></a><a name="p65517376197"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p19551153781918"><a name="p19551153781918"></a><a name="p19551153781918"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15511378194"><a name="p15511378194"></a><a name="p15511378194"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18551173791915"><a name="p18551173791915"></a><a name="p18551173791915"></a>Whether to query the <span id="ph151418415511"><a name="ph151418415511"></a><a name="ph151418415511"></a>Resilience Controller</span> version number.</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true: Query.</li><li>false: Do not query.</li></ul>
</td>
</tr>
<tr id="row8551137161913"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1155183715199"><a name="p1155183715199"></a><a name="p1155183715199"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p105511137141920"><a name="p105511137141920"></a><a name="p105511137141920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p755113373192"><a name="p755113373192"></a><a name="p755113373192"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6551123716195"><a name="p6551123716195"></a><a name="p6551123716195"></a>Log level:</p>
<a name="ul655113715194"></a><a name="ul655113715194"></a><ul id="ul655113715194"><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>
</td>
</tr>
<tr id="row1455163771915"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p10551143710191"><a name="p10551143710191"></a><a name="p10551143710191"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p14551193781920"><a name="p14551193781920"></a><a name="p14551193781920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1655193715195"><a name="p1655193715195"></a><a name="p1655193715195"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p7551183716190"><a name="p7551183716190"></a><a name="p7551183716190"></a>Log backup retention period. The value ranges from 7 to 700, in days.</p>
</td>
</tr>
<tr id="row175527378195"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p455223751910"><a name="p455223751910"></a><a name="p455223751910"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p195521937131913"><a name="p195521937131913"></a><a name="p195521937131913"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15552137111920"><a name="p15552137111920"></a><a name="p15552137111920"></a>/var/log/mindx-dl/resilience-controller/run.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18552143713199"><a name="p18552143713199"></a><a name="p18552143713199"></a>Log file. When a single log file exceeds 20 MB, automatic rotation is triggered. The maximum file size cannot be modified. The naming format of rotated files is: run-rotation_time.log, for example, run-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row1655213379191"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p11552163741920"><a name="p11552163741920"></a><a name="p11552163741920"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1552137171918"><a name="p1552137171918"></a><a name="p1552137171918"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p7552193711192"><a name="p7552193711192"></a><a name="p7552193711192"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p555233718199"><a name="p555233718199"></a><a name="p555233718199"></a>Maximum number of rotated log files to retain. The value ranges from 1 to 30.</p>
</td>
</tr>
<tr id="row33119022219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1532160192215"><a name="p1532160192215"></a><a name="p1532160192215"></a>-h or -help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p123213019227"><a name="p123213019227"></a><a name="p123213019227"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1832100102210"><a name="p1832100102210"></a><a name="p1832100102210"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1328016224"><a name="p1328016224"></a><a name="p1328016224"></a>Display help information.</p>
</td>
</tr>
</tbody>
</table>
