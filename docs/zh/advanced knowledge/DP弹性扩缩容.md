# DP弹性扩缩容流程

本文档描述了DP（Data Parallelism）弹性缩容的完整流程，包括训练拉起、故障检测、策略决策和恢复训练等阶段。

## 流程图

```plantuml

@startuml
participant device_plugin
participant noded
participant clusterd
participant ascend_operator
participant volcano
participant taskd_manager
participant mindio_controller
participant taskd_agent
participant mindspeed_llm
participant mindio_processor
participant torch_npu
participant cann

== 训练拉起与故障检测 ==
activate device_plugin
activate noded
device_plugin -> clusterd : 芯片故障上报(deviceinfo cm)
activate clusterd
noded -> clusterd : 节点故障上报(nodeinfo cm)

taskd_manager -> mindio_controller : 拉起
activate mindio_controller
taskd_agent -> mindspeed_llm : 拉起训练进程
activate mindspeed_llm
mindspeed_llm -> mindio_processor : 拉起、注册回调callback
deactivate mindspeed_llm
mindio_controller -> taskd_manager : 软件故障上报report_process_fault
deactivate mindio_controller
activate taskd_manager
taskd_manager -> clusterd : 软件故障上报ReportProcessFault
deactivate taskd_manager

clusterd -> clusterd : 故障汇总分析
note right of clusterd
内部故障聚合与决策
end note

== 停止训练 ==
clusterd -> taskd_manager : stop train
activate taskd_manager
taskd_manager -> mindio_controller : tft_notify_controller_stop_train(controller流转STATE_OP_NORMAL状态)
deactivate taskd_manager
activate mindio_controller
mindio_controller -> mindio_processor : OP_PRELOCK

== 停止完成及策略上报 ==
mindio_controller -> taskd_manager : report_stop_complete
deactivate mindio_controller
activate taskd_manager
taskd_manager -> clusterd : ReportStopComplete
deactivate taskd_manager

clusterd -> taskd_manager : notify all fault ranks & exit fault nodes
activate taskd_manager
taskd_manager -> taskd_agent : exit fault nodes
activate taskd_agent
taskd_agent -> volcano : 删除故障pod ERROR，volcano删除ERROR pod
deactivate taskd_agent
activate volcano
volcano -> ascend_operator : operator检测到pod不满足最小副本要求创建新pod
deactivate volcano
taskd_manager -> mindio_controller : tft_notify_controller_on_global_rank
deactivate taskd_manager
activate mindio_controller

mindio_controller -> taskd_manager : report_recover_strategy
deactivate mindio_controller
activate taskd_manager
taskd_manager -> clusterd : ReportRecoverStrategy
deactivate taskd_manager

== 策略下发 ==
clusterd -> taskd_manager : changeStrategy:recover
activate taskd_manager
taskd_manager -> mindio_controller : tft_notify_controller_change_strategy:recover(controller流转STATE_OP_ENV_CLEAR状态)
deactivate taskd_manager
activate mindio_controller

== 恢复流程 ==
mindio_controller -> mindio_processor : 通知 OP_DEVICE_STOP
activate mindio_processor
mindio_processor -> mindspeed_llm : stop_callback
deactivate mindio_processor

activate mindspeed_llm
mindspeed_llm -> torch_npu : stop_device
deactivate mindspeed_llm

activate torch_npu
torch_npu -> cann : AclRtDeviceTaskAbort
deactivate torch_npu

mindio_controller -> mindio_processor : 通知 OP_DEVICE_CLEAN
activate mindio_processor
mindio_processor -> mindspeed_llm : clean_callback
deactivate mindio_processor

activate mindspeed_llm
mindspeed_llm -> torch_npu : restart_device
deactivate mindspeed_llm

mindio_controller -> mindio_controller : controller流转STATE_OP_REPAIR状态，等待新节点训练进程全部注册成功

mindio_controller -> mindio_processor : OP_PT_COMM
activate mindio_processor
mindio_processor -> mindspeed_llm : arf_rebuild_process_group_callback
deactivate mindio_processor

activate mindspeed_llm
mindspeed_llm -> torch_npu : reinit_process_group
deactivate mindspeed_llm
activate torch_npu
torch_npu -> cann : HcclCommDestroy
deactivate torch_npu

mindio_controller -> mindio_processor : 通知 OP_REPAIR
activate mindio_processor
mindio_processor -> mindspeed_llm : repair_callback
deactivate mindio_processor

mindio_controller -> mindio_processor : 通知 OP_ROLLBACK
activate mindio_processor
mindio_processor -> mindspeed_llm : rollback_callback
deactivate mindio_processor

mindio_controller -> mindio_processor : 通知 OP_NOTIFY_NORMAL
activate mindio_processor
mindio_processor -> mindspeed_llm : 恢复训练
deactivate mindio_processor

mindio_controller -> taskd_manager : report_recover_status(controller流转STATE_OP_NORMAL状态)
deactivate mindio_controller
activate taskd_manager
taskd_manager -> clusterd : ReportRecoverStatus
deactivate taskd_manager

deactivate device_plugin
deactivate noded
deactivate clusterd
@enduml
```

## 流程说明

### 1. 训练拉起与故障检测阶段
- **device_plugin**：负责芯片故障上报
- **noded**：负责节点故障上报
- **clusterd**：接收并汇总所有故障信息，进行内部故障聚合与决策
- **taskd_manager**：协调训练进程的拉起
- **mindio_controller**：管理训练进程，注册故障回调

### 2. 停止训练阶段
- **clusterd**：下发停止训练命令
- **mindio_controller**：通知处理器进入prelock状态

### 3. 停止完成及策略上报阶段
- **taskd_agent**：执行故障节点退出
- **volcano**：删除故障Pod
- **ascend_operator**：检测Pod副本数，若没有足够资源则新pod不会被调度

### 4. 策略下发阶段
- **clusterd**：决策恢复策略并下发

### 5. 恢复流程阶段
- **设备停止**：通过`AclRtDeviceTaskAbort`中止设备任务
- **设备清理**：清理设备状态
- **通信重建**：通过`HcclCommDestroy`和`reinit_process_group`重建分布式通信组
- **恢复训练**：执行rollback并恢复正常训练

## 关键组件说明

| 组件 | 职责 |
|------|------|
| **device_plugin** | NPU设备发现与故障上报 |
| **noded** | 节点健康监控与故障上报 |
| **clusterd** | 集群级故障聚合与策略决策 |
| **taskd_manager** | 任务生命周期管理 |
| **mindio_controller** | 训练进程控制器 |
| **mindspeed_llm** | LLM训练框架 |
| **torch_npu** | PyTorch NPU适配层 |
| **cann** | Ascend计算架构 |
