# 在K8s集成Containerd使用<a name="ZH-CN_TOPIC_0000002479227280"></a>

K8s集成Containerd场景下，用户需要安装Ascend Docker Runtime。

- **（二选一）申请NPU卡资源的情况下**。使用任务YAML下发训练或推理任务时，NPU芯片的分配由Volcano和Ascend Device Plugin组件自动完成；NPU芯片及相关文件目录的挂载由Ascend Docker Runtime组件自动完成。示例如下。

    ```Yaml
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                resources:
                  limits:
                    huawei.com/Ascend910: 1     # 资源名称和数量根据实际情况修改
                  requests:
                    huawei.com/Ascend910: 1     #  资源名称和数量根据实际情况修改
    ...
    ```

- **（二选一）如果未申请NPU资源**，由集成平台写入ASCEND\_VISIBLE\_DEVICES=void环境变量。示例如下。

    ```Yaml
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                env:
                - name: ASCEND_VISIBLE_DEVICES     # 未使用resources申请NPU资源时需增加此配置
                   value: "void"
    ...
    ```
