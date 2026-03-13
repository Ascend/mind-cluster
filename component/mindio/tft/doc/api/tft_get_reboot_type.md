# tft\_get\_reboot\_type<a name="ZH-CN_TOPIC_0000002511721146"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

提供给MindSpore调用，在故障重新拉起节点后，训练框架从mindio\_ttp获取节点重启场景类型，进程启动后仅支持调用一次。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_get_reboot_type()
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

无

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

str类型。

-   arf：代表进程重调度。
-   hot switch：代表亚健康热切。

