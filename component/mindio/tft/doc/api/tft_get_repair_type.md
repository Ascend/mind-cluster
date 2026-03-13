# tft\_get\_repair\_type<a name="ZH-CN_TOPIC_0000002543281103"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

提供给MindSpore调用，用于在stop/clean/repair阶段的回调中查询修复类型。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_get_repair_type()
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

无

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

str类型。

-   retry：执行UCE修复。
-   recover：执行ARF修复。
-   dump：执行临终遗言。
-   unknow：未找到修复类型。

