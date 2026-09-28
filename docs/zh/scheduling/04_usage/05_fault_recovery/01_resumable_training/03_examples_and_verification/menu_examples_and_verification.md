# 示例与验证

## 使用示例

### （可选）配置组件<a name="ZH-CN_TOPIC_0000002511346449"></a>

如果用户在安装Ascend Device Plugin和NodeD时，已经配置了断点续训相关功能，则可以跳过本章节；若没有配置，则需要对[Ascend Device Plugin](#section14208511958)和[NodeD](#section162092113510)进行相关配置。

**配置Ascend Device Plugin<a name="section14208511958"></a>**

只支持以镜像方式启动Ascend Device Plugin。

1. 根据所使用的故障处理模式，修改Ascend Device Plugin组件的启动YAML，修改如下所示加粗部分。
    1. 重调度模式

        >[!NOTE]
        >在重调度模式下，Ascend Device Plugin的异常也会触发故障重调度。

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin
                         <strong>-volcanoType=true                    # 重调度场景下必须使用Volcano</strong>
                         <strong>-autoStowing=true                    # 该字段已日落。是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品</strong>
                         <strong>-listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒</strong>
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

    2. （可选）<span style="color:#D80000;">【DEPRECATED】</span> 优雅容错模式：在重调度配置的基础上，新增“-hotReset”字段。

        >[!CAUTION]
        > **【DEPRECATED】本功能已日落，请勿使用！**
        >
        >- 优雅容错功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。
        >- “-hotReset”字段取值为1对应的在线热复位功能已经日落。
        >- 以下配置示例仅用于历史参考，不建议在生产环境中使用。

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin
                         -volcanoType=true                    # 重调度场景下必须使用Volcano
                         -autoStowing=true                    # 该字段已日落。是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                         <span style="color:#D80000;"><strong>-hotReset=1 # 【DEPRECATED】开启优雅容错模式，系统会尝试自动复位故障芯片（取值为1的在线热复位功能已日落）</strong></span>
                         -listWatchPeriod=5                   # 健康状态检查周期，范围[3,1800]；单位为秒
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

2. 在K8s管理节点执行以下命令，启动Ascend Device Plugin。

    ```shell
   # {version}请替换为实际版本号，如：26.1.0
    kubectl apply -f device-plugin-volcano-v{version}.yaml
    ```

**配置NodeD<a name="section162092113510"></a>**

配置节点状态发送间隔时间。用户可以通过手动修改NodeD的启动YAML，配置上报节点状态的间隔时间。

1. 进入组件解压目录，执行以下命令，打开NodeD组件的启动YAML文件。

    ```shell
   # {version}请替换为实际版本号，如：26.1.0
    vi noded-v{version}.yaml
    ```

2. 在YAML文件的“args”行修改“-**reportInterval**”参数，如下所示：

    <pre codetype="yaml">
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 <strong>-reportInterval=5</strong>" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...</pre>

## 场景示例

- [PyTorch场景示例](./01_pytorch_examples_and_verification.md)
- [MindSpore场景示例](./02_mindspore_examples_and_verification.md)
