# Infer Operator<a name="ZH-CN_TOPIC_0000002511346797"></a>

## Health Probe<a name="ZH-CN_TOPIC_healthz_operator"></a>

Infer Operator starts the built-in HTTP health probe service in the component, which is used by the K8s livenessProbe mechanism to detect the liveness status of the component.

**Table 1**  Health probe interface

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default port | 11254 |
| Protocol | HTTP (HTTPS when the `--tls-cert-file` and `--tls-private-key-file` parameters are correctly configured) |

**Table 2**  Response description

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | Component is running normally | Response body is `ok` |
| 404 Not Found | Request path is not `/` | The probe only responds to the root path|
| 405 Method Not Allowed | Request method is not GET  | K8s livenessProbe uses GET by default, so this is not normally triggered |
| 503 Service Unavailable | Custom health check callback is registered and the check fails | Response body contains specific error information |

**K8s LivenessProbe Configuration Example**:

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11254
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

For details about the probe parameters, see the [Parameter Description](../05_developer_guide/00_installation_deployment/00_manual_installation/07_infer_operator.md#parameter-description).
