# MindCluster链路故障诊断工具

- [简介](#简介)
- [目录结构](#目录结构)
- [使用指导](#使用指导)
- [使用场景](#使用场景)
- [API参考](#API参考)

# 简介

**MindCluster Ascend FaultDiag Toolkit提供昇腾AI集群链路故障诊断能力的轻量工具，提供从PC或单台服务器远程访问集群设备采集数据诊断的能力。**
- **本工具仅用作参考设计，请勿在商用环境使用**

# 目录结构
```
ascend-faultdiag
└─toolkit_src
  ├─doc
  └─ascend_fd_tk
     ├─core
     │  ├─cli_module
     │  ├─collect
     │  │  ├─collector
     │  │  ├─fetcher
     │  │  │  ├─dump_log_fetcher
     │  │  │  │  ├─bmc
     │  │  │  │  ├─host
     │  │  │  │  └─switch
     │  │  │  │      ├─cli_output_txt
     │  │  │  │      └─diag_info_output
     │  │  │  └─ssh_fetcher
     │  │  └─parser
     │  ├─common
     │  ├─config
     │  ├─context
     │  ├─crypto
     │  ├─fault_analyzer
     │  │  ├─bmc
     │  │  ├─common
     │  │  ├─hccs
     │  │  ├─host
     │  │  └─switch
     │  ├─inspection
     │  │  ├─check_items
     │  │  └─config
     │  ├─log_parser
     │  │  └─parse_config
     │  ├─model
     │  └─service
     ├─examples
     │  ├─auto_diag
     │  │  ├─linux
     │  │  └─windows
     │  ├─cmd
     │  ├─inspection
     │  │  ├─linux
     │  │  └─windows
     │  ├─loopback_diag
     │  └─scripts
     ├─test
     └─utils
```

# 使用指导

## 工具获取

* 获取MindCluster Ascend FaultDiag软件包（暂未发布）。

* Linux环境源码编译生成whl包
```
git clone https://gitcode.com/Ascend/mind-cluster.git
cd mind-cluster/component/ascend-faultdiag/toolkit_src
python3 setup.py bdist_wheel
# 生成的whl存放路径：mind-cluster/component/ascend-faultdiag/toolkit_src/dist/ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```
* Windows环境生成exe
```
# 手动生成exe需要在Windows上安装pyinsatller
pip3 install pyinstaller -i https://mirrors.huaweicloud.com/repository/pypi/simple/ --trusted-host mirrors.huaweicloud.com

git clone https://gitcode.com/Ascend/mind-cluster.git
cd mind-cluster\component\ascend-faultdiag\toolkit_src
.\build_cli_exe.bat
# 生成的exe存放路径：mind-cluster/component/ascend-faultdiag/toolkit_src/dist/ascend-faultdiag-toolkit.exe
```

## 环境准备

MindCluster Ascend FaultDiag ToolKit支持的Python版本需≥3.8。在安装该工具前，请检查依赖的Python版本和[requirements.txt](requirements.txt)文件中的三方依赖是否满足要求。
- linux使用，环境准备请参考[在linux上使用的环境准备.md](doc/在linux上使用的环境准备.md)
- windows使用，环境准备请参考[在windows上使用的环境准备.md](doc/在windows上使用的环境准备.md)

## 安装与卸载

执行如下命令进行安装
```
pip3 install ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

执行如下命令进行卸载
```
pip3 uninstall ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

## 快速入门

MindCluster Ascend FaultDiag ToolKit支持两种命令执行方式：交互式模式和非交互式模式。

* 交互式模式使用指南：[交互式命令执行.md](doc/交互式命令执行.md)
* 非交互式模式使用指南：[非交互式命令执行.md](doc/非交互式命令执行.md)

# 使用场景

## 场景一：在线故障诊断
```
1.准备conn.ini配置文件（包含设备IP、用户名、密码或者密钥文件）
2.set_conn_config conn.ini
3.auto_collect_diag
4.查看生成的诊断报告
```

## 场景二：离线日志分析
```
1.将采集的日志放到指定目录
2.set_host_dump_log /your_path/host_logs
3.set_bmc_dump_log /your_path/bmc_logs
4.set_bmc_dump_log /your_path/switch_logs
5.auto_collect_diag
6.查看生成的诊断报告
```

## 场景三：跨网络平面采集

当分别部署在不同网络平面的诊断工具采集的数据需要汇总诊断时，可以将所有网络平面的信息收集汇总后，统一诊断。
```
1.在线或者离线方式获取信息
2.在网络A执行 auto_collect
3.在网络B执行 auto_collect
4.汇总后执行 auto_diag
5.查看生成的诊断报告
```

## 场景四：客户定制化巡检

```
1.在线或者离线方式获取信息
2.auto_collect
3.auto_inspection <客户类型>  # 目前支持客户类型：mayi
4.查看生成的诊断报告
```

# API参考


| 命令                    | 功能描述 | 用法说明                                                       |
|-----------------------| -------- |------------------------------------------------------------|
| `help`                | 显示帮助信息 | 直接执行<br>`help ?` 查看详情                |
| `exit`                | 退出程序 | 直接执行<br>`exit ?` 查看详情                          |
| `clear`               | 清屏 | 直接执行<br>`clear ?` 查看详情                        |
| `about`               | 查看关于诊断工具 | 直接执行<br>`about ?` 查看详情                                    |
| `guide`               | 获取向导信息 | 直接执行<br>`guide ?` 查看详情                             |
| `set_conn_config`     | 设置连接文件地址 | `set_conn_config <文件地址>`<br>`set_conn_config ?` 查看详情       |
| `set_host_dump_log`   | 设置服务器导出日志目录 | `set_host_dump_log <目录>`<br>`set_host_dump_log ?` 查看详情     |
| `set_bmc_dump_log`    | 设置BMC导出日志目录 | `set_bmc_dump_log <目录>`<br>`set_bmc_dump_log ?` 查看详情       |
| `set_switch_dump_log` | 设置交换机命令回显导出目录 | `set_switch_dump_log <目录>`<br>`set_switch_dump_log ?` 查看详情 |
| `collect_bmc_dump_info` | 在线收集BMC dump info日志 | 直接执行<br>`collect_bmc_dump_info ?` 查看详情                                            |
| `auto_collect`        | 启动自动信息采集，支持离线、在线采集，适用于不同网络平面分批收集 | 直接执行<br>`auto_collect ?` 查看详情                                     |
| `auto_inspection`     | 启动巡检结果诊断，适用于分批收集后统一诊断 | 直接执行<br>`auto_inspection ?` 查看详情                                        |
| `auto_diag`           | 启动自动诊断，适用于分批收集后统一诊断 | 直接执行<br>`auto_diag ?` 查看详情                               |
| `auto_collect_diag`   | 启动一键式自动收集诊断（在线设备采集或离线日志收集） | 直接执行<br>`auto_collect_diag ?` 查看详情                                    |
| `clear_cache`         | 清理缓存，新诊断任务前务必执行，避免干扰诊断结果 | 直接执行<br>`clear_cache ?` 查看详情                                   |

