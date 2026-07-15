# auto_diag

## 命令功能

启动自动诊断，适用于分批收集后统一诊断。

- 工具读取缓存目录中的结构化数据，运行全部[诊断项](../../07_references/02_diagnosis_items.md)，识别链路故障并定位根因。
- 需先通过 `auto_collect` 完成数据采集（可多次分批执行），再执行 `auto_diag` 启动统一诊断。
- 也可用于 `auto_collect_diag` 诊断失败后（如报告文件被占用），单独重新生成诊断报告。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `auto_diag` | 启动自动诊断，适用于分批收集后统一诊断 |
| `auto_diag ?` | 查看详情 |

## 参数说明

无业务参数，`?` 为内置帮助标识，用于查看命令用法。

## 执行前提

执行 `auto_diag` 前需确保缓存目录中存在结构化数据（通过 `auto_collect` 或 `auto_collect_diag` 采集生成）。

## 输出说明

- 成功：`诊断完成`
- 报告生成失败：`生成报告失败，解除占用后，可使用 'auto_diag' 重新生成报告。`

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect auto_diag
# 其他日志输出...
诊断完成
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> auto_diag
诊断完成
```
