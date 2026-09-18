# Uninstallation

<!-- md-trans-meta sourceCommit=3ecb02f7ecbf391029443f4d074a7e520371f838 translatedAt=2026-08-24T02:26:54.454Z pushedAt=2026-08-24T02:51:43.239Z -->

## Uninstallation for Package Installed via Command-Line

Run the following command as the installation user of the component:

```shell
pip3 uninstall ascend-faultdiag -y
```

After the uninstallation is complete, run `pip3 list | grep ascend-faultdiag` to verify. The absence of `ascend-faultdiag` indicates successful uninstallation.

## Uninstallation for Package Installed via MindCluster Ascend Deployer

Run the following command as the component's installation user to delete the binary file:

```shell
rm /usr/local/bin/ascend-fd
```

## Cleaning Up Residual Files

The `~/.ascend_faultdiag` directory stores logs and other information and is not automatically removed during uninstallation. If it is no longer needed, delete it manually:

```shell
rm -rf ~/.ascend_faultdiag
```
