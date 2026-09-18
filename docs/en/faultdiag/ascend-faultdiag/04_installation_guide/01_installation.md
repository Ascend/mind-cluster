# Installation and Deployment

<!-- md-trans-meta sourceCommit=cc2549abde258f50a946dbab123b7ff2035ed91e translatedAt=2026-08-24T02:26:50.048Z pushedAt=2026-08-24T02:51:43.237Z -->

## Before You Begin

- Deploy ascend-fd independently on each server. If it is deployed in a shared directory for use by multiple servers, functional exceptions or performance issues may occur.
- ascend-fd requires Python 3.7 or later. To use the performance degradation feature (device resource analysis and network congestion analysis), Python 3.8 or later is required. Before installation, check whether the Python version meets the requirement.
- Before installation, check whether the remaining drive space is sufficient (5 GB or more is recommended) and whether the network connection is normal. The installation process requires network access to download third-party dependency libraries.
- Update security patches or upgrade third-party software to the latest version in a timely manner.

## Installation Methods

ascend-fd supports installation via the Whl package and via MindCluster Ascend Deployer. Installation via the Whl package is recommended.

### Installation via Whl (Recommended)

#### Obtaining the Software Package

You can obtain the software package by downloading the zip package from the open-source community or by compiling the Whl package from source code.

1. Obtain the software package from the open-source community.

    | Software Package                                      | Subfile                                                         | Description                          | Link                                                         |
    |-------------------------------------------------------|-----------------------------------------------------------------|--------------------------------------|--------------------------------------------------------------|
    | `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip` | `ascend_faultdiag-{version}-py3-none-linux_{arch}.whl`          | ascend-fd installation package | [Download link](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0) |

    > [!NOTE]
    >
    > - `{version}` is the software package version. Change it as needed, for example, 26.1.0.
    > - `{arch}` is the software package architecture, which can be `x86_64` or `aarch64`. Change it as needed. You can run the `arch` command to view it.
    > - To prevent the software package from being maliciously tampered with during transfer or storage, you are advised to verify the SUM value of the software package. To verify the SUM value, refer to [Software Package SUM Value Verification](#software-package-sum-value-verification-procedure).

2. Compile the Whl package from source code.

    1. Clone the repository.

        ```shell
        git clone https://gitcode.com/Ascend/mind-cluster
        # Switch to the ascend-faultdiag directory
        cd mind-cluster/component/ascend-faultdiag
        ```

    2. Use `pip3` to install the third-party dependency libraries required for compilation.

        ```shell
        pip3 install -r src/requirements.txt && pip3 install 'setuptools>=60.3.0' 'wheel>=0.45.1'
        ```

    3. To customize version information, create a `service_config.ini` file in the current directory (`mind-cluster/component/ascend-faultdiag`) and fill in the version number, such as `26.1.0`. The version number format must comply with [Python version specifications](https://packaging.pythonlang.cn/en/latest/specifications/version-specifiers/).

    4. Execute the build script.

        ```shell
        bash build/build.sh
        ```

    After running the build script, the `ascend_faultdiag-{version}-py3-none-linux_{arch}.whl` file is generated in the `output/` directory.

#### Installation Procedure

1. It is recommended to run `umask 027` to improve security. For permanent modification, refer to [System Security Configuration](../07_references/03_security.md#system-security-configuration).

2. Upload the obtained software package to any directory on the server (for example, `~/software`).

3. Decompress the software package. If it is compiled from source code, skip this step.

    ```shell
    unzip Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip
    ```

4. Perform the installation.

    ```shell
    pip3 install ascend_faultdiag-{version}-py3-none-linux_{arch}.whl --log ~/.ascend_faultdiag/install.log
    ```

    > If the performance degradation feature is required, install the third-party libraries: `pip3 install 'scikit-learn>=1.3.0' 'pandas>=2.0.3'`

5. It is recommended that you modify the directory permissions to improve security.

    ```shell
    chmod 700 ~/.ascend_faultdiag
    chmod 600 ~/.ascend_faultdiag/*.log
    ```

6. Verify the installation.

    ```shell
    ascend-fd version
    ```

    The version number in the output indicates that the installation is successful, for example:

    ```shell
    ascend-fd v26.1.0
    ```

#### Log Description

- Default directory for runtime logs: `$HOME/.ascend_faultdiag/RUN_LOG/`.
- Default file for operation logs: `$HOME/.ascend_faultdiag/ascend_faultdiag_operation.log`.
- The size of a log file does not exceed 10 MB. When the limit is exceeded, logs are automatically dumped to another log file.
- To change the log storage path, set the environment variable `ASCEND_FD_HOME_PATH`. For details, refer to [Environment Variables](../07_references/01_common_operations.md).

### Installation via MindCluster Ascend Deployer

MindCluster Ascend Deployer supports installation of ascend-fd version 7.1.RC1 or later.

To install ascend-fd on a single server or multiple servers in batches, refer to the "[Installing Ascend Software](https://gitcode.com/Ascend/ascend-deployer/blob/branch_v26.1.0/docs/en/05_installation_and_upgrade/02_install_softwares.md)" section in the *MindCluster Ascend Deployer User Guide*.

## Reference

### Software Package SUM Value Verification Procedure

1. Download the [Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0) file.

2. Place `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum` and `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip` in the same directory, and run the following command to perform verification.

    ```bash
    sha256sum -c Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum
    ```

3. Check the verification result.

    The output is as follows, indicating that the software package has passed verification.

    ```bash
    Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip: OK
    ```
