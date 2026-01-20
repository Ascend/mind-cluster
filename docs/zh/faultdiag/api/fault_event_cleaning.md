# 故障事件清洗接口<a name="ZH-CN_TOPIC_0000002322636661"></a>

**接口原型<a name="section1652101232010"></a>**

```
parse_knowledge_graph(input_log_list: list, custom_entity: dict = None)
```

**功能说明<a name="section15129533192"></a>**

清洗故障日志。

**输入参数说明<a name="section12416184719242"></a>**

|参数名|是否必选|参数类型|说明|
|--|--|--|--|
|*input_log_list*|是|List|用户输入的故障日志。|
|*custom_entity*|否|Dictionary|用户自行输入的自定义故障实体，此参数为临时使用，将不会落盘至JSON文件中。|


**返回参数说明<a name="section14225151742812"></a>**

|参数名|参数类型|说明|
|--|--|--|
|results|List|清洗整合的结果。|
|err_msg_list|List|接口执行过程中产生的错误信息。|

