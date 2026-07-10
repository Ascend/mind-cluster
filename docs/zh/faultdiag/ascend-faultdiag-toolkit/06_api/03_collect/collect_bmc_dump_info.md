# collect_bmc_dump_info

## 命令功能

在线收集 BMC dump info 日志。仅触发 BMC 一侧的采集，不会自动启动诊断。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `collect_bmc_dump_info` | 在线收集 BMC dump info 日志。 |
| `collect_bmc_dump_info ?` | 查看详情。 |

## 参数说明

无参数。执行前需通过 `set_conn_config` 命令配置 BMC 设备连接信息。

## 输出说明

收集完成后控制台返回：`收集完成，请查看日志路径{path}`，日志默认存放于 `{家目录}/cache/bmc_dump_cache`。

## 示例

非交互式方式：

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini collect_bmc_dump_info
设置成功，请尽快删除包含明文密码的配置文件
收集完成，请查看日志路径/home/user/.ascend-faultdiag-toolkit/cache/bmc_dump_cache
```

交互式方式：

```bash
ascend-fd-tk
>>> set_conn_config /home/user/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> collect_bmc_dump_info
收集完成，请查看日志路径/home/user/.ascend-faultdiag-toolkit/cache/bmc_dump_cache
```
