# 健康探针<a name="ZH-CN_TOPIC_healthz_npu_exporter"></a>

NPU Exporter启动组件内的HTTP健康探针服务，用于K8s livenessProbe机制探测组件存活状态。探针服务与NPU Exporter的业务HTTP服务完全解耦，使用独立端口。

**表 1**  健康探针接口

| 项目 | 说明 |
|------|------|
| 路径 | `/` |
| 方法 | GET |
| 默认端口 | 11256 |
| 协议 | HTTP（正确配置--tls-cert-file和--tls-private-key-file参数时为HTTPS） |

**表 2**  响应说明

| 状态码 | 触发条件 | 说明 |
|--------|---------|------|
| 200 OK | 组件正常运行 | 响应体为 `ok` |
| 404 Not Found | 请求路径非 `/` | 探针只响应根路径 |
| 405 Method Not Allowed | 请求方法非 GET | K8s livenessProbe默认使用GET，正常不会触发 |
| 503 Service Unavailable | 注册了自定义健康检查回调且检查失败 | 响应体包含具体错误信息 |

**K8s livenessProbe配置示例：**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11256
    scheme: HTTP
  initialDelaySeconds: 20
  periodSeconds: 15
  timeoutSeconds: 5
  failureThreshold: 3
```

探针启动参数说明详见[NPU Exporter启动参数](../../05_developer_guide/00_installation_deployment/00_manual_installation/03_npu_exporter.md#参数说明)。
