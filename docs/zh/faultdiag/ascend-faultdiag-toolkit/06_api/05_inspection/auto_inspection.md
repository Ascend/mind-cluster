# auto_inspection

## 命令功能

启动巡检诊断，使用 `auto_collect` 收集后产生的中间数据进行批量规则检查。适用于客户定制化巡检场景，按预定义规则对清洗后的设备数据进行健康检查。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `auto_inspection` | 使用默认客户类型启动巡检。 |
| `auto_inspection <客户类型>` | 使用指定客户类型启动巡检。 |
| `auto_inspection ?` | 查看支持的客户类型。 |

## 参数说明

| 参数 | 说明 |
|------|------|
| `<客户类型>` | 客户类型枚举值。目前仅支持 `default`。 |

## 输出说明

- 成功：`诊断完成`
- 客户类型不支持：`{args}为不支持的客户类型，请使用 ' auto_inspection ? ' 查看支持的客户类型`

## 示例

非交互式方式：

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect auto_inspection default
# 其他日志输出...
诊断完成
```

交互式方式：

```bash
ascend-fd-tk
>>> auto_inspection default
诊断完成
```
