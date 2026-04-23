# NPU Exporter主页<a name="ZH-CN_TOPIC_0000002479386854"></a>

## 功能说明<a name="zh-cn_topic_0000001497524785_section1617874274411"></a>

NPU Exporter的基本信息页面。

## URL<a name="zh-cn_topic_0000001497524785_section103113034014"></a>

`GET http://ip:port/`

>[!NOTE]
>
>- IP：在容器化部署场景中，使用容器IP；在二进制部署场景中，使用启动NPU Exporter的IP入参。如果IP为ipv6格式，访问格式调整为：http://[IP]:port/。
>- port：默认为8082，部署时如有修改，使用实际部署时使用的port入参。

## 请求参数<a name="zh-cn_topic_0000001497524785_section162719122175"></a>

无

## 响应说明<a name="zh-cn_topic_0000001497524785_section1433551894112"></a>

返回一个简单的html页面。

```html
<html>
   <head><title>NPU-Exporter</title></head>
   <body>
   <h1 align="center">NPU-Exporter</h1>
   <p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is http://ip:8082/metrics: <a href="./metrics">Metrics</a></p>
   </body>
   </html>
```
