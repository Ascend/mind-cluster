# 配置推理任务的弹性扩缩容<a name="ZH-CN_TOPIC_0000002479226430"></a>

MindIE Motor推理任务中，用户可通过配置Job级别弹性扩缩容功能，在发生硬件或软件故障且当前资源不满足所有实例拉起时，降低运行的实例数量，尽量保证推理任务继续运行。在故障恢复或新的硬件加入时，等待拉起的Job实例会重新被调度。

**使用约束<a name="zh-cn_topic_0000002356673977_section270417201799"></a>**

当前仅支持MindIE Motor推理任务使用本功能。

**支持的产品型号<a name="zh-cn_topic_0000002356673977_section618313391397"></a>**

- Atlas 800I A2 推理服务器
- Atlas 800I A3 超节点服务器

**原理说明<a name="zh-cn_topic_0000002356673977_section1445672111019"></a>**

**图 1**  弹性扩缩容原理<a name="zh-cn_topic_0000002356673977_fig685814101278"></a>  
![](../../../figures/scheduling/弹性扩缩容原理.png "弹性扩缩容原理")

1. 用户配置多个Job属于同一个推理任务，并将Job分成多个组别，并配置一个扩缩容规则（scaling-rule）。
2. 弹性扩缩容规则以ConfigMap形式部署在集群中，不同类别的实例对应scaling-rule中的不同group。例如可以将所有的Prefill实例分类为group0，所有的Decode实例分类为group1。
3. 配置重调度的场景下，当发生硬件或软件故障时，Ascend Device Plugin和NodeD对故障进行上报，Volcano删除该实例下的所有Pod。
4. ClusterD将global-ranktable发送给MindIE Controller，关于global-ranktable的说明请参见[SubscribeRankTable](../../api/clusterd.md#subscriberanktable)中"global-ranktable文件说明"表。
5. MindIE Controller根据global-ranktable确定需要退出的实例，通知容器中的进程非0退出。
6. Volcano-Scheduler感知到Pod异常后，将实例的所有Pod删除。
7. Ascend Operator感知到Pod被删除后，会收集当前MindIE Motor对应scaling-rule下的所有实例运行情况。
8. Ascend Operator根据scaling-rule确认当前实例是否需要创建Pod。
9. 如果可以创建Pod，待Pod创建完成后，由调度器完成调度或处于Pending状态等待调度。
10. 处于Pending状态的Pod待资源充足时，自动完成调度。
11. 如果当前不可以创建Pod，则等待其他实例成功运行后再进行创建。

**创建扩缩容规则ConfigMap<a name="zh-cn_topic_0000002356673977_section476902931213"></a>**

用户需要设置特定的扩缩容规则，将其以ConfigMap的形式部署到k8s集群中，示例如下

```Yaml
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                 # 一个可以正常运行的任务配比状态            {
              "group_name": "group0",      # 用户自行设置
              "group_num": "2",            # 用户自行设置，要求从上往下不能增加
              "server_num_per_group": "2"  # 用户自行设置，要求相同的group_name，该值保持不变
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                 # 另一个可以正常运行的任务配比状态
            {
              "group_name": "group0",
              "group_num": "1",
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

>[!NOTE] 
>
>- 例如当前运行的group\_name为group0和group1的Job都为0个，则会选择索引为1的group\_list，即group0和group1都需要运行1个，那么此时group0或group1对应的Job就会创建对应的Pod，然后等待调度。
>- 如果当前group\_name为group0的Job运行了1个，group\_name为group1的Job运行了0个，此时只会为group\_name为group1的Job创建Pod，group\_name为group0的Job会等待group\_name为group1的Job成功运行后才创建Pod。

在以上ConfigMap中，可以修改的字段说明如下表所示。

**表 2**  参数说明

<a name="zh-cn_topic_0000002356673977_zh-cn_topic_0000002193288232_table985012534578"></a>

|参数|说明|取值|是否必填|
|--|--|--|--|
|metadata.name|承载scaling-rule的ConfigMap的名称。<p>用户可以自行设置，Job的label“mind-cluster/scaling-rule”的值需要与之对应，表明该Job受该scaling-rule控制。</p>|string|是|
|metadata.namespace|承载scaling-rule的ConfigMap的命名空间。<p>用户可以自行设置，但需要与推理任务保持一致。如果不设置，那么命名空间默认是"default"。</p>|string|否|
|group_name|group组名称。<p>Job的label "mind-cluster/group-name"，需要与之对应，表明该Job属于该group组。</p>|string|是|
|group_num|group组目标Job数量。<p>若当前运行中的该group下的Job数量未达该目标，会尝试拉起该group下的一个Job。</p>|string|是|
|server_num_per_group|group组目标Job的副本数。<p>不同group_list中相同group_name下，该值需保持一致。</p>|string|是|

**修改扩缩容规则<a name="zh-cn_topic_0000002356673977_section1769411616405"></a>**

如果此时已经运行了2个group0和1个group1的Job，用户需要增加运行一个group0的Job，那么用户需提前修改扩缩容模板，再下发新的任务，修改示例如下：

```Yaml
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                    # 新增一个条目到elastic_scaling_list中             
            {
              "group_name": "group0",     
              "group_num": "3",              # 修改group0的group_num
              "server_num_per_group": "2"  
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                             
            {
              "group_name": "group0",      
              "group_num": "2",            
              "server_num_per_group": "2"  
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                 # 另一个可以正常运行的任务配比状态
            {
              "group_name": "group0",
              "group_num": "1",
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

如果在任务正常运行的情况下，需要减少其中一个group0的Job，用户需要修改模板，再删除任务，修改示例如下。

```Yaml
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                 # 删除了一个group_list
            {
              "group_name": "group0",
              "group_num": "1",           # group0目标group_num为"1"
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

**准备任务YAML<a name="zh-cn_topic_0000002356673977_zh-cn_topic_0000002098814658_section463203519254"></a>**

在任务YAML中，修改或新增以下字段，开启Job级别弹性扩缩容。

```Yaml
... 
metadata:  
   labels:  
     ...  
     fault-scheduling: "force"
     fault-retry-times: "100000000"    # 处理业务面故障，必须配置业务面无条件重试的次数
     jobID: mindie-xxx      # 由用户自行定义
     app: mindeie-ms-server
     mind-cluster/scaling-rule: scaling-rule   # 需与扩缩容规则ConfigMap的名称保持一致
     mind-cluster/group-name: group0           # 需与扩缩容规则ConfigMap中的group_name取值保持一致
spec:
  schedulerName: volcano      # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
  runPolicy:
    backoffLimit: 3         # 任务重调度次数
...
```
