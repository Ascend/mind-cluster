# 卸载

## 命令行方式安装的卸载

以组件的安装用户执行以下命令：

```shell
pip3 uninstall ascend-faultdiag -y
```

卸载完成后，可以通过 `pip3 list | grep ascend-faultdiag` 查看，无回显则表示成功卸载。

## 使用 MindCluster Ascend Deployer 方式安装后进行工具的卸载

以组件的安装用户执行删除二进制文件命令

```shell
rm /usr/local/bin/ascend-fd
```

## 清理残留文件

`~/.ascend_faultdiag` 目录保存了日志等信息，不会随着卸载自动删除。如果不再需要，请手动删除：

```shell
rm -rf ~/.ascend_faultdiag
```
