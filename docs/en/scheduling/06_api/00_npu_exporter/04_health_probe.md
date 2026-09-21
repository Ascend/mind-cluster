# Health Probe<a name="ZH-CN_TOPIC_healthz_npu_exporter"></a>

NPU Exporter starts an HTTP health probe service within the component, which is used by the K8s livenessProbe mechanism to detect the component's liveness status. The probe service is fully decoupled from NPU Exporter's business HTTP service and uses an independent port.

**Table 1** Health probe interface

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default port | 11256 |
| Protocol | HTTP (HTTPS when the `--tls-cert-file` and `--tls-private-key-file` parameters are correctly configured) |

**Table 2** Response description

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | The component is running normally | The response body is `ok` |
| 404 Not Found | The request path is not `/` | The probe only responds to the root path |
| 405 Method Not Allowed | The request method is not GET | K8s livenessProbe uses GET by default, so this is normally not triggered |
| 503 Service Unavailable | A custom health check callback is registered and the check fails | The response body contains the specific error information |

**K8s LivenessProbe Configuration Example**

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

For details about the probe startup parameters, see [NPU Exporter Startup Parameters](../../05_developer_guide/00_installation_deployment/00_manual_installation/03_npu_exporter.md#parameter-description).
