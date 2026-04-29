# 简介

## 概述

链路诊断工具（Ascend-faultdiag-toolkit）通过在线采集或离线日志分析设备链路故障，包含服务器、交换设备（L1/L2灵衢交换机、RoCE交换机）、BMC管理。

**在线采集**

用户输入待访问设备的连接信息（账号、密码/密钥/免密），工具访问设备采集信息。

**离线分析**

用户导入采集服务器带内日志、BMC dump日志和交换设备diagnostic information日志，分析关键信息。

**故障分析**

结合在线/离线采集信息，继承故障模式，进行故障分析。

## 使用指导

详细请参见[使用指导](../../../../component/ascend-faultdiag/toolkit_src/README.md)。
