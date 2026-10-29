# （可选）配置自定义挂载内容<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime会为用户默认挂载驱动以及基础配置文件“/etc/ascend-docker-runtime.d/base.list”和“/etc/ascend-docker-runtime.d/ub_driver.list”中的全部内容，并且Ascend Docker Runtime默认会检查路径“/usr/local/Ascend/driver/lib64/common”、“/usr/local/Ascend/driver/lib64/driver”和“/usr/lib64”，如果存在，则会自动添加到LD_LIBRARY_PATH环境变量中，以便npu-smi info命令自动可用。若用户需要挂载文件里的全部路径，则跳过本小节；若用户不需要挂载基础配置文件base.list和ub_driver.list中的全部内容时，可新增自定义配置文件，减少挂载的内容。自定义配置文件挂载内容须基于base.list文件，操作如下：

1. 进入配置文件目录。

    ```shell
    cd /etc/ascend-docker-runtime.d/
    ```

    该目录下已存在基础配置文件base.list和ub_driver.list，内容即Ascend Docker Runtime默认挂载内容，具体可参见[Ascend Docker Runtime默认挂载内容](../../07_references/05_appendix.md#ascend-docker-runtime默认挂载内容)，原则上不允许用户修改base.list和ub_driver.list文件。

2. 创建新的配置文件，文件名可自定义（如hostlog.list）。

    ```shell
    vi hostlog.list
    ```

3. 将需要挂载的文件或目录写入hostlog.list，保存并退出。
4. 执行命令，使自定义配置文件hostlog.list生效。示例如下：

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} /bin/bash
    ```

    >[!NOTE]
    >- ub_driver.list属于UB驱动的挂载配置文件，仅当UB总线场景下，才会默认挂载ub_driver.list中的内容。若用户不需要挂载ub_driver.list中的内容，可以指定ASCEND_UB_DRV_MOUNT参数为"False"，不建议通过ASCEND_RUNTIME_MOUNTS参数挂载ub_driver.list中的内容。
    >- 示例中的\{image-name:tag\}为镜像名称与标签，其他参数说明请参见[表1](./02_usage_on_the_docker_client.md#参数说明)。
    >- 支持配置主机路径、主机路径:容器内路径两种方式。

## 使用约束<a name="section178269291000015"></a>

- 自定义配置文件须存放于“/etc/ascend-docker-runtime.d/”目录下，且文件名须为小写；通过ASCEND_RUNTIME_MOUNTS参数引用多个配置文件时，文件名之间使用逗号分隔，参数取值总长度不能超过128个字符，超出后仅挂载默认内容。
- 单个配置文件中的挂载条目（不含空行和以“#”开头的注释行）不能超过128条。
- 主机源路径仅支持字母、数字和“._-/~”字符，不支持包含空格或其他特殊字符（如“@”、“$”等）的路径。
- 主机源路径不存在时，该条目会被自动跳过。若挂载内容未生效，请检查主机源路径是否存在。
- 主机源路径支持使用通配符（*、?、[）匹配多个文件或目录，但通配符路径不支持同时配置容器内路径；通配符匹配到的软链接，其目标文件必须与匹配项位于同一目录，否则该条目会被跳过。
- 默认不允许挂载软链接。若配置文件中包含软链接，须指定ASCEND_ALLOW_LINK=True；软链接及其目标文件的属主必须为root，否则该条目会被跳过。
- 容器内路径必须为以“/”开头的绝对路径，不能为“/”，不能包含“..”或“:”，长度不能超过1023个字符。
- 自定义挂载内容均以只读方式挂载至容器，不支持配置为可写挂载。
