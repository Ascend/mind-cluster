# 根因节点诊断接口<a name="ZH-CN_TOPIC_0000002288027410"></a>

**接口原型<a name="section1652101232010"></a>**

```
diag_root_cluster(input_log_list: list)
```

**功能说明<a name="section15129533192"></a>**

使用MindCluster Ascend FaultDiag组件诊断集群中发生错误的根因节点信息。

**输入参数说明<a name="section12416184719242"></a>**

|参数名|是否必选|参数类型|说明|
|--|--|--|--|
|*input_log_list*|是|List|使用<a href="./root_cause_cleaning.md">根因节点清洗接口</a>获得的results结果数据。|


**返回参数说明<a name="section14225151742812"></a>**

|参数名|参数类型|说明|
|--|--|--|
|results|Dictionary|发生错误的根因节点信息。|
|err_msg_list|List|接口执行过程中产生的错误信息。|

