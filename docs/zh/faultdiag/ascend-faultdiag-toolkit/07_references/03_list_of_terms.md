# 术语表

| 术语 | 说明 |
|------|------|
| AIV | NPU 算子执行 / 总线访问相关的故障类型，触发后可能导致 NPU 热复位 |
| BER | Bit Error Rate，误码率，衡量链路传输质量的指标，用于诊断链路稳定性 |
| BMC | Baseboard Management Controller，基板管理控制器 |
| CDR | Clock Data Recovery，时钟数据恢复 |
| ECC | Error Checking and Correction，错误检查和纠正，用于检测和修正内存中的位翻转错误 |
| FaultLevel | 故障等级分类，包含故障态、次故障态、亚健康态三级 |
| FRU | Field Replaceable Unit，现场可更换单元 |
| HBM | High Bandwidth Memory，高带宽存储器 |
| HCCS | Huawei Cache Coherent System，华为缓存一致性系统互联总线，用于 NPU 之间的高速互联 |
| Host | 主机服务器，运行训练或推理任务的计算节点 |
| iBMC | Intelligent Baseboard Management Controller，华为服务器智能基板管理控制器 |
| IIC | Inter-Integrated Circuit，集成电路总线 |
| L1 交换机 | 灵衢网络第一层交换设备，实现单机内多 NPU 高速互通 |
| L2 交换机 | 灵衢网络第二层交换设备，完成跨机柜算力节点互联 |
| LLD.xlsx | 机房位置配置文件，含「灵衢L1网络对应关系」「灵衢L2网络对应关系」两个 Sheet |
| LLDP | Link Layer Discovery Protocol，链路层发现协议 |
| NPU | Neural Network Processing Unit，神经网络处理器 |
| PSIP | Power Supply Integrated Package，NPU 供电集成模块（如 6A PSIP、20A PSIP），故障时需联系运维处理 |
| PSU | Power Supply Unit，电源模块，工具可检测 PSU 过温等告警 |
| RoCE | RDMA over Converged Ethernet，基于以太网的 RDMA 技术 |
| RoCE 交换机 | RoCE 参数平面使用的叶脊以太网交换机，承载参数同步、数据读写等业务流量 |
| Serdes | Serializer / Deserializer，串行器 / 解串器 |
| SEL | System Event Log，系统事件日志（BMC 侧） |
| SNR | Signal-to-Noise Ratio，信噪比 |
| SPOD | Single Port Of Death，NPU 故障端口定位信息 |
| TX LoL / RX LoL | 发送 / 接收失锁（Loss of Lock） |
| TX Los / RX Los | 发送 / 接收信号丢失（Loss of Signal） |
| uncorr_cw_cnt | uncorrectable codeword count，不可纠正码字数 |
| VRP | Versatile Routing Platform，华为通用路由平台操作系统，交换机需运行 VRP 以支持工具采集命令 |
| 灵衢 | 华为专用高速总线交换网络方案，包含 L1 / L2 两级交换设备 |
| 网络平面 | 集群中不同的网络子网，多网络平面场景下需分批采集 |
