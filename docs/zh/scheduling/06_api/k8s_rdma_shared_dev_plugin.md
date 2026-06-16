# K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_k8s_rdma_shared_dev_plugin"></a>

## 健康探针<a name="ZH-CN_TOPIC_healthz_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin启动组件内的HTTP健康探针服务，用于K8s livenessProbe机制探测组件存活状态。

**表 1**  健康探针接口

| 项目 | 说明 |
|------|------|
| 路径 | `/` |
| 方法 | GET |
| 默认端口 | 11257 |
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
    port: 11257
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

> 探针启动参数说明详见[K8s RDMA Shared Dev Plugin 启动参数](../05_developer_guide/installation_deployment/manual_installation/12_k8s_rdma_shared_dev_plugin.md#参数说明)。
