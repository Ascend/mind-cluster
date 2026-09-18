# Upgrade

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:26:47.556Z pushedAt=2026-08-24T02:51:43.235Z -->

1. Run the following command to install the new software package.

    ```shell
    pip3 install --upgrade ascend_faultdiag-{version}-py3-none-linux_{arch}.whl --log ~/.ascend_faultdiag/install.log
    ```

    > [!NOTE]
    >
    > - `{version}` is the ascend-fd version.
    > - `{arch}` is the software package architecture, which can be `x86_64` or `aarch64`. Modify it based on the actual requirement. You can run the `arch` command to view it.
    > - If the version number is the same, you can use the `--force-reinstall` parameter to force reinstallation.
    > - Ensure that no cleaning or diagnosis task is in progress.

2. Run the following command to verify whether the upgrade is successful.

    ```shell
    ascend-fd version
    ```

    If the new version number is displayed, the upgrade is successful:

    ```shell
    ascend-fd v26.1.0
    ```

    If the upgrade fails, refer to the [uninstallation guide](03_uninstallation.md) to uninstall it, and then refer to the [installation guide](01_installation.md) to reinstall it.
