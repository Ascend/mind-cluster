# Installation

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:42.757Z pushedAt=2026-08-24T02:22:41.527Z -->

## Before You Begin

- The ascend-fd-tk tool requires Python 3.8 or later. Before installation, check whether the Python version meets the requirement.
- Before installation, check whether the remaining drive space is sufficient (5 GB or more is recommended).
- Before installation, confirm that the network connection is normal. The installation process requires an internet connection to download third-party dependency libraries.
- The installation process has declared the relevant third-party library dependencies and verified the minimum compatible versions.
- The installation steps on Windows are similar to those on Linux. Refer to the following steps (on Linux, run the `unzip` command to decompress; on Windows, decompress manually). Using the tool on Windows is a beta feature and is not recommended for production environments.

## Installation Steps

### 1. Obtain the Whl Package

Obtain the Whl package in either of the following ways.

**Method 1: Download from a release**

| Package                                                | Subfile                                                         | Description          | Link                                                         |
|-------------------------------------------------------|-----------------------------------------------------------------|-------------|--------------------------------------------------------------|
| `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip` | `ascend_faultdiag_toolkit-{version}-py3-none-any.whl`          | Link fault diagnosis component installation package | [Download link](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0) |

>[!NOTE]
>
> - `{version}` is the software package version number, which defaults to the latest version.
> - `{arch}` is the software package architecture, which can be `x86_64` or `aarch64`. Modify it as needed. You can run the `arch` command to check it.
> - The ascend-fd-tk Whl package is architecture-independent.
> - To prevent the software package from being maliciously tampered with during transfer or storage, you are advised to verify the checksum of the software package. To verify the checksum, see [Software Package Checksum Verification](#reference).

Decompress to obtain the Whl package:

```bash
# Decompress
unzip Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip
# Obtain the Whl package: ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

**Method 2: Build from source code**

Use `pip` to install the third-party dependency libraries required for compilation:

```bash
pip3 install 'setuptools>=60.3.0' 'wheel>=0.45.1'
```

Clone the source code and compile it into a package:

```bash
git clone https://gitcode.com/Ascend/mind-cluster.git
cd mind-cluster/component/ascend-faultdiag/toolkit_src
# Specify the version number before building and packaging.
python3 setup.py --version {version} bdist_wheel
```

>[!NOTE]
>
> - `{version}` is the version number and must be replaced with the actual version, for example, `v1.0.0`.
> - The Whl package name is generated according to the Wheel standard: `ascend_faultdiag_toolkit-{remove the 'v' prefix from version}-py3-none-any.whl`, for example, `ascend_faultdiag_toolkit-1.0.0-py3-none-any.whl`

The generated Whl package is located in the `dist/` directory:

```text
dist/ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

### 2. Installing the Whl Package

Install the following third-party dependency libraries. During the installation process, the required third-party dependency libraries are automatically downloaded from the network.

| Dependency | Version Requirement | Purpose |
|------|----------|------|
| `paramiko` | \>= 3.0.0 | Online collection via SSH |
| `scp` | \>= 0.14.0 | Remote file transfer (for BMC log retrieval) |
| `cryptography` | \>= 41.0.0 | Connection configuration encryption |
| `openpyxl` | \>= 3.1.0 | `.xlsx` file parsing and Excel report generation |

Install the package:

```bash
pip3 install ascend_faultdiag_toolkit-{version}-py3-none-any.whl
```

Output if installation succeeds

```txt
Successfully installed ascend-faultdiag-toolkit-{version}
```

### 3. Verify the Installation

Run the `about` command to view the version information. If the command is executed successfully and returns `MindCluster ascend-faultdiag-toolkit diagnostic tool version: {version}`, the installation is successful.

```bash
ascend-fd-tk about
```

## Reference

**Software package checksum verification steps**:

1. Download the [Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0) file.

2. Place `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum` and `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip` in the same directory, and run the following command to perform verification.

    ```bash
    sha256sum -c Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip.sha256sum
    ```

3. Check the verification result.

    The following output example indicates that the software package has passed verification.

    ```bash
    Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip: OK
    ```
