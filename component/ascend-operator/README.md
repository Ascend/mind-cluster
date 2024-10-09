# Ascend Operator
Ascend Operator 是MindX DL支持mindspore、pytorch、tensorflow三个AI框架在Kubernetes上进行分布式训练的插件。CRD（Custom Resource Definition）中定义了AscendJob任务，用户只需配置yaml文件，即可轻松实现分布式训练。

# 安装
安装方法可以有以下几种
## 1. 编译
```
bash build/build.sh
```
## 2. 制作镜像
```
cd ./output
docker build -t ascend-operator:<version> .
```
## 3. 部署到k8s集群
```
kubectl apply -f ascend-operator-<version>.yaml
```
安装后：
使用`kubectl get pods --all-namespaces`,即可看到namespace为mindxdl-system的部署任务。
使用`kubectl describe pod ascned-operator-controller-manager-xxx-xxx -n mindx-dl`，可查看pod的详细信息。
# Samples
当前ms-operator支持普通单Worker训练、和自动并行（例如数据并行、模型并行等）的Scheduler、Worker启动。

在`config/samples/`中有运行样例。
以数据并行的Scheduler、Worker启动为例，其中数据集和网络脚本需提前准备：
```
kubectl apply -f examples/ms_ascendjob.yaml
```
使用`kubectl get all -o wide`即可看到集群中启动的Scheduler和Worker
# 开发指南
## 核心代码：
`pkg/api/v1/ascendjob_types.go`中为AscendJob的CRD定义。
`pkg/controllers/v1/*`中为AscendJob controller的核心逻辑。

