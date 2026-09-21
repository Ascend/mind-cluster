# Hardware Fault Recovery<a name="ZH-CN_TOPIC_0000002479227136"></a>

<!-- md-trans-meta sourceCommit=f8c727966e9dd4576ec4ccafd8e5d7ef6397cead translatedAt=2026-08-31T01:42:37.499Z pushedAt=2026-08-31T01:43:04.751Z -->

The hardware fault recovery feature leverages the offline hot reset capability of Ascend Device Plugin to automatically restore chips to a healthy state when partial hardware faults occur on chip resources in the cluster.

## Before You Start<a name="section_precondition"></a>

**Prerequisites<a name="section_prerequisite"></a>**

- The offline hot reset feature must be used together with the **full-NPU scheduling** feature. For details, see [Full NPU Scheduling](./03_full_npu_scheduling.md).
- To enable the offline hot reset feature, you only need to set the value of the Ascend Device Plugin startup parameter `-hotReset` to `0` or `2` (the default value is `-1`, which means the offline hot reset feature is not supported). For details about how to enable it, see step 6 in [Manually Installing Ascend Device Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md).

**Offline Hot Reset Process**

1. When using the offline hot reset feature, jobs must be configured with a rescheduling policy. For details about the rescheduling upon faults feature, see [Rescheduling upon Faults](./05_rescheduling_upon_inference_card_faults.md).

2. After a chip fault occurs, the cluster scheduling components first reschedule the Pods running on the faulty chip to other healthy nodes. After all Pods on the faulty chip are scheduled away (that is, the faulty chip becomes idle), Ascend Device Plugin performs a hot reset on the faulty chip.

>[!NOTE]
>
>- The Atlas 800I A2 inference server supports the following two fault recovery modes. An Atlas 800I A2 inference server can use only one fault recovery mode, which is automatically identified by the cluster scheduling components.
>
>    - Mode 1: If no HCCS ring exists on the device, when an NPU fault occurs during inference job execution, Ascend Device Plugin waits until the NPU becomes idle and then resets the NPU.
>
>    - Mode 2: If an HCCS ring exists on the device, when one or more NPU faults occur on the server during inference job execution, Ascend Device Plugin waits until all NPUs on the ring become idle and then resets all NPUs on the ring at once.
>
>- For the Atlas 9000 A3 SuperPoD cluster computing system, the NPU module where the specified chip resides is reset. For the Atlas 900 A3 SuperPoD, Atlas 800T A3 SuperPoD server, Atlas 800I A3 SuperPoD server, and A200T A3 Box8  SuperPoD server, the NPU module where the specified chip resides and the NPU modules that have a network port mutual assistance relationship with it are reset.
>
>- Hot reset recovery cannot cover all faults. Some faults may fail to be recovered, for example, faults that cause device disconnection or device OS hang.
