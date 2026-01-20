# 故障事件诊断接口<a name="ZH-CN_TOPIC_0000002287924160"></a>

**接口原型<a name="section1652101232010"></a>**

```
diag_knowledge_graph(input_log_list: list)
```

**功能说明<a name="section15129533192"></a>**

诊断清洗后的故障事件，输出诊断报告。

**输入参数说明<a name="section12416184719242"></a>**

|参数名|是否必选|参数类型|说明|
|--|--|--|--|
|*input_log_list*|是|List|使用<a href="./fault_event_cleaning.md">故障事件清洗接口</a>获得的各个节点results结果数据。<p>注：若节点参数中出现"source": "ccae"时，可能导致该节点清洗结果不准确。</p>|


**返回参数说明<a name="section14225151742812"></a>**

|参数名|参数类型|说明|
|--|--|--|
|results|List|整合后的故障事件诊断报告。|
|err_msg_list|List|接口执行过程中产生的错误信息。|



