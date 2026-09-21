# NPU Exporter Home Page<a name="ZH-CN_TOPIC_0000002479386854"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:21.039Z pushedAt=2026-06-09T02:05:50.667Z -->

## Function Description<a name="zh-cn_topic_0000001497524785_section1617874274411"></a>

Basic information page of NPU Exporter.

## URL<a name="zh-cn_topic_0000001497524785_section103113034014"></a>

`GET http://ip:port/`

>[!NOTE]
>
>- `ip`: In image-based deployment scenarios, use the container IP; in binary-based deployment scenarios, use the IP that starts NPU Exporter. If the IP is in IPv6 format, adjust the access format to: `http://[IP]:port/`.
>- `port`: Defaults to 8082. If modified during deployment, use the actual port.

## Request Parameters<a name="zh-cn_topic_0000001497524785_section162719122175"></a>

None

## Response Description<a name="zh-cn_topic_0000001497524785_section1433551894112"></a>

Returns a simple HTML page.

```html
<html>
   <head><title>NPU-Exporter</title></head>
   <body>
   <h1 align="center">NPU-Exporter</h1>
   <p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is http://ip:8082/metrics: <a href="./metrics">Metrics</a></p>
   </body>
   </html>
```
