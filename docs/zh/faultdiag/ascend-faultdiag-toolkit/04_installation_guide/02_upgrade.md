# 升级

本文档介绍链路诊断工具的版本升级流程。

1. 升级 WHL 包

   - 若当前环境工具版本与升级版本不一致，直接使用 `--upgrade` 参数升级工具：

    ```bash
    pip3 install --upgrade ascend_faultdiag_toolkit-{new_version}-py3-none-any.whl
    ```

   - 若当前环境工具版本与升级版本一致，使用 `--force-reinstall` 参数安装新版本：

    ```bash
    pip3 install --force-reinstall ascend_faultdiag_toolkit-{new_version}-py3-none-any.whl
    ```

    升级成功回显示例：

    ```txt
    Successfully installed ascend-faultdiag-toolkit-{version}
    ```

2. 验证升级

    执行 `about` 命令查看版本信息，若执行成功并回显 `MindCluster ascend-faultdiag-toolkit诊断工具版本：{version}`，则说明升级成功。

    ```bash
    ascend-fd-tk about
    ```
