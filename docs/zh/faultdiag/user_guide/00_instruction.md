# 使用说明<a name="ZH-CN_TOPIC_0000001677859701"></a>

## 使用建议<a name="section181951321155817"></a>

- 使用诊断功能时，因Linux系统最大进程数限制（默认为1024），故集群规格建议≤128台服务器（1024卡）。若服务器数量超过此规格时，需使用**ulimit -n**命令调整文件描述符上限。
- 用户在使用MindCluster Ascend FaultDiag工具命令时，尽量不使用管道命令，可能会影响用户IP的获取、影响日志审计。

## 支持的场景<a name="section12381154319315"></a>

- MindCluster Ascend FaultDiag工具仅支持对整机满卡训练及推理任务提供故障诊断能力，若非满卡训练及推理场景执行诊断可能导致故障根因定位错误或失败。
- MindCluster Ascend FaultDiag工具当前仅支持IPv4，不支持使用IPv6。

## 系统时间说明<a name="section98531395015"></a>

- 请用户同步各训练及推理服务器的系统时间，系统时间不一致可能会导致分析结果不准确。
- 请用户同步每个训练及推理服务器上Host系统时间与Device的系统时间，系统时间不一致可能会导致分析结果不准确。
- 若使用容器执行训练及推理任务，请用户同步宿主机与训练及推理容器的系统时间，系统时间不一致可能会导致分析结果不准确。

## 故障诊断日志版本配套表<a name="section087884485711"></a>

**表 1**  日志对应软件配套表

|日志文件|对应软件|软件版本|说明|
|--|--|--|--|
|CANN应用类日志|CANN|7.0.RC1及以上|CANN打印的Host侧应用类日志和Device侧应用类日志。更多相关信息请参见《CANN 日志参考》中的“[查看日志（Ascend EP）](https://www.hiascend.com/document/detail/zh/canncommercial/850/maintenref/logreference/logreference_0002.html)”章节。|
|PyTorch框架训练及推理日志|PyTorch1.11.0框架适配插件|5.0.RC3及以上|-|
|MindSpore框架训练日志|MindSpore|2.1.0及以上|部分故障类型描述中包含对应的MindSpore版本说明，请以实际故障诊断描述为主。|
|Host OS日志|-|-|<ul><li>支持检测Host OS日志包括但不限于CentOS 7.6、Debian 10.0、EulerOS 2.10、EulerOS 2.12和CTyunOS 22.06的HOST OS日志。不同操作系统日志打印关键字可能存在差异。</li><li>建议Host OS日志大小在512MB以内。</li></ul>|
|Device侧日志|Ascend HDK|23.0.RC3及以上|-|
|MindCluster组件日志|Ascend Device Plugin、NodeD、Ascend Docker Runtime、NPU Exporter、Volcano|6.0.RC3及以上|-|
|MindIE组件日志|MindIE Server、MindIE LLM、MindIE SD、MindIE RT、MindIE Torch、MindIE MS、MindIE Benchmark、MindIE Client|6.0.0及以上|-|
|AMCT组件日志|AMCT模型压缩组件|7.0.RC1及以上|AMCT集成在CANN包中进行发布。更多相关信息请参见《[AMCT模型压缩工具用户指南](https://www.hiascend.com/document/detail/zh/canncommercial/850/devaids/amct/atlasamct_16_0001.html)》。|
|MindIE Pod控制台日志|MindIE Pod控制台日志|-|-|
|MindIO组件日志|MindIO组件日志|-|-|
