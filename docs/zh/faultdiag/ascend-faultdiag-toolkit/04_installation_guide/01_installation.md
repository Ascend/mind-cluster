# 安装

## 安装前必读

- ascend-fd-tk 要求 Python 版本不低于 3.8，安装前请检查 Python 版本是否满足要求。
- 安装前请检查磁盘剩余空间是否充足（建议 5GB 以上）。
- 安装前请确认网络连接正常，安装过程需联网下载三方依赖库。
- 安装过程已声明相关三方库依赖，并验证了最低兼容版本。

## 安装步骤

### 1. 获取 WHL 包

#### 方式 1：从发行版本下载

访问 [MindCluster 发行版本](https://gitcode.com/Ascend/mind-cluster/releases)，下载对应版本的链路故障诊断组件安装包。

| 软件包 | 子文件 | 说明 |
|--------|--------|------|
| `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip` | `ascend_faultdiag_toolkit-{version}-py3-none-any.whl` | `{version}` 为软件包版本号，默认为最新版本。<br/>`{arch}` 为软件包架构，分为 x86_64 和 aarch64，ascend-fd-tk 不区分架构 |

```bash
# 解压
unzip Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip
# 获取 WHL 包：ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

#### 方式 2：源码编译生成

使用 pip 安装编译所需三方依赖库：

```bash
pip3 install setuptools>=60.3.0 wheel>=0.45.1
```

克隆源码并编译打包：

```bash
git clone https://gitcode.com/Ascend/mind-cluster.git
cd mind-cluster/component/ascend-faultdiag/toolkit_src
# 指定版本号，编译打包
python3 setup.py --version {version} bdist_wheel
```

> **说明**：`{version}` 为版本号，需替换为实际版本，如 `v1.0.0`。

生成的 WHL 包位于 `dist/` 目录下：

```text
dist/ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

### 2. 安装 WHL 包

安装所需三方依赖库：

| 依赖 | 版本要求 | 用途 |
|------|----------|------|
| `paramiko` | \>= 3.0.0 | SSH 在线采集 |
| `scp` | \>= 0.14.0 | 远程文件传输（用于 BMC 日志获取） |
| `cryptography` | \>= 41.0.0 | 连接配置加密 |
| `openpyxl` | \>= 3.1.0 | `.xlsx` 文件解析、Excel 报告生成 |

> **说明**：安装过程中会自动联网下载所需的三方依赖库。

执行安装操作：

```bash
pip3 install ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

### 3. 验证安装

执行 `about` 命令查看版本信息，若执行成功并回显 `MindCluster ascend-faultdiag-toolkit诊断工具版本：{version}`，则说明安装成功。

```bash
ascend-fd-tk about
```
