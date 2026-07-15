# about

## 命令功能

查看诊断工具的版本信息，常用于安装验证或问题定位时提供版本信息。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `about` | 查看关于诊断工具 |
| `about ?` | 查看详情 |

## 参数说明

无业务参数，`?` 为内置帮助标识，用于查看命令用法。

## 输出说明

控制台返回版本信息： `MindCluster ascend-faultdiag-toolkit诊断工具版本：<version>` 。

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk about
        MindCluster ascend-faultdiag-toolkit诊断工具版本：v0.10
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> about
        MindCluster ascend-faultdiag-toolkit诊断工具版本：v0.10
```
