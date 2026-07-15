# 卸载

本文档介绍链路诊断工具的卸载流程。

1. 卸载 WHL 包

    ```bash
    pip3 uninstall ascend-faultdiag-toolkit
    ```

    卸载成功回显：

    ```txt
    Successfully uninstalled ascend-faultdiag-toolkit-{version}
    ```

2. 清理缓存与运行数据

   - Linux 平台：工具数据统一存放在用户主目录下，执行 `rm -rf ~/.ascend-faultdiag-toolkit/`。
   - Windows 平台：工具数据存放在**各次启动工具时所在的当前工作目录**下，每个工作目录下均可能生成 `.ascend-faultdiag-toolkit` 文件夹。可通过 PowerShell 批量查找后删除：

     ```powershell
     # 批量查找所有残留目录
     Get-ChildItem -Path C:\ -Directory -Recurse -Filter ".ascend-faultdiag-toolkit" -ErrorAction SilentlyContinue | Select-Object FullName

     # 确认后批量删除
     Get-ChildItem -Path C:\ -Directory -Recurse -Filter ".ascend-faultdiag-toolkit" -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force
     ```

     也可在常用工作目录下逐一找到 `.ascend-faultdiag-toolkit` 文件夹并手动删除。

3. 卸载验证

    执行以下命令验证是否卸载成功：

   - Linux 平台：执行 `which ascend-fd-tk`，若命令不存在则表示卸载成功。
   - Windows 平台：执行 `where ascend-fd-tk`，若命令不存在则表示卸载成功。
