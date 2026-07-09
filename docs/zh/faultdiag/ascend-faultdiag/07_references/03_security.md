# 安全加固

## 文件权限

ascend-fd 安装和使用过程中，建议设置以下文件权限：

| 目录/文件                         | 建议权限 | 说明                     |
|-----------------------------------|----------|--------------------------|
| `~/.ascend_faultdiag`             | 700      | 仅所有者可读写执行       |
| `~/.ascend_faultdiag/*.log`       | 600      | 仅所有者可读写           |
| `~/.ascend_faultdiag/config.json` | 600      | 配置文件，仅所有者可读写 |

设置权限：

```shell
chmod 700 ~/.ascend_faultdiag
chmod 600 ~/.ascend_faultdiag/*.log
chmod 600 ~/.ascend_faultdiag/config.json
```

## 系统安全配置

建议将 umask 设置为 027，以提高安全性：

```shell
vim /etc/profile
# 在文件末尾添加 umask 027
source /etc/profile
```

## 日志安全

- ascend-fd 安装和使用过程中产生的日志文件包含操作记录，建议定期清理
- 不要将日志文件暴露给未经授权的用户
- 操作日志记录了用户的操作行为，请妥善保管

## 用户权限

- 建议使用同一用户完成安装和使用
- 如果必须使用 root 安装、普通用户使用，需要确保普通用户有正确的 PATH 配置和文件权限
