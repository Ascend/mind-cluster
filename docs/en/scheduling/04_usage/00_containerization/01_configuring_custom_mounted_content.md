# (Optional) Configuring Custom Mounted Content<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime automatically mounts the driver and all contents from the base configuration files `/etc/ascend-docker-runtime.d/base.list` and `/etc/ascend-docker-runtime.d/ub_driver.list` by default. Additionally, it checks the paths `/usr/local/Ascend/driver/lib64/common`, `/usr/local/Ascend/driver/lib64/driver`, and `/usr/lib64` by default. If they exist, they are automatically added to the `LD_LIBRARY_PATH` environment variable, making the npu-smi info command available by default. If you need to mount all paths in the file, skip this section. If you do not need to mount all contents from the base configuration files `base.list` and `ub_driver.list`, a custom configuration file can be added to reduce the mounted content. The custom configuration file mounts must be based on the `base.list` file. The steps are as follows:

1. Enter the configuration file directory.

    ```shell
    cd /etc/ascend-docker-runtime.d/
    ```

    `base.list` and `ub_driver.list` already exist in this directory. The content is the default mounted content of Ascend Docker Runtime. For details, see [Content Mounted by Ascend Docker Runtime](../../07_references/05_appendix.md#content-mounted-by-ascend-docker-runtime). In principle, you are not allowed to modify the `base.list` and `ub_driver.list` files.

2. Create a new configuration file. The file name can be customized, for example, `hostlog.list`.

    ```shell
    vi hostlog.list
    ```

3. Write the files or directories to be mounted into `hostlog.list`, then save and exit.
4. Run the command to make `hostlog.list` take effect. Example:

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} /bin/bash
    ```

    >[!NOTE]
    >- `ub_driver.list` is a mount configuration file for the UB driver. Only in UB scenarios will the contents of `ub_driver.list` be mounted by default. If you do not need to mount the contents of `ub_driver.list`, set the `ASCEND_UB_DRV_MOUNT` parameter to `False`. It is not recommended to mount the contents of `ub_driver.list` through the `ASCEND_RUNTIME_MOUNTS` parameter.
    >- `{image-name:tag}` is the image name and tag. For other parameters, see [Table 1](./02_usage_on_the_docker_client.md#parameter-description).
