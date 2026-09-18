# 特性说明<a name="ZH-CN_TOPIC_00000026faultdiagnosis01"></a>

训练或推理任务在运行过程中出现故障时，需要快速定位故障根因。任务通常分布在多个节点，故障相关日志（系统日志、设备日志、业务日志、plog等）分散在各节点的宿主机上，人工排查成本高、效率低。MindCluster提供集群运维Agent特性，通过`kubectl ascend_diag`命令按任务维度一键触发诊断，自动完成节点日志采集、日志清洗、集中诊断的全流程，帮助用户快速定位任务故障根因。

## 功能特点<a name="sectionfaultdiagnosisfeatures"></a>

- **一键诊断**：通过`kubectl ascend_diag`命令按任务维度触发诊断，无需登录各节点手动采集日志。
- **自动采集调度**：Agent Core作为故障诊断的集中控制中心，自动维护任务与Pod的映射关系，并向各计算节点的Node Collector下发采集指令。
- **本地清洗上报**：Node Collector按采集契约从宿主机采集任务日志，本地调用ascend-fd parse完成日志清洗，并将清洗结果打包上报给Agent Core。
- **集中诊断**：Agent Core聚合各节点上报的清洗结果，组装为诊断输入目录，调用ascend-fd diag执行集中诊断，生成诊断报告。
- **结果缓存**：每次诊断之后缓存诊断结果，重复诊断直接返回缓存，支持通过`--refresh`参数强制刷新。
- **智能总结**：可选对接LLM服务，对诊断报告进行智能总结，输出根因报告；未配置LLM时自动回退确定性诊断报告。

## 前提条件<a name="sectionfaultdiagnosisprereq"></a>

- 已完成Agent Core、Node Collector和Kubectl Plugin的安装部署，操作请参见[手动安装](../../03_installation_guide/02_installation/01_manual_installation.md)章节。
- 本机已安装kubectl，且kubeconfig具备访问集群的权限。
- 本机已安装socat，确保`kubectl port-forward`功能正常。
- 需要诊断的任务（如acjob、inferjob）已在集群中创建。若任务类型不在Agent Core默认支持范围内，请先完成[配置任务类型](./01_configuring_task_crds.md)章节的操作。
- 需要采集任务日志（如plog）时，请先完成[配置日志采集](./02_configuring_log_collection.md)章节的操作。

## 使用方式<a name="sectionfaultdiagnosisusage"></a>

1. [配置任务类型](./01_configuring_task_crds.md)：确认或配置Agent Core支持检测的任务类型。
2. [配置日志采集](./02_configuring_log_collection.md)：配置任务日志的采集方式及采集契约。
3. [使用故障诊断](./03_using_fault_diagnosis.md)：通过`kubectl ascend_diag`命令对异常任务执行诊断。
4. [配置和清除LLM](./04_configuring_llm.md)（可选）：对接LLM服务，对诊断报告进行智能总结。
