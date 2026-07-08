# flush接口<a name="ZH-CN_TOPIC_0000002468994002"></a>

## 接口功能<a name="zh-cn_topic_0000002112429502_section107101141937"></a>

等待后台异步刷盘任务全部执行成功。

## 接口格式<a name="zh-cn_topic_0000002112429502_section13362162011417"></a>

```
mindio_acp.flush()
```

## 接口参数<a name="zh-cn_topic_0000002112429502_section171201830749"></a>

无

## 使用样例<a name="zh-cn_topic_0000002112429502_section81115380412"></a>

```
>>> # flush all data to disk
>>> mindio_acp.flush()
```

## 返回值<a name="zh-cn_topic_0000002112429502_section17538071458"></a>

-   0：刷盘成功。
-   1：刷盘失败。
