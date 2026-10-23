# Chip Fault Recovery<a name="ZH-CN_TOPIC_0000002479227136"></a>

Chip fault recovery refers to the offline hot reset function of Ascend Device Plugin. When certain hardware faults occur on chip resources in a cluster, the chips can be automatically restored to a healthy state through a hot reset.

## Before You Start<a name="section_precondition"></a>

**Prerequisites<a name="section_prerequisite"></a>**

- The offline hot reset feature is usually used together with the **full-NPU scheduling** and **rescheduling** features. For details about how to use them, see [Full-NPU Scheduling](../03_basic_scheduling/03_full_npu_scheduling.md) and [Rescheduling](../03_basic_scheduling/05_rescheduling.md).
- To enable the offline hot reset feature, Ascend Device Plugin must be installed and the startup parameter `-hotReset` must be set to `0` or `2` (`-1` by default, which does not support the offline hot reset feature). For details, see [Manually Installing Ascend Device Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md).
- Only chip faults support recovery through a hot reset.

## Supported Product Forms

- Atlas 800 training server (model 9000) (fully populated with NPUs)
- Atlas 800 training server (model 9010) (fully populated with NPUs)
- Atlas 900T PoD Lite
- Atlas 900 PoD (model 9000)
- Atlas 800T A2
- Atlas 900 A2 PoD
- Atlas 900 A3 SuperPoD
- Atlas 800T A3
- Atlas 850E
- Atlas 850
- Atlas 650E
- Atlas 650
- Atlas 950 SuperPoD
- Atlas 350 accelerator card
- Atlas 300I Pro inference card
- Atlas 300V video analysis card
- Atlas 300V Pro video analysis card
- Atlas 300I Duo inference card
- Atlas 300I inference card (model 3000) (full card)
- Atlas 300I inference card (model 3010)
- Atlas 800I A2
- A200I A2 Box
- Atlas 800I A3

## Working Principle

The fault levels that support chip reset are as follows:

- RestartRequest
- RestartBusiness
- FreeRestartNPU
- RestartNPU

For details about fault levels, see [Fault Levels](../../06_api/02_ascend_device_plugin.md#custom-chip-faults).

### Offline Hot Reset Process

1. After Ascend Device Plugin is started, it obtains the chip health status through the DCMI of the driver. When a chip is faulty, the fault information can be obtained from the driver interface. For details, see [Chip Faults](../04_fault_detection_and_diagnosis/02_fault_types/02_chip_faults.md).

2. Ascend Device Plugin determines the fault level of the obtained fault information, and then determines whether a hot reset is required for the chip.

3. Some chip faults are theoretically recovered after the job ends. Therefore, Ascend Device Plugin waits for 60 seconds and then checks whether the faulty chip has returned to a healthy state.

4. If the faulty chip is still not in a healthy state after 60 seconds, Ascend Device Plugin checks, based on the server type, whether the faulty chip and the associated chips that need to be reset together are idle.

5. If the faulty chip and the associated chips that need to be reset together are both idle, Ascend Device Plugin calls the driver interface to perform a hot reset on them.

>[!NOTE]
>
>- For the Atlas 300I Duo inference card, only card-based reset is supported, that is, both chips are reset at the same time.
>
>- If a chip is faulty and has been scheduled and mounted to a job container, to make it idle, you need to configure a job rescheduling policy to reschedule the faulty Pod to another idle NPU. For details, see [Rescheduling](../03_basic_scheduling/05_rescheduling.md).
>
>- An Atlas 800I A2 has the following two fault recovery methods. Only one fault recovery method can be used on one Atlas 800I A2. The cluster scheduling components automatically identify which fault recovery method to use.
>
>    - Method 1: If no HCCS ring exists on the device, when an NPU is faulty during inference job execution, Ascend Device Plugin waits until the NPU is idle and then performs a reset on the NPU.
>
>    - Method 2: If an HCCS ring exists on the device, when one or more NPUs on the server are faulty during inference job execution, Ascend Device Plugin waits until all NPUs on the ring are idle and then resets all NPUs on the ring at a time.
>
>- When a hot reset is performed on an Atlas 900 A3 SuperPoD, an Atlas 800T A3, or an Atlas 800I A3, the NPU module where the specified chip resides and the NPU module that has a NIC port mutual assistance relationship with it are reset.
>
>- On an Atlas 900 A3 SuperPoD, an Atlas 800T A3, or an Atlas 800I A3, if Ascend Device Plugin fails to perform a hot reset by calling the in-band hot reset interface of the driver, it attempts to call the out-of-band hot reset interface to perform a hot reset again.
>
>- Hot reset recovery cannot cover all faults. Some faults may fail to be recovered, for example, card loss caused by a fault and a hang of the device OS.
