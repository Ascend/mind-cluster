# Glossary

| Term | Description |
|------|------|
| AIV | A fault type related to NPU operator execution/bus access. When triggered, it may cause an NPU warm reset. |
| BER | Bit Error Rate, a metric for measuring link transmission quality, used to diagnose link stability. |
| BMC | Baseboard Management Controller. |
| CDR | Clock Data Recovery. |
| FaultLevel | Fault level classification, including three levels: fault state, sub-fault state, and sub-health state. |
| FRU | Field Replaceable Unit. |
| HBM | High Bandwidth Memory. |
| Host | Host server, the compute node that runs training or inference tasks. |
| hinicadm5       | Huawei hinic NIC management command-line tool, used to query the NIC list, port count, and SFP optical module information. |
| iBMC | Intelligent Baseboard Management Controller. |
| IIC | Inter-Integrated Circuit. |
| L1 switch | The first-layer switching device of the UnifiedBus network, enabling high-speed interconnection of multiple NPUs within a single server. |
| L2 switch | The second-layer switching device of the UnifiedBus network, interconnecting compute nodes across cabinets. |
| LLD.xlsx | The equipment room location configuration file, containing two sheets: "UnifiedBus L1 Network Mapping" and "UnifiedBus L2 Network Mapping". |
| NIC             | Network Interface Card. The `hinicadm5` command is used to collect lane-level information from the SFP optical module on the NIC port for diagnosis. |
| PoDManager      | Unified device management entry for Atlas 950 SuperPoD. After its IP is configured, the tool logs in and collects information from each device one by one by slot. |
| PSIP | Power Supply Integrated Package, an integrated NPU power supply module (such as 6A PSIP and 20A PSIP). Contact O&M personnel when a fault occurs. |
| PSU | Power Supply Unit. The tool can detect alarms such as PSU over-temperature. |
| RoCE | RDMA over Converged Ethernet. |
| RoCE switch | The leaf-spine Ethernet switch used in the RoCE parameter plane, carrying service traffic such as parameter synchronization and data read/write. |
| Serdes | Serializer/Deserializer. |
| SEL | System Event Log (on the BMC side). |
| SNR | Signal-to-Noise Ratio. |
| SFP             | Small Form-factor Pluggable, a hot-pluggable package for optical modules. NIC ports use SFP optical modules, and lane-level metrics are collected through the `hinicadm5` command. |
| SPOD | Single Port Of Death, the location information of a faulty NPU port. |
| TX LoL/RX LoL | Transmit/Receive loss of lock. |
| TX Los/RX Los | Transmit/Receive loss of signal. |
| uncorr_cw_cnt | Uncorrectable codeword count. |
| VRP | Versatile Routing Platform, the general-purpose routing platform operating system of Huawei. The switch must run VRP to support command collection by the tool. |

>[!NOTE]
> For more terms, see [Ascend Glossary](https://www.hiascend.com/document/detail/en/Glossary/gls/gls_0001.html).
