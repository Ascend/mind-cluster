# Setting Up the Development and Compilation Environment

This document describes how to set up a development and compilation environment, and provides the compilation entry point for each component. The following assumes that the developer has cloned the [MindCluster source code](https://gitcode.com/Ascend/mind-cluster).

## Environment Setup

### Method 1: Using the Official Development Container

The `.devcontainer` directory in the project root directory provides a `Dockerfile` and a `devcontainer.json`, which come with Go, Python, gcc, cmake, musl, swig, the Volcano source code, and so on preinstalled, and with environment variables such as `GOPROXY`, `GOTOOLCHAIN`, and `TZ` already configured. It works out of the box.

1. Install Docker locally and install the `Dev Containers` extension in VS Code.
2. Open the project root directory and choose "Reopen in Container" as the opening method.
3. Wait until the image build is complete, and then you can develop and compile inside the container.

>[!NOTE]
>
> To adjust the toolchain version or add dependencies or extensions, modify the `Dockerfile` and `devcontainer.json` in the `.devcontainer` directory, and then rebuild the container.

### Method 2: Manually Preparing the Environment

Taking the openEuler 24.03 LTS system as an example, manually install the basic toolchain, Go, Python component dependencies, yq, Helm packaging tools, and so on. For details, see [openEuler Environment Preparation for Building from Scratch](https://gitcode.com/Ascend/mind-cluster/blob/master/build/README.md#openeuler%E4%BB%8E%E9%9B%B6%E6%9E%84%E5%BB%BA%E7%8E%AF%E5%A2%83%E5%87%86%E5%A4%87).

## Compilation

>[!NOTE]
>
>- Before compiling or running test cases, it is recommended to run `go mod tidy` once in the corresponding component directory (the directory containing go.mod) to complete and correct the dependencies in go.mod and go.sum, so as to avoid compilation or test failures caused by missing or inconsistent dependencies.
>- Scripts pulled from Git may contain Windows (CRLF) line breaks in some environments, and executing them directly will report line break exceptions such as `\r` or bad interpreter. Before compiling or testing, convert the line break format of the scripts with `dos2unix`. For example, run `dos2unix build.sh` for the compilation file build.sh.

### Compiling All Components

1. Change the value of the `mind-cluster-version` field in `build/service_config.ini` to the target version.
2. Go to the `build` directory and run `./build_all.sh $GOPATH` to compile all components at once.
3. After compilation, the build artifacts of each component are located in the `component/<component>/output/` directory (for example, `component/ascend-device-plugin/output/`).

### Compiling a Single Component

For the detailed compilation steps of each component, see the README.md in the `component/<component>/` directory (for example, `component/ascend-device-plugin/`).
