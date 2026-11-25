# FAQ

## 安装时提示Python版本不满足要求

ascend-fd要求Python 3.7及以上版本，如需使用性能劣化功能，要求Python 3.8及以上版本。请检查Python版本：

```shell
python3 --version
```

如果版本不满足，请升级Python或使用conda创建独立环境。

## 清洗时提示磁盘空间不足

清洗输出目录需要至少5GB的可用磁盘空间。请清理磁盘或更换输出目录。

## root安装后普通用户无法使用

需要配置PATH环境变量。以root用户查询ascend-fd位置：

```shell
which ascend-fd
```

以普通用户添加PATH环境变量（假设ascend-fd安装在`/usr/local/python3.7.5/bin`）：

```shell
export PATH=$PATH:/usr/local/python3.7.5/bin
```

## 集群规模较大时诊断失败

Linux系统默认最大文件描述符数为1024。集群规模超过128台服务器（1024卡）时，需要调整文件描述符上限：

```shell
ulimit -n 65535
```

## 诊断报告中的故障设备过多

终端默认仅展示16条故障设备信息。完整信息可以在诊断结果文件 `diag_report.json` 中查看。

## 安装完ascend-fd提示command not found

- ascend-fd安装失败，重新安装ascend-fd即可。
- 当前设备可能有多个Python版本，ascend-fd安装在非默认Python下。

  安装在非默认Python下时，会有类似如下提示：

  ```output
  WARNING: The script ascend-fd is installed in '/usr/local/python3.8/bin' which is not on PATH.
    Consider adding this directory to PATH or, if you prefer to suppress this warning, use --no-warn-script-location.
  Successfully installed ascend-faultdiag-26.2.0
  ```

  可按以下方式解决：
  1. 查找ascend-fd：

      ```bash
      find / -name ascend-fd
      ```

      回显示例如下：

      ```output
      /usr/local/python3.8/bin/ascend-fd
      ```

  2. 将步骤1查找到的目录添加到PATH中：

      ```bash
      export PATH=$PATH:/usr/local/python3.8/bin/
      ```

  3. 再次尝试执行ascend-fd命令即可。
