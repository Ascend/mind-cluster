# 升级

## 升级步骤

1. 执行以下命令安装新的软件包：

    ```shell
    pip3 install --upgrade ascend_faultdiag-{version}-py3-none-linux_{arch}.whl
    ```

    > [!NOTE]
    >
    > - 如果版本号相同，可以使用 `--force-reinstall` 参数强制重新安装。
    > - 请确保没有未完成的清洗或诊断任务。

2. 执行以下命令验证升级是否成功：

    ```shell
    ascend-fd version
    ```

    回显新的版本号即表示升级成功。

3. 升级失败处理

    请参考[卸载指南](03_uninstallation.md)进行卸载，然后参考[安装指南](01_installation.md)重新进行安装。
