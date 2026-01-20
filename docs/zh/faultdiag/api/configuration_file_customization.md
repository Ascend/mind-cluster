# 自定义配置文件接口<a name="ZH-CN_TOPIC_0000002447169561"></a>

**接口原型<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>**

```
ascend-fd config 子命令
```

**功能说明<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>**

提供自定义配置文件，包括新增或修改、查询自定义配置信息，同时支持校验自定义故障实体custom-fd-config.json文件。

**参数说明<a name="zh-cn_topic_0000001511538701_section122149111390"></a>**

**表 1**  子命令参数说明

|**参数**|**缩写**|是否必选|**值类型**|**说明**|
|--|--|--|--|--|
|--update|-u|必选。--update、--show和--check三者之间互斥，即只能指定一个且必须指定一个参数。|String|以JSON文件格式新增或修改自定义配置信息。JSON文件的相关参数说明请参见<a href="../user_guide.md#可选自定义配置文件">（可选）自定义配置文件</a>中“参数说明”表。|
|--show|-s|必选。--update、--show和--check三者之间互斥，即只能指定一个且必须指定一个参数。|Bool|查看用户自定义的配置信息。|
|--check|-c|必选。--update、--show和--check三者之间互斥，即只能指定一个且必须指定一个参数。|Bool|校验custom-fd-config.json文件的合法性，主要校验每个自定义配置的字段属性的有效性。|
|--help|-h|可选|-|查询使用说明。|


**返回说明<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>**

例如通过JSON文件，新增自定义配置文件实体。

```
ascend-fd config -u custom-config.json
Updated entity successfully.
```
