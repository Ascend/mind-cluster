# Performance Degradation, Slow Nodes, and Slow Networks

## Performance Degradation Faults<a name="ZH-CN_TOPIC_0000002479386488"></a>

### Using TaskD 7.1.RC1 or Later<a name="ZH-CN_TOPIC_0000002511346475"></a>

Together with the profiling capability provided by MindStudio, MindCluster cluster scheduling components provide diagnosis for performance degradation faults (slow nodes) in a cluster. This feature provides dynamic profiling and persistence of profile data. The profiling of a training job can be dynamically started and stopped, and diagnosis does not require restarting the job, causing no loss to training.

The currently supported profile data is shown in [Table 1](#zh-cn_topic_0000002194466236_table5530103025919).

**Table 1** Profile data description

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>

| Data Type|Supported AI Framework|Involved Component|
|--|--|--|
|<p>FP</p><p>(forward propagation data)</p>|<p>PyTorch</p><p>Only the single-operator scenario is supported.</p>|mstx_torch_plugin|
|<p>Step</p><p>(step latency)</p>|PyTorch, MindSpore|<ul><li>PyTorch<ul><li>Native optimizer scenario: If TorchNPU is 7.1.RC1, mstx_torch_plugin is required. If TorchNPU is later than 7.1.RC1, mstx_torch_plugin is not required, and TorchNPU provides Step data.</li><li>Custom optimizer scenario: Add profile data manually.</li></ul></li><li>MindSpore<ul><li>MindFormers scenario: Step data is provided by MindFormers.</li><li>MindSpeed scenario: Step data is not provided.</li></ul></li></ul>|
|<p>Communication</p><p>(communication operators)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|
|<p>SaveCheckpoint</p><p>(SaveCheckpoint time)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|
|<p>DataLoader</p><p>(Indicates the DataLoader time)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|

#### Constraints<a name="zh-cn_topic_0000002194466236_section487603614"></a>

- Currently, Step, SaveCheckpoint, FP, and DataLoader can be enabled only at the same time. To disable the preceding four data types, disable Communication at the same time.
- Communication operator data can be enabled and disabled separately.
- The dynamic lightweight profiling and the full profiling functions of MindStudio cannot be enabled at the same time. Enabling full profiling will cause abnormal data collection for performance degradation faults.

### Using Another TaskD Version<a name="ZH-CN_TOPIC_0000002511346483"></a>

MindCluster cluster scheduling components use the profiling capability provided by MindStudio to provide diagnosis for performance degradation faults (slow nodes) in a cluster. This feature provides dynamic profiling and persistence of profile data. The profiling of a training job can be dynamically started and stopped, and diagnosis does not require restarting the job, causing no loss to training.

The currently supported profile data is shown in [Table 2](#zh-cn_topic_0000002194466236_table553010302591923).

**Table 2** Profile data description

<a name="zh-cn_topic_0000002194466236_table553010302591923"></a>

|Data Type|Supported AI Framework|Involved Component|
|--|--|--|
|<p>FP</p><p>(forward propagation data)</p>|<p>PyTorch</p><p>Only the single-operator scenario is supported.</p>|mstx_torch_plugin|
|<p>Step</p><p>(step latency)</p>|PyTorch, MindSpore|<ul><li>PyTorch<ul><li>Native optimizer scenario: If TorchNPU is 7.1.RC1 or earlier, mstx_torch_plugin is required. If TorchNPU is later than 7.1.RC1, mstx_torch_plugin is not required, and TorchNPU provides Step data.</li><li>Custom optimizer scenario: Add profile data manually.</li></ul></li><li>MindSpore<ul><li>MindFormers scenario: Step data is provided by MindFormers.</li><li>MindSpeed scenario: Step data is not provided.</li></ul></li></ul>|
|<p>Communication</p><p>(Indicates communication operators)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|
|<p>SaveCheckpoint</p><p>(Indicates the SaveCheckpoint time)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|
|<p>DataLoader</p><p>(Indicates the DataLoader time)</p>|PyTorch, MindSpore|<ul><li>PyTorch: TorchNPU</li><li>MindSpore: MindSpore framework</li></ul>|

#### Constraints<a name="zh-cn_topic_0000002194466236_section487603614-duplicate-2"></a>

- Currently, Step, SaveCheckpoint, FP, and DataLoader can be enabled only at the same time. To disable the preceding four data types, disable Communication at the same time.
- Communication operator data can be enabled and disabled separately.
- The dynamic lightweight profiling and the full profiling functions of MindStudio cannot be enabled at the same time. Enabling full profiling will cause abnormal data collection for performance degradation faults.

## Slow Node and Slow Network Faults<a name="ZH-CN_TOPIC_0000002511426421"></a>

<a name="ZH-CN_TOPIC_0000002532640773"></a>

MindCluster cluster scheduling components use the online diagnosis capability provided by MindCluster Ascend FaultDiag (fault diagnosis tool) to provide diagnosis for slow node and slow network faults in a cluster.

### Slow Node Diagnosis<a name="ZH-CN_TOPIC_0000002500880704"></a>

#### Function Description<a name="zh-cn_topic_0000002278667326_section27999216294"></a>

For the node training performance degradation phenomenon in an AI cluster, real-time detection of slow nodes caused by computing domain problems or network problems is supported, so that you can isolate slow nodes by switching or other methods.

Currently, only online deployment integrated with ClusterD and NodeD is supported. For details, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md) to complete ClusterD and NodeD deployment.

- Slow node algorithm: Perceives the real-time degradation state based on key performance indicators in training scenarios. For the synchronization relationship between communication operators and computing operators, it demarcates slow computing card and slow communication domain problems.
- Slow node cleansing: Converts and cleanses the incremental data inside a node and generates a cleansing result csv file.
- Slow node scheduling: Schedules the entire slow node process and controls data cleansing and the slow node algorithm.

### Slow Network Diagnosis<a name="ZH-CN_TOPIC_0000002500720860"></a>

#### Function Description<a name="zh-cn_topic_0000002313236861_section27999216294"></a>

Detection of the connectivity of the parameter plane network is supported, and network monitoring and exception reporting are performed in real time, assisting fault analysis and demarcation and locating, giving early warning of network faults and sub-health risks, and ensuring long-term stable running of the cluster network.

Currently, only online deployment integrated with ClusterD and NodeD is supported. For details, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md) to complete ClusterD and NodeD deployment.

- Slow network algorithm: Analyzes and detects the network dial test data between nodes and outputs network diagnosis results.
- Slow network scheduling: Controls the start and stop of detection processes, reports fault results, and schedules the entire slow network process.

## Related Operations

- [Configuring Performance Degradation Diagnosis](../03_configuration/06_performance_diagnosis.md)
- [Querying and Verifying Faults](../04_querying_and_verifying_faults.md)
