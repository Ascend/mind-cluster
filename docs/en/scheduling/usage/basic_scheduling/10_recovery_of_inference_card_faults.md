# Inference Card Fault Recovery<a name="ZH-CN_TOPIC_0000002479227136"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:22:02.104Z pushedAt=2026-06-30T12:23:24.384Z -->

The inference card fault recovery feature must be used in conjunction with the full NPU scheduling feature. To enable the inference card fault recovery feature, simply set the Ascend Device Plugin startup parameter "-hotReset" to "0" or "2" (the default value is "-1", which does not support the fault recovery function). For specific usage, refer to [Full NPU Scheduling or Static vNPU Scheduling (Inference)](./04_full_npu_scheduling_and_static_vnpu_scheduling_inference.md).

When this feature is enabled on the Atlas 800I A2 inference server and A200I A2 Box heterogeneous component, only single-server single-processor jobs can be delivered. Distributed jobs are not supported. Additionally, the [infer-vcjob-910-hotreset.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-910-hotreset.yaml) sample must be used separately to submit jobs.

>[!NOTE]
>There are two fault recovery modes for the Atlas 800I A2 inference server. One Atlas 800I A2 inference server can use only one fault recovery mode, which is automatically identified by cluster scheduling components.
>
>- Method 1: If no HCCS ring exists on the server, when an NPU is faulty during inference, Ascend Device Plugin waits until the NPU is idle and resets it.
>- Method 2: If an HCCS ring exists on the server, when one or more NPUs are faulty during inference, Ascend Device Plugin waits until all NPUs on the ring are idle and resets them at a time.
