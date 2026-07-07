# Introduction

## About This Document<a name="ZH-CN_TOPIC_0000002017481558"></a>

**Disclaimer<a name="zh-cn_topic_0000001461453746_section582264385620"></a>**

- This document may contain third-party information, products, services, software, components, data, or content (collectively referred to as "Third-Party Content"). Huawei does not control and assumes no responsibility for Third-Party Content, including but not limited to accuracy, compatibility, reliability, availability, legality, appropriateness, performance, non-infringement, update status, etc., unless otherwise expressly stated in this document. Mentioning or referencing any Third-Party Content in this document does not imply Huawei's endorsement or guarantee of such Third-Party Content.
- This feature reads and processes the relevant original logs and monitoring metric files collected by users in the input directory. Users must ensure that the relevant files contain no sensitive information or personal data. Huawei does not control and assumes no responsibility for the content of the input data.

- If users require third-party licenses, they must obtain them through legal channels, unless otherwise explicitly stated in this document.

**Audience<a name="section55821246876"></a>**

This document is intended for the following personnel:

- Huawei technical support engineers
- Technical support engineers of channel partners
- Enterprise administrators
- Enterprise end users

## Overview<a name="ZH-CN_TOPIC_0000002005252541"></a>

MindCluster Ascend FaultDiag provides log cleaning and fault diagnosis capabilities. It extracts key information from logs generated during training and inference, then analyzes the cleaned data from all cluster nodes to identify the root cause node and fault events.

### Key Features<a name="section14223166131518"></a>

MindCluster Ascend FaultDiag mainly provides the following two major functions:

**Log Cleaning**

After a training or inference task fails, MindCluster Ascend FaultDiag performs a series of cleaning operations on the original logs and monitoring metric information. The cleaning results are dumped together with the original information to the same path, providing data support for diagnosis tasks.

Currently, the supported cleaning content mainly includes: original training and inference logs such as user training and inference logs, CANN App logs, host-side resource information, and NPU network port resource information, as well as monitoring metrics.

**Fault Diagnosis**

MindCluster Ascend FaultDiag provides diagnostic functions for the following two types of issues, and supports root cause node analysis, fault event analysis, device resource analysis, and network congestion analysis.

|Fault Classification|Diagnostic Content|
|--|--|
|Abnormal exit of training and inference tasks|<ul><li>Root cause node analysis: Based on the HCCL error message of cluster communication, locate the root cause node that triggered the error.</li><li>Fault event analysis: Analyze the root cause error of the device where the root cause node resides based on the fault patterns contained in the fault knowledge graph.</li></ul>|
|Performance degradation during training and inference|<ul><li>Device resource analysis for device resource status: By analyzing the device-related metric files collected by the user, locate issues such as computing frequency reduction and CPU resource contention.</li><li>Network congestion analysis: Analyze the network status between nodes, typically used to locate network issues in Spine + Leaf networking scenarios. By analyzing the NPU network port's monitoring metric files collected by the user, it analyzes whether network congestion anomalies occur on node links.</li></ul>|

> **NOTE**
> Performance degradation issues are diagnosed only when the training and inference tasks have not exited abnormally.

**Usage Process<a name="section7779135035518"></a>**

The usage process of MindCluster Ascend FaultDiag is shown in the following table.

|Step|Description|Reference|
|--|--|--|
|Log collection|<p>When a training or inference task fails or becomes abnormal, collect logs from each training or inference device and store them according to a predefined structure.</p><p>For details about the logs to be collected, see the "Table. Training and inference task log and metric information" in [Application Scenarios and Solutions](#application-scenarios-and-solutions).</p>|For details, see the [Log Collection](./user_guide/03_collecting_logs.md) section.|
|Log cleaning|After log collection is complete, use the cleaning function of MindCluster Ascend FaultDiag on each training or inference device to clean the collected raw logs and metric data, filtering and extracting valid information.|For details, see [Log Cleaning and Dumping](./user_guide/06_cleaning_and_dumping_logs.md).|
|Cleaning result dDumping|After log cleaning is complete, dump and aggregate the cleaning results from each training or inference device to a single training device or general-purpose device, and store them according to a predefined structure.|For details, see [Log Cleaning and Dumping](./user_guide/06_cleaning_and_dumping_logs.md).|
|Fault diagnosis|Based on the aggregated cleaning results, use the diagnosis function of MindCluster Ascend FaultDiag to analyze the root cause of the training or inference task failure or abnormality.|For details, see [Fault Diagnosis](./user_guide/07_diagnosing_faults.md).|

> **NOTE**
>
>In the preceding usage process, log collection and cleaning result dumping are not functions provided by MindCluster Ascend FaultDiag. This document only provides operation guidance for them.

## Application Scenarios and Solutions<a name="ZH-CN_TOPIC_0000001592414013"></a>

The intelligent fault diagnosis feature addresses the challenges of fault locating and demarcation in cluster training and inference tasks. Given the large volume of cluster logs, the complexity of AI full-stack log analysis, and the potential need for cross-domain analysis spanning compute, network, and storage, issues encountered by users are typically difficult to locate, time-consuming, and require collaboration across multiple domains.

This feature can effectively improve problem‑locating capabilities for training and inference tasks, increasing user adoption and promoting the expansion of the product ecosystem.

Specifically, this feature provides log cleaning and fault diagnosis functions for each device in training and inference clusters. Users need to complete log collection and cleaning, then dump the cleaned information files to a specific path for diagnosis, enabling rapid fault demarcation by analyzing the diagnostic results. It also supports user-defined fault entities or ERROR message masking in CANN App logs.

Based on actual service requirements, the following two scenarios are currently available.

|Scenario |User|Task Type|Characteristics|
|--|--|--|--|
|[Full-Process Application Scenario](#section1511514596338)|Enterprises, governments, public institutions, etc. (with AI cluster O&M platform capabilities)|Training and inference tasks|The collection content is relatively complex, due to its dependency on training and inference logs, CANN, host-side resources, and hardware-related data. This scenario is suitable for AI cluster O&M platform users performing complex task diagnosis.|
|[Basic Application Scenario](#section587911381388)|Individuals|Training and inference tasks|The collection content and method are simple, because only the training and inference logs, CANN logs are required. This scenario is suitable for individual users performing basic task diagnosis.|

**Full-Process Application Scenario<a name="section1511514596338"></a>**

**In training scenarios**, multiple types of logs and metric data, such as training logs, host-side resource logs, NPU logs, and hardware logs, are required.

**In inference scenarios**, inference task logs, CANN App logs, device-side logs, and MindIE component logs are required.

Some metric data needs to be collected through additional operations. Therefore, the full-process application scenario is recommended for users with AI cluster O&M platform capabilities to integrate and use.

As shown in the following figure, install MindCluster Ascend FaultDiag on all training or inference devices. After the training or inference task ends, each device needs to collect all the aforementioned logs and metric data information, then use the cleaning function of MindCluster Ascend FaultDiag to filter and extract valid information, and finally dump the original logs, metric information, and cleaning results from all devices to the AI cluster O&M platform. The platform uses the diagnostic function of MindCluster Ascend FaultDiag to analyze the root cause of the fault. At the same time, it supports users in customizing fault entities or masking ERROR messages in CANN App logs.

**Figure 1** Full-process application scenario solution<a name="fig21091944182810"></a>
![](../figures/faultdiag/full-process%20application%20solution.png)

In the full-process application scenario, the data sources and purposes corresponding to the log and metric data information to be collected are shown in [Table 1](#table7211162233417).

**Table 1** Training and inference task logs and metric information

<a name="table7211162233417"></a>

|Data Category|Log Description|Data Source|Data Purpose|
|--|--|--|--|
|Training task logs|Logs generated by the model training process|Training task|Used for fault event analysis|
|NPU port check before and after training|Before and after executing training tasks, use `hccn_tool` to check the port information of each NPU.|Training task|Used for fault event analysis|
|Host-side resource information|Metrics such as NPU status monitoring, including the CPU usage (`%CPU`) and physical memory (`RES`) used by the main training process of each NPU.|Training task|Used for device resource analysis|
|NPU port resource information|Metrics such as NPU port packet sending and receiving statistics|Training task|Used for network congestion analysis|
|OS logs|Linux system logs|Training task|Used for fault event analysis|
|MindCluster component logs|Logs of SuperPoD devices, AI servers, and components collected by Ascend Device Plugin, NodeD, Ascend Docker Runtime, NPU Exporter, and Volcano.|Training task|Used for fault event analysis|
|Inference task logs|Logs generated by the inference task process|Inference task|Used for fault event analysis|
|NPU device operation logs|Device-side logs and files, including slog logs, hisi_logs, etc.|Training and inference tasks|Used for fault event analysis|
|CANN App logs|Operation logs generated by CANN.|Training and inference tasks|Used for root cause node analysis and fault event analysis|
|MindIE component logs|Logs generated by MindIE components, including MindIE Server, MindIE LLM, MindIE SD, MindIE RT, MindIE Torch, MindIE MS, MindIE Benchmark, and MindIE Client.|Inference task|Used for fault event analysis|
|AMCT logs|Logs generated by the AMCT model compression process|Model compression|Used for AMCT tool fault event analysis|
|MindIE Pod console logs|Logs of MindIE Pod console.|Inference task|Used for root cause node analysis|
|MindIO component logs|Logs generated by MindIO components.|Training and inference tasks|Used for fault event analysis|

> [!NOTE]
> For details about the collection methods for all logs and metric data, see [Log Collection](./user_guide/03_collecting_logs.md).

**Basic Application Scenario<a name="section587911381388"></a>**

Considering the application requirements of different users, a basic application scenario is supported that relies solely on training or inference logs and CANN App logs for diagnosis. These logs are generated by training or inference tasks and require no additional collection.

As shown in the following figure, install MindCluster Ascend FaultDiag on all training or inference devices. After a training or inference task is completed, each device needs to collect at least the training or inference logs and CANN App logs. Then, the cleaning function of MindCluster Ascend FaultDiag is used to filter and extract valid information. Finally, the original logs and cleaning results from all devices are dumped to the same general-purpose device, and users use the diagnosis function of MindCluster Ascend FaultDiag to analyze the root cause of the fault. At the same time, it supports users in customizing fault entities or masking ERROR messages in CANN App logs.

**Figure 2**  Basic application scenario solution<a name="fig3750917713"></a>
![](../figures/faultdiag/basic%20application%20solution.png)

The data sources and purposes corresponding to the logs and metric data to be collected are shown in the following table.

For the collection methods of all logs and metric data, see [Log Collection](./user_guide/03_collecting_logs.md).

**Table 2** Training and inference task log and metric information

|Data Category|Log Description|Data Source|Data Purpose|
|--|--|--|--|
|Training task logs|Logs generated by the training task process|Training task|Used for fault event analysis|
|CANN App logs|Operation logs generated by CANN |Training and inference tasks|Used for root cause node analysis and fault event analysis|
|Inference task logs|Logs generated by the inference task process|Inference task|Used for fault event analysis|
