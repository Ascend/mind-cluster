# Container Manager

# 组件介绍
支持故障管理和故障芯片及故障容器的自动恢复。

# 编译

1.  通过git拉取源码，并切换master分支。

    示例：源码放在/home/mind-cluster/component/container-manager目录下

2.  执行以下命令，进入构建目录，选择构建脚本执行，在“output”目录下生成二进制container-manager、故障码配置文件faultCode.json。

    ```bash
    cd /home/mind-cluster/component/container-manager/build/
    chmod +x build.sh
    ./build.sh
    ```

3.  执行以下命令，查看output生成的软件列表。

    ```bash
    ls /home/mind-cluster/component/container-manager/output
    ```
    回显示例如下：
    ```
    container-manager faultCode.json
    ```

# 使用

1.  可通过以下命令获取帮助信息。
    ```bash
    ./container-manager -h
    ./container-manager -help
    ```
    回显示例如下：
    ```
    Container Manager, supports fault management and automatic recovery.
    
    Usage: [OPTIONS...] COMMAND
    
    Options:
	    -h,-help	Print help information
	    -v,-version	Print version information

    Commands:
        run         Run container manager
        status      Display container status information and container abnormal information
    ```

2. 可通过以下命令查看版本信息。
    ```bash
    ./container-manager -v
    ./container-manager -version
    ```
   回显示例如下：
    ```
    container-manager version: v7.3.0_linux-x86-64
    ```

3. 可通过以下命令启动故障管理和故障芯片及故障容器的自动恢复功能。
    ```bash
    ./container-manager run -ctrStrategy=<ctrStrategy> -sockPath=<sockPath>
    ```

4. 查看容器恢复进度及提示信息。
    ```bash
    ./container-manager status -containerID=<containerID>
    ```
   回显示例如下：
    ```
    +==============================================================================================+
    | Container ID             : e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  |
    | Container Status         : resuming                                                          |
    | Container Description    : The device has been recovered, but the container failed to be     |
    |                            resumed. Please manually pull up the container                    |
    +==============================================================================================+
    ```