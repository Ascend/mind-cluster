# Destroying a vNPU<a name="ZH-CN_TOPIC_0000002479386366"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:21:20.705Z pushedAt=2026-06-30T12:23:24.372Z -->

This section describes how to destroy a specified vNPU.

**Command<a name="section397122431219"></a>**

**npu-smi set -t destroy-vnpu -i** _id_ **-c** _chip\_id_ **-v** _vnpu\_id_

**Usage Example<a name="section198531444111215"></a>**

Run **npu-smi set -t destroy-vnpu -i 0 -c 0 -v 103** to destroy vNPU 103 of chip 0 on device 0. If the following information is displayed, the vNPU is successfully destroyed:

```ColdFusion
       Status : OK
       Message : Destroy vnpu 103 success
```

>[!NOTE]
>Before destroying the specified vNPU, ensure that the device is not in use.
