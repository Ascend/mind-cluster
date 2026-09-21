# Health Probe<a name="ZH-CN_TOPIC_healthz_clusterd"></a>

ClusterD starts an HTTP health probe service within the component, which is used by the K8s livenessProbe mechanism to detect the liveness status of the component. The probe service is decoupled from ClusterD's business gRPC service and uses an independent port.

**Table 1** Health probe interface

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default port | 11253 |
| Protocol | HTTP (HTTPS when the --tls-cert-file and --tls-private-key-file parameters are correctly configured) |

**Table 2**  Response description

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | The component is running normally | The response body is `ok` |
| 404 Not Found | The request path is not `/` | The probe only responds to the root path |
| 405 Method Not Allowed | The request method is not GET | K8s livenessProbe uses GET by default, so this is not normally triggered |
| 503 Service Unavailable | A custom health check callback is registered and the check fails | The response body contains the specific error information |

**K8s LivenessProbe Configuration Example**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11253
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

For details about the probe startup parameters, see [ClusterD Startup Parameters](../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md#parameter-description).
