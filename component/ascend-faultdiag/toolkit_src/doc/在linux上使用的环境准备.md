# Linux 环境准备指南

## 一、Python 环境准备

### 1. 检查系统已安装 Python 版本

打开终端（Terminal），执行以下命令检查是否已安装 Python 3.8+（项目推荐最低版本，可根据实际需求调整）：

```
python3 --version
```

* 若输出类似 `Python 3.9.7` 且版本 ≥ 3.8，直接进入 **第二步 安装依赖**。

* 若提示 “command not found” 或版本低于 3.8，需安装 / 升级 Python。

### 2. 安装 / 升级 Python

#### 2.1 适用于 Debian/Ubuntu 系列
```
\# 更新软件包列表

sudo apt update

\# 安装 Python 3.8+ 及 pip（以 Python 3.9 为例，可替换为 3.10 等更高版本）

sudo apt install -y python3.9 python3-pip python3.9-venv

\# （可选）设置 Python 3.9 为默认版本（若系统存在多个 Python 版本）

sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.9 100

sudo update-alternatives --config python3  # 按提示选择 Python 3.9
```

### 2.2 适用于CentOS/RHEL/OpenEuler 系列
```
\# 1. 安装 EPEL 仓库（提供额外 Python 版本）

sudo yum install -y epel-release

\# 2. 安装 Python 3.8+ 及 pip（以 Python 3.9 为例）

sudo yum install -y python39 python39-pip

\# 3. 建立 Python3 软链接（避免与系统默认 Python 2 冲突）

sudo ln -s /usr/bin/python3.9 /usr/bin/python3

sudo ln -s /usr/bin/pip3.9 /usr/bin/pip3

\# 4. 验证软链接是否生效

ls -l /usr/bin/python3 /usr/bin/pip3
```

### 3. 验证 Python 安装成功

```
python3 --version  # 应输出安装的版本号

pip3 --version     # 应输出对应 pip 版本
```

## 二、安装项目依赖（requirements.txt）

### 1. 进入项目根目录

打开终端，通过 `cd` 命令切换到项目源码所在文件夹：


### 2. 安装依赖包

项目根目录下需存在[requirements.txt](../requirements.txt)文件，执行以下命令安装所有依赖：

```
pip3 install -r requirements.txt -i https://mirrors.huaweicloud.com/repository/pypi/simple/ --trusted-host mirrors.huaweicloud.com
```

### 3. 验证依赖安装成功

无报错提示即安装完成，可通过以下命令查看已安装依赖：

```
pip3 list
```
