# (Optional) Configuring Custom Mounted Content<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime automatically mounts the driver and all contents from the base configuration files `/etc/ascend-docker-runtime.d/base.list` and `/etc/ascend-docker-runtime.d/ub_driver.list` by default. Additionally, it checks the paths `/usr/local/Ascend/driver/lib64/common`, `/usr/local/Ascend/driver/lib64/driver`, and `/usr/lib64` by default. If they exist, they are automatically added to the `LD_LIBRARY_PATH` environment variable, making the npu-smi info command available by default. The contents of `base.list` and `ub_driver.list` must be mounted to the container normally; otherwise, related functions inside the container will be affected. The existing default mount path configuration items in these files do not need to be modified. If users need to add mounts for other files or directories, they can create a custom configuration file. The steps are as follows:

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
   >- Supports two configuration methods: host path, and host path:container path.

## Constraints<a name="section178269291000015"></a>

- Custom configuration files must be stored in the `/etc/ascend-docker-runtime.d/` directory, and file names must be lowercase. When multiple configuration files are referenced via the ASCEND_RUNTIME_MOUNTS parameter, separate the file names with commas. The total length of the parameter value must not exceed 128 characters; if exceeded, only the default content is mounted.
- A single configuration file must not contain more than 128 mount entries (excluding blank lines and comment lines starting with "#").
- Host source paths support only letters, digits, and the characters "._-/~". Paths containing spaces or other special characters (such as "@" and "$") are not supported.
- If the host source path does not exist, the entry is automatically skipped. If the mounted content does not take effect, check whether the host source path exists.
- Host source paths support wildcards (*, ?, [) to match multiple files or directories, but wildcard paths cannot be configured together with a container path.
- Only the mount entry itself (the last level) is supported as a symlink, and `ASCEND_ALLOW_LINK=True` must be specified. The owner of the symlink and its target file must be root; otherwise, the entry is skipped.
- Container paths must be absolute paths starting with "/", cannot be "/", cannot contain ".." or ":", and must not exceed 1023 characters in length.
- Custom mount content is always mounted to the container as read-only; writable mounts are not supported.
