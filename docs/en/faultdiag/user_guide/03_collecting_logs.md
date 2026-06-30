# Collecting Logs<a name="ZH-CN_TOPIC_0000001628171657"></a>

## Log Collection Directory Structure<a name="ZH-CN_TOPIC_0000001629558433"></a>

This section describes the structure of the directory to be cleaned. You can collect logs and store them in the corresponding structure.

>[!NOTE]
>
>- The size of the log file in the `Ascend-fd parse` input directory affects the efficiency of running the cleaning command. The total file size must be less than 5 GB, and the total number of files cannot exceed 1,000,000.
>- The size of a CANN App log file must be less than 20 MB.
>- The size of the NPU status monitoring metric file, monitoring metric file of NPU network port statistics, and host resource information file must be less than 512 MB.
>- The size of a user training or inference log is not limited. By default, only the last log file of 1 MB size is read.
>- The host OS logs include `messages`, `dmesg`, `vmcore_dmesg.txt`, and `sysmonitor.log`. The dump size of a single file must be less than 512 MB. The latest `dmesg` log is used, and the maximum number of lines is 100,000.
>- The locations of `process_log`, `environment_check`, `device_log`, `dl_log`, `mindie`, and `amct_log` are not restricted. They can be stored in any location in the collection directory.
>- If you perform training or inference in a container, save logs, such as user training or inference logs and CANN App logs, to the host in a timely manner.
>- Collect the NPU environment check file before or after training/inference, NPU network port statistics monitoring metric file, NPU status monitoring metric file, host resource information, host OS logs, device logs, MindCluster logs, MindIE logs, and AMCT logs on the host.
>- After dump is triggered by volcano-scheduler and volcano-controller, the dumped logs compressed in gzip format will not be read. Ensure that related logs are contained in the `volcano-scheduler.log` and `volcano-controller.log` files that are not dumped during collection.
>- You can collect console logs of all pods on the master node of the Kubernetes cluster and store all MindIE Pod console logs in a specified directory on a node.
>- An aging mechanism is introduced to MindIE Pod console logs. If the collected MindIE Pod console logs do not contain instance node information, multi-instance fault diagnosis will not be supported.
>- After MindIO logs are collected, they are dumped to the `/dl_log` directory. In the later release, they will be dumped to the collection directory.

- You can summarize all logs to the same collection directory for cleaning. The following is an example of the directory structure of the files to be cleaned.
    - Host log directory structure

        ```text
        Collection directory
        |-- messages              # Host OS logs
        |-- dmesg                # Host kernel message logs
        |-- crash
            |--Host + Fault timestamp (eg:127.xx.xx.1-2024-09-23-11:25:29)
                |-- vmcore_dmesg.txt     # Host kernel message log saved when the system breaks down
        |-- sysmonitor.log       # System monitoring log
        |-- rank-0.txt           # Training and inference console log file
        |-- dmidecode.txt        # dmidecode output log file
        ...
        |-- rank-7.txt           # Training and inference console log file
        |-- process_log          # Original App logs of CANN. The directory name must be process_log.
        |-- device_log           # Device logs, which must be stored in the device_log directory.
        |-- dl_log               #  MindCluster logs. The directory must be dl_log.
            |-- devicePlugin       # Ascend Device Plugin logs
            |-- noded              # NodeD logs
            |-- ascend-docker-runtime        # Ascend Docker Runtime logs
            |-- volcano-scheduler            # volcano-scheduler logs
            |-- volcano-controller           # volcano-controller logs
            |-- npu-exporter                # NPU Exporter logs
            |-- ttp_log                      # MindIO logs
        |-- mindie               # MindIE logs
            |-- log
                -- debug        # MindIE run logs
                |-- security     # MindIE audit logs
                |-- mindie_cluster_log     # MindIE Pod console logs
        |-- amct_log             # AMCT logs
        |-- bus_log              # LCNE logs (Ascend 950)
        |-- environment_check # Information about the NPU network port, status, and resource
            |-- npu_smi_0_details.csv   # NPU status monitoring metrics
             ...
            |-- npu_smi_7_details.csv   # NPU status monitoring metrics
            |-- npu_0_details.csv         # NPU network port monitoring metrics
             ...
            |-- npu_7_details.csv       # NPU network port monitoring metrics
            |-- npu_info_before/after.txt   # NPU environment check file before or after training and inference
            |-- host_metrics_{core_num}.json  # Host resource monitoring metrics
        ```

    - BMC log directory structure

        ```text
        Collection_directory/dump_info/AppDump/*/*.log
        Collection_directory/dump_info/DeviceDump/*/*.log
        Collection_directory/dump_info/LogDump/*/*.log
        Collection_directory/dump_info/AppDump/frudata/fruinfo.txt # BMC extension board SNs
        Collection_directory/dump_info/AppDump/chassis/mdb_info.log # BMC SuperPoD information
        ```

    - LCNE log directory structure

        ```text
        Collection_directory/*/diagnostic_information/slot_1/tempdir/devm_bddrvadp.log # LCNE extension board SNs
        Collection_directory/*/diag_display_info.txt # LCNE SuperPoD information
        Collection_directory/*/log.log
        Collection_directory/*/log_1_*.log
        ```

        The table below describes the log files stored in each directory.

        **Table 1** Log file list

        <a name="table12937722195315"></a>
        <table><thead align="left"><tr id="row693332215532"><th class="cellrowborder" valign="top" width="16.150000000000002%" id="mcps1.2.5.1.1"><p id="p139331922175315"><a name="p139331922175315"></a><a name="p139331922175315"></a>File Type</p>
        </th>
        <th class="cellrowborder" valign="top" width="21.26%" id="mcps1.2.5.1.2"><p id="p493392214536"><a name="p493392214536"></a><a name="p493392214536"></a><strong id="b493316221533"><a name="b493316221533"></a><a name="b493316221533"></a>Log File</strong></p>
        </th>
        <th class="cellrowborder" valign="top" width="20.39%" id="mcps1.2.5.1.3"><p id="p13933142265319"><a name="p13933142265319"></a><a name="p13933142265319"></a><strong id="b17933102265312"><a name="b17933102265312"></a><a name="b17933102265312"></a>ile Description</strong></p>
        </th>
        <th class="cellrowborder" valign="top" width="42.199999999999996%" id="mcps1.2.5.1.4"><p id="p993372275310"><a name="p993372275310"></a><a name="p993372275310"></a><strong id="b189334229535"><a name="b189334229535"></a><a name="b189334229535"></a>Storage Directory</strong></p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row17933112220536"><td class="cellrowborder" rowspan="2" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p10933022195316"><a name="p10933022195316"></a><a name="p10933022195316"></a>CANN App logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p6933622185319"><a name="p6933622185319"></a><a name="p6933622185319"></a>plog-<em id="i1493313229537"><a name="i1493313229537"></a><a name="i1493313229537"></a>{pid}</em>_<em id="i149334227536"><a name="i149334227536"></a><a name="i149334227536"></a>{time}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p1093313229534"><a name="p1093313229534"></a><a name="p1093313229534"></a>Host-side App log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p1493312235317"><a name="p1493312235317"></a><a name="p1493312235317"></a>Collection_directory/process_log/debug or run/plog/plog-<em id="i2933622125316"><a name="i2933622125316"></a><a name="i2933622125316"></a>{pid}</em>_<em id="i6933222145320"><a name="i6933222145320"></a><a name="i6933222145320"></a>{time}</em>.log</p>
        </td>
        </tr>
        <tr id="row159331225539"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10933192214532"><a name="p10933192214532"></a><a name="p10933192214532"></a>device-<em id="i19933162225310"><a name="i19933162225310"></a><a name="i19933162225310"></a>{pid}</em>_<em id="i129331223539"><a name="i129331223539"></a><a name="i129331223539"></a>{time}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7933722165312"><a name="p7933722165312"></a><a name="p7933722165312"></a>Device-side App log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p4933722175314"><a name="p4933722175314"></a><a name="p4933722175314"></a>Collection_directory/process_log/debug or run/device-<em id="i139331122135318"><a name="i139331122135318"></a><a name="i139331122135318"></a>{id}</em>/device-<em id="i293313226535"><a name="i293313226535"></a><a name="i293313226535"></a>{pid}</em>_<em id="i17933222185310"><a name="i17933222185310"></a><a name="i17933222185310"></a>{time}</em>.log</p>
        </td>
        </tr>
        <tr id="row1993482285313"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p1493317220533"><a name="p1493317220533"></a><a name="p1493317220533"></a>User training and inference logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p9934322185315"><a name="p9934322185315"></a><a name="p9934322185315"></a>rank<em id="i99341622195317"><a name="i99341622195317"></a><a name="i99341622195317"></a>-{id}</em>.txt</p>
        <p id="p19934722135313"><a name="p19934722135313"></a><a name="p19934722135313"></a>rank<em id="i2934162217537"><a name="i2934162217537"></a><a name="i2934162217537"></a>-{id}</em>.log</p>
        <p id="p1934192215536"><a name="p1934192215536"></a><a name="p1934192215536"></a>worker<em id="i493422219532"><a name="i493422219532"></a><a name="i493422219532"></a>-{id}</em>.txt</p>
        <p id="p793419224533"><a name="p793419224533"></a><a name="p793419224533"></a>worker<em id="i09341722175313"><a name="i09341722175313"></a><a name="i09341722175313"></a>-{id}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p693432219531"><a name="p693432219531"></a><a name="p693432219531"></a>Training and inference console logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><a name="ul493492215531"></a><a name="ul493492215531"></a><ul id="ul493492215531"><li>Collection_directory/rank-<em id="i993414221535"><a name="i993414221535"></a><a name="i993414221535"></a>{id</em><em id="i1893410222530"><a name="i1893410222530"></a><a name="i1893410222530"></a>}</em>.*?.txt</li><li>Collection_directory/rank-<em id="i1934112220532"><a name="i1934112220532"></a><a name="i1934112220532"></a>{id</em><em id="i993472245319"><a name="i993472245319"></a><a name="i993472245319"></a>}</em>.*?.log</li><li>Collection_directory/worker-<em id="i1693413224532"><a name="i1693413224532"></a><a name="i1693413224532"></a>{id}</em>.*?.log</li><li>Collection_directory/worker-<em id="i129341322145314"><a name="i129341322145314"></a><a name="i129341322145314"></a>{id}</em>.*?.txt</li></ul>
        </td>
        </tr>
        <tr id="row6934122265317"><td class="cellrowborder" rowspan="4" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p119341722135319"><a name="p119341722135319"></a><a name="p119341722135319"></a>NPU network port resource information</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p7934202225314"><a name="p7934202225314"></a><a name="p7934202225314"></a>npu_info_before.txt</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p29340221534"><a name="p29340221534"></a><a name="p29340221534"></a>NPU network port check before training and inference</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p15934192245318"><a name="p15934192245318"></a><a name="p15934192245318"></a>Collection_directory/environment_check/npu_info_before.txt</p>
        </td>
        </tr>
        <tr id="row189341222165320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p13934102295314"><a name="p13934102295314"></a><a name="p13934102295314"></a>npu_info_after.txt</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1493411228532"><a name="p1493411228532"></a><a name="p1493411228532"></a>NPU network port check after training and inference</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1993452218536"><a name="p1993452218536"></a><a name="p1993452218536"></a>Collection_directory/environment_check/npu_info_after.txt</p>
        </td>
        </tr>
        <tr id="row12934182295318"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p493442275316"><a name="p493442275316"></a><a name="p493442275316"></a>npu_smi_<em id="i493482205317"><a name="i493482205317"></a><a name="i493482205317"></a>{npu_id}</em>_details.csv</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1193462265313"><a name="p1193462265313"></a><a name="p1193462265313"></a>NPU status monitoring metric file</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p3934022125312"><a name="p3934022125312"></a><a name="p3934022125312"></a>Collection_directory/environment_check/npu_smi_<em id="i129341722135313"><a name="i129341722135313"></a><a name="i129341722135313"></a>{npu_id}</em>_details.csv</p>
        </td>
        </tr>
        <tr id="row1593418221533"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p12934192255315"><a name="p12934192255315"></a><a name="p12934192255315"></a>npu_<em id="i199341122115317"><a name="i199341122115317"></a><a name="i199341122115317"></a>{npu_id}</em>_details.csv</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1793482214532"><a name="p1793482214532"></a><a name="p1793482214532"></a>NPU network port monitoring metric file</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p18934122265316"><a name="p18934122265316"></a><a name="p18934122265316"></a>Collection_directory/environment_check/npu_<em id="i14934622125317"><a name="i14934622125317"></a><a name="i14934622125317"></a>{npu_id}</em>_details.csv</p>
        </td>
        </tr>
        <tr id="row89341122195315"><td class="cellrowborder" rowspan="2" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p693462215320"><a name="p693462215320"></a><a name="p693462215320"></a>Host resource information</p>
        <p id="p14185102720010"><a name="p14185102720010"></a><a name="p14185102720010"></a></p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p1934622125311"><a name="p1934622125311"></a><a name="p1934622125311"></a>host_metrics_<em id="i19934122125311"><a name="i19934122125311"></a><a name="i19934122125311"></a>{core_num}</em>.json</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p4934192215538"><a name="p4934192215538"></a><a name="p4934192215538"></a>Host resource monitoring metric file</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p4934192216538"><a name="p4934192216538"></a><a name="p4934192216538"></a>Collection_directory/environment_check/host_metrics_<em id="i17934622105318"><a name="i17934622105318"></a><a name="i17934622105318"></a>{core_num}</em>.json</p>
        </td>
        </tr>
        <tr id="row118418271603"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1518513276018"><a name="p1518513276018"></a><a name="p1518513276018"></a>dmidecode.txt</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1518502716012"><a name="p1518502716012"></a><a name="p1518502716012"></a>DMI log file on the host</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1418515272016"><a name="p1418515272016"></a><a name="p1418515272016"></a>Collection_directory/dmidecode.txt</p>
        </td>
        </tr>
        <tr id="row1993412255313"><td class="cellrowborder" rowspan="4" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p793413223536"><a name="p793413223536"></a><a name="p793413223536"></a>Host-side logs</p>
        <p id="p5934152275312"><a name="p5934152275312"></a><a name="p5934152275312"></a></p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p9934102218537"><a name="p9934102218537"></a><a name="p9934102218537"></a>dmesg</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p3934142215531"><a name="p3934142215531"></a><a name="p3934142215531"></a>Host kernel message file</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p2934122214533"><a name="p2934122214533"></a><a name="p2934122214533"></a>Collection_directory/dmesg</p>
        </td>
        </tr>
        <tr id="row19935152245315"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p8934162214536"><a name="p8934162214536"></a><a name="p8934162214536"></a>sysmonitor.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p993412245316"><a name="p993412245316"></a><a name="p993412245316"></a>Host system monitoring file</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p179347221530"><a name="p179347221530"></a><a name="p179347221530"></a>Collection_directory/sysmonitor.log</p>
        </td>
        </tr>
        <tr id="row0935422155317"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p2093582265313"><a name="p2093582265313"></a><a name="p2093582265313"></a>messages-*?</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p493516222539"><a name="p493516222539"></a><a name="p493516222539"></a>Host OS log file</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1193532265318"><a name="p1193532265318"></a><a name="p1193532265318"></a>Collection_directory/messages-*?</p>
        </td>
        </tr>
        <tr id="row293522265317"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p69351922185310"><a name="p69351922185310"></a><a name="p69351922185310"></a>vmcore_dmesg.txt</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1893512218538"><a name="p1893512218538"></a><a name="p1893512218538"></a>Host kernel message file saved during a system crash</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p9935132235315"><a name="p9935132235315"></a><a name="p9935132235315"></a>Collection_directory/crash/Host + Fault timestamp (eg: 127.xx.xx.1-2024-09-23-11:25:29)/vmcore_dmesg.txt</p>
        </td>
        </tr>
        <tr id="row1193562265314"><td class="cellrowborder" rowspan="7" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p149350225535"><a name="p149350225535"></a><a name="p149350225535"></a>Device-side logs</p>
        <p id="p293522215530"><a name="p293522215530"></a><a name="p293522215530"></a></p>
        <p id="p16935182211533"><a name="p16935182211533"></a><a name="p16935182211533"></a></p>
        <p id="p493572255315"><a name="p493572255315"></a><a name="p493572255315"></a></p>
        <p id="p3935422115314"><a name="p3935422115314"></a><a name="p3935422115314"></a></p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p1693552225316"><a name="p1693552225316"></a><a name="p1693552225316"></a>device-os_<em id="i11935202218535"><a name="i11935202218535"></a><a name="i11935202218535"></a>{time}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p1393510222534"><a name="p1393510222534"></a><a name="p1393510222534"></a>Device-side Ctrl CPU system log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p99351722105318"><a name="p99351722105318"></a><a name="p99351722105318"></a>Collection_directory/device_log/slog/dev-os-<em id="i7935182215316"><a name="i7935182215316"></a><a name="i7935182215316"></a>{id}</em>/debug或run/device-os/device-os_<em id="i1093510226536"><a name="i1093510226536"></a><a name="i1093510226536"></a>{time}</em>.log</p>
        </td>
        </tr>
        <tr id="row159351022125320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p2935122218532"><a name="p2935122218532"></a><a name="p2935122218532"></a>event_<em id="i10935162295311"><a name="i10935162295311"></a><a name="i10935162295311"></a>{time}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12935132215316"><a name="p12935132215316"></a><a name="p12935132215316"></a>EVENT-level Ctrl CPU system log on the device</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p11935142219539"><a name="p11935142219539"></a><a name="p11935142219539"></a>Ascend HDK 23.0.3 and later:</p>
        <p id="p293572212532"><a name="p293572212532"></a><a name="p293572212532"></a>Collection_directory/device_log/slog/dev-os-<em id="i1293517221539"><a name="i1293517221539"></a><a name="i1293517221539"></a>{id}</em>/run/event/event_<em id="i1935182215539"><a name="i1935182215539"></a><a name="i1935182215539"></a>{time}</em>.log</p>
        </td>
        </tr>
        <tr id="row493642295315"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p19351222125310"><a name="p19351222125310"></a><a name="p19351222125310"></a>device-<em id="i99351722195317"><a name="i99351722195317"></a><a name="i99351722195317"></a>{id}</em>_<em id="i16935422175310"><a name="i16935422175310"></a><a name="i16935422175310"></a>{time}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p119351522125319"><a name="p119351522125319"></a><a name="p119351522125319"></a>Device-side non-Ctrl CPU system log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p4935722185312"><a name="p4935722185312"></a><a name="p4935722185312"></a>Ascend HDK 23.0.RC3:</p>
        <p id="p109351225536"><a name="p109351225536"></a><a name="p109351225536"></a>Collection_directory/device_log/slog/dev-os-<em id="i99354227530"><a name="i99354227530"></a><a name="i99354227530"></a>{id}</em>/device-<em id="i793592216534"><a name="i793592216534"></a><a name="i793592216534"></a>{id}</em>/device-<em id="i89354225538"><a name="i89354225538"></a><a name="i89354225538"></a>{id}</em>_<em id="i129356224537"><a name="i129356224537"></a><a name="i129356224537"></a>{time}</em>.log</p>
        <p id="p129353225539"><a name="p129353225539"></a><a name="p129353225539"></a>Ascend HDK 23.0.3 and later:</p>
        <p id="p1793662255315"><a name="p1793662255315"></a><a name="p1793662255315"></a>Collection_directory/device_log/slog/dev-os-<em id="i19935152295311"><a name="i19935152295311"></a><a name="i19935152295311"></a>{id}</em>/debug/device-<em id="i16935102219531"><a name="i16935102219531"></a><a name="i16935102219531"></a>{id}</em>/device-<em id="i18936182211536"><a name="i18936182211536"></a><a name="i18936182211536"></a>{id}</em>_<em id="i14936422105318"><a name="i14936422105318"></a><a name="i14936422105318"></a>{time}</em>.log</p>
        </td>
        </tr>
        <tr id="row169363227534"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p49363229530"><a name="p49363229530"></a><a name="p49363229530"></a>history.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7936132215531"><a name="p7936132215531"></a><a name="p7936132215531"></a>Black Box log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p18936322155316"><a name="p18936322155316"></a><a name="p18936322155316"></a>Collection_directory/device_log/hisi_logs/device-<em id="i59369226534"><a name="i59369226534"></a><a name="i59369226534"></a>{id}</em>/history.log</p>
        </td>
        </tr>
        <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>kernel.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>NPU kernel log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>Collection_directory/device_log/hisi_logs/device-<em>{id}/{time}</em>/log/kernel.log</p>
        </td>
        </tr>
        <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>os_info.txt</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>Device OS information</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>Collection_directory/device_log/hisi_logs/device-<em>{id}/{time}</em>/bbox/os/os_info.txt</p>
        </td>
        </tr>
        <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>hbm.txt</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>Device-side on-chip memory log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>Collection_directory/device_log/hisi_logs/device-<em>{id}/{time}</em>/mntn/hbm.txt</p>
        </td>
        </tr>
        <tr id="row119369226534"><td class="cellrowborder" rowspan="7" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p139367226537"><a name="p139367226537"></a><a name="p139367226537"></a><span id="ph19936162211535"><a name="ph19936162211535"></a><a name="ph19936162211535"></a>MindCluster</span> logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p149361822115320"><a name="p149361822115320"></a><a name="p149361822115320"></a>devicePlugin*.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p1936222165316"><a name="p1936222165316"></a><a name="p1936222165316"></a>SuperPoD device log and <span id="ph59365226535"><a name="ph59365226535"></a><a name="ph59365226535"></a>Ascend Device Plugin</span> log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p993616221532"><a name="p993616221532"></a><a name="p993616221532"></a>Collection_directory/dl_log/devicePlugin/devicePlugin*.log</p>
        </td>
        </tr>
        <tr id="row2093672255310"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1493642215320"><a name="p1493642215320"></a><a name="p1493642215320"></a>noded*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p109361822165314"><a name="p109361822165314"></a><a name="p109361822165314"></a>AI server log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p793602225317"><a name="p793602225317"></a><a name="p793602225317"></a>Collection_directory/dl_log/noded/noded*.log</p>
        </td>
        </tr>
        <tr id="row1793652265316"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9936102213539"><a name="p9936102213539"></a><a name="p9936102213539"></a>runtime-run*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12936132210532"><a name="p12936132210532"></a><a name="p12936132210532"></a><span id="ph193622217536"><a name="ph193622217536"></a><a name="ph193622217536"></a></span>Log generated when ascend-docker-runtime is executed</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p2936172225314"><a name="p2936172225314"></a><a name="p2936172225314"></a>Collection_directory/dl_log/ascend-docker-runtime/runtime-run*.log</p>
        </td>
        </tr>
        <tr id="row13936172212537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p199365220535"><a name="p199365220535"></a><a name="p199365220535"></a>hook-run*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p393682217536"><a name="p393682217536"></a><a name="p393682217536"></a><span id="ph7936222145311"><a name="ph7936222145311"></a><a name="ph7936222145311"></a></span>Log generated when ascend-docker-hook is executed</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p393672255310"><a name="p393672255310"></a><a name="p393672255310"></a>Collection_directory/dl_log/ascend-docker-runtime/</p>
        <p id="p19936202213537"><a name="p19936202213537"></a><a name="p19936202213537"></a>hook-run*.log</p>
        </td>
        </tr>
        <tr id="row159361322195317"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p2936102285313"><a name="p2936102285313"></a><a name="p2936102285313"></a>volcano-scheduler*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p09364226530"><a name="p09364226530"></a><a name="p09364226530"></a><span id="ph109361622135319"><a name="ph109361622135319"></a><a name="ph109361622135319"></a></span>volcano-scheduler log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p6936172255314"><a name="p6936172255314"></a><a name="p6936172255314"></a>Collection_directory/dl_log/volcano-scheduler/</p>
        <p id="p119364226534"><a name="p119364226534"></a><a name="p119364226534"></a>volcano-scheduler*.log</p>
        </td>
        </tr>
        <tr id="row1593692245310"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3936112245314"><a name="p3936112245314"></a><a name="p3936112245314"></a>volcano-controller*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2093632212539"><a name="p2093632212539"></a><a name="p2093632212539"></a><span id="ph79361122165314"><a name="ph79361122165314"></a><a name="ph79361122165314"></a></span>volcano-controller log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p159362022175316"><a name="p159362022175316"></a><a name="p159362022175316"></a>Collection_directory/dl_log/volcano-controller/</p>
        <p id="p8936922185317"><a name="p8936922185317"></a><a name="p8936922185317"></a>volcano-controller*.log</p>
        </td>
        </tr>
        <tr id="row16936102255311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1193611221535"><a name="p1193611221535"></a><a name="p1193611221535"></a>npu-exporter*.log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15936122219532"><a name="p15936122219532"></a><a name="p15936122219532"></a><span id="ph89361220535"><a name="ph89361220535"></a><a name="ph89361220535"></a>NPU Exporter</span> log</p>
        </td>
        <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p19936122175317"><a name="p19936122175317"></a><a name="p19936122175317"></a>Collection_directory/dl_log/npu-exporter/</p>
        <p id="p493610224532"><a name="p493610224532"></a><a name="p493610224532"></a>npu-exporter*.log</p>
        </td>
        </tr>
        <tr id="row14937122218532"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p19365227531"><a name="p19365227531"></a><a name="p19365227531"></a><span id="ph1936112216538"><a name="ph1936112216538"></a><a name="ph1936112216538"></a>MindIE</span> logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p1093652215535"><a name="p1093652215535"></a><a name="p1093652215535"></a>mindie-<em id="i79369222536"><a name="i79369222536"></a><a name="i79369222536"></a>{module}</em>_<em id="i19936102211536"><a name="i19936102211536"></a><a name="i19936102211536"></a>{pid}</em>_<em id="i693682205314"><a name="i693682205314"></a><a name="i693682205314"></a>{datetime}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p15937422205315"><a name="p15937422205315"></a><a name="p15937422205315"></a><span id="ph149361922175311"><a name="ph149361922175311"></a><a name="ph149361922175311"></a>MindIE Server</span>, <span id="ph1493672212536"><a name="ph1493672212536"></a><a name="ph1493672212536"></a>MindIE LLM</span>, <span id="ph2093712210533"><a name="ph2093712210533"></a><a name="ph2093712210533"></a>MindIE SD</span>, <span id="ph49372223533"><a name="ph49372223533"></a><a name="ph49372223533"></a>MindIE RT</span>, <span id="ph109374224535"><a name="ph109374224535"></a><a name="ph109374224535"></a>MindIE Torch</span>, span id="ph5937162285320"><a name="ph5937162285320"></a><a name="ph5937162285320"></a>MindIE MS, <span id="ph2093712205312"><a name="ph2093712205312"></a><a name="ph2093712205312"></a>MindIE Benchmark</span>, and <span id="ph1093752225310"><a name="ph1093752225310"></a><a name="ph1093752225310"></a>MindIE Client</span> logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p1393742217531"><a name="p1393742217531"></a><a name="p1393742217531"></a>Collection_directory/mindie/log/debug/mindie-<em id="i19937152285320"><a name="i19937152285320"></a><a name="i19937152285320"></a>{module}</em>_<em id="i4937192219539"><a name="i4937192219539"></a><a name="i4937192219539"></a>{pid}</em>_<em id="i0937152219532"><a name="i0937152219532"></a><a name="i0937152219532"></a>{datetime}</em>.log</p>
        </td>
        </tr>
        <tr id="row39371522175316"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p4937122225319"><a name="p4937122225319"></a><a name="p4937122225319"></a>AMCT log</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p7937192235317"><a name="p7937192235317"></a><a name="p7937192235317"></a>amct_<em id="i7937322135310"><a name="i7937322135310"></a><a name="i7937322135310"></a>{framework}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p109371922175319"><a name="p109371922175319"></a><a name="p109371922175319"></a>AMCT log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p1693732217532"><a name="p1693732217532"></a><a name="p1693732217532"></a>Collection_directory/amct_log/amct_<em id="i1393742212536"><a name="i1393742212536"></a><a name="i1393742212536"></a>{framework}</em>.log</p>
        </td>
        </tr>
        <tr id="row29371022185312"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p9937182225310"><a name="p9937182225310"></a><a name="p9937182225310"></a>BMC logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p79371422165316"><a name="p79371422165316"></a><a name="p79371422165316"></a>All out-of-band .log files</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p5937132215311"><a name="p5937132215311"></a><a name="p5937132215311"></a>All out-of-band logs collected in one-click mode</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p46581757586"><a name="p46581757586"></a><a name="p46581757586"></a>Collection_directory/dump_info/AppDump/*/*.log</p>
        <p id="p56581657582"><a name="p56581657582"></a><a name="p56581657582"></a>Collection_directory/dump_info/DeviceDump/*/*.log</p>
        <p id="p66589505817"><a name="p66589505817"></a><a name="p66589505817"></a>Collection_directory/dump_info/LogDump/*/*.log</p>
        <p id="p16581511586"><a name="p16581511586"></a><a name="p16581511586"></a>Collection_directory/dump_info/AppDump/frudata/fruinfo.txt</p>
        <p id="p14658205185813"><a name="p14658205185813"></a><a name="p14658205185813"></a>Collection_directory/dump_info/AppDump/chassis/mdb_info.log</p>
        </td>
        </tr>
        <tr id="row793719229533"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p19937922125320"><a name="p19937922125320"></a><a name="p19937922125320"></a>LCNE logs</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p6937122225315"><a name="p6937122225315"></a><a name="p6937122225315"></a>All LCNE .log files</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p1193714226532"><a name="p1193714226532"></a><a name="p1193714226532"></a>Logs collected by LCNE</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p4306192985912"><a name="p4306192985912"></a><a name="p4306192985912"></a>Collection_directory/*/diagnostic_information/slot_1/tempdir/devm_bddrvadp.log</p>
        <p id="p3306132910599"><a name="p3306132910599"></a><a name="p3306132910599"></a>Collection_directory/*/diag_display_info.txt</p>
        <p id="p2072813615546"><a name="p2072813615546"></a><a name="p2072813615546"></a>Collection_directory/*/log.log</p>
        <p id="p1972918610542"><a name="p1972918610542"></a><a name="p1972918610542"></a>Collection_directory/*/log_1_*.log</p>
        </td>
        </tr>
        <tr id="row1393710222537"><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p7937112275315"><a name="p7937112275315"></a><a name="p7937112275315"></a>MindIE Pod console log</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p id="p59372228536"><a name="p59372228536"></a><a name="p59372228536"></a><em id="i193712225530"><a name="i193712225530"></a><a name="i193712225530"></a>{podname}</em>.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p id="p10937192213539"><a name="p10937192213539"></a><a name="p10937192213539"></a>MindIE Pod console log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p id="p119371222125315"><a name="p119371222125315"></a><a name="p119371222125315"></a>Collection_directory/mindie/log/mindie_cluster_log/<em id="i793712229536"><a name="i793712229536"></a><a name="i793712229536"></a>{podname}</em>.log</p>
        </td>
        </tr>
        <tr><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p>MindIO log</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p>ttp_log.log.*</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p>MindIO log</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p>Collection_directory/dl_log/ttp_log/ttp_log.log.*</p>
        </td>
        </tr>
        <td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p>Bus log</p>
        </td>
        <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p>log.log</p>
        </td>
        <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p>LCNE log (Ascend 950)</p>
        </td>
        <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p>Collection_directory/lcne/*/log.log</p>
        </td>
        </tbody>
        </table>

- You can also enter the corresponding log directory in the cleaning command for cleaning. The following describes the log storage structure of each parameter. For details about the parameters of the cleaning command, see Table 1.

    ```text
    |-- ${--process_log paths}
            |-- debug/plog/plog-{pid}_{time}.log
            |-- run/plog/plog-{pid}_{time}.log
            |-- debug/device-*/device-{pid}_{time}.log
            |-- run/device-*/device-{pid}_{time}.log

    |-- ${--device_log paths}
            |-- slog/dev-os-*/debug/device-os/device-os_*.log
            |-- slog/dev-os-*/run/device-os/device-os_*.log
            |-- slog/dev-os-*/run/event/event_*.log # This path is displayed only in Ascend HDK 23.0.3 and later versions.
            |-- slog/dev-os-*/device-*/device-*_*.log # In Ascend HDK 23.0.RC3, the device-*_*.log file is stored in this path.
            |-- slog/dev-os-*/debug/device-*/device-*_*.log # In Ascend HDK 23.0.3 and later versions, the device-*_*.log file is stored in this path.
            |-- hisi_logs/device-*/history.log
            |-- hisi_logs/device-*/{time}/log/kernel.log
            |-- hisi_logs/device-*/{time}/bbox/os/os_info.txt
            |-- hisi_logs/device-*/{time}/mntn/hbm.txt
            ....

    |-- ${--env_check paths}
           |-- npu_info_before.txt
           |-- npu_info_after.txt
           |-- npu_smi_0_details.csv
            ...
           |-- npu_smi_0_details.csv
           |-- npu_0_details.csv
           ...
           |-- npu_7_details.csv

    |-- ${--train_log paths}
           |-- rank-0.txt
           ...
           |-- rank-7.txt

    |-- ${--host_log paths}
           |-- messages
           |-- crash
                  |--Host + Fault timestamp directory (eg:127.xx.xx.1-2024-09-23-11:25:29)
                         |-- vmcore_dmesg.txt
           |-- dmesg
           |-- sysmonitor.log

    |-- ${--dl_log paths}
           |-- devicePlugin/devicePlugin*.log
           |-- noded/noded*.log
           |-- ascend-docker-runtime/runtime-run*.log
           |-- ascend-docker-runtime/hook-run*.log
           |-- volcano-scheduler/volcano-scheduler*.log
           |-- volcano-controller/volcano-controller*.log

           |-- npu-exporter/npu-exporter*.log
           |-- ttp_log/ttp_log.log.*

    |-- ${--mindie_log paths}
           |-- log/debug/mindie-{module}_{pid}_{datetime}.log
           |-- log/mindie_cluster_log/{podname}.log

    |-- ${--amct_log path}
           |-- amct_{framework}.log
    |-- ${--bus_log path}
           |-- log.log
    ```

    <a name="table192794861215"></a>
    <table><thead align="left"><tr id="row1527204819125"><th class="cellrowborder" valign="top" width="15.64%" id="mcps1.1.5.1.1"><p id="p1027134812121"><a name="p1027134812121"></a><a name="p1027134812121"></a>File Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="21.790000000000003%" id="mcps1.1.5.1.2"><p id="p152784811211"><a name="p152784811211"></a><a name="p152784811211"></a><strong id="b2027104891218"><a name="b2027104891218"></a><a name="b2027104891218"></a>Log File</strong></p>
    </th>
    <th class="cellrowborder" valign="top" width="20.34%" id="mcps1.1.5.1.3"><p id="p12744815122"><a name="p12744815122"></a><a name="p12744815122"></a><strong id="b132710486129"><a name="b132710486129"></a><a name="b132710486129"></a>File Description</strong></p>
    </th>
    <th class="cellrowborder" valign="top" width="42.230000000000004%" id="mcps1.1.5.1.4"><p id="p172724817129"><a name="p172724817129"></a><a name="p172724817129"></a><strong id="b42719482129"><a name="b42719482129"></a><a name="b42719482129"></a>Storage Directory</strong></p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row192814810125"><td class="cellrowborder" rowspan="2" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p18282048131210"><a name="p18282048131210"></a><a name="p18282048131210"></a>CANN App logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p172824815128"><a name="p172824815128"></a><a name="p172824815128"></a>plog-<em id="i858114825510"><a name="i858114825510"></a><a name="i858114825510"></a>{pid}</em>_<em id="i1734617413395"><a name="i1734617413395"></a><a name="i1734617413395"></a>{time}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p12287484122"><a name="p12287484122"></a><a name="p12287484122"></a>Host-side logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><a name="ul18745112814134"></a><a name="ul18745112814134"></a><ul id="ul18745112814134"><li>${--process_log}/debug/plog/plog-{pid}_{time}.log</li><li>${--process_log}/run/plog/plog-{pid}_{time}.log</li></ul>
    </td>
    </tr>
    <tr id="row1928164861216"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p2281348121210"><a name="p2281348121210"></a><a name="p2281348121210"></a>device-<em id="i12211557397"><a name="i12211557397"></a><a name="i12211557397"></a>{pid}</em>_<em id="i7720949402"><a name="i7720949402"></a><a name="i7720949402"></a>{time}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p9281448201219"><a name="p9281448201219"></a><a name="p9281448201219"></a>Device-side logs</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><a name="ul1134249141411"></a><a name="ul1134249141411"></a><ul id="ul1134249141411"><li>${--process_log}/debug/device-{id}/device-{pid}_{time}.log</li><li>${--process_log}/run/device-{id}/device-{pid}_{time}.log</li></ul>
    </td>
    </tr>
    <tr id="row2028104811215"><td class="cellrowborder" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p202874831219"><a name="p202874831219"></a><a name="p202874831219"></a>User training and inference logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p127491449202210"><a name="p127491449202210"></a><a name="p127491449202210"></a>rank<em id="i9749154972211"><a name="i9749154972211"></a><a name="i9749154972211"></a>-{id}</em>.txt</p>
    <p id="p137490492226"><a name="p137490492226"></a><a name="p137490492226"></a>rank<em id="i18749104992218"><a name="i18749104992218"></a><a name="i18749104992218"></a>-{id}</em>.log</p>
    <p id="p10749849152214"><a name="p10749849152214"></a><a name="p10749849152214"></a>worker<em id="i19749749142219"><a name="i19749749142219"></a><a name="i19749749142219"></a>-{id}</em>.txt</p>
    <p id="p1774984972212"><a name="p1774984972212"></a><a name="p1774984972212"></a>worker<em id="i274904942219"><a name="i274904942219"></a><a name="i274904942219"></a>-{id}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p128748101219"><a name="p128748101219"></a><a name="p128748101219"></a>Training and inference console logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><a name="ul11356426182714"></a><a name="ul11356426182714"></a><ul id="ul11356426182714"><li>${--train_log}/rank-<em id="i10284481126"><a name="i10284481126"></a><a name="i10284481126"></a>id</em>.*?.txt</li><li>${--train_log}/rank-<em id="i4109203271013"><a name="i4109203271013"></a><a name="i4109203271013"></a>id</em>.*?.log</li><li>${--train_log}/worker-<em id="i3736195518112"><a name="i3736195518112"></a><a name="i3736195518112"></a>id</em>.*?.log</li><li>${--train_log}/worker-<em id="i51091325104"><a name="i51091325104"></a><a name="i51091325104"></a>id</em>.*?.txt</li></ul>
    </td>
    </tr>
    <tr id="row928104818125"><td class="cellrowborder" rowspan="4" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p18281448111210"><a name="p18281448111210"></a><a name="p18281448111210"></a>NPU network port resource information</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p82834811121"><a name="p82834811121"></a><a name="p82834811121"></a>npu_info_before.txt</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p1028548121213"><a name="p1028548121213"></a><a name="p1028548121213"></a>NPU network port check before training</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p18882956161517"><a name="p18882956161517"></a><a name="p18882956161517"></a>${--env_check}/npu_info_before.txt</p>
    </td>
    </tr>
    <tr id="row1428248191217"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p328144851218"><a name="p328144851218"></a><a name="p328144851218"></a>npu_info_after.txt</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p192810485120"><a name="p192810485120"></a><a name="p192810485120"></a>NPU network port check after training</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p1317894520179"><a name="p1317894520179"></a><a name="p1317894520179"></a>${--env_check}/npu_info_after.txt</p>
    </td>
    </tr>
    <tr id="row528114871210"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p7291648141218"><a name="p7291648141218"></a><a name="p7291648141218"></a>npu_smi_<em id="i19656201019408"><a name="i19656201019408"></a><a name="i19656201019408"></a>{npu_id}</em>_details.csv</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p122919480125"><a name="p122919480125"></a><a name="p122919480125"></a>NPU status monitoring metric file</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p10613346131713"><a name="p10613346131713"></a><a name="p10613346131713"></a>${--env_check}/npu_smi_{npu_id}_details.csv</p>
    </td>
    </tr>
    <tr id="row1429204820125"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p192944814121"><a name="p192944814121"></a><a name="p192944814121"></a>npu_<em id="i32967161405"><a name="i32967161405"></a><a name="i32967161405"></a>{npu_id}</em>_details.csv</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p172913481124"><a name="p172913481124"></a><a name="p172913481124"></a>NPU network port monitoring metric file</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p158241447111718"><a name="p158241447111718"></a><a name="p158241447111718"></a>${--env_check}/npu_{npu_id}_details.csv</p>
    </td>
    </tr>
    <tr id="row11291648171212"><td class="cellrowborder" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p1629948101218"><a name="p1629948101218"></a><a name="p1629948101218"></a>Host resource information</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p62974819125"><a name="p62974819125"></a><a name="p62974819125"></a>host_metrics_<em id="i1432162014018"><a name="i1432162014018"></a><a name="i1432162014018"></a>{core_num}</em>.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p14291448141213"><a name="p14291448141213"></a><a name="p14291448141213"></a>Host resource monitoring metric file</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p1132018552171"><a name="p1132018552171"></a><a name="p1132018552171"></a>${--env_check}/host_metrics_{core_num}.json</p>
    </td>
    </tr>
    <tr id="row2291548141214"><td class="cellrowborder" rowspan="4" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p191119549613"><a name="p191119549613"></a><a name="p191119549613"></a>Host-side logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p52954818124"><a name="p52954818124"></a><a name="p52954818124"></a>messages-*?</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p7297483129"><a name="p7297483129"></a><a name="p7297483129"></a>Host OS log file</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p71491818186"><a name="p71491818186"></a><a name="p71491818186"></a>${--host_log}/messages-*?</p>
    </td>
    </tr>
    <tr id="row104971028144016"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p68411848154012"><a name="p68411848154012"></a><a name="p68411848154012"></a>dmesg</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p68411648134014"><a name="p68411648134014"></a><a name="p68411648134014"></a>Host kernal message file</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p1841174854011"><a name="p1841174854011"></a><a name="p1841174854011"></a>${--host_log}/dmesg</p>
    </td>
    </tr>
    <tr id="row981943679"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p1482204315716"><a name="p1482204315716"></a><a name="p1482204315716"></a>vmcore-dmesg.txt</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p58274315720"><a name="p58274315720"></a><a name="p58274315720"></a>Host kernal message file saved during a system crash</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p17824439712"><a name="p17824439712"></a><a name="p17824439712"></a>${--host_log}/crash/Host + fault timestamp (eg: 127.xx.xx.1-2024-09-23-11:25:29)/vmcore_dmesg.txt</p>
    </td>
    </tr>
    <tr id="row1258783414010"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p8842648154011"><a name="p8842648154011"></a><a name="p8842648154011"></a>sysmonitor.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p8842184864018"><a name="p8842184864018"></a><a name="p8842184864018"></a>Host system monitoring file</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p3842048124010"><a name="p3842048124010"></a><a name="p3842048124010"></a>${--host_log}/sysmonitor.log</p>
    </td>
    </tr>
    <tr id="row12294488123"><td class="cellrowborder" rowspan="7" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p629144891211"><a name="p629144891211"></a><a name="p629144891211"></a>Device-side logs</p>
    <p id="p1248516164264"><a name="p1248516164264"></a><a name="p1248516164264"></a></p>
    <p id="p1248591612267"><a name="p1248591612267"></a><a name="p1248591612267"></a></p>
    <p id="p14485141682613"><a name="p14485141682613"></a><a name="p14485141682613"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p102904816128"><a name="p102904816128"></a><a name="p102904816128"></a>device-os_<em id="i19815122414403"><a name="i19815122414403"></a><a name="i19815122414403"></a>{time}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p329144819125"><a name="p329144819125"></a><a name="p329144819125"></a>Device-side Ctrl CPU system log</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p10878185661512"><a name="p10878185661512"></a><a name="p10878185661512"></a>${--device_log}/slog/dev-os-{id}/debug/device-os/device-os_{time}.log</p>
    </td>
    </tr>
    <tr id="row199869351304"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p16161374016"><a name="p16161374016"></a><a name="p16161374016"></a>event_<em id="i4381833174019"><a name="i4381833174019"></a><a name="i4381833174019"></a>{time}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p126161837100"><a name="p126161837100"></a><a name="p126161837100"></a>EVENT-level Ctrl CPU system log on the device</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p161612371017"><a name="p161612371017"></a><a name="p161612371017"></a>Ascend HDK 23.0.3 and later:</p>
    <p id="p10616113711018"><a name="p10616113711018"></a><a name="p10616113711018"></a>${--device_log}/slog/dev-os-{id}/run/event/event_{time}.log</p>
    </td>
    </tr>
    <tr id="row18291948191216"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p1429124811129"><a name="p1429124811129"></a><a name="p1429124811129"></a>device-id_<em id="i17917183954011"><a name="i17917183954011"></a><a name="i17917183954011"></a>{time}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p329748171215"><a name="p329748171215"></a><a name="p329748171215"></a>Device-side non-Ctrl CPU system log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p17532535016"><a name="p17532535016"></a><a name="p17532535016"></a>Ascend HDK 23.0.RC3: </p>
    <p id="p158782056101513"><a name="p158782056101513"></a><a name="p158782056101513"></a>${--device_log}/slog/dev-os-{id}/device-{id}/device-{id}_{time}.log</p>
    <p id="p628611308210"><a name="p628611308210"></a><a name="p628611308210"></a>Ascend HDK 23.0.3 and later:</p>
    <p id="p02861330921"><a name="p02861330921"></a><a name="p02861330921"></a>${--device_log}/slog/dev-os-{id}/debug/device-{id}/device-{id}_{time}.log</p>
    </td>
    </tr>
    <tr id="row12301848161217"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p83004810128"><a name="p83004810128"></a><a name="p83004810128"></a>history.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p19301748141219"><a name="p19301748141219"></a><a name="p19301748141219"></a>Black Box log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p13877145661520"><a name="p13877145661520"></a><a name="p13877145661520"></a>${--device_log}/hisi_logs/device-{id}/history.log</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>kernel.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>NPU kernel log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>${--device_log}/hisi_logs/device-{id}/{time}/log/kernel.log</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>os_info.txt</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>Device OS information</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>${--device_log}/hisi_logs/device-{id}/{time}/bbox/os/os_info.txt</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>hbm.txt</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>Device-side on-chip memory log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p>${--device_log}/hisi_logs/device-{id}/{time}/mntn/hbm.txt</p>
    </td>
    </tr>
    <tr id="row1096711207261"><td class="cellrowborder" rowspan="7" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p8338102742611"><a name="p8338102742611"></a><a name="p8338102742611"></a><span id="ph686313599221"><a name="ph686313599221"></a><a name="ph686313599221"></a>MindCluster</span> componengt logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p14338192792619"><a name="p14338192792619"></a><a name="p14338192792619"></a>devicePlugin*.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p17338152752619"><a name="p17338152752619"></a><a name="p17338152752619"></a>SuperPoD device log<span id="ph1297011103531"><a name="ph1297011103531"></a><a name="ph1297011103531"></a> and Ascend Device Plugin</span> log</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p0338112792614"><a name="p0338112792614"></a><a name="p0338112792614"></a>${--dl_log}/devicePlugin/devicePlugin*.log</p>
    </td>
    </tr>
    <tr id="row588102412264"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p1338162713267"><a name="p1338162713267"></a><a name="p1338162713267"></a>noded*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p7338162720269"><a name="p7338162720269"></a><a name="p7338162720269"></a>AI server log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p16338162713268"><a name="p16338162713268"></a><a name="p16338162713268"></a>${--dl_log}/noded/noded*.log</p>
    </td>
    </tr>
    <tr id="row13715151654817"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p196021757154811"><a name="p196021757154811"></a><a name="p196021757154811"></a>runtime-run*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p16718144795013"><a name="p16718144795013"></a><a name="p16718144795013"></a><span id="ph107402047545"><a name="ph107402047545"></a><a name="ph107402047545"></a></span>Log generated when ascend-docker-runtime is executed</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p07449236533"><a name="p07449236533"></a><a name="p07449236533"></a>${--dl_log}/ascend-docker-runtime/runtime-run*.log</p>
    </td>
    </tr>
    <tr id="row8933819164810"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p1360345715489"><a name="p1360345715489"></a><a name="p1360345715489"></a>hook-run*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p771894714509"><a name="p771894714509"></a><a name="p771894714509"></a><span id="ph0161474547"><a name="ph0161474547"></a><a name="ph0161474547"></a></span>Log generarted when ascend-docker-hook is executed</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p76041557104812"><a name="p76041557104812"></a><a name="p76041557104812"></a>${--dl_log}/ascend-docker-runtime/</p>
    <p id="p12604165718481"><a name="p12604165718481"></a><a name="p12604165718481"></a>hook-run*.log</p>
    </td>
    </tr>
    <tr id="row824762413484"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p11604165718480"><a name="p11604165718480"></a><a name="p11604165718480"></a>volcano-scheduler*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p76052579481"><a name="p76052579481"></a><a name="p76052579481"></a><span id="ph94341624105410"><a name="ph94341624105410"></a><a name="ph94341624105410"></a></span>volcano-scheduler log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p160520575488"><a name="p160520575488"></a><a name="p160520575488"></a>${--dl_log}/volcano-scheduler/</p>
    <p id="p7605657184819"><a name="p7605657184819"></a><a name="p7605657184819"></a>volcano-scheduler*.log</p>
    </td>
    </tr>
    <tr id="row13356162964820"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p46051157134819"><a name="p46051157134819"></a><a name="p46051157134819"></a>volcano-controller*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p14606125716487"><a name="p14606125716487"></a><a name="p14606125716487"></a><span id="ph1613023013547"><a name="ph1613023013547"></a><a name="ph1613023013547"></a></span>volcano-controller log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p2606057114813"><a name="p2606057114813"></a><a name="p2606057114813"></a>${--dl_log}/volcano-controller/</p>
    <p id="p1660685710486"><a name="p1660685710486"></a><a name="p1660685710486"></a>volcano-controller*.log</p>
    </td>
    </tr>
    <tr id="row1415202516523"><td class="cellrowborder" valign="top" headers="mcps1.1.5.1.1 "><p id="p19169252524"><a name="p19169252524"></a><a name="p19169252524"></a>npu-exporter*.log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.2 "><p id="p21617254529"><a name="p21617254529"></a><a name="p21617254529"></a><span id="ph1457105118547"><a name="ph1457105118547"></a><a name="ph1457105118547"></a>NPU Exporter</span> log</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.1.5.1.3 "><p id="p9608657124815"><a name="p9608657124815"></a><a name="p9608657124815"></a>${--dl_log}/npu-exporter/</p>
    <p id="p20608657144818"><a name="p20608657144818"></a><a name="p20608657144818"></a>npu-exporter*.log</p>
    </td>
    </tr>
    <tr id="row19745101722112"><td class="cellrowborder" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p1178823162110"><a name="p1178823162110"></a><a name="p1178823162110"></a>MindIE componeng logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p41781123132114"><a name="p41781123132114"></a><a name="p41781123132114"></a>mindie-<em id="i917915234212"><a name="i917915234212"></a><a name="i917915234212"></a>{module}</em>_<em id="i6179152313219"><a name="i6179152313219"></a><a name="i6179152313219"></a>{pid}</em>_<em id="i1417932362118"><a name="i1417932362118"></a><a name="i1417932362118"></a>{datetime}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p11179023172116"><a name="p11179023172116"></a><a name="p11179023172116"></a><span id="ph202226189104"><a name="ph202226189104"></a><a name="ph202226189104"></a>MindIE Server</span>, <span id="ph122221618111018"><a name="ph122221618111018"></a><a name="ph122221618111018"></a>MindIE LLM</span>, <span id="ph422211184106"><a name="ph422211184106"></a><a name="ph422211184106"></a>MindIE SD</span>, <span id="ph622211183102"><a name="ph622211183102"></a><a name="ph622211183102"></a>MindIE RT</span>, <span id="ph1422910436311"><a name="ph1422910436311"></a><a name="ph1422910436311"></a>MindIE Torch</span>, <span id="ph7973205355818"><a name="ph7973205355818"></a><a name="ph7973205355818"></a>MindIE MS</span>, <span id="ph684311814254"><a name="ph684311814254"></a><a name="ph684311814254"></a>MindIE Benchmark</span>, and <span id="ph377018316257"><a name="ph377018316257"></a><a name="ph377018316257"></a>MindIE Client</span> logs</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p129912577214"><a name="p129912577214"></a><a name="p129912577214"></a>${--mindie_log}/log/debug/mindie-{module}_{pid}_{datetime}.log</p>
    </td>
    </tr>
    <tr id="row1399841203612"><td class="cellrowborder" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p1899861213362"><a name="p1899861213362"></a><a name="p1899861213362"></a>MindIE Pod console log</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p13998312203616"><a name="p13998312203616"></a><a name="p13998312203616"></a><em id="i1333214713616"><a name="i1333214713616"></a><a name="i1333214713616"></a>{podname}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p1199851233611"><a name="p1199851233611"></a><a name="p1199851233611"></a>MindIE Pod console log</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p28071349378"><a name="p28071349378"></a><a name="p28071349378"></a>${--mindie_log}/log/mindie_cluster_log/<em id="i5570173643717"><a name="i5570173643717"></a><a name="i5570173643717"></a>{podname}</em>.log</p>
    </td>
    </tr>
    <tr id="row594320215217"><td class="cellrowborder" valign="top" width="15.64%" headers="mcps1.1.5.1.1 "><p id="p1724461918375"><a name="p1724461918375"></a><a name="p1724461918375"></a>AMCT log</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.790000000000003%" headers="mcps1.1.5.1.2 "><p id="p518002372113"><a name="p518002372113"></a><a name="p518002372113"></a>amct_<em id="i1518022382117"><a name="i1518022382117"></a><a name="i1518022382117"></a>{framework}</em>.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.34%" headers="mcps1.1.5.1.3 "><p id="p11180202362118"><a name="p11180202362118"></a><a name="p11180202362118"></a>AMCT log</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.230000000000004%" headers="mcps1.1.5.1.4 "><p id="p101808238218"><a name="p101808238218"></a><a name="p101808238218"></a>${--amct_log}/amct_{framework}.log</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p>MindIO log</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p>ttp_log.log.*</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p>MindIO log</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p>${--dl_log}/ttp_log/ttp_log.log.*</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="16.150000000000002%" headers="mcps1.2.5.1.1 "><p>Bus log</p>
    </td>
    <td class="cellrowborder" valign="top" width="21.26%" headers="mcps1.2.5.1.2 "><p>log.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.39%" headers="mcps1.2.5.1.3 "><p>LCNE log (Ascend 950)</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.5.1.4 "><p>${--bus_log}/lcne/*/log.log</p>
    </td>
    </tr>
    </tbody>
    </table>

## Collecting Logs Before Training and Inference

### NPU Environment Check File Before Training and Inference

**File Description <a name="section5664143619418"></a>**

- Before training and inference jobs are started, use `hccn_tool` or a script to query and record the IP address, mask, received and sent packets, and historical link statistics of each NPU network port. Before training is started, use `npu-smi` or a script to query the chip health information.
- Naming rule: `npu_info_before.txt`
- Constraints on the storage path:
    - `Collection_directory/environment_check/`
    - `${--env_check parameter-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

**Collection Methods <a name="section16941192518572"></a>**

MindCluster Ascend FaultDiag can collect logs before training or inference in either of the following ways:

- Script: Use the `npu_info_collect.sh` script to collect the NPU environment check file before training and inference. For details, see the [Log Collection Script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector).
- [Command](#section1020314437418): Before training and inference, use `hccn_tool` to query the NPU environment check file, and save the query command and result to the `npu_info_before.txt` file.

**Collection via Commands <a name="section1020314437418"></a>**

The involved commands and examples are as follows:

- Query the network health status.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -net_health -g
    ```

    Command output:

    ```ColdFusion
    net health status: Init
    ```

- Query the RoCE physical link connection status.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -link -g
    ```

    Command output:

    ```ColdFusion
    link status: UP
    ```

- Query information about the RoCE network optical module.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -optical -g
    ```

    Command output:

    ```ColdFusion
    optical info:
    present              : not present
    ...
    Tx Power             : 4.4035 mW
    Rx Power             : 1.0189 mW
    Vcc High Thres       : 3465.00 mV
    Vcc Low Thres        : 3135.00 mV
    Temp High Thres      : 70 C
    Temp Low Thres       : 0 C
    TxPower High Thres   : 3.5481 mW
    TxPower Low Thres    : 0.2818 mW
    RxPower High Thres   : 3.5481 mW
    RxPower Low Thres    : 0.1445 mW
    Tx Bias              : 7.9360 mA
    Tx Los Flag          : 0x0
    Rx Los Flag          : 0xff
    Tx LoL Flag          : 0x0
    Rx LoL Flag          : 0xff
    ...
    ```

- Query the TLS configuration.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -tls -g | grep switch
    ```

    Command output:

    ```ColdFusion
    dev_id:0, tls switch[0](0:disable, 1:enable), tls preconfigured[1](0:non-preset, 1:preset), tls alarm time threshold[60]days
    ```

- Query the FEC mode.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -fec -g
    ```

    Command output:

    ```ColdFusion
    fec mode: rs FEC mode
    ```

- Query the IP address and mask.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -ip -g
    ```

    Command output:

    ```ColdFusion
    ipaddr:10.xx.xx.10
    netmask:255.255.255.0
    ```

- Query statistics about sent and received packets.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -stat -g
    ```

    Command output:

    ```ColdFusion
    packet statistics:
    mac_tx_mac_pause_num:0
    mac_rx_mac_pause_num:0
    mac_tx_pfc_pkt_num:0
    ...
    roce_qp_status_err_num:0
    nic_tx_all_pkg_num:122404
    nic_tx_all_oct_num:16921741
    nic_rx_all_pkg_num:6414803
    nic_rx_all_oct_num:482237805
    ```

- Query the historical link statistics of the network port.

    ```shell
    /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -link_stat -g
    ```

    Command output:

    ```ColdFusion
    [device 0]current time        : Wed Jun  7 10:08:28 2023
    [device 0]link up count       : 2
    [device 0]link change records :
    [device 0]    Tue Jun  6 16:32:12 2023    LINK UP
    [device 0]    Tue Jun  6 16:32:10 2023    LINK DOWN
    [device 0]    Tue Jun  6 16:31:55 2023    LINK UP
    ```

    The following is an example of information about device 0. You need to collect information about all devices.

    ```ColdFusion
    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -net_health -g
    net health status: Init

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -link -g
    link status: UP

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -optical -g
    optical info:
    present              : not present
    ...
    Tx Power             : 4.4035 mW
    Rx Power             : 1.0189 mW
    Vcc High Thres       : 3465.00 mV
    Vcc Low Thres        : 3135.00 mV
    Temp High Thres      : 70 C
    Temp Low Thres       : 0 C
    TxPower High Thres   : 3.5481 mW
    TxPower Low Thres    : 0.2818 mW
    RxPower High Thres   : 3.5481 mW
    RxPower Low Thres    : 0.1445 mW
    Tx Bias              : 7.9360 mA
    Tx Los Flag          : 0x0
    Rx Los Flag          : 0xff
    Tx LoL Flag          : 0x0
    Rx LoL Flag          : 0xff
    ...

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -tls -g | grep switch
    dev_id:0, tls switch[0](0:disable, 1:enable), tls preconfigured[1](0:non-preset, 1:preset), tls alarm time threshold[60]days

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -fec -g
    fec mode: rs FEC mode

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -ip -g
    ipaddr:10.xx.xx.10
    netmask:255.255.255.0

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -stat -g
    packet statistics:
    mac_tx_mac_pause_num:0
    mac_rx_mac_pause_num:0
    mac_tx_pfc_pkt_num:0
    ...
    roce_qp_status_err_num:0
    nic_tx_all_pkg_num:122404
    nic_tx_all_oct_num:16921741
    nic_rx_all_pkg_num:6414803
    nic_rx_all_oct_num:482237805

    /usr/local/Ascend/driver/tools/hccn_tool -i 0 -link_stat -g
    [device 0]current time        : Wed Jun  7 10:08:28 2023
    [device 0]link up count       : 2
    [device 0]link change records :
    [device 0]    Tue Jun  6 16:32:12 2023    LINK UP
    [device 0]    Tue Jun  6 16:32:10 2023    LINK DOWN
    [device 0]    Tue Jun  6 16:31:55 2023    LINK UP
    ```

    >[!NOTE]
    >The result of each collection command must be separated by one line. For example:
    >
    >```shell
    >/usr/local/Ascend/driver/tools/hccn_tool -i 0 -ip -g
    >XXXX
    >/usr/local/Ascend/driver/tools/hccn_tool -i 0 -stat -g
    >```

- Before training and inference, use `npu-smi` to query the chip health information and save the query command and result to the `npu_info_before.txt` file. The involved commands and examples are as follows:

    - Query the basic information about a device.

        ```shell
        /usr/local/bin/npu-smi info
        ```

        Command output:

        ```ColdFusion
        +------------------------------------------------------------------------------------------------+
        | npu-smi 24.1.rc1                 Version: 24.1.rc1                                             |
        +---------------------------+---------------+----------------------------------------------------+
        | NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
        | Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | 7     xxx                | OK            | 67.0        44                0    / 0             |
        | 0                         | 0000:3D:00.0  | 0           2505 / 15567      0    / 32768         |
        +===========================+===============+====================================================+
        +---------------------------+---------------+----------------------------------------------------+
        | NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
        +===========================+===============+====================================================+
        | No running processes found in NPU 0                                                            |
        +===========================+===============+====================================================+
        ...
        | No running processes found in NPU 7                                                            |
        +===========================+===============+====================================================+
        ```

    - Query ECC of the high-bandwidth memory.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t ecc
        ```

        Command output:

        ```ColdFusion
        NPU ID                                   : 1
        Chip Count                               : 1

        DDR Single Bit Error Count               : 0
        DDR Double Bit Error Count               : 0
        DDR Single Bit Aggregate Total Err Cnt   : 0
        DDR Double Bit Aggregate Total Err Cnt   : 0
        DDR Single Bit Isolated Pages Count      : 0
        DDR Double Bit Isolated Pages Count      : 0
        HBM Single Bit Error Count               : 0
        HBM Double Bit Error Count               : 0
        HBM Single Bit Aggregate Total Err Cnt   : 0
        HBM Double Bit Aggregate Total Err Cnt   : 0
        HBM Single Bit Isolated Pages Count      : 0
        HBM Double Bit Isolated Pages Count      : 0
        Chip ID                                  : 0
        ```

    - Query the basic information about the hardware.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t board
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Software Version               : 23.0.5
        Firmware Version               : 7.1.0.7.220
        Compatibility                  : OK
        Board ID                       : 0x02
        PCB ID                         : A
        BOM ID                         : 1
        PCIe Bus Info                  : 0000:61:00.0
        Slot ID                        : 0
        Class ID                       : NA
        PCI Vendor ID                  : 0x19e5
        PCI Device ID                  : 0xD801
        Subsystem Vendor ID            : 0x0200
        Subsystem Device ID            : 0x0100
        Chip Count                     : 1
        ```

    - Query the basic hardware information and the name of the specified device.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -c 0 -t board
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Chip ID                        : 0
        Chip Type                      : Ascend
        Chip Name                      : xxx
        Chip Version                   : V1
        Board ID                       : 0x02
        PCB ID                         : NA
        BOM ID                         : 1
        VDie ID                        : 42C711D4 20B03704 4A10C8D4 14CC040A D2102003
        NDie ID                        : 27216594 20401010 4E10C8D4 14CC040A A4102003
        Chip Position ID               : 0
        PCIe Bus Info                  : 0000:61:00.0
        Firmware Version               : 7.1.0.7.220
        ```

    - Query the memory usage.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t usages
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Chip Count                     : 1

        DDR Capacity(MB)               : 13553
        DDR Usage Rate(%)              : 6
        DDR Hugepages Total(page)      : 0
        DDR Hugepages Usage Rate(%)    : 0
        HBM Capacity(MB)               : 32768
        HBM Usage Rate(%)              : 0
        Aicore Usage Rate(%)           : 0
        Aicpu Usage Rate(%)            : 0
        Ctrlcpu Usage Rate(%)          : 0
        DDR Bandwidth Usage Rate(%)    : 0
        HBM Bandwidth Usage Rate(%)    : 0
        Chip ID                        : 0
        ```

    - Query the chip health information.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -c 0 -t health
        ```

        Command output:

        ```ColdFusion
         Health Status                  : OK
         Error Code                     : NA
         Error Information              : NA
        ```

        The following is an example of obtained information. You need to collect information about all devices.

        ```ColdFusion
        /usr/local/bin/npu-smi info
        +------------------------------------------------------------------------------------------------+
        | npu-smi 23.0.5                   Version: 23.0.5                                               |
        +---------------------------+---------------+----------------------------------------------------+
        | NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
        | Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
        +===========================+===============+====================================================+
        | 0     xxx                 | OK            | 73.1        37                0    / 0             |
        | 0                         | 0000:61:00.0  | 0           920  / 13553      0    / 32768         |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | 7     xxx                 | OK            | 67.0        38                0    / 0             |
        | 0                         | 0000:3D:00.0  | 0           2346 / 15567      0    / 32768         |
        +===========================+===============+====================================================+
        +---------------------------+---------------+----------------------------------------------------+
        | NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
        +===========================+===============+====================================================+
        | No running processes found in NPU 0                                                            |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | No running processes found in NPU 7                                                            |
        +===========================+===============+====================================================+

        /usr/local/bin/npu-smi info -i 0 -c 0 -t health
        Health Status                  : OK
        Error Code                     : NA
        Error Information              : NA

        /usr/local/bin/npu-smi info -i 0 -t ecc
        NPU ID                                   : 0
        Chip Count                               : 1

        DDR Single Bit Error Count               : 0
        DDR Double Bit Error Count               : 0
        DDR Single Bit Aggregate Total Err Cnt   : 0
        DDR Double Bit Aggregate Total Err Cnt   : 0
        DDR Single Bit Isolated Pages Count      : 0
        DDR Double Bit Isolated Pages Count      : 0
        HBM Single Bit Error Count               : 0
        HBM Double Bit Error Count               : 0
        HBM Single Bit Aggregate Total Err Cnt   : 0
        HBM Double Bit Aggregate Total Err Cnt   : 0
        HBM Single Bit Isolated Pages Count      : 0
        HBM Double Bit Isolated Pages Count      : 0
        Chip ID                                  : 0

        /usr/local/bin/npu-smi info -i 0 -t board
        NPU ID                         : 0
        Software Version               : 23.0.5
        Firmware Version               : 7.1.0.7.220
        Compatibility                  : OK
        Board ID                       : 0x02
        PCB ID                         : A
        BOM ID                         : 1
        PCIe Bus Info                  : 0000:61:00.0
        Slot ID                        : 0
        Class ID                       : NA
        PCI Vendor ID                  : 0x19e5
        PCI Device ID                  : 0xD801
        Subsystem Vendor ID            : 0x0200
        Subsystem Device ID            : 0x0100
        Chip Count                     : 1

        /usr/local/bin/npu-smi info -i 0 -c 0 -t board
        NPU ID                         : 0
        Chip ID                        : 0
        Chip Type                      : Ascend
        Chip Name                      : xxx
        Chip Version                   : V1
        Board ID                       : 0x02
        PCB ID                         : NA
        BOM ID                         : 1
        VDie ID                        : 42C711D4 20B03704 4A10C8D4 14CC040A D2102003
        NDie ID                        : 27216594 20401010 4E10C8D4 14CC040A A4102003
        Chip Position ID               : 0
        PCIe Bus Info                  : 0000:61:00.0
        Firmware Version               : 7.1.0.7.220

        /usr/local/bin/npu-smi info -i 0 -t usages
        NPU ID                         : 0
        Chip Count                     : 1

        DDR Capacity(MB)               : 13553
        DDR Usage Rate(%)              : 6
        DDR Hugepages Total(page)      : 0
        DDR Hugepages Usage Rate(%)    : 0
        HBM Capacity(MB)               : 32768
        HBM Usage Rate(%)              : 0
        Aicore Usage Rate(%)           : 0
        Aicpu Usage Rate(%)            : 0
        Ctrlcpu Usage Rate(%)          : 0
        DDR Bandwidth Usage Rate(%)    : 0
        HBM Bandwidth Usage Rate(%)    : 0
        Chip ID                        : 0

        /usr/local/bin/npu-smi info -i 0 -c 0 -t health
         Health Status                  : OK
         Error Code                     : NA
         Error Information              : NA
        ...
        ```

    >[!NOTE]
    >The result of each collection command must be separated by one line. For example:
    >
    > ```shell
    > /usr/local/bin/npu-smi info -i 0 -c 0 -t health
    > XXXX
    > /usr/local/bin/npu-smi info -i 1 -c 0 -t health
    > ```

- Before training and inference, run other related commands to query the NPU environment check files, and save the query commands and results to the `npu_info_before.txt` file. The involved commands and examples are as follows:
    - Query the current system time.

        ```shell
        datetime=$(date "+%Y-%m-%d %H:%M:%S")
        echo "Datetime: $datetime">>${save_file}
        echo -e "\n">>${save_file}
        ```

        Command output:

        ```ColdFusion
        Datetime: 2024-06-26 01:13:36
        ```

    - Query the driver version.

        ```shell
        cat /usr/local/Ascend/driver/version.info
        ```

        Command output:

        ```ColdFusion
        Version=24.1.rc1
        ascendhal_version=7.35.19
        aicpu_version=1.0
        tdt_version=1.0
        log_version=1.0
        prof_version=2.0
        dvppkernels_version=1.1
        tsfw_version=1.0
        Innerversion=V100R001C15SPC006B220
        compatible_version=[V100R001C30],[V100R001C13],[V100R001C15],[V100R001C17]
        compatible_version_fw=[7.0.0,7.2.99]
        ```

    - Query the firmware version.

        ```shell
        cat /usr/local/Ascend/firmware/version.info
        ```

        Command output:

        ```ColdFusion
        Version=7.1.0.11.220
        firmware_version=1.0
        package_version=23.0.7
        compatible_version_drv=[23.0.rc3,23.0.rc3.],[23.0.0,23.0.0.]
        ```

    - Query the CANN version (AArch64).

        ```shell
        cat /usr/local/Ascend/cann/aarch64-linux/ascend_toolkit_install.info
        ```

        Command output:

        ```ColdFusion
        package_name=Ascend-cann-toolkit
        version=8.5.0
        innerversion=V100R001C25SPC001B212
        compatible_version=[V100R001C15],[V100R001C18],[V100R001C19],[V100R001C20],[V100R001C21],[V100R001C23]
        arch=aarch64
        os=linux
        path=/usr/local/Ascend/cann-8.5.0/aarch64-linux
        ```

    - Query the CANN version (x86_64).

        ```shell
        cat /usr/local/Ascend/cann/x86_64-linux/ascend_toolkit_install.info
        ```

        Command output:

        ```ColdFusion
        package_name=Ascend-cann-toolkit
        version=8.5.0
        innerversion=V100R001C25SPC001B212
        compatible_version=[V100R001C15],[V100R001C18],[V100R001C19],[V100R001C20],[V100R001C21],[V100R001C23]
        arch=x86_64
        os=linux
        path=/usr/local/Ascend/cann-8.5.0/x86_64-linux
        ```

    - Query the AI framework version.

        ```shell
        pip list | grep "torch"
        pip list | grep torch-npu
        pip list | grep "mindspore"
        ```

        Command output:

        ```ColdFusion
        torch              1.11.0
        torch-npu          2.1.0.post8.dev20241009
        mindspore          2.3.0
        ```

    - Query the firmware version details.

        ```shell
        /usr/local/Ascend/driver/tools/upgrade-tool --device_index -1 --component -1 --version
        ```

        Command output:

        ```ColdFusion
        {
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(0).
        {"device_id":0, "component":nve, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(3).
        {"device_id":0, "component":uefi, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(8).
        {"device_id":0, "component":imu, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(9).
        {"device_id":0, "component":imp, "version":7.1.0.7.220}
        ...
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(0).
        {"device_id":7, "component":nve, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(3).
        {"device_id":7, "component":uefi, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(8).
        {"device_id":7, "component":imu, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(9).
        {"device_id":7, "component":imp, "version":7.1.0.7.220}
        }
        ```

## Collecting Logs During Training and Inference

### NPU Network Port Monitoring Metric File <a name="ZH-CN_TOPIC_0000001579238638"></a>

**File Description <a name="section56641436194180001"></a>**

- Use `hccn_tool` or a script to collect statistics on the number of packets received and sent by the NPU network port.
- Naming rule: `npu_(\d+)_details.csv`, for example, `npu_0_details.csv`, where `0` indicates the NPU device ID.
- Constraints on the storage path:
    - `Collection directory/environment_check/`
    - `${--env_check parameter-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

>[!NOTE]
>You need to create a monitoring metric file of network port statistics for each NPU.

**Collection Methods<a name="section207215361658"></a>**

MindCluster Ascend FaultDiag can collect logs of training and inference jobs in the following ways:

- Script: Use the `net_data_collect.py` script to collect the NPU network port monitoring metric file. For details, see  [Log Collection Script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector).
- [Command](#section1020314437418_001): During training and inference jobs, use the `hccn_tool` to query the NPU network port statistics every 15 seconds.

**Collection via Commands <a name="section1020314437418_001"></a>**

Example:

```shell
/usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -stat -g
```

Record all metrics and their values and save them as a CSV file, as shown in [Table 1].

Command output:

```ColdFusion
packet statistics:
mac_tx_mac_pause_num:0
mac_rx_mac_pause_num:0
mac_tx_pfc_pkt_num:0
...
roce_qp_status_err_num:0
nic_tx_all_pkg_num:122404
nic_tx_all_oct_num:16921741
nic_rx_all_pkg_num:6414803
nic_rx_all_oct_num:482237805
```

Save parameter metrics in each command output to a CSV file.

**Table 1** Storage format

<a name="table205133240413"></a>

|timestamp|mac_tx_mac_pause_num|...|mac_rx_mac_pause_num|mac_tx_pfc_pkt_num|mac_tx_pfc_pri0_pkt_num|...|
|--|--|--|--|--|--|--|
|1684460336|0|...|0|0|0|...|
|1684460354|0|...|0|0|0|...|

### NPU Status Monitoring Metric File <a name="ZH-CN_TOPIC_0000001579717794"></a>

**File Description <a name="section56641436194180002"></a>**

- This file is collected using `npu-smi` or script to monitor the rated frequency, current power, and temperature of NPUs.
- Naming rule: `npu_smi_(\d+)_details.csv`, for example, `npu_smi_0_details.csv`, where `0` indicates the NPU device ID.
- Constraints on the storage path:
    - `Collection directory/environment_check/`
    - `${--env\_check parameter-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

>[!NOTE]
>You need to create a monitoring metric file of network port status for each NPU.

**Collection Methods<a name="section20721536165801"></a>**

MindCluster Ascend FaultDiag can collect NPU network port status monitoring files in either of the following ways:

- Script: Use the `npu_data_collect.py` script to collect the NPU status monitoring metric files. For details, see [Log Collection Script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector).
- [Command](#section0729525121013): During training and inference, use `npu-smi` to query the NPU status every 15 seconds.

**Collection via Commands<a name="section0729525121013"></a>**

Example:

```shell
/usr/local/bin/npu-smi info -t common -i ${device_id}
```

Record the values of `NPU ID`, `Aicore Usage Rate`, `Aicore Freq(MHZ)`, `Aicore curFreq(MHZ)`, `Temperature`, `NPU Real-time Power(W)`, and `HBM Usage Rate` of all NPUs in sequence, and save the values as a CSV file. The format is shown in [Table 1](#table9968833174718).

Command output:

```ColdFusion
        NPU ID                         : 0
        Chip Count                     : 1
        Chip ID                        : 0
        Memory Usage Rate(%)           : 6
        HBM Usage Rate(%)              : 0
        Aicore Usage Rate(%)           : 0
        Aicore Freq(MHZ)               : 900
        Aicore curFreq(MHZ)            : 900
        Aicore Count                   : 30
        Temperature(C)                 : 41
        NPU Real-time Power(W)         : 71.7
```

Save the parameter metrics in each command output to a CSV file.

**Table 1** Storage format

<a name="table9968833174718"></a>

|**time**|**dev_id**|hbm_rate|aicore_rate|**rated_freq**|**freq**|**temp**|power|
|--|--|--|--|--|--|--|--|
|1683862905|2|0|0|1000|1000|42|70.3|
|1683862925|2|0|0|1000|1000|42|70.5|

- `time`: current collection time of the UNIX system
- `dev_id`: NPU ID, which corresponds to the `NPU ID` in the command output.
- `hbm_rate`: on-chip memory usage, which corresponds to `HBM Usage Rate(%)` in the command output.
- `aicore_rate`: AI Core usage, which corresponds to `Aicore Usage Rate(%)` in the command output.
- `rated_freq`: rated frequency of the NPU, which corresponds to `Aicore Freq(MHZ)` in the command output.
- `freq`: real-time frequency of the NPU, which corresponds to `Aicore curFreq(MHZ)` in the command output.
- `temp`: NPU temperature, which corresponds to `Temperature(C)` in the command output.
- `power`: NPU power consumption, which corresponds to `NPU Real-time Power(W)` in the command output.

### Host Resource Information <a name="ZH-CN_TOPIC_0000001629758409"></a>

**File Description <a name="section56641436194180003"></a>**

- The `top` command or script is used to collect information about the total physical memory used by the host, the percentage of CPU usage (`%CPU`) and physical memory (`RES`) used by the main training and inference processes of each NPU. The data is stored in a JSON file in the format of `host_metrics_${core_num}.json`.
- Naming rule: `host_metrics_${core_num}.json`, for example, `host_metrics_64.json`, where `64` indicates the number of CPU cores.
- Constraints on the storage path:
    - `Collection directory/environment_check/`
    - `${--env_check parameter-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

**Collection Methods<a name="section20721536165802"></a>**

MindCluster Ascend FaultDiag can collect the host resource information in either of the following ways:

- Script: Use the `host_resource_collect.py` script to collect host resource information. For details, see [Log Collection Script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector).
- [Command](#section157555973015): Collect the host resource information by running commands.

**Collection via Commands<a name="section157555973015"></a>**

- Before training or inference, run the following command to query the total number of CPU cores of the training or inference device:

    ```shell
    cat /proc/cpuinfo | grep "processor" | wc -l
    ```

- During training and inference, run the `npu-smi info` command to query the process ID of each NPU and record the process IDs of all processes as `{pid_list}`.

    ```shell
    /usr/local/bin/npu-smi info
    ```

    Command output:

    ```ColdFusion
    +------------------------------------------------------------------------------------------------+
    | npu-smi 23.0.rc3          Version: 23.0.rc2.3                                      |
    +---------------------------+---------------+----------------------------------------------------+
    | NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
    | Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
    +===========================+===============+====================================================+
    | 0     xxx                | OK            | 73.4        44                1123 / 1123          |
    | 0                         | 0000:C1:00.0  | 0           4565 / 15137      30710/ 32768         |
    +===========================+===============+====================================================+
    | 1     xxx                | OK            | 69.6        39                1123 / 1123          |
    | 0                         | 0000:81:00.0  | 0           4483 / 15137      30710/ 32768         |
    +===========================+===============+====================================================+
    | 2     xxx                | OK            | 70.0        36                1123 / 1123          |
    | 0                         | 0000:41:00.0  | 0           4437 / 15137      30710/ 32768         |
    +===========================+===============+====================================================+
    | 3     xxx                | OK            | 69.6        44                1123 / 1123          |
    | 0                         | 0000:01:00.0  | 0           3845 / 15039      30709/ 32768         |
    +===========================+===============+====================================================+
    | 4     xxx                | OK            | 71.3        40                1123 / 1123          |
    | 0                         | 0000:C2:00.0  | 0           4296 / 15137      30709/ 32768         |
    +===========================+===============+====================================================+
    | 5     xxx                | OK            | 67.0        36                1123 / 1123          |
    | 0                         | 0000:82:00.0  | 0           3758 / 15137      30709/ 32768         |
    +===========================+===============+====================================================+
    | 6     xxx                | OK            | 71.7        37                1123 / 1123          |
    | 0                         | 0000:42:00.0  | 0           4581 / 15137      30710/ 32768         |
    +===========================+===============+====================================================+
    | 7     xxx                | OK            | 69.1        42                1123 / 1123          |
    | 0                         | 0000:02:00.0  | 0           4690 / 15039      30710/ 32768         |
    +===========================+===============+====================================================+
    +---------------------------+---------------+----------------------------------------------------+
    | NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
    +===========================+===============+====================================================+
    | 0       0                 | 139667        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 1       0                 | 139577        | python                   | 30782                   |
    +===========================+===============+====================================================+
    | 2       0                 | 139446        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 3       0                 | 139372        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 4       0                 | 139258        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 5       0                 | 139163        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 6       0                 | 139126        | python                   | 30780                   |
    +===========================+===============+====================================================+
    | 7       0                 | 139090        | python                   | 30780                   |
    +===========================+===============+====================================================+
    ```

- During training or inference, run the `top` command to query the resource usage and record the total physical memory used by the host, PID of each process, physical memory used by each process, and CPU usage of each process.

    ```shell
    top -p {pid_list} -n 1 -b
    ```

    Example:

    ```shell
    top -p 139667,139577,139446,139372,139258,139163,139126,139090 -n 1 -b
    ```

    Command output:

    ```ColdFusion
    top - 14:15:53 up 39 days, 22:54,  9 users,  load average: 28.32, 10.28, 5.44
    Tasks: 2727 total,   9 running, 1261 sleeping,   1 stopped,   0 zombie
    %Cpu(s):  5.6 us,  5.4 sy,  0.0 ni, 89.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
    KiB Mem : 80358528+total, 57884742+free, 70817856 used, 15392000+buff/cache
    KiB Swap:        0 total,        0 free,        0 used. 67941792+avail Mem

       PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
    139667 root      20   0 8203.5g   3.4g 526208 R 309.5  0.4   1:46.26 python
    139577 root      20   0 8203.5g   3.4g 526208 R 214.3  0.4   1:25.03 python
    139446 root      20   0 8203.5g   3.4g 526144 R 204.8  0.4   1:54.20 python
    139372 root      20   0 8203.5g   3.4g 526144 R 314.3  0.4   2:10.20 python
    139258 root      20   0 8203.5g   3.4g 526144 R 209.5  0.4   1:23.53 python
    139163 root      20   0 8203.5g   3.4g 526144 R 309.5  0.4   2:18.71 python
    139126 root      20   0 8203.5g   3.4g 526144 R 109.5  0.4   0:58.54 python
    139090 root      20   0 8203.5g   3.4g 526144 R 409.5  0.4   2:07.01 python
    ```

    The format of the saved file is as follows:

    The total physical memory used by the host and the `PID`, `RES`, and `%CPU` information of the training and inference processes are recorded in the format of `[Unix timestamp, Metric value]`. All information is stored in a JSON file named `host_metrics_$ {core_num}.json` in the following format:

    ```json
    host_metrics_${core_num}.json:
    {
    "node_mem_used": [[*Unix timestamp*, *Metric value*],...],
    "node_rss_{pid}": [[*Unix timestamp*, *Metric value*],...],
    "node_cpu_{pid}": [[*Unix timestamp*, *Metric value*],...],
    }
    ```

    - `core_num`: total number of CPU cores on the device.
    - `node_rss_${pid}`: list of metrics indicating the physical memory used by a process, corresponding to the `RES` data. The data is stored by PID.
    - `node_cpu_${pid}`: list of metrics indicating the percentage of CPU used by a process, corresponding to the `%CPU` data. The data is stored by PID.
    - `node_mem_used:` list of metrics indicating the total physical memory used by a host, corresponding to `KiB Mem: xxx used`.

    >[!NOTE]
    >If the collected host resource information contains a large amount of abnormal data, the device resource analysis result for further fault diagnosis may be abnormal, hindering the identification of the actual problem.

    Storage example:

    ```ColdFusion
    {
    "node_mem_used": [[1689732534, 10259988480],[1689732594, 10259988481]],
    "node_rss_139667": [[1689732534, 353370112],[1689732594, 353370115]],
    "node_cpu_139667": [[1689732534, "12.0"],[1689732594, "13.0"]],
    "node_rss_139577": [[1689732534, 224591872],[1689732594, 224591877]],
    "node_cpu_139577": [[1689732534, "24.0"],[1689732594, "27.0"]],
    "node_rss_139446": [[1689732534, 127008768],[1689732594, 127008769]],
    "node_cpu_139446": [[1689732534, "16.0"],[1689732534, "19.0"]]
    ...
    }
    ```

### MindIE Pod Log Collection <a name="ZH-CN_TOPIC_0000002358336673"></a>

**File Description <a name="section7111203341415"></a>**

- The MindIE Pod console logs are collected using Kubernetes commands or collection scripts. The logs contain instance node information and are stored in JSON files.
- Naming rule: `${pod_name}.json`
- Constraints on the storage path:
    - `Collection_directory/mindie/log/mindie_cluster_log/`
    - `${--mindie_log parameter-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

**Example <a name="section2640626141918"></a>**

1. Compile a script for collection by referring to [pod_log_collect.sh](https://gitcode.com/Ascend/mindxdl-deploy/blob/master/mindie/pod_log_collect.sh).
2. Ensure that the output path of the script is `Collection_directory/mindie/log/mindie_cluster_log/`. You can run the command in any directory.

    Output path example:

    ```shell
    log_dir="Collection_directory/mindie/log/mindie_cluster_log/"
    ```

    Example command:

    ```shell
    bash pod_log_collect.sh
    ```

    The `${pod_name}.json` file is generated in the output path directory.

**Collection Methods <a name="section9137173815206"></a>**

MindCluster Ascend FaultDiag can collect MindIE Pod console logs in either of the following ways:

- Script: Use the `pod_log_collect.sh` script to collect MindIE Pod console logs.
- Command: Collect MindIE Pod console logs using commands.

**Collection via Commands <a name="section3975440173719"></a>**

- After the MindIE Service is stably started, run the following command to collect MindIE Pod console logs.

    ```shell
    kubectl logs -f -n ${namespace} ${podname} | head -n 1000 > ${log_dir}/${podname}.log 2>&1 &
    ```

     View the `${podname}.log` file in the `${log_dir}` directory.

    The log content is as follows:

    ```ColdFusion
    ......
    INFO:root:status of ranktable is not completed, waiting for file update.
    INFO:root:status of ranktable is not completed, waiting for file update.
    INFO:root:status of ranktable is not completed, waiting for file update.
    {"IsMindIEEPJob":true,"status":"completed","server_list":[{"device":[{"device_id":"0","device_ip":"10.0.2.41","super_device_id":"113246208","rank_id":"0"},{"device_id":"1","device_ip":"10.0.3.41","super_device_id":"113311745","rank_id":"1"},{"device_id":"2","device_ip":"10.0.2.42","super_device_id":"113508354","rank_id":"2"},{"device_id":"3","device_ip":"10.0.3.42","super_device_id":"113573891","rank_id":"3"},{"device_id":"4","device_ip":"10.0.2.43","super_device_id":"113770500","rank_id":"4"},{"device_id":"5","device_ip":"10.0.3.43","super_device_id":"113836037","rank_id":"5"},{"device_id":"6","device_ip":"10.0.2.44","super_device_id":"114032646","rank_id":"6"},{"device_id":"7","device_ip":"10.0.3.44","super_device_id":"114098183","rank_id":"7"},{"device_id":"8","device_ip":"10.0.2.45","super_device_id":"114294792","rank_id":"8"},{"device_id":"9","device_ip":"10.0.3.45","super_device_id":"114360329","rank_id":"9"},{"device_id":"10","device_ip":"10.0.2.46","super_device_id":"114556938","rank_id":"10"},{"device_id":"11","device_ip":"10.0.3.46","super_device_id":"114622475","rank_id":"11"},{"device_id":"12","device_ip":"10.0.2.47","super_device_id":"114819084","rank_id":"12"},{"device_id":"13","device_ip":"10.0.3.47","super_device_id":"114884621","rank_id":"13"},{"device_id":"14","device_ip":"10.0.2.48","super_device_id":"115081230","rank_id":"14"},{"device_id":"15","device_ip":"10.0.3.48","super_device_id":"115146767","rank_id":"15"}],"server_id":"141.61.57.128","container_ip":"192.168.247.11"}],"server_count":"1","version":"1.2","super_pod_list":[{"super_pod_id":"1","server_list":[{"server_id":"141.61.57.128"}]}]}
    ......
    ```

    - `server_list`: list containing all nodes of the instance where the pod is located.
    - `container_ip`: container IP address
    - `device_id`: device ID

>[!NOTE]
>After the MindIE Pod is started, instance logs are recorded. Due to the aging mechanism of logs, if the collected MindIE Pod logs do not contain instance logs, multi-instance fault diagnosis will not be supported.

## Collecting Logs After Training and Inference

### NPU Environment Check File After Training and Inference

**File Description <a name="section56641436194180005"></a>**

- After the training and inference jobs are complete, use `hccn_tool` or script to query the IP address, mask, statistics on received and sent packets, and historical link statistics of each NPU network port. After the training is complete, use `npu-smi` or script to query the chip health information.
- Naming restriction: `npu_info_after.txt`.
- Constraints on the storage path:
    - `Collection directory/environment_check/`
    - `${--env_check-specified path}/`
    - For details, see [Log Collection Directory Structure](#log-collection-directory-structure).

**Collection Methods <a name="section225344011339"></a>**

MindCluster Ascend FaultDiag can collect NPU environment check files after training or inference jobs are finished in either of the following ways:

- Script: Use the `npu_info_collect.sh` script to collect the NPU environment check files. For details, see [log collection script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector).
- [Command](#section1020314437418_002): Run commands to collect NPU environment check files.

**Collection via Commands <a name="section1020314437418_002"></a>**

- After the training and inference jobs are complete, run the corresponding command to query the NPU environment check file, and save the query command and result to the `npu_info_after.txt` file. The involved commands and examples are as follows:
    - Query the network health status.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -net_health -g
        ```

        Command output:

        ```ColdFusion
        net health status: Init
        ```

    - Query the RoCE physical link connection status.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -link -g
        ```

        Command output:

        ```ColdFusion
        link status: UP
        ```

    - Query information about the RoCE network optical module.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -optical -g
        ```

        Command output:

        ```ColdFusion
        optical info:
        present              : not present
        ...
        Tx Power             : 4.4035 mW
        Rx Power             : 1.0189 mW
        Vcc High Thres       : 3465.00 mV
        Vcc Low Thres        : 3135.00 mV
        Temp High Thres      : 70 C
        Temp Low Thres       : 0 C
        TxPower High Thres   : 3.5481 mW
        TxPower Low Thres    : 0.2818 mW
        RxPower High Thres   : 3.5481 mW
        RxPower Low Thres    : 0.1445 mW
        Tx Bias              : 7.9360 mA
        Tx Los Flag          : 0x0
        Rx Los Flag          : 0xff
        Tx LoL Flag          : 0x0
        Rx LoL Flag          : 0xff
        ...
        ```

    - Query the TLS configuration.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -tls -g | grep switch
        ```

        Command output:

        ```ColdFusion
        dev_id:0, tls switch[0](0:disable, 1:enable), tls preconfigured[1](0:non-preset, 1:preset), tls alarm time threshold[60]days
        ```

    - Query the FEC mode.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -fec -g
        ```

        Command output:

        ```ColdFusion
        fec mode: rs FEC mode
        ```

    - Query the IP address and mask.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -ip -g
        ```

        Command output:

        ```ColdFusion
        ipaddr:10.xx.xx.10
        netmask:255.255.255.0
        ```

    - Query statistics about sent and received packets.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -stat -g
        ```

        Command output:

        ```ColdFusion
        packet statistics:
        mac_tx_mac_pause_num:0
        mac_rx_mac_pause_num:0
        mac_tx_pfc_pkt_num:0
        ...
        roce_qp_status_err_num:0
        nic_tx_all_pkg_num:122404
        nic_tx_all_oct_num:16921741
        nic_rx_all_pkg_num:6414803
        nic_rx_all_oct_num:482237805
        ```

    - Query the historical link statistics of the network port.

        ```shell
        /usr/local/Ascend/driver/tools/hccn_tool -i ${device_id} -link_stat -g
        ```

        Command output:

        ```ColdFusion
        [device 0]current time        : Wed Jun  7 10:08:28 2023
        [device 0]link up count       : 2
        [device 0]link change records :
        [device 0]    Tue Jun  6 16:32:12 2023    LINK UP
        [device 0]    Tue Jun  6 16:32:10 2023    LINK DOWN
        [device 0]    Tue Jun  6 16:31:55 2023    LINK UP
        ```

        The following is an example of information about device 0. You need to collect information about all devices.

        ```ColdFusion
        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -net_health -g
        net health status: Init

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -link -g
        link status: UP

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -optical -g
        optical info:
        present              : not present
        ...
        Tx Power             : 4.4035 mW
        Rx Power             : 1.0189 mW
        Vcc High Thres       : 3465.00 mV
        Vcc Low Thres        : 3135.00 mV
        Temp High Thres      : 70 C
        Temp Low Thres       : 0 C
        TxPower High Thres   : 3.5481 mW
        TxPower Low Thres    : 0.2818 mW
        RxPower High Thres   : 3.5481 mW
        RxPower Low Thres    : 0.1445 mW
        Tx Bias              : 7.9360 mA
        Tx Los Flag          : 0x0
        Rx Los Flag          : 0xff
        Tx LoL Flag          : 0x0
        Rx LoL Flag          : 0xff
        ...

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -tls -g | grep switch
        dev_id:0, tls switch[0](0:disable, 1:enable), tls preconfigured[1](0:non-preset, 1:preset), tls alarm time threshold[60]days

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -fec -g
        fec mode: rs FEC mode

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -ip -g
        ipaddr:10.xx.xx.10
        netmask:255.255.255.0

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -stat -g
        packet statistics:
        mac_tx_mac_pause_num:0
        mac_rx_mac_pause_num:0
        mac_tx_pfc_pkt_num:0
        ...
        roce_qp_status_err_num:0
        nic_tx_all_pkg_num:122404
        nic_tx_all_oct_num:16921741
        nic_rx_all_pkg_num:6414803
        nic_rx_all_oct_num:482237805

        /usr/local/Ascend/driver/tools/hccn_tool -i 0 -link_stat -g
        [device 0]current time        : Wed Jun  7 10:08:28 2023
        [device 0]link up count       : 2
        [device 0]link change records :
        [device 0]    Tue Jun  6 16:32:12 2023    LINK UP
        [device 0]    Tue Jun  6 16:32:10 2023    LINK DOWN
        [device 0]    Tue Jun  6 16:31:55 2023    LINK UP
        ```

        >[!NOTE]
        >The result of each collection command must be separated by one line. For example:
        >
        >```shell
        >/usr/local/Ascend/driver/tools/hccn_tool -i 0 -ip -g
        >XXXX
        >/usr/local/Ascend/driver/tools/hccn_tool -i 0 -stat -g
        >```

- After the training and inference jobs are complete, use `npu-smi` to query the chip health information and save the query command and result to the `npu_info_after.txt` file.
    - Query the basic information about the training or inference device.

        ```shell
        /usr/local/bin/npu-smi info
        ```

        Command output:

        ```ColdFusion
        +------------------------------------------------------------------------------------------------+
        | npu-smi 24.1.rc1                 Version: 24.1.rc1                                             |
        +---------------------------+---------------+----------------------------------------------------+
        | NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
        | Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | 7     xxx                | OK            | 67.0        44                0    / 0             |
        | 0                         | 0000:3D:00.0  | 0           2505 / 15567      0    / 32768         |
        +===========================+===============+====================================================+
        +---------------------------+---------------+----------------------------------------------------+
        | NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
        +===========================+===============+====================================================+
        | No running processes found in NPU 0                                                            |
        +===========================+===============+====================================================+
        ...
        | No running processes found in NPU 7                                                            |
        +===========================+===============+====================================================+
        ```

    - Query ECC of the high-bandwidth memory.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t ecc
        ```

        Command output:

        ```ColdFusion
        NPU ID                                   : 1
        Chip Count                               : 1
        ```

    - Query the basic information about the hardware.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t board
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Software Version               : 23.0.5
        Firmware Version               : 7.1.0.7.220
        Compatibility                  : OK
        Board ID                       : 0x02
        PCB ID                         : A
        BOM ID                         : 1
        PCIe Bus Info                  : 0000:61:00.0
        Slot ID                        : 0
        Class ID                       : NA
        PCI Vendor ID                  : 0x19e5
        PCI Device ID                  : 0xD801
        Subsystem Vendor ID            : 0x0200
        Subsystem Device ID            : 0x0100
        Chip Count                     : 1
        ```

    - Query the basic hardware information and the name of the specified device.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -c 0 -t board
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Chip ID                        : 0
        Chip Type                      : Ascend
        Chip Name                      : xxx
        Chip Version                   : V1
        Board ID                       : 0x02
        PCB ID                         : NA
        BOM ID                         : 1
        VDie ID                        : 42C711D4 20B03704 4A10C8D4 14CC040A D2102003
        NDie ID                        : 27216594 20401010 4E10C8D4 14CC040A A4102003
        Chip Position ID               : 0
        PCIe Bus Info                  : 0000:61:00.0
        Firmware Version               : 7.1.0.7.220
        ```

    - Query the memory usage.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -t usages
        ```

        Command output:

        ```ColdFusion
        NPU ID                         : 0
        Chip Count                     : 1

        DDR Capacity(MB)               : 13553
        DDR Usage Rate(%)              : 6
        DDR Hugepages Total(page)      : 0
        DDR Hugepages Usage Rate(%)    : 0
        HBM Capacity(MB)               : 32768
        HBM Usage Rate(%)              : 0
        Aicore Usage Rate(%)           : 0
        Aicpu Usage Rate(%)            : 0
        Ctrlcpu Usage Rate(%)          : 0
        DDR Bandwidth Usage Rate(%)    : 0
        HBM Bandwidth Usage Rate(%)    : 0
        Chip ID                        : 0
        ```

    - Query the chip health information.

        ```shell
        /usr/local/bin/npu-smi info -i ${device_id} -c 0 -t health
        ```

        Command output:

        ```ColdFusion
         Health Status                  : OK
         Error Code                     : NA
         Error Information              : NA
        ```

        The following is an example of obtained information. You need to collect information about all devices.

        ```ColdFusion
        /usr/local/bin/npu-smi info
        +------------------------------------------------------------------------------------------------+
        | npu-smi 23.0.5                   Version: 23.0.5                                               |
        +---------------------------+---------------+----------------------------------------------------+
        | NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
        | Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
        +===========================+===============+====================================================+
        | 0     xxx                 | OK            | 73.1        37                0    / 0             |
        | 0                         | 0000:61:00.0  | 0           920  / 13553      0    / 32768         |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | 7     xxx                 | OK            | 67.0        38                0    / 0             |
        | 0                         | 0000:3D:00.0  | 0           2346 / 15567      0    / 32768         |
        +===========================+===============+====================================================+
        +---------------------------+---------------+----------------------------------------------------+
        | NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
        +===========================+===============+====================================================+
        | No running processes found in NPU 0                                                            |
        +===========================+===============+====================================================+
        ...
        +===========================+===============+====================================================+
        | No running processes found in NPU 7                                                            |
        +===========================+===============+====================================================+

        /usr/local/bin/npu-smi info -i 0 -c 0 -t health
        Health Status                  : OK
        Error Code                     : NA
        Error Information              : NA

        /usr/local/bin/npu-smi info -i 0 -t ecc
        NPU ID                                   : 0
        Chip Count                               : 1

        DDR Single Bit Error Count               : 0
        DDR Double Bit Error Count               : 0
        DDR Single Bit Aggregate Total Err Cnt   : 0
        DDR Double Bit Aggregate Total Err Cnt   : 0
        DDR Single Bit Isolated Pages Count      : 0
        DDR Double Bit Isolated Pages Count      : 0
        HBM Single Bit Error Count               : 0
        HBM Double Bit Error Count               : 0
        HBM Single Bit Aggregate Total Err Cnt   : 0
        HBM Double Bit Aggregate Total Err Cnt   : 0
        HBM Single Bit Isolated Pages Count      : 0
        HBM Double Bit Isolated Pages Count      : 0
        Chip ID                                  : 0

        /usr/local/bin/npu-smi info -i 0 -t board
        NPU ID                         : 0
        Software Version               : 23.0.5
        Firmware Version               : 7.1.0.7.220
        Compatibility                  : OK
        Board ID                       : 0x02
        PCB ID                         : A
        BOM ID                         : 1
        PCIe Bus Info                  : 0000:61:00.0
        Slot ID                        : 0
        Class ID                       : NA
        PCI Vendor ID                  : 0x19e5
        PCI Device ID                  : 0xD801
        Subsystem Vendor ID            : 0x0200
        Subsystem Device ID            : 0x0100
        Chip Count                     : 1

        /usr/local/bin/npu-smi info -i 0 -c 0 -t board
        NPU ID                         : 0
        Chip ID                        : 0
        Chip Type                      : Ascend
        Chip Name                      : xxx
        Chip Version                   : V1
        Board ID                       : 0x02
        PCB ID                         : NA
        BOM ID                         : 1
        VDie ID                        : 42C711D4 20B03704 4A10C8D4 14CC040A D2102003
        NDie ID                        : 27216594 20401010 4E10C8D4 14CC040A A4102003
        Chip Position ID               : 0
        PCIe Bus Info                  : 0000:61:00.0
        Firmware Version               : 7.1.0.7.220

        /usr/local/bin/npu-smi info -i 0 -t usages
        NPU ID                         : 0
        Chip Count                     : 1

        DDR Capacity(MB)               : 13553
        DDR Usage Rate(%)              : 6
        DDR Hugepages Total(page)      : 0
        DDR Hugepages Usage Rate(%)    : 0
        HBM Capacity(MB)               : 32768
        HBM Usage Rate(%)              : 0
        Aicore Usage Rate(%)           : 0
        Aicpu Usage Rate(%)            : 0
        Ctrlcpu Usage Rate(%)          : 0
        DDR Bandwidth Usage Rate(%)    : 0
        HBM Bandwidth Usage Rate(%)    : 0
        Chip ID                        : 0

        /usr/local/bin/npu-smi info -i 0 -c 0 -t health
         Health Status                  : OK
         Error Code                     : NA
         Error Information              : NA
        ...
        ```

        >[!NOTE]
        >The result of each collection command must be separated by one line. For example:
        >
        >```shell
        >/usr/local/bin/npu-smi info -i 0 -c 0 -t health
        >XXXX
        >/usr/local/bin/npu-smi info -i 1 -c 0 -t health
        >```

- After the training and inference jobs are complete, run other related commands to query the environment check files of each NPU and save the query commands and results to the `npu_info_after.txt` file. The involved commands and examples are as follows:
    - Query the current system time.

        ```shell
        datetime=$(date "+%Y-%m-%d %H:%M:%S")
        echo "Datetime: $datetime">>${save_file}
        echo -e "\n">>${save_file}
        ```

        Command output:

        ```ColdFusion
        Datetime: 2024-06-26 01:13:36
        ```

    - Query the driver version.

        ```shell
        cat /usr/local/Ascend/driver/version.info
        ```

        Command output:

        ```ColdFusion
        Version=24.1.rc1
        ascendhal_version=7.35.19
        aicpu_version=1.0
        tdt_version=1.0
        log_version=1.0
        prof_version=2.0
        dvppkernels_version=1.1
        tsfw_version=1.0
        Innerversion=V100R001C15SPC006B220
        compatible_version=[V100R001C30],[V100R001C13],[V100R001C15],[V100R001C17]
        compatible_version_fw=[7.0.0,7.2.99]
        ```

    - Query the firmware version.

        ```shell
        cat /usr/local/Ascend/firmware/version.info
        ```

        Command output:

        ```ColdFusion
        Version=7.1.0.11.220
        firmware_version=1.0
        package_version=23.0.7
        compatible_version_drv=[23.0.rc3,23.0.rc3.],[23.0.0,23.0.0.]
        ```

    - Query the CANN version (AArch64).

        ```shell
        cat /usr/local/Ascend/cann/aarch64-linux/ascend_toolkit_install.info
        ```

        Command output:

        ```ColdFusion
        package_name=Ascend-cann-toolkit
        version=8.5.0
        innerversion=V100R001C25SPC001B212
        compatible_version=[V100R001C15],[V100R001C18],[V100R001C19],[V100R001C20],[V100R001C21],[V100R001C23]
        arch=aarch64
        os=linux
        path=/usr/local/Ascend/cann-8.5.0/aarch64-linux
        ```

    - Query the CANN version (x86_64).

        ```shell
        cat /usr/local/Ascend/cann/x86_64-linux/ascend_toolkit_install.info
        ```

        Command output:

        ```ColdFusion
        package_name=Ascend-cann-toolkit
        version=8.5.0
        innerversion=V100R001C25SPC001B212
        compatible_version=[V100R001C15],[V100R001C18],[V100R001C19],[V100R001C20],[V100R001C21],[V100R001C23]
        arch=x86_64
        os=linux
        path=/usr/local/Ascend/cann-8.5.0/x86_64-linux
        ```

    - Query the AI framework version.

        ```shell
        pip list | grep "torch"
        pip list | grep torch-npu
        pip list | grep "mindspore"
        ```

        Command output:

        ```ColdFusion
        torch              1.11.0
        torch-npu          2.1.0.post8.dev20241009
        mindspore          2.3.0
        ```

    - Query the firmware version details.

        ```shell
        /usr/local/Ascend/driver/tools/upgrade-tool --device_index -1 --component -1 --version
        ```

        Command output:

        ```ColdFusion
        {
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(0).
        {"device_id":0, "component":nve, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(3).
        {"device_id":0, "component":uefi, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(8).
        {"device_id":0, "component":imu, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(0), componentType(9).
        {"device_id":0, "component":imp, "version":7.1.0.7.220}
        ...
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(0).
        {"device_id":7, "component":nve, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(3).
        {"device_id":7, "component":uefi, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(8).
        {"device_id":7, "component":imu, "version":7.1.0.7.220}
        Get component version(7.1.0.7.220) succeed for deviceId(7), componentType(9).
        {"device_id":7, "component":imp, "version":7.1.0.7.220}
        }
        ```

### User Training and Inference Logs <a name="ZH-CN_TOPIC_0000001629917781"></a>

**File Description <a name="section56641436194180006"></a>**

- Include printed console logs generated by a training or inference job.
- Naming rule: The file name contains `rank-` or `worker-` and ends with `.txt` or `.log`.
- Constraints on the storage path: The data is stored in the collection directory. For details about the collection directory, see [Log Collection Directory Structure](#log-collection-directory-structure).

    **Figure 1**
    ![](../../figures/faultdiag/example.png)

**Collection Methods <a name="section1020314437418"></a>**

If the AI framework is used for training and inference, the Python logs printed on the screen are stored on the local PC in redirection mode. In the PyTorch framework, there is only one copy of console logs.

This feature requires that the training or inference logs of each user be dumped to the collection directory, and the training and inference logs of each card be named in the format of `/rank-(rank_id).log`, `/rank-(rank_id).txt`, `/worker-(worker_id).log`, or `/worker-(worker_id).txt`, for example, `Collection_directory/rank-0.txt`.

### CANN App Logs <a name="ZH-CN_TOPIC_0000001579557866"></a>

**File Description <a name="section56641436194180007"></a>**

- Include host-side App log `plog-{pid}-{time}.log` and device-side App log `device-{pid}-{time}.log`. For details, see [Viewing Logs (Ascend EP)](https://www.hiascend.com/document/detail/en/canncommercial/900/maintenref/logreference/logreference_0002.html) in the *CANN Log Reference*.
- Naming rule: `plog-{pid}-{time}.log`, `device-{pid}-{time}.log`
- Constraints on the storage path:
    - `Collection directory/`
    - `${--process_log}/`

- Directory structure

    ```text
    |-- process_log
        |-- debug
            |--plog # Directory for storing host-side App logs
               |--plog-{pid}-{unix time}.log
            |--device-0 # Directory for storing device-side App logs
                |--device-{pid}-{unix time}.log
            |--device-1
            |--device-2
            |--...
        |-- run
        |--operation
        |--security
    ```

**Collection Methods <a name="section1020314437418"></a>**

By default, CANN App logs are stored in the `${HOME}/ascend/log` directory. You can also use the environment variable `ASCEND_PROCESS_LOG_PATH` to customize the log storage path.

```shell
export ASCEND_PROCESS_LOG_PATH=${Custom_directory_path}
```

This feature requires that CANN App logs be dumped to `Collection_directory/process_log`. For details about the format restrictions, see [File Description](#section56641436194180007).

### Host-side Logs<a name="ZH-CN_TOPIC_0000002069726921"></a>

**File Description <a name="section44071732162918"></a>**

<a name="table11351184432912"></a>
<table><thead align="left"><tr id="row17351944152912"><th class="cellrowborder" valign="top" width="25.679999999999996%" id="mcps1.1.4.1.1"><p id="p19351114492919"><a name="p19351114492919"></a><a name="p19351114492919"></a>Log Name</p>
</th>
<th class="cellrowborder" valign="top" width="27.779999999999998%" id="mcps1.1.4.1.2"><p id="p8988155553014"><a name="p8988155553014"></a><a name="p8988155553014"></a>Naming Rule</p>
</th>
<th class="cellrowborder" valign="top" width="46.54%" id="mcps1.1.4.1.3"><p id="p193511844102913"><a name="p193511844102913"></a><a name="p193511844102913"></a>Storage Path</p>
</th>
</tr>
</thead>
<tbody><tr id="row12351744202916"><td class="cellrowborder" valign="top" width="25.679999999999996%" headers="mcps1.1.4.1.1 "><p id="p10351444142912"><a name="p10351444142912"></a><a name="p10351444142912"></a>Host OS log</p>
</td>
<td class="cellrowborder" valign="top" width="27.779999999999998%" headers="mcps1.1.4.1.2 "><p id="p11672035144216"><a name="p11672035144216"></a><a name="p11672035144216"></a>messages-*?</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="46.54%" headers="mcps1.1.4.1.3 "><p id="p63517444293"><a name="p63517444293"></a><a name="p63517444293"></a><em id="i44547719306"><a name="i44547719306"></a><a name="i44547719306"></a>Collection_directory</em></p>
</td>
</tr>
<tr id="row12351124412296"><td class="cellrowborder" valign="top" headers="mcps1.1.4.1.1 "><p id="p1624610033412"><a name="p1624610033412"></a><a name="p1624610033412"></a>Host kernel message log</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.4.1.2 "><p id="p139880554303"><a name="p139880554303"></a><a name="p139880554303"></a>dmesg</p>
</td>
</tr>
<tr id="row235294472916"><td class="cellrowborder" valign="top" headers="mcps1.1.4.1.1 "><p id="p177060159402"><a name="p177060159402"></a><a name="p177060159402"></a>Host system monitoring log</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.4.1.2 "><p id="p9601372512"><a name="p9601372512"></a><a name="p9601372512"></a>sysmonitor.log</p>
</td>
</tr>
<tr id="row397419114342"><td class="cellrowborder" valign="top" headers="mcps1.1.4.1.1 "><p id="p1497517119348"><a name="p1497517119348"></a><a name="p1497517119348"></a>Host kernel message log during a system crash</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.4.1.2 "><p id="p2975411163413"><a name="p2975411163413"></a><a name="p2975411163413"></a>vmcore-dmesg.txt</p>
</td>
</tr>
</tbody>
</table>

**Collecting Host OS Logs<a name="section20224123123311"></a>**

1. Go to the log storage directory and open the `messages` file.

    ```shell
    cd /var/log && vi messages
    ```

2. Obtain the corresponding log information based on the training and inference start and end time, create a `messages` file in the collection directory, and dump the log content.

    ```shell
    cd Collection directory / && vi messages
    ```

    Dump log information. A log example is as follows:

    ```ColdFusion
    Aug 13 03:19:24 # A training or inference job starts.
    ...
    Aug 13 04:14:39 # A training or inference job ends.
    ```

    Run the `:wq` command to save the file and exit. Log content is subject to the actual files.

**Collecting Host Kernel Logs<a name="section188199227357"></a>**

Run the following command to collect the latest `dmesg` log and place it in the collection directory. A maximum of 100,000 lines can be collected.

```shell
dmesg -T | tail -n 100000 > *Collection directory*/dmesg
```

A log example is as follows:

```ColdFusion
[Fri Aug 30 16:42:49 2024] Log printing
...
[Fri Aug 30 16:42:49 2024] Log printing
```

**Collecting Host System Monitoring Logs<a name="section060324111387"></a>**

Copy the `sysmonitor.log` file to the collection directory.

```shell
cp -r /var/log/sysmonitor.log Collection directory/
```

A log example is as follows:

```ColdFusion
2024-08-27T19:54:48.242959+00:00|info|sysmonitor[xxxxx]: Log printing
     ...
2024-08-27T19:54:48.343493+00:00|info|sysmonitor[xxxxx]: Log printing
```

**Collecting Host Kernel Logs During a System Crash<a name="section4793251445"></a>**

These logs are saved when the system breaks down. Perform the following steps to capture these logs:

Copy the `vmcore-dmesg.txt` file to the collection directory.

```shell
cp -r /var/crash/Collection_directory/
```

A log example is as follows:

```ColdFusion
[292.448078] Log printing
......
[292.448080] Log printing
```

**Collecting Host dmidecode Logs<a name="section73671119181810"></a>**

The host-side dmidecode logs contain DMI hardware information.

Run the following command to collect them:

```shell
dmidecode > dmidecode.txt
```

### Device-side Logs

**File Description <a name="section11629750194115"></a>**

- Include logs on the device.
- Naming rule: `device-os_{time}.log`, `event_{time}.log`, d`evice-{id}_{time}.log`, and `history.log`.
- Constraints on the storage path:
    - `Collection directory/device_log`
    - `${--device_log}/`

- Directory structure

    Ascend HDK 23.0.RC3:

    ```text
    |--device_log
        |-- slog
            |-- dev-os-3
                |-- debug
                    |--device-os
                        |-- device-os_{time}.log # System logs of the Ctrl CPU on the device
                |-- run
                    |--device-os
                        |-- device-os_{time}.log # System logs of the Ctrl CPU on the device
                |--device-0
                    |--device-0_{time}.log # System logs of the non-Ctrl CPU on the device
                |--device-2
                |--...
                |--slogd
                |--device_sys_init_ext.log
            |-- dev-os-7
            |-- ...
        |--hisi_logs
            |-- device-0
                |-- ...
                |-- history.log # Black Box log
            |-- device-2
            |-- ...
            |-- device_info.txt
    ```

    Ascend HDK 23.0.3 and later versions:

    ```text
    |--device_log
        |-- slog
            |-- dev-os-3
                |-- debug
                    |--device-os
                        |-- device-os_{time}.log # System logs of the Ctrl CPU on the device
                    |--device-0
                        |--device-0_{time}.log # System logs of the non-Ctrl CPU on the device
                    |--device-2
                    |--...
                |-- run
                    |--device-os
                        |-- device-os_{time}.log # System logs of the Ctrl CPU on the device
                    |--event
                        |-- event_{time}.log # Event-level system logs of the Ctrl CPU on the device
                |--...
                |--slogd
                |--device_sys_init_ext.log
            |-- dev-os-7
            |-- ...
        |--hisi_logs
            |-- device-0
                |-- ...
                |-- history.log # Black Box log
                |-- {time}/log/kernel.log # NPU kernel log
                |-- {time}/bbox/os/os_info.txt # Basic OS information on the device
                |-- {time}/mntn/hbm.txt # On-chip memory log on the device
            |-- device-2
            |-- ...
            |-- device_info.txt
    ```

**Collection Methods<a name="section07821997403"></a>**

Run the following command to collect device-side logs:

```shell
msnpureport
```

- After the command is executed, a folder with the timestamp is generated in the current directory. You need to dump the `slog` and `hisi_logs` folders in the timestamp directory to `Collection directory/device_log`.
- If collection fails, refer to [FAQs > Logs Not Flushed](https://www.hiascend.com/document/detail/en/canncommercial/900/maintenref/logreference/logreference_0024.html) in the *CANN Log Reference*.

### MindCluster Logs <a name="ZH-CN_TOPIC_0000002045702997"></a>

**File Description <a name="section4835161911300"></a>**

- File content: MindCluster logs.
- Naming rule: `devicePlugin*.log`, `noded*.log`, `runtime-run*.log`, `hook-run*.log`, `volcano-scheduler*.log`, `volcano-controller*.log`, and `npu-exporter*.log`
- Constraints on the storage path: The logs are stored in the `Collection directory/dl_log`.

**Collection Methods <a name="section2835619103019"></a>**

Go to the log storage directory and copy related component logs.

```shell
cp -r /var/log/mindx-dl/devicePlugin Collection_directory/dl_log
cp -r /var/log/mindx-dl/noded Collection_directory/dl_log
cp -r /var/log/ascend-docker-runtime Collection_directory/dl_log
cp -r /var/log/mindx-dl/volcano-scheduler Collection_directory/dl_log
cp -r /var/log/mindx-dl/volcano-controller Collection_directory/dl_log
cp -r /var/log/mindx-dl/npu-exporter Collection_directory/dl_log
```

### MindIE Logs <a name="ZH-CN_TOPIC_0000002071852084"></a>

**File Description <a name="section18774132472711"></a>**

- File content: MindIE logs
- Naming rule: `mindie-{module}_{pid}_{datetime}.log`
- Constraints on the storage path: The logs are stored in `Collection_directory/mindie/log/debug`.

**Collection Methods <a name="section1429254183012"></a>**

Before the collection, check whether the log flushing path of MindIE is set in the environment.

```shell
env | grep "MINDIE_LOG_PATH"
```

- If no result is displayed or the result does not contain an absolute path, for example, the following information is displayed:

    ```shell
    MINDIE_LOG_PATH="llm: llm"
    ```

    The logs are stored in the default path. Run the following command to go to the default log storage directory and copy the logs of related components:

    ```shell
    cp -r ~/mindie Collection directory
    ```

- If the command output is displayed and contains an absolute path, for example, the following information is displayed:

    ```shell
    MINDIE_LOG_PATH="llm: /home/working/"
    ```

    You need to go to the log storage directory in the command output and copy the logs of related components.

    ```shell
    cp -r /home/working collection directory
    ```

### AMCT Logs <a name="ZH-CN_TOPIC_0000002107731865"></a>

**File Description<a name="section16387124912381"></a>**

- File content: logs generated during model compression by using AMCT.
- Naming rule: `amct_{framework}.log`
- Constraints on the storage path: The file is stored in `Collection_directory/amct_log/`.

**Collection Methods<a name="section125257584397"></a>**

During model compression, the number of logs generated corresponds to the number of quantization processes. Typically, only one quantization process is initiated, resulting in a single log. This feature requires that the AMCT logs be dumped to the `Collection_directory/amct_log` directory.

Go to the log storage directory (process execution directory) and copy related log.

```shell
cp -r ~/amct_log collection directory/amct_log
```

### MindIO Logs <a name="ZH-CN_TOPIC_0000002107731865_01"></a>

**File Description**

- File content: logs generated during the running of MindIO components.
- Naming rule: `ttp_log.log.*`
- Constraints on the storage path: The file is stored in the `Collection directory/dl_log/ttp_log/` directory.

**Collection Methods**

When MindIO is running, each process generates a `ttp_log.log.* log` file. This feature requires that the MindIO logs be dumped to the `Collection_directory/dl_log/ttp_log/` directory.

Go to the directory where the console logs are stored and copy the logs of related components.

```shell
cp -r ~/ttp_log collection directory/dl_log/ttp_log
```

### Bus Logs <a name="ZH-CN_TOPIC_00000021077318650023"></a>

**File Description**

- File content: This log file is generated when the LCNE component (Ascend 950) is running.
- Naming restrictions: `log.log`
- Constraints on the storage path: The log file is stored in the `Collection_directory/lcne/` directory.

**Collection Methods**

When the the LCNE component (Ascend 950) is running, log files are generated.

- Method 1: Log in to the Ascend 950 1213 background and obtain the `log.log` file from the `/opt/vrpv8/home/logfile` directory.
- Method 2: Log in to the Ascend 950 1213 foreground, run the `collect diagnostic information` command to collect logs, and obtain the `diagnostic_information_*.zi`p compressed log file from the `/opt/vrpv8/home/logfile` directory on the Ascend 950 1213 background. You need to manually decompress all compressed logs.
