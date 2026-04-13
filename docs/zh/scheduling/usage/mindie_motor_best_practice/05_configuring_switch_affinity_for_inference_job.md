# 配置推理任务交换机亲和性

当前仅支持Atlas 800I A2 推理服务器配置交换机亲和性功能。开启此功能，可规避Spine交换机下行流量冲突问题。如需了解该功能的原理，请参见[交换机亲和性调度1.0](../basic_scheduling.md#交换机亲和性调度10)章节。

**前提条件**

已完成[（可选）使用Volcano交换机亲和性调度](../../installation_guide/03_installation.md#可选使用volcano交换机亲和性调度)。

**配置推理任务交换机亲和性**

将交换机亲和性tor-affinity配置为normal-schema，YAML示例如下：

```Yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch        
    app: mindie-ms-server        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test        # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    tor-affinity: normal-schema    # 开启交换机亲和性
    ring-controller.atlas: ascend-910b
```
