# Key Features

**Key features of MindCluster cluster scheduling components**

- **Enhanced resumable training and fault recovery**: Supports resumable training triggered by DPU faults and DPU sub-health faults, and supports instance-level recovery in reinforcement learning (RL) scenarios. The Atlas 950 SuperPoD supports process-level online recovery for UB network faults.
- **Enhanced inference high availability**: Supports composite metric definitions for elastic scaling, container snapshots during elastic scaling, and inference job recovery by reducing the number of NPUs or Pods.
- **Scheduling and device management**: Supports ignoring the RoCE network health status during job scheduling, and supports generic affinity scheduling within a node and cross-iteration hybrid scheduling. Supports affinity scheduling of the Atlas 950 SuperPoD Flex.
- **Ecosystem version adaptation**: Supports Kubernetes 1.36 and Volcano 1.15.
- **Runtime capability extension**: Ascend Docker Runtime adds support for CRI-O and supports mounting UMDK and UB driver user-mode files.
- **Enhanced observability**: Adds DPU Exporter component, which supports 1825 metric collection. NPU Exporter supports reporting the number of available chips, and Ascend Device Plugin supports reporting device NUMA information.
- **Hardware adaptation**: Adapts the basic capabilities of the Atlas 950 SuperPoD Flex.

**Key features of MindCluster Ascend FaultDiag**

- **New K8s ClusterOps Agent**: Supports collecting the logs of failed Ascend training/inference jobs in a K8s cluster and performing fault diagnosis according to user instructions, and supports connecting to model APIs to optimize diagnosis reports.
- **Enhanced link diagnosis tool**: Adds support for optical link fault diagnosis of <term>Ascend 950PR&950DT products</term>. Fault thresholds can be modified through the configuration file, and different thresholds can be set by optical module type and metric.

# Version Mapping Description

## Version Information

| Product | MindCluster |
|--|--|
|Product Version|26.2.0|
|Version Type|Release|
|Maintenance Period|1 year|

> [!NOTE]
>
> MindCluster 26.0 version planning: MindCluster 26.0.0, MindCluster 26.1.0, MindCluster 26.2.0, and MindCluster 26.3.0.

## Related Product Version Compatibility

**Table 1** MindCluster software version compatibility

|MindCluster|CANN| HDK    |MindSpeed-LLM|TorchNPU|
|--|--|--------|--|--|
|26.2.0|9.2.0| 26.2.0 |26.2.0|26.2.0|

# Version Compatibility

MindCluster components must be used together as a matched set. Do not mix components across different versions.
>[!NOTE]
>In the tables in this section, "/" indicates that versions are not compatible, and "Y" indicates versions are compatible.

**Table 2** MindCluster and CANN version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 98px">
<col style="width: 98px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">CANN</th>
  </tr>
  <tr>
    <th>8.5.X</th>
    <th>9.0.X</th>
    <th>9.1.X</th>
    <th>9.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 3** MindCluster and HDK version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 232px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="1">HDK of <term>Ascend 950PR&950DT Products</term>(Before June 30, 2026)</th>
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
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">HDK</th>
  </tr>
  <tr>
    <th>25.5.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 4** MindCluster and MindSpeed-LLM version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">MindSpeed-LLM</th>
  </tr>
  <tr>
    <th>2.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 5** MindCluster and TorchNPU version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">TorchNPU</th>
  </tr>
  <tr>
    <th>7.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**Table 6** MindCluster and MindSpore version compatibility

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">MindSpore</th>
  </tr>
  <tr>
    <th>2.7.2</th>
    <th>2.9.X</th>
    <th>2.10.X</th>
    <th>2.11.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

## Version Usage Notes

- HDK version numbers vary with hardware and generations. Determine the hardware form first, then find the corresponding HDK version for that hardware form, and finally find the MindCluster version that matches it in the compatibility tables.
- MindSpore 2.11.0 no longer supports advanced resumable training capabilities such as process-level rescheduling, process-level online recovery, online stress testing, and link failover communication when used with MindCluster. Only the basic Job-level and Pod-level rescheduling capabilities are retained.

## Update Notes

### New Features

**MindCluster cluster scheduling components**

**1. Resumable training and fault recovery**

- Supports resumable training triggered by DPU faults.
- Supports proactively triggering resumable training through a configurable policy when a DPU is sub-healthy.
- Supports instance-level recovery in RL scenarios.
- The Atlas 950 SuperPoD supports process-level online recovery for UB network faults.

**2. High inference reliability**

- Supports composite metric definitions for elastic scaling.
- Supports enabling container snapshots during elastic scaling.
- Supports inference job recovery by reducing the number of NPUs or Pods.

**3. Reliability and robustness**

- Completes the reliability verification of fast recovery capabilities.
- Completes the reliability verification of proactively isolating nodes for sub-healthy faults such as lane degradation.
- Supports 100% recovery of self-healing faults in appliances.

**4. Scheduling**

- K8s resource scheduling supports heterogeneous PD instances.
- Pods that do not request NPUs can be scheduled normally without configuring the `skip` flag.
- Supports the kube-scheduler and worker node components of Kubernetes 1.36.
- Supports Volcano 1.15.
- The ranktable in `job-summary` supports <term>Ascend 950PR&950DT products</term>.
- Supports generic affinity scheduling within a node and cross-iteration hybrid scheduling.
- Supports affinity scheduling of the Atlas 950 SuperPoD Flex.
- Affinity scheduling supports Pod types.
- Provides hardware basic capability labels.
- Infer Operator supports automatic instance affinity configuration.
- Infer Operator supports pulling up MetaService.
- Supports ignoring the RoCE network health status during job scheduling.

**5. Device management**

- hccn_tool is changed to the DCMI to optimize scheduling duration.
- Supports output of component external version information.

**6. Ascend Docker Runtime**

- Supports CRI-O.
- Supports mounting UMDK and UB driver user-mode files.

**7. Observability**

- DPU Exporter supports 1825 metric collection.
- NPU Exporter supports reporting the number of available chips.
- Ascend Device Plugin supports reporting device NUMA information.
- vNPU hard partitioning metric monitoring is enhanced.
- Supports a new NPU utilization metric interface for <term>Ascend 950PR&950DT products</term>.

**8. Hardware adaptation**

- Adapts the basic capabilities of the Atlas 950 SuperPoD Flex.

**9. Architecture optimization**

- Completes the decoupling of internal MindCluster components.
- Reduces third-party dependencies.

**10. Documentation**

- Optimizes the documentation related to resumable training and fault monitoring.

**MindCluster Ascend FaultDiag**

**1. Log diagnostic tool**

- Enhances the UB link fault diagnosis capability based on the Cqe error status fault type and combined with log refinement.
- Adds support for fault diagnosis of Atlas 950 SuperPoD Flex.

**2. Link diagnostic tool**

- Adds support for optical link fault diagnosis of <term>Ascend 950PR&950DT products</term>.
- Fault thresholds can be modified through the configuration file, and different thresholds can be set by optical module type and metric.

**3. K8s ClusterOps Agent**

- Supports collecting the logs of failed Ascend training/inference jobs in a K8s cluster and performing fault diagnosis according to user instructions, and supports connecting to model APIs to optimize diagnosis reports.

## Key Feature Changes

This release inherits some features of MindCluster 26.1.0 and earlier releases. The main changes are as follows:

- The hccn_tool interface is changed to be implemented based on the DCMI to optimize scheduling duration.

## Service Interface Changes

- Ascend Docker Runtime adds support for CRI-O.
- Ascend Device Plugin reports device NUMA information.
- NPU Exporter reports the number of available chips.
- The hccn_tool interface is changed to be implemented based on the DCMI to optimize scheduling duration.
- Pods that do not request NPUs can be scheduled normally without configuring the `skip` flag (scheduling behavior adjustment).
- Volcano supports ignoring the RoCE network health status during job scheduling.
- For other compatibility issues, see [MindCluster Component Compatibility Issue Announcement](https://gitcode.com/Ascend/mind-cluster/issues/588).

## Fixed Issues

- In earlier versions, the recovery detection period for fault code 81078603 was 5 minutes. In this release, it is optimized to a short-period fast detection mechanism with a backoff mechanism, which can detect recovery events and clear faults more quickly.
- In MindCluster 26.1.0 and earlier versions, when operator re-execution was enabled for <term>Atlas A3 products</term>, process-level rescheduling could fail in linkdown fault scenarios. This issue has been fixed in this release. For details, see the [PR 4253](https://gitcode.com/Ascend/mind-cluster/pull/4253).

## Known Issues

After multiple rescheduling recoveries, the process-level rescheduling feature is subject to a gloo segmentation fault in the native PyTorch component, with a probability of approximately 0.00125. For details, see [issue 188266](https://github.com/pytorch/pytorch/issues/188266). You can configure Job/Pod-level rescheduling as a fallback measure.

# Upgrade Impact

## Impact of the Upgrade Process on the Current System

None

## Impact on the Current System After Upgrade

When upgrading the Infer Operator component from a version earlier than 26.1.0 to 26.1.0 or later, you need to delete the log directory and recreate it with `root` permissions, or change the permissions of the log directory and log files to `root`.

# Document Reference

|Document|Description|Release Notes|
|--|--|--|
|[*MindCluster 26.2.0 Cluster Scheduling User Guide*](./scheduling/01_introduction/00_overview.md)|Provides component descriptions, feature principles, and usage references for cluster scheduling, including installation and deployment of each component, integration and adaptation examples, API references, and principle introductions for some scheduling solutions.|Adds guidance for features such as resumable training (triggered by DPU faults or sub-health faults) and heterogeneous PD instances, and restructures the resumable training chapter. For details about other changes, see the [*MindCluster 26.2.0 Cluster Scheduling User Guide*](./scheduling/01_introduction/00_overview.md).|
|[*MindCluster 26.2.0 Fault Diagnosis User Guide*](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md)|Provides usage guidance for log collection, log cleaning and dumping, fault diagnosis, and other functions.|Optimizes the guidance related to resumable training and fault monitoring. For details about other changes, see the [*MindCluster 26.2.0 Fault Diagnosis User Guide*](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md).|

# Virus Scan Results

The virus scan passed.

# Fixed Vulnerabilities List

None

# Change History

|Document Version|Release Date|Change Description|
|--|--|--|
|01|2026-09-30|First official release|
