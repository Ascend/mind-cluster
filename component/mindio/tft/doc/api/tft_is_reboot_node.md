# tft\_is\_reboot\_node<a name="ZH-CN_TOPIC_0000002511721142"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

MindIO ARF功能流程中，判断当前进程是否为故障后重新拉起的节点，仅支持在tft\_start\_processor接口调用成功后立即调用，且仅支持调用一次。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_is_reboot_node()
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

无

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

bool值，表示是否为故障后重新拉起的节点。

