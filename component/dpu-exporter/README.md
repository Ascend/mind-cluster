# DPU-Exporter

# 组件介绍

DPU-Expoter是华为自研的专门收集华为DPU各种监测信息和指标，并封装成Prometheus专用数据格式的一个服务组件。

# 编译DPU-Exporter

1. 通过git拉取源码，获得dpu-exporter。

   示例：Dpu-Exporter源码放在/home/mind-cluster/component/dpu-exporter目录下

2. 执行以下命令，进入Dpu-Exporter构建目录，执行构建脚本，在“output“目录下生成二进制dpu-exporter、yaml文件和Dockerfile等文件。

   **cd** _/home/mind-cluster/component/_**dpu-exporter/build/**

   **chmod +x build.sh**

   **./build.sh**

3. 执行以下命令，查看**output**生成的软件列表。

   **ll** _/home/mind-cluster/component/_**dpu-exporter/output**

    ```text
    -rw-r--r--. 1 root root      166 Aug 22 17:31 config.json
    -rw-r--r--. 1 root root      777 Aug 22 17:31 Dockerfile
    -rw-r--r--. 1 root root 10944814 Aug 22 17:31 dpu-exporter
    -rw-r--r--. 1 root root     3981 Aug 22 17:31 dpu-exporter-v6.0.0.yaml
    ```

# 说明

1. 当前Dpu-Exporter仅支持http启动，如果需要使用https启动，请自行完成代码修改并适配Prometheus
