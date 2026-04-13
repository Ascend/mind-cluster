# 销毁vNPU<a name="ZH-CN_TOPIC_0000002479386366"></a>

销毁指定vNPU。

**命令格式<a name="section397122431219"></a>**

**npu-smi set -t destroy-vnpu -i** _id_ **-c** _chip\_id_ **-v** _vnpu\_id_

**使用示例<a name="section198531444111215"></a>**

执行**npu-smi set -t destroy-vnpu -i  0  -c 0 -v 103**销毁设备0的芯片编号0中编号为103的vNPU设备。显示以下信息表示销毁成功。

```ColdFusion
       Status : OK
       Message : Destroy vnpu 103 success
```

>[!NOTE] 
>在销毁指定vNPU之前，请确保此设备未被使用。
