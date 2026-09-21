# Release Notes

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T07:50:59.233Z pushedAt=2026-08-24T08:49:07.134Z -->

## Version Mapping Description

### Version Information

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>

<table><tbody><tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.1.1"><p>Product Name</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p>MindCluster</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.2.1"><p>Product Version</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p>26.1.0</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.3.1"><p>Version Type</p></th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>Release Version</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>MindCluster 26.0 version planning: MindCluster 26.0.0, MindCluster 26.1.0, MindCluster 26.2.0, and MindCluster 26.3.0.

### Related Product Version Compatibility

**Table 1**  MindCluster version compatibility

|MindCluster|CANN| HDK                                                                   |MindSpeed-LLM|TorchNPU|MindSpore|
|--|--|-----------------------------------------------------------------------|--|--|--|
|26.1.0|9.1.0| <ul><li>Atlas 350 accelerator card: 25.7.RC1</li><li>Atlas 950 SuperPoD: 25.1.RC1</li><li>Atlas 850E SuperPoD/Atlas 650E server: 25.6.RC1</li><li>Other products: 26.1.0</li></ul> |26.1.0|26.1.0|2.10.0|

## Version Compatibility

MindCluster components must be used together as a matched set. Do not mix components across different versions.

>[!NOTE]
>In the tables in this section, "/" indicates that versions are not compatible, and "Y" indicates versions are compatible.

**Table 2**  MindCluster and CANN version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">CANN</th>
  </tr>
  <tr>
    <th>8.5.X</th>
    <th>9.0.X</th>
    <th>9.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 3**  MindCluster and HDK version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="1">HDK of <term>Ascend 950 Products</term></th>
  </tr>
  <tr>
    <th>25.1.RC1/25.6.RC1/25.7.RC1</th>
  </tr></thead>
<tbody>

  <tr>
    <td>26.0.0</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">HDK of Other Products</th>
  </tr>
  <tr>
    <th>25.5.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 4**  MindCluster and MindSpeed-LLM version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">MindSpeed-LLM</th>
  </tr>
  <tr>
    <th>2.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 5**  MindCluster and TorchNPU version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">TorchNPU</th>
  </tr>
  <tr>
    <th>7.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 6**  MindCluster and MindSpore version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">MindSpore Version</th>
  </tr>
  <tr>
    <th>2.7.2</th>
    <th>2.9.X</th>
    <th>2.10.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

## Version Usage Notes

None

## Update Notes

### New Features

|Component| Feature Description |
|--|----------|
|MindCluster Ascend FaultDiag| <ul><li>In the existing fault mode library, fault modes are supplemented for <term>Ascend 950 products</term>, and fault modes based on pyMotor+vLLM are newly supported.</li><li>The content of the output report of Ascend FaultDiag Toolkit is optimized.</li><li>Fault diagnosis supports IPv6 scenarios.</li></ul>   |
|MindCluster cluster scheduling components| <ul><li>Support storage of DTFS faults.</li><li>The Atlas 950 SuperPoD supports IPv6 scenarios (except the parameter plane). Atlas A2/A3 products support IPv6.</li><li>Volcano supports priority scheduling of pods back to their original running nodes.</li><li>Ascend Device Plugin supports pluggable fault handling.</li><li>Support in-band detection and reporting of 1825 faults.</li><li>Provide RDMA device plugin for 1825 NICs.</li><li>One-click helm deployment is supported for improved usability.</li><li>NPU Exporter supports grouped configuration of collection periods.</li><li>Hot switching is supported in multi-level scheduling scenarios.</li><li>Support configuration of liveness probes.</li><li>Add hang detection and recovery functionality.</li><li>Basic capabilities such as device management, affinity scheduling, metric monitoring, fault detection, and RankTable generation are supported for the Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD.</li><li>Basic resumable training is supported for the Atlas 850E SuperPoD and Atlas 650E server; full resumable training is supported for the Atlas 950 SuperPoD.</li><li>Containerization is supported for the Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD.</li><li>Inference fault recovery is supported, including priority scheduling, reducing prefill instances to preserve decode instances, and instance-level rescheduling.</li><li>Load-based elastic scaling and container snapshot capabilities are supported for inference scenarios.</li></ul> |

### Key Feature Changes

MindCluster cluster scheduling components:

- Ascend Docker Runtime supports configuring the `LD_LIBRARY_PATH` environment variable of the HDK driver by default, so that the npu-smi tool can be used properly.
- When components such as Ascend Device Plugin, NPU Exporter, and NodeD are started, if the number of chips is insufficient, the default value of the `-deviceResetTimeout` parameter — which controls the maximum wait time for the driver to report all chips — has been increased from 60 seconds to 600 seconds.
- The log permission of the Infer Operator component is updated to `root` permission.

### Service Interface Changes

|Component|Interface Change|
|--|--|
|MindCluster Ascend FaultDiag|<ul><li>Ascend FaultDiag Toolkit adds the configuration command `set_config_dir`, which currently supports only setting the path where the networking configuration file `LLD.xlsx` resides.</li><li>Features that impact performance — such as resource preemption and network congestion — can affect training and inference workloads during metric data collection. These features will be deprecated and sunset in a future release.</li></ul>|
|MindCluster cluster scheduling components|A liveness probe is added to all components deployed on K8s.|

### Fixed Issues

- Fixed the soft link verification error after the Atlas 350 accelerator card driver was deployed.
- Fixed an issue where the Ascend Device Plugin incorrectly determined whether an NPU was occupied due to the `huawei.com/AscendReal` annotation being assigned with different meanings for `phyID` and `logicID` on <term>Ascend 950 products</term>.
- Fixed an issue in `device-cm` where the chip name in `ManuallySeparateNPU` was not adapted to NPU on <term>Ascend 950 products</term>.
- Fixed an issue where `optical_index` existed but was not reported when NPU Exporter failed to collect optical module metrics.

### Known Issues

After multiple rescheduling recoveries, the process-level rescheduling triggers a gloo segmentation fault in the native PyTorch component, with a probability of approximately 0.00125 (see [issue 188266](https://github.com/pytorch/pytorch/issues/188266)). You can configure Job/Pod-level rescheduling as a fallback measure.

When operator re-execution is enabled on Atlas A3 products, process-level rescheduling may fail in linkdown fault scenarios. You can configure process-level online recovery to avoid this, or configure Job/Pod-level rescheduling as a fallback measure.

## Upgrade Impact

### Impact of the Upgrade Process on the Current System

None

### Impact on the Current System After Upgrade

When upgrading the Infer Operator component from a version earlier than 26.1.0 to 26.1.0 or later, you need to delete the log directory and recreate it with `root` permissions, or change the permissions of the log directory and log files to `root`.

## Document Reference

|Document|Description|Release Notes|
|--|--|--|
|[*MindCluster Cluster Scheduling User Guide*](./scheduling/01_introduction/00_overview.md)|Provides component descriptions, feature principles, and usage references for cluster scheduling, including installation and deployment of each component, integration and adaptation examples, API references, and principle introductions for some scheduling solutions.|Added installation of components using Helm, developer guide, container snapshot deployment and usage, etc. For details on other changes, see the [*MindCluster Cluster Scheduling User Guide*](./scheduling/01_introduction/00_overview.md).|
|[*MindCluster Fault Diagnosis User Guide*](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md)|Provides usage guidance for log collection, log cleaning and dumping, fault diagnosis, and other functions.|Added support for Ascend 950 products, fault modes based on pyMotor+vLLM, etc. For details on other changes, see the [*MindCluster Fault Diagnosis User Guide*](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md).|

## Virus Scan Results

The virus scan passed.

## Fixed Vulnerabilities List

None
