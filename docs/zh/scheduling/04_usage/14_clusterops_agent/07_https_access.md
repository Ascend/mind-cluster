# HTTPS 安全接入

本功能包含双监听、证书初始化、共享 token、登录认证、安全响应头、部署与测试，为 `/ui` 参考界面、`/api/v1/*` 与 `/chat/stream` 提供 HTTPS 接入与共享 token 认证。

访问 `https://<node-ip>:30443/ui/`：未携带 token 的页面导航返回 200 与登录页；`/ui/login` 为匿名登录页，登录后业务接口使用 `POST /api/v1/auth/check` 校验。容器 HTTPS 端口 9743，独立 Service 为 `agent-core-ui`。HTTP 9700 保留 CLI 诊断，不需要 WebUI token，不能作为网页入口。

通用模板使用 `mindx-dl`。

- `agent-core-webui-token`：启动内存快照，仅不存在时生成 256-bit 随机值；轮换需 patch Secret 并重启 Pod。原值匹配，不去空白、不改大小写和连字符。
- `agent-core-webui-tls`：通过 K8s API 初始化，RSA 4096、SHA256、全部节点 InternalIP、Service DNS、serverAuth。检查 PEM、密钥配对、有效期与用途。
- 证书由组件生成并独占管理，不做来源判定、不设计人工替换流程；过期、SAN 漂移或临期自动重签；材料非法（PEM、配对、EKU、尚未生效）时关闭 HTTPS。
- 自签有效期 365 天；自签证书临期不足 30 天、过期或 SAN 漂移时自动重签。`/ui` 无 token 返回 200 + 登录页，`/ui/login`、登录静态资源与参考界面静态资源公开，业务接口需 token。TLS Secret 的 patch 权限按名称限制，token bootstrap 不更新已有值。
- 额外 SAN 可用 `WEBUI_TLS_SANS`（逗号分隔 IP/DNS）补充；命名空间与节点 IP 由 Downward API 的 `POD_NAMESPACE` / `HOST_IP` 注入。bootstrap 或 HTTPS 绑定失败时只告警并关闭 HTTPS，本次进程内不重试；恢复需重启 Pod。
- 路径存在但 token 无效时返回 **200**，响应体为 `{"ok": false, "error": {"code": "UNAUTHENTICATED"}}`，并携带 `X-WebUI-Auth: unauthenticated` 头；不存在的路径在认证之前判定，一律返回 404。
- HTTPS 请求共享进程内窗口限额（`WEBUI_RATELIMIT_MAX_REQUESTS` 默认 600 / `WEBUI_RATELIMIT_WINDOW_SECONDS` 默认 60 秒）；受保护路径超限返回 429，登录页与静态资源仍可访问。
- TLS 最低 1.2，同一 app 只初始化一次；证书、API、RBAC 或 HTTPS 监听异常不停止 HTTP。
- HTTPS 响应携带 HSTS，所有响应携带 nosniff、DENY、CSP 和 no-store。证书和 token 不热加载、不后台轮换。

浏览器共享 WebUI token 与 LLM API Key 是不同凭据；LLM 配置沿用主仓服务端机制。故障演练必须先确认独立命名空间。
