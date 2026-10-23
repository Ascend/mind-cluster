# Destroying vNPUs<a name="ZH-CN_TOPIC_0000002479386366"></a>

This section describes how to destroy a specified vNPU.

## Command<a name="section397122431219"></a>

**npu-smi set -t destroy-vnpu -i** _id_ **-c** _chip\_id_ **-v** _vnpu\_id_

## Usage Example<a name="section198531444111215"></a>

Run **npu-smi set -t destroy-vnpu -i 0 -c 0 -v 100** to destroy vNPU 100 of chip 0 on device 0. If the following information is displayed, the vNPU is successfully destroyed:

```ColdFusion
       Status : OK
       Message : Destroy vnpu 100 success
```

>[!NOTE]
>Before destroying the specified vNPU, ensure that the device is not in use.
