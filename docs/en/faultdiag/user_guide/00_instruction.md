# Instructions<a name="ZH-CN_TOPIC_0000001677859701"></a>

## Notes<a name="section181951321155817"></a>

- To enable the diagnosis function, it is recommended that the number of servers in a cluster be fewer than or equal to 128 (1024 cards) due to the limitation of the maximum number of processes (1024 by default) in Linux. If the number of servers exceeds the upper limit, run the `ulimit -n` command to adjust the upper limit of file descriptors.
- Do not use pipe commands when using MindCluster Ascend FaultDiag commands. Otherwise, user IP address acquisition and log audit may be affected.

## Supported Scenarios<a name="section12381154319315"></a>

- MindCluster Ascend FaultDiag provides fault diagnosis capabilities exclusively for training and inference tasks on servers with full-card configurations. In other scenarios, root cause localization may be inaccurate or may fail.
- MindCluster Ascend FaultDiag supports only IPv4 addresses.

## Notes for System Time<a name="section98531395015"></a>

- Synchronize the system time of each training or inference server. If the system time is inconsistent, the analysis result may be inaccurate.
- Synchronize the system time of the host on each training or inference server with that of the device. If the system time is inconsistent, the analysis result may be inaccurate.
- If a container is used to execute training or inference tasks, synchronize the system time of the host with that of the training or inference container. If the system time is inconsistent, the analysis result may be inaccurate.

## Version Mapping<a name="section087884485711"></a>

**Table 1** Software versions corresponding to logs

|Log|Software|Software Version|Description|
|--|--|--|--|
|CANN App logs|CANN|7.0.RC1 or later|Host App logs and device App logs printed by CANN. For more information, see [Viewing Logs (Ascend EP)](<https://www.hiascend.com/document/detail/en/canncommercial/850/maintenref/logreference/logreference_0002.html>) in the *CANN Log Reference*.|
|Training and inference logs of the PyTorch framework|PyTorch adaptation plugin 1.11.0|5.0.RC3 or later|-|
|Training logs of the MindSpore framework|MindSpore|2.1.0 or later|Certain fault type descriptions may contain relevant MindSpore version notes. The actual fault diagnosis description shall prevail.|
|Host OS logs|-|-|<ul><li>The supported host OS logs include but are not limited to CentOS 7.6, Debian 10.0, EulerOS 2.10, EulerOS 2.12, and CTyunOS 22.06. The keywords in logs may vary according to the OS. </li><li>It is recommended that the host OS log size be less than 512 MB.</li></ul>|
|Device logs|Ascend HDK|23.0.RC3 or later|-|
|MindCluster component logs|Ascend Device Plugin, NodeD, Ascend Docker Runtime, NPU Exporter, and Volcano|6.0.RC3 or later|-|
|MindIE component logs|MindIE Server, MindIE LLM, MindIE SD, MindIE RT, MindIE Torch, MindIE MS, MindIE Benchmark, and MindIE Client|6.0.0 or later|-|
|AMCT logs|Model compression toolkit|7.0.RC1 or later|AMCT is integrated into the CANN package for release. For more information, see [AMCT User Guide](<https://www.hiascend.com/document/detail/en/canncommercial/850/devaids/amct/atlasamct_16_0001.html>).|
|MindIE Pod console logs|MindIE Pod console logs|-|-|
|MindIO logs|MindIO logs|-|-|
