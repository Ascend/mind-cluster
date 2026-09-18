# Upgrade

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:28.425Z pushedAt=2026-08-24T02:22:41.521Z -->

This section describes the version upgrade process of ascend-fd-tk.

1. Upgrade the Whl package.

   - If the tool version in the current environment differs from the upgrade version, directly use the `--upgrade` parameter to upgrade the tool.

        ```bash
        pip3 install --upgrade ascend_faultdiag_toolkit-{new_version}-py3-none-any.whl
        ```

   - If the tool version in the current environment is the same as the upgrade version, use the `--force-reinstall` parameter to install the new version.

        ```bash
        pip3 install --force-reinstall ascend_faultdiag_toolkit-{new_version}-py3-none-any.whl
        ```

    Example of a successful upgrade:

    ```txt
    Successfully installed ascend-faultdiag-toolkit-{version}
    ```

2. Verify the upgrade.

    Run the `about` command to view the version information. If the command succeeds and returns `MindCluster ascend-faultdiag-toolkit diagnostic tool version: {version}`, the upgrade is successful.

    ```bash
    ascend-fd-tk about
    ```
