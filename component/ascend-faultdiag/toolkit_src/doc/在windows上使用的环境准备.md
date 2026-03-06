# Windows 环境准备指南

## 一、关键前置操作：移除 Windows Store 路径干扰（必做）

Windows 系统默认会在 `PATH` 中添加 Windows Store 相关路径，导致输入 `python` 时自动跳转至应用商店，需先移除该路径：

### 1. 打开 “环境变量” 设置

* 按下 `Win + R` 键，输入 `sysdm.cpl` 并按回车，打开 “系统属性” 窗口；

* 切换到 **高级** 选项卡，点击右下角的 **环境变量** 按钮；
![进入环境变量.png](picture/%E8%BF%9B%E5%85%A5%E7%8E%AF%E5%A2%83%E5%8F%98%E9%87%8F.png)
* 在弹出的 “环境变量” 窗口中，找到 “用户变量” 或 “系统变量” 下的 **Path** 变量（建议先修改 “用户变量”，无需管理员权限），点击
  **编辑**。
![进入path.png](picture/%E8%BF%9B%E5%85%A5path.png)

### 2. 移除 Windows Store 相关路径

在 “编辑环境变量” 窗口中，找到包含以下内容的路径，选中后点击 **删除**：

* `%USERPROFILE%\AppData\Local\Microsoft\WindowsApps`
![删除windowsapp.png](picture/%E5%88%A0%E9%99%A4windowsapp.png)
* 若存在其他类似 “WindowsApps” 的路径，也一并删除（确保无关联路径残留）。

### 3. 保存设置

* 点击 **确定** 关闭 “编辑环境变量” 窗口；

* 再次点击 “环境变量” 和 “系统属性” 窗口的 **确定**，设置生效；

* （重要）关闭当前所有已打开的命令提示符（CMD）或 PowerShell 窗口，重新打开新窗口（环境变量修改需重启终端生效）。

## 二、Python 环境准备
提供两种python环境安装方式（常规安装/免安装）可选，适配不同的场景。
- 常规安装python：系统全局可用，支持所有 Python 功能（如系统服务、第三方工具调用）。
- 免安装python：无需安装，可随文件夹迁移，可批量部署到多台机器，实现多版本隔离，避免系统环境冲突。

### 方式1. 免安装python和依赖（嵌入版, 推荐使用该方式）
确保可直连外网, 执行脚本[download_embed_python.bat](../ascend_fd_tk/examples/scripts/download_embed_python.bat)，脚本自动完成：
- 下载指定版本的 Python Embed 版（脚本默认安装python3.10.11）；
- 修复 pip 环境；
- 安装预设依赖和项目需要依赖requirements.txt（不需要再执行步骤三、安装项目依赖）；
- 生成环境配置文件

下载的Python Embed版和依赖文件会保存在用户个人目录的python_portable目录中（%USERPROFILE%\python_portable\）。

### 方式2.常规安装python（系统级python）
#### 2.1. 下载 Python 安装包

* 打开浏览器，访问 Python 官方下载页：[python windows版本下载列表](https://www.python.org/downloads/windows/)
  ![下载python.png](picture/%E4%B8%8B%E8%BD%BDpython.png)
* 滚动到 “Stable Releases”（稳定版）区域，选择 **Python 3.8 及以上版本**（推荐 3.9/3.10，需与项目兼容）；

* 根据系统位数选择安装包：


* 64 位系统：点击 “Windows Installer (64-bit)”；

* 32 位系统：点击 “Windows Installer (32-bit)”（若不确定系统位数，可右键 “此电脑”→“属性” 查看 “系统类型”）。

#### 2.2. 安装 Python（关键选项必选）

* 双击下载的 `.exe` 安装包，打开安装界面；

* （**必须勾选**）在安装界面底部勾选 **Add Python 3.x to PATH**（自动添加 Python 到环境变量，避免手动配置）；

* 点击 **Install Now**（默认安装，推荐新手），或选择 “Customize installation”（自定义安装路径，适合有经验用户，建议安装到非系统盘如
  `D:\Python39`）；

* 安装完成后，勾选 **Disable path length limit**（移除路径长度限制，避免后续依赖安装报错），点击 **Close**。

#### 2.3. 验证 Python 安装成功

* 打开 **新的命令提示符（CMD）** 或 **PowerShell**（旧窗口已关闭，确保环境变量生效）；

* 执行以下命令检查版本：

```
python --version
```

* 若输出类似 `Python 3.9.13`（版本 ≥ 3.8），说明安装成功；

* 检查 pip（Python 包管理工具）是否正常：

```
pip --version
```

* 若输出类似 `pip 22.0.4 from ... (python 3.9)`，说明 pip 正常。


## 三、安装项目依赖（requirements.txt）

### 1. 进入项目根目录

* 到根目录下 [..](..)/ 

* 在 CMD/PowerShell 中通过 `cd` 命令切换到该目录（示例路径需替换为实际路径）：
* （验证）执行 `dir` 命令，若能看到 `requirements.txt` 文件，说明路径正确。

### 3. 安装依赖包

确保终端已进入项目根目录，执行以下命令安装[requirements.txt](../requirements.txt)中的所有依赖：

```
pip install -r requirements.txt -i https://mirrors.huaweicloud.com/repository/pypi/simple/ --trusted-host mirrors.huaweicloud.com
```

* 若无法在线安装, 请到[python依赖仓](https://pypi.org/)下载requirements.txt文件中的所有依赖

### 4. 验证依赖安装成功

无报错提示即安装完成，执行以下命令查看已安装的依赖列表：

```
pip list
```

* 列表中应包含[requirements.txt](../requirements.txt)中指定的所有包及其版本。
