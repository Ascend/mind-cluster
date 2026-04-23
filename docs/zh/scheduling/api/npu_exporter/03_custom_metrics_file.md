# 自定义指标文件<a name="ZH-CN_TOPIC_0000002501343480"></a>

## 字段说明<a name="section42791696421"></a>

自定义指标文件中的字段说明如[表1](#ZH-CN_TOPIC_0000002501343480_table5395205714441)所示。通过自定义指标文件可开发自定义指标，详情请参见[自定义指标开发](../../appendix.md#自定义指标开发)。

**表 1**  自定义指标文件字段说明

<a name="ZH-CN_TOPIC_0000002501343480_table5395205714441"></a>

|字段名称|类型|说明|
|--|--|--|
|version|string|固定值：1.0。|
|name|string|指标名称。不能为空，长度不能超过128。|
|desc|string|指标的详情介绍。不能为空，长度不能超过1024。|
|timestamp|timestamp|指标的更新时间戳，单位为us。|
|data_list|list|非空数组，长度不能超过128。|
|-value|float|指标的值。|
|-label|json|指标的标签，JSON的key和value都必须是字符串类型，JSON的子元素个数不能超过10个。|

## 约束说明<a name="section342514924413"></a>

- 不支持软链接。
- 不支持指定一个目录。
- 支持指定的文件路径不能超过10个，多个文件之间以逗号分隔。
- 当指定多个文件时，若其中部分文件不存在或为空文件，合并等待1分钟后仍不满足条件，则取消对应文件的指标采集。
- 文件中字段的格式需满足[表1](#ZH-CN_TOPIC_0000002501343480_table5395205714441)的要求。格式不正确会取消对应文件的指标采集。
- 自动获取指标的label时，以data\_list中的第一条数据为准。
- 自定义指标文件的属组需为npu-exporter进程属组，且具有读权限，不具有任何执行权限。
- 程序运行中不支持修改文件的name、desc和version。
- 程序运行中若修改label名称，程序不会感知，仍按初始化时的label名称上报。
- 自定义指标文件大小限制为100KB。
- 容器场景部署时，需确保指标文件正确挂载到容器中。
- 指标采集过程中，若某批次数据不满足约束，则忽略本批次数据，按缓存数据进行上报。

## 指标文件样例

```json
{
  "version": "1.0",
  "name": "hccs_bandwidth",
  "desc": "hccs bandwidth info, unit is 'MB/s'.",
  "timestamp": 1766456419845127,
  "data_list": [
    {
      "value": 190.02,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "0",
        "direction": "in",
        "path": "P0->P1"
      }
    },
    {
      "value": 143.09,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "1",
        "direction": "in",
        "path": "P2->P1"
      }
    }
  ]
}
```

## Telegraf中自定义指标上报样例

```ColdFusion
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=0,numa=2,path=P0->P1 hccs_bandwidth=190.02 1766456419845127000
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=1,numa=2,path=P2->P1 hccs_bandwidth=143.09 1766456419845127000
```
