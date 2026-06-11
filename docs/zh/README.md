# 欢迎使用MindCluster

<div style="text-align: center; margin: 0.5rem 0 0.3rem 0; font-family: 'Avenir Next', 'Avenir', 'Century Gothic', 'Segoe UI', sans-serif;">
  <span style="font-size: 4.5rem; font-weight: 300; letter-spacing: 0.02em;">MindCluster</span>
</div>

MindCluster是支持NPU的构建的深度学习系统组件，专为训练和推理任务提供集群级解决方案。本docs仓提供MindCluster集群调度及Ascend FaultDiag（故障诊断）工具相关指导文档。

**表 1**  文档目录结构

|文件夹|文档|文档名称|内容介绍|跳转链接|
|--|--|--|--|--|
|-|MindCluster是什么|overview.md|整体介绍MindCluster的定义、特性说明和组件说明。|[MindCluster是什么](./overview.md)|
|集群调度（scheduling）|简介|01_introduction|整体介绍集群调度是什么，以及适配的软硬件环境。|[简介](./scheduling/01_introduction/_menu_introduction.md)|
|-|快速入门|02_quick_start|提供快速完成集群调度组件安装及训练任务下发的示例。|[快速入门](./scheduling/02_quick_start/quick_start.md)|
|-|安装部署|03_installation_guide|介绍集群调度组件的多种安装方式及维护。|[安装部署](./scheduling/03_installation_guide/menu_installation_guide.md)|
|-|特性指南|04_usage|介绍集群调度的相关特性。包括但不限于：<ul><li>容器化支持特性指南</li><li>资源监测特性指南</li><li>虚拟化实例特性指南</li><li>调度特性指南</li><li>断点续训特性指南</li><li>一体机特性指南</li><li>MindIE Motor推理任务最佳实践</li><li>SGLang推理任务最佳实践</li><li>vLLM推理任务最佳实践</li><li>Infer Operator推理任务最佳实践</li></ul>|[特性指南](./scheduling/04_usage/menu_usage.md)|
|-|开发者指南|05_developer_guide|介绍集群调度组件的手动安装方式及维护、自定义指标开发、插件开发等内容。|[开发者指南](./scheduling/05_developer_guide/menu_developer_guide.md)|
|-|API参考|06_api|介绍集群调度的API接口。|[API参考](./scheduling/06_api/menu_api.md)|
|-|参考|07_references|介绍集群调度的使用参考、常用操作、FAQ、安全加固等内容。|[参考](./scheduling/07_references/menu_references.md)|
|-|目录结构|menu_scheduling_user_guide.md|提供《用户指南》整体目录结构。|[目录结构](./scheduling/menu_scheduling_user_guide.md)|
|故障诊断（faultdiag）|简介|introduction.md|整体介绍故障诊断工具是什么、应用场景及方案。|[简介](./faultdiag/introduction.md)|
|-|支持的产品形态|supported_products.md|介绍故障诊断工具适配的硬件环境。|[支持的产品形态](./faultdiag/supported_products.md)|
|-|安装与升级|installation_guide.md|介绍故障诊断工具的安装及维护。|[安装与升级](./faultdiag/installation_guide.md)|
|-|使用指导|user_guide|介绍故障诊断工具的使用方法。|[使用指导](./faultdiag/user_guide/menu_user_guide.md)|
|-|API接口说明|api|介绍故障诊断工具的API接口。|[API接口说明](./faultdiag/api/menu_api.md)|
|-|常用操作|common_operations.md|介绍故障诊断工具使用涉及的常用操作。|[常用操作](./faultdiag/common_operations.md)|
|-|安全加固|security_hardening.md|介绍故障诊断工具的操作系统安全加固、防火墙配置等内容。|[安全加固](./faultdiag/security_hardening.md)|
|-|FAQ|faq.md|介绍故障诊断工具的常见问题。|[FAQ](./faultdiag/faq.md)|
|-|链路诊断工具|ascend-faultdiag-toolkit|介绍链路诊断工具是什么以及使用指导。|[链路诊断工具](./faultdiag/ascend-faultdiag-toolkit/menu_ascend-faultdiag-toolkit.md)|
|-|附录|appendix.md|介绍故障诊断工具支持的错误码和故障类型。|[附录](./faultdiag/appendix.md)|
|-|目录结构|menu_faultdiag_user_guide.md|提供《用户指南》整体目录结构。|[目录结构](./faultdiag/menu_faultdiag_user_guide.md)|
|-|版本说明书|release_notes.md|介绍MindCluster的版本配套说明、兼容性说明、特性更新说明等内容。|[版本说明书](./release_notes.md)|
|-|目录结构|menu_release_notes.md|提供《版本说明书》整体目录结构。|[目录结构](./menu_release_notes.md)|
|资源（resource）|-|-|提供MindCluster公网地址等资源。|[资源](./resource)|
|图片（figures）|-|-|提供文档中所使用到的图片参考。|[图片](./figures)|
