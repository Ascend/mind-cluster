# 业务流清洗接口<a name="ZH-CN_TOPIC_0000002256436654"></a>

**接口原型<a name="section1652101232010"></a>**

```
parse_fault_type(input_log_list: list)
```

**功能说明<a name="section15129533192"></a>**

使用MindCluster Ascend FaultDiag组件定义的故障模式清洗业务流日志。

**输入参数说明<a name="section12416184719242"></a>**

|参数名|是否必选|参数类型|说明|
|--|--|--|--|
|*input_log_list*|是|List|用户输入的业务流。|


**返回参数说明<a name="section14225151742812"></a>**

|参数名|参数类型|说明|
|--|--|--|
|results|List|清洗整合的结果。|
|err_msg_list|List|接口执行过程中产生的错误信息。|

