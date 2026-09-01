# 开发与编译环境搭建

本文介绍开发者如何搭建开发与编译环境，并提供各组件的编译入口，下文默认开发者已经克隆了[MindCluster源码](https://gitcode.com/Ascend/mind-cluster)。

## 环境搭建

### 方式一：使用官方开发容器

项目根目录 `.devcontainer/` 下已提供 `Dockerfile` 与 `devcontainer.json`，预装了 Go、Python、gcc、cmake、musl、swig及 volcano 源码等，并配置好了 `GOPROXY`、`GOTOOLCHAIN`、`TZ` 等环境变量，开箱即用。

1. 本地安装 Docker，并在 VS Code 中安装 `Dev Containers` 插件。
2. 打开项目根目录，选择打开方式为“在容器中重新打开（Reopen in Container）”。
3. 等待镜像构建完成，即可在容器内进行开发与编译。

>[!NOTE]
>
>- 如需调整工具链版本、追加依赖或插件，可直接修改 `.devcontainer/Dockerfile` 与 `.devcontainer/devcontainer.json`，然后重新构建容器。

### 方式二：手动准备环境

以 openEuler 24.03 LTS 系统为例，手动安装基础工具链、Go 1.26、Python 组件依赖及 yq、helm 打包工具等，请参阅[openeuler从零构建环境准备](https://gitcode.com/Ascend/mind-cluster/blob/master/build/README.md#openeuler%E4%BB%8E%E9%9B%B6%E6%9E%84%E5%BB%BA%E7%8E%AF%E5%A2%83%E5%87%86%E5%A4%87)。

## 编译

>[!NOTE]
>
>- 编译或执行用例前，建议先在对应组件（含 go.mod 的目录）下执行一次 `go mod tidy`，补全/校正 go.mod 与 go.sum 的依赖，避免因依赖缺失或不一致导致编译、测试失败。
>- 从 git 拉取的脚本在部分环境下可能带有 Windows（CRLF）换行符，直接执行会报 `\r`、bad interpreter 等换行符异常。编译、测试前建议先用 `dos2unix` 转换脚本换行格式，例如：对编译文件build.sh执行 `dos2unix build.sh`。

### 编译所有组件

1. 修改 `build/service_config.ini` 中 `mind-cluster-version` 字段值为目标版本。
2. 进入 `build` 目录，执行 `./build_all.sh $GOPATH`，一次编译全部组件。
3. 编译完成后，各组件的产物位于 `component/<组件>/output/` （如 `component/ascend-device-plugin/output/`）目录下。

### 编译单个组件

各组件的详细编译步骤，请参考 `component/<组件>/` （如 `component/ascend-device-plugin/`）目录下的 README.md。
