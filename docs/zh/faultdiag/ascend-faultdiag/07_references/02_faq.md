# FAQ

## 使用问题

### Q1：安装时提示 Python 版本不满足要求

**A**：ascend-fd 要求 Python 3.7 及以上版本，如需使用性能劣化功能，要求 Python 3.8 及以上版本。请检查 Python 版本：

```shell
python3 --version
```

如果版本不满足，请升级 Python 或使用 conda 创建独立环境。

### Q2：清洗时提示磁盘空间不足

**A**：清洗输出目录需要至少 5GB 的可用磁盘空间。请清理磁盘或更换输出目录。

### Q3：root 安装后普通用户无法使用

**A**：需要配置 PATH 环境变量。以 root 用户查询 ascend-fd 位置：

```shell
which ascend-fd
```

以普通用户添加 PATH 环境变量（假设 ascend-fd 安装在 `/usr/local/python3.7.5/bin`）：

```shell
export PATH=$PATH:/usr/local/python3.7.5/bin
```

### Q4：集群规模较大时诊断失败

**A**：Linux 系统默认最大文件描述符数为 1024。集群规模超过 128 台服务器（1024 卡）时，需要调整文件描述符上限：

```shell
ulimit -n 65535
```

### Q5：诊断报告中的故障设备过多

**A**：终端默认仅展示 16 条故障设备信息。完整信息可以在诊断结果文件 `diag_report.json` 中查看。
