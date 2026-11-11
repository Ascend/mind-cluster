# （可选）配置自定义挂载内容<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime会为用户默认挂载驱动以及基础配置文件“/etc/ascend-docker-runtime.d/base.list”和“/etc/ascend-docker-runtime.d/ub_driver.list”中的全部内容，并且Ascend Docker Runtime默认会检查路径“/usr/local/Ascend/driver/lib64/common”、“/usr/local/Ascend/driver/lib64/driver”和“/usr/lib64”，如果存在，则会自动添加到LD_LIBRARY_PATH环境变量中，以便npu-smi info命令自动可用。base.list和ub_driver.list中的内容正常挂载至容器，否则影响容器内相关功能使用，文件中已有的默认挂载路径配置项无需修改。若用户需要新增挂载其他文件或目录，可创建自定义配置文件，操作如下：

1. 进入配置文件目录。

    ```shell
    cd /etc/ascend-docker-runtime.d/
    ```

    该目录下已存在基础配置文件base.list和ub_driver.list，内容即Ascend Docker Runtime默认挂载内容，具体可参见[Ascend Docker Runtime默认挂载内容](../../07_references/05_appendix.md#ascend-docker-runtime默认挂载内容)，原则上不允许用户修改base.list和ub_driver.list文件。

2. 创建新的配置文件，文件名可自定义（如hostlog.list）。

    ```shell
    vi hostlog.list
    ```

3. 将需要挂载的文件或目录写入hostlog.list，保存并退出。支持的配置方式如下表所示。

    | 配置方式 | 条目格式 | 示例 | 挂载结果 |
    |---|---|---|---|
    | 挂载文件或目录 | 主机路径 | /var/log/npu | 挂载到容器内相同路径 |
    | 挂载到容器内指定路径 | 主机路径:容器内路径 | /var/log/npu:/home/npu/log | 挂载到容器内“/home/npu/log”路径 |
    | 通配符匹配多个文件或目录 | 含通配符的主机路径 | /usr/local/Ascend/driver/lib64/*.so | 各命中项挂载到容器内相同路径 |
    | 通配符命中项挂载到容器内指定目录 | 含通配符的主机路径:容器内路径 | /usr/lib64/libummu*:/usr/lib64 | 各命中项以原文件名挂载到该目录下，如“/usr/lib64/libummu.so” |

    >[!NOTE]
    >- 使用“主机路径:容器内路径”方式时，容器内路径须写全目标路径：源路径为文件时，容器内路径须包含目标文件名（如/a/b.txt:/c/dir/b.txt）；源路径为目录时，容器内路径为目标目录路径（如/a/logdir:/c/logdir）。
4. 执行命令，使自定义配置文件hostlog.list生效。示例如下：

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} /bin/bash
    ```

    >[!NOTE]
    >- ub_driver.list属于UB驱动的挂载配置文件，仅当UB总线场景下，才会默认挂载ub_driver.list中的内容。若用户不需要挂载ub_driver.list中的内容，可以指定ASCEND_UB_DRV_MOUNT参数为"False"，不建议通过ASCEND_RUNTIME_MOUNTS参数挂载ub_driver.list中的内容。
    >- 示例中的\{image-name:tag\}为镜像名称与标签，其他参数说明请参见[表1](./02_usage_on_the_docker_client.md#参数说明)。
    >- 挂载条目的配置方式与约束请参见下方“使用约束”。

## 使用约束<a name="section178269291000015"></a>

- 自定义配置文件须存放于“/etc/ascend-docker-runtime.d/”目录下，且文件名须为小写；通过ASCEND_RUNTIME_MOUNTS参数引用多个配置文件时，文件名之间使用逗号分隔，参数取值总长度不能超过128个字符，超出后仅挂载默认内容。
- 单个配置文件中的挂载条目（不含空行和以“#”开头的注释行）不能超过128条。
- 主机源路径仅支持字母、数字和“._-/~”字符，不支持包含空格或其他特殊字符（如“@”、“$”等）的路径。
- 主机源路径不存在时，该条目会被自动跳过。若挂载内容未生效，请检查主机源路径是否存在。
- 主机源路径支持使用通配符（*、?、[）匹配多个文件或目录。
- “主机路径:容器内路径”方式下，容器内路径须为以“/”开头的绝对路径，长度不能超过1023个字符；其类型须与主机源路径一致（文件源对应文件目标，目录源对应目录目标），目标路径不存在时按主机源类型创建。
- 仅支持挂载条目本身（最后一级）为软链接，且须指定ASCEND_ALLOW_LINK=True；软链接及其目标文件的属主必须为root。
- 自定义挂载内容均以只读方式挂载至容器，不支持配置为可写挂载。
<!-- npu="950" id1 -->
- UB网卡默认驱动挂载清单及校验规则请参见[使用前必读](./00_before_you_start.md#使用说明)。
<!-- end id1 -->
