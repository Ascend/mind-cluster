# (Optional) Configuring Custom Mounted Content<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime automatically mounts the driver and all contents listed in the base configuration file `/etc/ascend-docker-runtime.d/base.list` by default. If you need to mount all paths in this file, you can skip this section. If you do not need to mount all contents in the `base.list` file, you can create a custom configuration file to reduce the mounted content. The custom configuration file must be based on the `base.list` file. The procedure is as follows:

1. Enter the configuration file directory.

    ```shell
    cd /etc/ascend-docker-runtime.d/
    ```

    `base.list` already exists in this directory. Its content is the default mounted content of Ascend Docker Runtime. For details, see [Content Mounted by Ascend Docker Runtime](../../appendix.md#content-mounted-by-ascend-docker-runtime). In principle, you are not allowed to modify the `base.list` file.

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
    >- For descriptions of the `ASCEND_VISIBLE_DEVICES` and `ASCEND_RUNTIME_MOUNTS` parameters, see [Table 1](./02_usage_on_the_docker_client.md#parameter-description).
    >- Custom mounted content is restricted by the default mount whitelist of Ascend Docker Runtime. For details, see [Default Mount Whitelist of Ascend Docker Runtime](../../appendix.md#default-mount-whitelist-of-ascend-docker-runtime).
