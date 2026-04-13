# 通过Prometheus使用<a name="ZH-CN_TOPIC_0000002511426931"></a>

本章节指导用户安装部署Prometheus相关软件，并通过Prometheus查看资源监测的相关数据信息，数据信息的相关说明可参见[Prometheus Metrics接口](../../api/npu_exporter.md#prometheus-metrics接口)章节。

- [直接对接Prometheus](#zh-cn_topic_0000001447284876_section875071183215)：NPU Exporter可以直接将NPU设备的数据信息导入到Prometheus中，无需额外的中间件或代理，架构更加简单。
- [通过Prometheus Operator对接Prometheus](#section1031014512341)：NPU Exporter通过Prometheus Operator插件对接Prometheus，帮助用户快速、简便地实现Prometheus服务的平台化，提高监测系统的可靠性和可维护性。

**直接对接Prometheus<a name="zh-cn_topic_0000001447284876_section875071183215"></a>**

1. 进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/utils/prometheus/base”目录下的prometheus.yaml文件。
2. <a name="zh-cn_topic_0000001447284876_li127175170321"></a>在管理节点执行以下命令获取镜像。

    ```shell
    docker pull prom/prometheus:v2.10.0
    ```

    >[!NOTE] 
    >- 获取镜像前，请确保能够正常访问互联网。
    >- 若不使用集群调度提供的prometheus.yaml，需要参考该YAML在相应位置加上app: prometheus字段，否则可能出现NPU Exporter连接超时。

3. prometheus.yaml已经默认包含获取NPU-Exporter metrics的相关的配置文件，用户可以根据需求自行修改相应的配置。以下从job_name开始之后的内容为获取的NPU-Exporter metrics的相关配置。

    ```Yaml
    ...
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: prometheus-config
      namespace: kube-system
    data:
      prometheus.yml: |
        global:
          scrape_interval:     15s
          evaluation_interval: 15s
        scrape_configs:
    ...
        - job_name: 'kubernetes-npu-exporter'
          kubernetes_sd_configs:
          - role: pod
          scheme: http
          relabel_configs:
          - action: keep
            source_labels: [__meta_kubernetes_namespace]
            regex: npu-exporter
          - source_labels: [__meta_kubernetes_pod_node_name]
            target_label: job
            replacement: ${1}
    ...
    ```

4. 执行以下命令，给管理节点打标签。

    ```shell
    kubectl label nodes <管理节点Hostname> masterselector=dls-master-node --overwrite=true
    ```

5. 将“prometheus.yaml”上传至[步骤2](#zh-cn_topic_0000001447284876_li127175170321)节点的任意路径下。
6. 在“prometheus.yaml”存放路径，执行以下命令，安装Prometheus服务。

    ```shell
    kubectl apply -f prometheus.yaml
    ```

    回显如下，表示安装成功。

    ```ColdFusion
    [root@centos check_env]# kubectl apply -f prometheus.yaml 
    clusterrole.rbac.authorization.k8s.io/prometheus created
    serviceaccount/prometheus created
    clusterrolebinding.rbac.authorization.k8s.io/prometheus created
    service/prometheus created
    deployment.apps/prometheus created
    configmap/prometheus-config created
    ```

7. 执行以下命令，查看Prometheus是否启动成功。

    ```shell
    kubectl get pods --all-namespaces | grep prometheus
    ```

    回显示例如下，出现Running状态表示Prometheus启动成功。

    ```ColdFusion
    kube-system      prometheus-58c69548b4-rhxsc                1/1     Running            0          6d14h
    ```

8. 登录Prometheus服务，查看监测的数据信息。
    1. 打开浏览器。
    2. 在浏览器中输入“http://_管理节点IP地址_:_端口号_”并按“Enter”。

        在prometheus.yaml文件中找到nodePort字段，该字段的值为Prometheus服务的端口号，默认为30003。

    3. 选择NPU的相关标签，查看对应数据信息。

**通过Prometheus Operator对接Prometheus<a name="section1031014512341"></a>**

1. 执行以下命令，获取Prometheus Operator插件源码。

    ```shell
    git clone https://github.com/prometheus-operator/kube-prometheus.git
    ```

    >[!NOTE] 
    >- 请根据[官方文档](https://github.com/prometheus-operator/kube-prometheus/tree/release-0.7)的兼容性列表，获取与K8s配套的Prometheus Operator源码分支。
    >- 若已经安装Prometheus Operator和Prometheus，可以直接执行[步骤4](#li15822115020428)。

2. 安装Prometheus Operator插件。
    1. 执行以下命令，安装Prometheus Operator。

        ```shell
        kubectl create -f manifests/setup/
        ```

        回显示例如下，表示Prometheus Operator安装成功。

        ```ColdFusion
        namespace/monitoring created
        ...
        deployment.apps/prometheus-operator created
        service/prometheus-operator created
        serviceaccount/prometheus-operator created
        ```

    2. 执行以下命令，查看Prometheus Operator是否启动成功。

        ```shell
        kubectl get pod -A -o wide|grep prometheus-operator
        ```

        回显示例如下，出现**Running**表示Prometheus Operator启动成功。

        ```ColdFusion
        monitoring     prometheus-operator-7649c7454f-wp84n       2/2     Running   0          58s   192.168.xx.xx   node133   <none>           <none>
        ```

3. 安装Prometheus。
    1. <a name="li601241164212"></a>进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/utils/prometheus/base”目录下的prometheus.yaml文件。
    2. 将[步骤1](#li601241164212)中获取到的prometheus.yaml上传至环境任意路径。
    3. 在“prometheus.yaml”存放路径，执行以下命令，安装Prometheus。

        ```shell
        kubectl apply -f prometheus.yaml
        ```

        回显如下，表示安装成功。

        ```ColdFusion
        service/prometheus created
        prometheus.monitoring.coreos.com/prometheus created
        serviceaccount/prometheus-service-account created
        clusterrole.rbac.authorization.k8s.io/prometheus-cluster-role created
        clusterrolebinding.rbac.authorization.k8s.io/prometheus-cluster-role-binding created
        ```

    4. 执行以下命令，查看Prometheus是否启动成功。

        ```shell
        kubectl get pods --all-namespaces | grep prometheus
        ```

        回显示例如下：

        ```ColdFusion
        kube-system    prometheus-prometheus-0                    2/2     Running   1          3m47s   192.168.xx.xx   node133   <none>           <none>
        monitoring     prometheus-operator-7649c7454f-wp84n       2/2     Running   0          5m52s   192.168.xx.xx   node133   <none>           <none>
        ```

4. <a name="li15822115020428"></a>NPU Exporter通过Prometheus Operator对接Prometheus。
    1. 获取[npu-exporter-svc.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.3.0/samples/utils/prometheus/prometheus_operator/npu-exporter-svc.yaml)和[servicemonitor.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.3.0/samples/utils/prometheus/prometheus_operator/servicemonitor.yaml)。

        >[!NOTE] 
        >若已经提前安装Prometheus，需要确保servicemonitor.yaml的以下字段，和已经部署的Prometheus中serviceMonitorSelector配置的matchLabels标签一致。
        >
        >```Yaml
        >...
        >  labels:                               
        >    serviceMonitorSelector: prometheus
        >...
        >```
        >
        >matchLabels标签可通过执行以下命令进行查询。
        >
        >```shell
        >kubectl describe pod <pod-name>
        >```

    2. （可选）可根据实际情况修改NPU Exporter的标签，不修改则直接跳过该步骤。
        1. 在npu-exporter-svc.yaml中，根据实际情况修改标签。

            ```Yaml
            apiVersion: v1
            kind: Service
            metadata:
              namespace: npu-exporter   # 命名空间为npu-exporter
              name: npu-exporter             
              labels:                        
                app: npu-exporter-svc   # NPU Exporter service的标签
            spec:
              type: ClusterIP
              ports:
              - port: 8082             # NPU Exporter的服务端口号
                targetPort: 8082      
            ...
            ```

        2. 在servicemonitor.yaml中，根据实际情况修改NPU Exporter的标签，并确保修改内容与npu-exporter-svc.yaml中一致。

            ```Yaml
            ...
            spec:
              endpoints:
              - interval: 10s
                targetPort: 8082                                 # NPU Exporter的服务端口号
                path: /metrics
              namespaceSelector:
                matchNames:
                - npu-exporter                                   # 命名空间为npu-exporter
              selector:
                matchLabels:                                     
                  app: npu-exporter-svc                          # NPU Exporter service的标签
            ```

    3. 依次执行以下命令，使用NPU Exporter通过Prometheus Operator对接Prometheus。

        ```shell
        kubectl apply -f servicemonitor.yaml
        kubectl apply -f npu-exporter-svc.yaml
        ```

    4. 执行以下命令，查看NPU Exporter对接Prometheus Operator是否成功。

        ```shell
        kubectl get svc -A|grep npu-exporter
        ```

        回显示例如下，表示NPU Exporter对接Prometheus Operator成功。

        ```ColdFusion
        npu-exporter   npu-exporter          ClusterIP   10.98.xx.xx     <none>        8082/TCP                       31s
        ```

    5. 执行以下命令，查看Prometheus Operator对接Prometheus是否成功。

        ```shell
        kubectl get servicemonitor -A|grep npu-exporter
        ```

        回显示例如下，表示Prometheus Operator对接Prometheus成功。

        ```ColdFusion
        kube-system   npu-exporter   55s
        ```

5. 登录Prometheus服务，查看监测的数据信息。
    1. 打开浏览器。
    2. 在浏览器中输入“http://_管理节点IP地址_:_端口号_”并按“Enter”。

        在prometheus.yaml文件中找到nodePort字段，该字段的值为Prometheus服务的端口号，默认为30003。

    3. 选择NPU的相关标签，查看对应数据信息。
