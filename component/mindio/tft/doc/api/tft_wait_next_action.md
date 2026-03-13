# tft\_wait\_next\_action<a name="ZH-CN_TOPIC_0000002511721126"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

修复期间，训练主线程在装饰器中调用该接口等待从线程完成业务数据修复。

>**说明：** 
>该接口为阻塞接口，在未获取到下一次action前，会一直阻塞。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.framework_ttp.tft_wait_next_action()
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

无

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

-   0：成功
-   1：失败

