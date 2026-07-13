# 常见问题

## 安装与启动

### Q1：启动工具时报 "ModuleNotFoundError: No module named 'ascend_fd_tk'" 如何处理？

A：依次排查：① 是否已安装 WHL 包或处于源码目录；② 当前 Python 环境是否与安装时一致（多版本 Python 场景建议使用虚拟环境）；③ WHL 包安装方式下，确认 `ascend-fd-tk` 命令所在路径已加入系统 PATH。

### Q2：WHL 包安装时报 "error: invalid command 'bdist_wheel'" 如何处理？

A：缺少 `wheel` 打包工具，执行 `pip3 install wheel --upgrade` 后重新打包即可。

## 采集与诊断

### Q1：SSH 采集时频繁报"连接超时"如何处理？

A：依次排查：① 目标设备 IP / 端口是否可达；② `conn.ini` 中账号密码 / 私钥是否正确；③ 目标设备 `~/.ssh/authorized_keys` 是否已添加工具所在节点的公钥；④ 网络 ACL / 防火墙是否放通 22 端口。

### Q2：`auto_collect_diag` 报"生成报告失败"如何处理？

A：通常是上次生成的报告被 Excel / WPS 等进程占用。关闭占用进程后重新执行 `auto_diag` 即可重新生成报告。

### Q3：离线日志诊断时只设置了一个日志目录，能否正常诊断？

A：可以。工具支持单设备类型（仅主机 / 仅 BMC / 仅交换机）的独立诊断，至少设置一个日志目录后执行 `auto_collect_diag` 即可。诊断范围仅覆盖已配置设备类型。

## 报告解读

### Q1：诊断报告中的颜色标记代表什么含义？

A：诊断报告 Excel 文件中，每个 Sheet 的 Tab 颜色用于区分报告类型（链路分析、故障诊断、光模块信息等），Sheet 内部单元格颜色用于标识故障等级：浅红色表示存在故障 / 异常需优先处理，浅黄色表示告警 / 需关注建议排查，浅绿色表示正常 / 通过无需处理。详细说明参见[诊断 / 巡检报告说明](../05_usage/06_fault_analysis_report.md)。

### Q2：巡检报告（CSV）与诊断报告（XLSX）的区别是什么？

A：① 巡检报告由 `auto_inspection` 命令生成，基于预定义规则批量检查设备健康状态，输出 CSV 格式，仅记录异常项；② 诊断报告由 `auto_diag` 命令生成，针对已发生故障进行深度检测和根因分析，输出 XLSX 格式，包含链路分析、故障诊断、光模块信息等多个 Sheet。两者输入输出相互独立。

### Q3：报告生成后找不到文件在哪里如何处理？

A：① Linux 平台默认输出至 `~/.ascend-faultdiag-toolkit/report/` 目录；② Windows 平台默认输出至当前工作目录下 `.ascend-faultdiag-toolkit/report/` 目录；③ 诊断报告文件名格式为 `diag_report_{YYYYMMDD_HHMMSS}.xlsx`，巡检报告为 `inspection_errors.csv`。
