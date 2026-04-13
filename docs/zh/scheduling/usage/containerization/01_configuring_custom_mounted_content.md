# （可选）配置自定义挂载内容<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime会为用户默认挂载驱动以及基础配置文件“/etc/ascend-docker-runtime.d/base.list”中的全部内容。若用户需要挂载文件里的全部路径，则跳过本小节；若用户不需要挂载基础配置文件base.list中的全部内容时，可新增自定义配置文件，减少挂载的内容。自定义配置文件挂载内容须基于base.list文件，操作如下：

1. 进入配置文件目录。

    ```shell
    cd /etc/ascend-docker-runtime.d/
    ```

    该目录下已存在基础配置文件base.list，内容即Ascend Docker Runtime默认挂载内容，具体可参见[Ascend Docker Runtime默认挂载内容](../../appendix.md#ascend-docker-runtime默认挂载内容)，原则上不允许用户修改base.list文件。

2. 创建新的配置文件，文件名可自定义（如hostlog.list）。

    ```shell
    vi hostlog.list
    ```

3. 将需要挂载的文件或目录写入hostlog.list，保存并退出。
4. 执行命令，使自定义配置文件hostlog.list生效。示例如下：

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag /bin/bash
    ```

    >[!NOTE]  
    >- ASCEND\_VISIBLE\_DEVICES和ASCEND\_RUNTIME\_MOUNTS参数说明，请参见[表1](./02_usage_on_the_docker_client.md)。
    >- 自定义挂载内容受Ascend Docker Runtime的默认挂载白名单限制，具体白名单列表请参见[Ascend Docker Runtime默认挂载白名单](../../appendix.md#ascend-docker-runtime默认挂载白名单)。
