# 所有组件统一编译说明

- **[编译](#编译)**
- **[openEuler从零构建环境准备](#openeuler从零构建环境准备)**
- **[mindio组件编译依赖](#mindio组件编译依赖)**
- **[自动拉取源码失败](#自动拉取源码失败)**

# 编译

1. 拉取mind-cluster整体源码，例如放在/home目录下。

2. 修改组件版本配置文件service_config.ini中mind-cluster-version字段值为所需编译版本，默认值如下：

        mind-cluster-version=26.1.0

3. 执行以下命令，进入/home/mind-cluster/build目录，选择构建脚本执行：

    **cd /home/mind-cluster/build**

        dos2unix *.sh && chmod +x *.sh

        ./build_all.sh $GOPATH

4. 执行完成后进入/home/mind-cluster，在component下的各组件的“output”目录下生成编译完成的文件。

5. 此处使用的go版本为1.26。

6. 若只需要编译部分组件，而非全量组件时，请在component下的各组件的“build”目录下执行build.sh脚本，若组件编译失败，可以参考该组件的README.md文件。

7. 若需要执行用例，请在component下的各组件的“build”目录下执行test.sh脚本。

> 建议：编译或执行用例前，最好先在对应组件（含 go.mod 的目录）下执行一次 `go mod tidy`，以补全/校正 go.mod 与 go.sum 的依赖，避免因依赖缺失或不一致导致编译、测试失败。
>
> 建议：从 git 拉取的脚本在部分环境下可能带有 Windows（CRLF）换行符，直接执行会报 `\r`/`bad interpreter` 等换行符异常。编译、测试前建议先用 `dos2unix` 转换脚本换行格式，例如：`find . -name '*.sh' -print0 | xargs -0 dos2unix`（或对目标目录执行 `dos2unix *.sh`）。

# openEuler从零构建环境准备

以下为在 **openEuler 24.03 LTS** 基础镜像/系统上从零全量构建（build_all.sh，涵盖12个组件及 helm-deploy-tool 打包工具）实测所需的依赖与环境改动。已在 [mindio组件编译依赖](#mindio组件编译依赖) 单列的项这里不再重复。

## 1. 系统工具链（dnf）

裸镜像需先安装以下基础包（Go/C++/Python 组件均会用到）：

    dnf install -y gcc gcc-c++ make cmake git wget unzip dos2unix \
        python3 python3-pip python3-devel swig numactl-devel zlib-devel

实测版本参考：gcc 12.3.1、make 4.4.1、cmake 3.27.9、git 2.43、python3 3.11.6。

> 注：`zlib-devel` 为 mindio 链接期依赖（`-lz`），缺失时 ockiod 等二进制链接阶段会报 `ld.gold: cannot find -lz`。
>
> 注：dnf 安装的 swig 版本为 4.1.1，仅能满足非 mindio 组件；**mindio 的 abi3 wheel 要求 SWIG>=4.2.0**，需另用 pip 安装（`pip3 install "swig>=4.2,<4.4"`），详见 [mindio组件编译依赖](#mindio组件编译依赖)。

## 2. Go 1.26（重点：dnf 版本不满足）

本仓库各组件 go.mod 要求 **go 1.26**，但 openEuler dnf 仓库当前仅提供 golang 1.21.4，**无法直接用 dnf 安装满足版本的 Go**，需手动安装官方发行版：

    wget https://mirrors.aliyun.com/golang/go1.26.4.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

建议同时配置以下环境变量（写入 /etc/profile.d/go.sh）：

    export GOPROXY=https://goproxy.cn,direct   # 加速模块下载
    export GOSUMDB=off
    export GOTOOLCHAIN=local                    # 已装 1.26.x，禁止 go 再自动下载工具链

> 注：build_all.sh / 各组件 build.sh 内部已使用 `go build -mod=mod`，构建阶段无需额外设置；`-mod=mod` 主要在手动执行单元测试时需要（见下）。

## 3. Python 组件依赖（ascend-faultdiag、taskd）

这两个组件为 Python 工程，构建产物为 wheel 包，需按其 requirements.txt 预装依赖，否则构建会报 `ModuleNotFoundError` 或 `invalid command 'bdist_wheel'`：

    pip3 install wheel setuptools \
        scikit-learn pandas numpy ply paramiko scp cryptography openpyxl \
        grpcio protobuf pyOpenSSL

（可通过 `-i https://mirrors.aliyun.com/pypi/simple/` 使用国内 PyPI 镜像加速。）

## 4. helm-deploy-tool 打包依赖（yq、helm）

build_all.sh 末尾会构建 helm-deploy-tool（打包 chart），该组件依赖 **yq** 与 **helm** 两个命令行工具，openEuler 基础镜像与上述依赖清单均不含，缺失时构建会报 `yq: command not found` / `helm: command not found`：

- **yq**：脚本使用 mikefarah 版 yq 语法（如 `yq eval "select(documentIndex == $i)"`），务必安装该实现（非 python-yq）。
- **helm**：`build.sh` 通过 `helm package` 打包 app/app-crds。

裸镜像无 github release 直连时，可用 go install 安装（走 GOPROXY 镜像）：

    export GOBIN=/usr/local/bin
    go install github.com/mikefarah/yq/v4@v4.44.3
    go install helm.sh/helm/v3/cmd/helm@v3.15.4

安装后可执行 `yq --version`、`helm version` 验证。

## 5. 单元测试执行

各组件用例可在其 build 目录执行 test.sh；如需直接用 go 命令逐模块运行 Go 单测，需注意：

    # 执行用例前先补全依赖，避免 go.sum 不全导致测试拉取/校验失败
    go mod tidy
    # 执行用例前需将时区配置为东八区，否则时间相关用例的预期值会因时区差异而失败
    export TZ=Asia/Shanghai
    # gomonkey 打桩要求禁用内联
    export GOFLAGS="-gcflags=all=-l -mod=mod"
    go test -count=1 ./...

> 说明：部分用例（如 clusterd 的 `TestReadableMsTime`）的预期值按东八区（UTC+8）时间写定，而容器默认时区通常为 UTC，未设置 `TZ=Asia/Shanghai` 时会因时区相差 8 小时导致断言失败，故执行测试前务必配置时区为东八区。
>
> 说明：暂时不跑ascend-faultdiag-online包的单元测试。

## 6. 容器内构建注意（Docker + seccomp）

在较旧版本 Docker（如 20.10.x）容器内构建时，其默认 seccomp 配置会拦截 `clone3` 系统调用，导致 openEuler（glibc 2.38）内 dnf/curl 报 `getaddrinfo() thread failed to start`（DNS/线程创建失败）。启动构建容器时需放开 seccomp 并显式指定 DNS：

    docker run -d --name build-env \
        --security-opt seccomp=unconfined \
        --dns 223.5.5.5 --dns 114.114.114.114 \
        -v /path/to/mind-cluster:/workspace -w /workspace \
        openeuler/openeuler:24.03-lts tail -f /dev/null

若 dnf 从官方源下载超时，可将 /etc/yum.repos.d/openEuler.repo 的 baseurl 切换为国内镜像（如 <https://mirrors.huaweicloud.com/openeuler> ）。

# mindio组件编译依赖

编译mindio组件（acp与tft）除go外，还依赖以下系统软件包，缺失时CMake配置阶段会报错（例如`Could NOT find SWIG (missing: SWIG_EXECUTABLE SWIG_DIR)`），请在编译前提前安装：

| 依赖                           | 用途                                                                                                                                             |
|------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| swig（**>=4.2.0**）             | 生成C++与Python接口绑定（acp、tft的python_sdk均通过`find_package(SWIG REQUIRED)`强依赖）。acp构建固定启用`Py_LIMITED_API`以产出跨Python 3.7+通用的abi3 wheel，因此**SWIG必须>=4.2.0**；低于该版本（如openEuler dnf自带的4.1.1）生成的绑定代码会因`PyTuple_GET_SIZE`/`PyTuple_GET_ITEM`宏在受限API下不可用而编译报错。openEuler dnf仅提供4.1.1，需改用pip安装：`pip3 install "swig>=4.2,<4.4"`，并确保`/usr/local/bin/swig`在PATH中优先于dnf版本 |
| python3、python3开发包（头文件）      | CMake通过`find_package(Python3 COMPONENTS Interpreter Development REQUIRED)`查找Python解释器与头文件                                                      |
| python3-pip、wheel            | 打包whl（`python3 setup.py bdist_wheel`），编译脚本会自动安装wheel                                                                                           |
| libnuma开发包                   | acp内存管理及三方库ubs-comm依赖`<numa.h>`（Ubuntu为libnuma-dev，openEuler为numactl-devel）                                                                    |
| cmake（>=3.14.1）、make、gcc/g++ | C++编译构建                                                                                                                                        |
| git                          | 自动拉取三方依赖（ubs-comm、libboundscheck、spdlog）                                                                                                       |
| dos2unix                     | 转换脚本换行格式（build目录脚本会调用）                                                                                                                         |

参考安装命令：

- Ubuntu / Debian：

        sudo apt update
        sudo apt install -y swig python3-dev python3-pip libnuma-dev cmake make g++ git dos2unix
        pip3 install wheel

- openEuler / CentOS / RHEL：

        sudo dnf install -y python3-devel python3-pip numactl-devel cmake make gcc-c++ git dos2unix
        # openEuler dnf 的 swig 为 4.1.1，不满足 abi3 wheel 要求（>=4.2.0），改用 pip 安装
        pip3 install wheel "swig>=4.2,<4.4"

安装完成后可执行`swig -version`验证（应 >=4.2.0，且 `which swig` 指向 pip 安装的 `/usr/local/bin/swig`）。重新编译时直接重跑build.sh即可，脚本会清理旧的Build目录缓存，无需手动删除CMakeCache.txt。

# 自动拉取源码失败

1. 参考以下命令，分别在/opt/buildtools/volcano_opensource/volcano_1.9/与
    /opt/buildtools/volcano_opensource/volcano_1.7/目录下手动拉取Volcano v1.9.0与v1.7.0版本官方开源代码。

        cd /opt/buildtools/volcano_opensource/volcano_1.9/
        git clone -b release-1.9 https://github.com/volcano-sh/volcano.git

    > 注：国内网络从 github 克隆 volcano 限速严重。可改用 gitee 镜像加速：
    >
    > `git clone -b release-1.9 https://gitee.com/mirrors/volcano.git`（v1.7 对应 `-b release-1.7`）。
2. 进入$GOPATH/mind-cluster/ascend-docker-runtime目录，执行ascend-docker-runtime 组件readme
    中编译部分2,3命令手动拉取编译所需包，其中ascend-docker-runtime目录修改为当前目录

3. 编译 mindio（acp 与 tft）时，`build.sh` 会从 `https://atomgit.com/openeuler/libboundscheck.git` 拉取 libboundscheck（`v1.1.16`），该源在国内网络下不稳定，常出现 `curl 56 Recv failure: Connection timed out` 导致构建中断。可采取以下任一方式规避：

    - **复用已下载源码**：acp 与 tft 需要的是同一份 libboundscheck（同为 `v1.1.16`），若其中一个已成功拉取，可将其 `3rdparty/libboundscheck/libboundscheck` 目录复制到另一个组件对应目录下，脚本检测到目录已存在即跳过 clone：

      ```bash
      cp -rf ${MINDIO}/acp/3rdparty/libboundscheck/libboundscheck \
             ${MINDIO}/tft/3rdparty/libboundscheck/
      ```

    - **预置镜像源**：提前手动 clone 到对应 `3rdparty/libboundscheck/libboundscheck` 目录并 `git checkout v1.1.16`（可用可达的镜像地址替换 atomgit），再执行 build.sh。

    > 注：spdlog（`v1.12.0`）已使用 gitcode 镜像（`https://gitcode.com/GitHub_Trending/sp/spdlog.git`），通常无需额外处理。
