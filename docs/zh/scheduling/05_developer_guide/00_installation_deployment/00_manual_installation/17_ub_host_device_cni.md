# UB Host Device CNI

UB Host Device CNI以二进制方式安装在各计算节点的CNI插件目录，无需单独部署Pod。容器化网络使用ROCE网卡时需安装该插件。

## 安装前准备

- 请参见[获取软件包](./00_obtaining_software_packages.md)获取与节点CPU架构匹配的UB Host Device CNI软件包，并解压获取`ub-host-device`二进制。使用源码编译时，请参见[开发与编译环境搭建](../../04_development_environment.md)准备环境，在源码根目录执行以下命令，产物位于`component/ub-host-device-cni/output/ub-host-device`。

    ```shell
    cd component/ub-host-device-cni
    bash build/build.sh
    ```

## 操作步骤

1. 将`ub-host-device`分发至所有运行1825业务的计算节点，以root用户在二进制所在目录执行以下命令。

    ```shell
    cp ub-host-device /opt/cni/bin/ub-host-device
    chmod 0755 /opt/cni/bin/ub-host-device
    ```

    `/opt/cni/bin`需与Multus及容器运行时使用的CNI二进制目录一致。

2. 验证二进制已安装。

    ```shell
    ls -l /opt/cni/bin/ub-host-device
    ```

    NAD中的`type`为`ub-host-device`，二进制名称必须与其一致。
