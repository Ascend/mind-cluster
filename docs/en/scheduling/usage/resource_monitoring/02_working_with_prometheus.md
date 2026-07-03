# Using with Prometheus<a name="ZH-CN_TOPIC_0000002511426931"></a>

This chapter guides users through installing and deploying Prometheus-related software, and viewing resource monitoring data information through Prometheus. For details about the data information, see [Prometheus Metrics API](../../api/npu_exporter/01_prometheus_metrics_api.md).

- [Directly Connecting to Prometheus](#zh-cn_topic_0000001447284876_section875071183215): NPU Exporter can directly import NPU device data information into Prometheus without additional middleware or agents, resulting in a simpler architecture.
- [Connecting to Prometheus via Prometheus Operator](#section1031014512341): NPU Exporter connects to Prometheus through the Prometheus Operator plugin, helping users quickly and easily platformize the Prometheus service, improving the reliability and maintainability of the monitoring system.

## Directly Connecting to Prometheus<a name="zh-cn_topic_0000001447284876_section875071183215"></a>

1. Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository, switch to the corresponding branch according to the [mindcluster-deploy Open-Source Repository Version Description](../../references/appendix.md#mindcluster-deploy-open-source-repository-version-description), and obtain the `prometheus.yaml` file in the `samples/utils/prometheus/base` directory.
2. <a name="zh-cn_topic_0000001447284876_li127175170321"></a>Run the following command on the management node to obtain the image.

    ```shell
    docker pull prom/prometheus:v2.10.0
    ```

    **NOTE**
    - Before obtaining the image, ensure that you can access the internet normally.
    - If you do not use the `prometheus.yaml` file provided by MindCluster, refer to this YAML and add the `app: prometheus` field in the corresponding location. Otherwise, NPU Exporter connection may time out.

3. The `prometheus.yaml` file already includes the default configuration for obtaining NPU Exporter metrics. You can modify the corresponding configuration as needed. The content starting from `job_name` below is the configuration for obtaining NPU Exporter metrics.

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

4. Run the following command to label the management node.

    ```shell
    kubectl label nodes <Hostname of the management node> masterselector=dls-master-node --overwrite=true
    ```

5. Upload `prometheus.yaml` to any path on the node in [step 2](#zh-cn_topic_0000001447284876_li127175170321).
6. In the directory where `prometheus.yaml` is stored, run the following command to install the Prometheus service.

    ```shell
    kubectl apply -f prometheus.yaml
    ```

    The sample output is as follows, which indicates a successful installation.

    ```ColdFusion
    [root@centos check_env]# kubectl apply -f prometheus.yaml
    clusterrole.rbac.authorization.k8s.io/prometheus created
    serviceaccount/prometheus created
    clusterrolebinding.rbac.authorization.k8s.io/prometheus created
    service/prometheus created
    deployment.apps/prometheus created
    configmap/prometheus-config created
    ```

7. Run the following command to check whether Prometheus has started successfully.

    ```shell
    kubectl get pods --all-namespaces | grep prometheus
    ```

    The sample output is as follows. A `Running` status indicates that Prometheus has started successfully.

    ```ColdFusion
    kube-system      prometheus-58c69548b4-rhxsc                1/1     Running            0          6d14h
    ```

8. Log in to the Prometheus service and view the monitored data information.
    1. Open a browser.
    2. Enter `http://management node IP address:port number` in the browser and press `Enter`.

        Find the `nodePort` field in the `prometheus.yaml` file. The value of this field is the port number of the Prometheus service, which defaults to `30003`.

    3. Select the relevant NPU labels to view the corresponding data information.

## Connecting to Prometheus via Prometheus Operator<a name="section1031014512341"></a>

1. Run the following command to obtain the Prometheus Operator plugin source code.

    ```shell
    git clone https://github.com/prometheus-operator/kube-prometheus.git
    ```

    >[!NOTE]
    >- Refer to the compatibility list in the [official documentation](https://github.com/prometheus-operator/kube-prometheus/tree/release-0.7) to obtain the Prometheus Operator source code branch that matches your K8s version.
    >- If Prometheus Operator and Prometheus are already installed, you can proceed directly to [step 4](#li15822115020428).

2. Install the Prometheus Operator plugin.
    1. Run the following command to install Prometheus Operator.

        ```shell
        kubectl create -f manifests/setup/
        ```

        The sample output is as follows, which indicates that Prometheus Operator has been installed successfully.

        ```ColdFusion
        namespace/monitoring created
        ...
        deployment.apps/prometheus-operator created
        service/prometheus-operator created
        serviceaccount/prometheus-operator created
        ```

    2. Run the following command to check whether Prometheus Operator has started successfully.

        ```shell
        kubectl get pod -A -o wide|grep prometheus-operator
        ```

        The sample output is as follows. A `Running` state indicates that Prometheus Operator has started successfully.

        ```ColdFusion
        monitoring     prometheus-operator-7649c7454f-wp84n       2/2     Running   0          58s   192.168.xx.xx   node133   <none>           <none>
        ```

3. Install Prometheus.
    1. <a name="li601241164212"></a>Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository, switch to the corresponding branch according to [mindcluster-deploy Open-Source Repository Version Description](../../references/appendix.md#mindcluster-deploy-open-source-repository-version-description), and obtain the `prometheus.yaml` file from the `samples/utils/prometheus/base` directory.
    2. Upload the `prometheus.yaml` obtained in [Step 1](#li601241164212) to any path in the environment.
    3. In the directory where `prometheus.yaml` is stored, run the following command to install Prometheus.

        ```shell
        kubectl apply -f prometheus.yaml
        ```

        The sample output is as follows, which indicates a successful installation.

        ```ColdFusion
        service/prometheus created
        prometheus.monitoring.coreos.com/prometheus created
        serviceaccount/prometheus-service-account created
        clusterrole.rbac.authorization.k8s.io/prometheus-cluster-role created
        clusterrolebinding.rbac.authorization.k8s.io/prometheus-cluster-role-binding created
        ```

    4. Run the following command to check whether Prometheus has started successfully.

        ```shell
        kubectl get pods --all-namespaces | grep prometheus
        ```

        Sample output:

        ```ColdFusion
        kube-system    prometheus-prometheus-0                    2/2     Running   1          3m47s   192.168.xx.xx   node133   <none>           <none>
        monitoring     prometheus-operator-7649c7454f-wp84n       2/2     Running   0          5m52s   192.168.xx.xx   node133   <none>           <none>
        ```

4. <a name="li15822115020428"></a>Connect NPU Exporter to Prometheus through Prometheus Operator.
    1. Obtain [npu-exporter-svc.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/utils/prometheus/prometheus_operator/npu-exporter-svc.yaml) and [servicemonitor.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/utils/prometheus/prometheus_operator/servicemonitor.yaml).

        **NOTE**
        If Prometheus has been installed in advance, ensure that the following fields in `servicemonitor.yaml` are consistent with `matchLabels` configured in `serviceMonitorSelector` of the deployed Prometheus.
        >
        >```Yaml
        >...
        >  labels:
        >    serviceMonitorSelector: prometheus
        >...
        >```
        >
        >`matchLabels` can be queried by running the following command.
        >
        >```shell
        >kubectl describe pod <pod-name>
        >```

    2. (Optional) Modify the labels of NPU Exporter based on the actual situation. If no modification is required, skip this step.
        1. In `npu-exporter-svc.yaml`, modify the labels based on the actual situation.

            ```Yaml
            apiVersion: v1
            kind: Service
            metadata:
              namespace: npu-exporter   # The namespace is npu-exporter
              name: npu-exporter
              labels:
                app: npu-exporter-svc   # Labels of the NPU Exporter service
            spec:
              type: ClusterIP
              ports:
              - port: 8082             # Service port number of NPU Exporter
                targetPort: 8082
            ...
            ```

        2. In `servicemonitor.yaml`, modify the labels of NPU Exporter according to the actual situation, and ensure that the modifications are consistent with those in `npu-exporter-svc.yaml`.

            ```Yaml
            ...
            spec:
              endpoints:
              - interval: 10s
                targetPort: 8082                                 # Service port number of NPU Exporter
                path: /metrics
              namespaceSelector:
                matchNames:
                - npu-exporter                                   # The namespace is npu-exporter
              selector:
                matchLabels:
                  app: npu-exporter-svc                          # Labels of the NPU Exporter service
            ```

    3. Run the following commands in sequence to connect NPU Exporter to Prometheus through Prometheus Operator.

        ```shell
        kubectl apply -f servicemonitor.yaml
        kubectl apply -f npu-exporter-svc.yaml
        ```

    4. Run the following command to check whether NPU Exporter is successfully connected to Prometheus Operator.

        ```shell
        kubectl get svc -A|grep npu-exporter
        ```

        The sample output is as follows, which indicates that NPU Exporter is successfully connected to Prometheus Operator.

        ```ColdFusion
        npu-exporter   npu-exporter          ClusterIP   10.98.xx.xx     <none>        8082/TCP                       31s
        ```

    5. Run the following command to check whether Prometheus Operator is successfully connected to Prometheus.

        ```shell
        kubectl get servicemonitor -A|grep npu-exporter
        ```

        The sample output is as follows, which indicates that Prometheus Operator is successfully connected to Prometheus.

        ```ColdFusion
        kube-system   npu-exporter   55s
        ```

5. Log in to the Prometheus service to view the monitored data information.
    1. Open a browser.
    2. Enter `http://management node IP address:port number`" in the browser and press `Enter`.

        Find the `nodePort` field in the `prometheus.yaml` file. The value of this field is the port number of the Prometheus service, which defaults to `30003`.

    3. Select the relevant labels for the NPU to view the corresponding data information.
