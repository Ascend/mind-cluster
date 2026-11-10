# Public Faults<a name="ZH-CN_TOPIC_0000002511426387"></a>

Public faults mainly refer to faults reported by other fault providers (not MindCluster components). Public faults include the following types: NPU faults, node faults, network faults, and storage faults. In addition, silent faults determined by ClusterD based on comprehensive software and hardware fault information also belong to public faults.

>[!NOTE]
>For ClusterD to receive public faults, Ascend Device Plugin must be installed on the target node, and the corresponding `device-info-cm` must be generated.

## Reporting Mechanism<a name="zh-cn_topic_0000002216292813_section64469192378"></a>

After a public fault provider detects a fault, it sends the fault information to ClusterD through a ConfigMap or gRPC. ClusterD summarizes the received information, writes it to `cluster-info-device-cm`, and reports it to ascend-volcano-plugin.

- Obtaining through a ConfigMap: The fault detector writes the fault information into a ConfigMap, and ClusterD obtains the fault information from it. You can inject public faults by calling the ConfigMap API. For details, see [ConfigMap](../../../06_api/04_clusterd/03_public_fault_apis.md#configmap).
- Obtaining through gRPC: The fault detector sends the fault information to ClusterD through a gRPC channel, and ClusterD obtains the fault information from it. You can inject public faults by calling the gRPC API. For details, see [gRPC Interface](../../../06_api/04_clusterd/03_public_fault_apis.md#grpc-interface).

**Figure 1** Public fault reporting<a name="fig72618571585"></a>

![](../../../../figures/scheduling/public-fault-reporting.png)

## Required Components<a name="zh-cn_topic_0000002194466236_section138036504533"></a>

To ensure that the public fault detection function works properly, install the following components.

- Mandatory components: Volcano, Ascend Operator, Ascend Device Plugin, and ClusterD
- Optional component: NodeD

## Supported Fault Handling Types<a name="zh-cn_topic_0000002216292813_section177211923175116"></a>

Job-level rescheduling, Pod-level rescheduling, and process-level rescheduling

>[!NOTE]
> A silent fault is a special type of public faults. Its handling policy is to append the fault to the fault list only without changing the chip health status, so it does not trigger job-level, Pod-level, or process-level rescheduling.

## (Optional) Configuring the Fault Detection Level and Provider<a name="zh-cn_topic_0000002216292813_section1343172016386"></a>

The fault detection feature provides a default fault level and supported fault providers for **public faults**. If you need to change the level and fault provider of public faults, see [Public Faults](../03_configuration/08_public_faults.md#ZH-CN_TOPIC_0000002479386564). Unless otherwise required, do not change them at will.

## Related Operations

- [Configuring Public Faults](../03_configuration/08_public_faults.md)
- [Querying and Verifying Faults](../04_querying_and_verifying_faults.md)
