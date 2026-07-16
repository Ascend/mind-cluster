# help

## 命令功能

显示所有可用命令的帮助信息，便于快速检索工具能力。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `help` | 显示帮助信息 |
| `help ?` | 查看详情 |

## 参数说明

无业务参数，`?` 为内置帮助标识，用于查看命令用法。

## 输出说明

控制台按列对齐输出所有可用命令及简短帮助。

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk help
help                       - 显示帮助信息
exit                       - 退出程序
clear                      - 清屏
about                      - 查看关于诊断工具
guide                      - 获取向导信息
set_config_dir             - 设置配置文件目录路径，支持 " set_config_dir <目录路径> " 设置，或 " set_config_dir ? " 查看详情
set_conn_config            - 设置连接文件地址，支持 " set_conn_config <文件地址> " 设置，或 " set_conn_config ? " 查看详情
set_host_dump_log          - 设置服务器导出日志目录，支持 " set_host_dump_log <目录> " 设置目录，或 " set_host_dump_log ? " 查看详情
set_bmc_dump_log           - 设置BMC导出日志目录，支持 " set_bmc_dump_log <目录> " 设置目录，或 " set_bmc_dump_log ? " 查看详情
set_switch_dump_log        - 设置交换机命令回显导出目录，支持 " set_switch_dump_log <目录> " 设置目录，或 " set_switch_dump_log ? " 查看详情
collect_bmc_dump_info      - 在线收集BMC dump info日志
auto_collect               - 启动自动信息采集，支持离线、在线采集，适用于不同网络平面分批收集
auto_inspection            - 启动巡检结果诊断，适用于分批收集后统一诊断
auto_diag                  - 启动自动诊断，适用于分批收集后统一诊断
auto_collect_diag          - 启动一键式自动收集（在线设备采集或离线日志收集）诊断
clear_cache                - 清理缓存，请在执行新诊断任务前务必执行！避免干扰诊断结果（若清理未生效请用管理员模式打开工具）
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> help
# 回显结果同上
```
