# 配置和清除LLM<a name="ZH-CN_TOPIC_00000026faultdiagnosis07"></a>

集群运维Agent可以对接LLM服务，对ascend-fd生成的诊断报告进行智能总结，输出简明根因报告（故障事件、根因节点、建议处置）。LLM配置为可选项，未配置或配置清除后，诊断自动回退确定性诊断报告，不影响诊断主流程。

## 添加LLM配置<a name="sectionfaultdiagnosisllmcreate"></a>

通过`kubectl clusterops --create-llm-config`命令添加LLM配置，支持交互输入、命令行参数和环境变量三种方式。

**交互输入（推荐）**

```shell
kubectl clusterops --create-llm-config
LLM API Key: ********
LLM Base URL:
LLM Model:
secret/llm-secret created
llm-secret updated, effective immediately
```

**命令行参数**

```shell
kubectl clusterops --create-llm-config \
  --base-url https://open.bigmodel.cn/api/paas/v4 \
  --model glm-4-plus
```

**环境变量**

配置`LLM_API_KEY`、`LLM_BASE_URL`、`LLM_MODEL`环境变量后，执行`kubectl clusterops --create-llm-config`命令将直接读取环境变量，不再交互输入。

其中：

- `api-key`为必填项，由大模型服务提供商提供。
- `base-url`和`model`为必填项，由大模型服务提供商指定，可通过`--base-url`、`--model`参数或`LLM_BASE_URL`、`LLM_MODEL`环境变量提供，否则命令会交互输入；留空将报错退出。
- 配置通过k8s Secret `llm-secret`存储，默认命名空间`mindx-dl`，可通过环境变量`ASCEND_CLUSTEROPS_NS`覆盖。

## 清除LLM配置<a name="sectionfaultdiagnosisllmclear"></a>

通过`kubectl clusterops --clear-llm-config`命令清除全部LLM配置（api-key/base-url/model一并删除）：

```shell
kubectl clusterops --clear-llm-config
secret "llm-secret" deleted
LLM config cleared, effective immediately
```

> [!NOTE]
> `--create-llm-config`与`--clear-llm-config`互斥，不能同时使用。

## 实时生效说明<a name="sectionfaultdiagnosisllmeffect"></a>

Agent Core实时监听`llm-secret`（对应RBAC `agent-core-llm`已内置在agent-core.yaml中），因此：

- 添加或更新LLM配置后**立即生效，无需重启Agent Core**。
- 清除LLM配置后，Agent Core自动回退确定性诊断报告，不调用LLM。
- LLM调用失败（配置错误、网络不通、api-key无效等）时，诊断结果会附带`llm_error`提示原因，并自动回退确定性诊断报告，不影响诊断结果返回。
