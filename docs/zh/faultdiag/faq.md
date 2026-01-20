# FAQ<a name="ZH-CN_TOPIC_0000001681379661"></a>

## 诊断失败，日志提示\[Errno 24\] Too many open files<a name="ZH-CN_TOPIC_0000001681140317"></a>

**问题描述<a name="section846317157147"></a>**

在使用诊断功能时，若集群规格较大，输入目录下日志文件数较多。可能会导致诊断失败，日志报“Too many open files.”。

![](figures/zh-cn_image_0000001632860848.png)

**解决方案<a name="section18590815191418"></a>**

1.  执行**ulimit -n**命令，查看允许同时打开的最大文件描述符数量。

    ![](../figures/faultdiag/zh-cn_image_0000001632540996.png)

2.  执行**ulimit -n** <i>num</i>命令，调整文件描述符上限，如ulimit -n 2048。

    ![](../figures/faultdiag/zh-cn_image_0000001632700908.png)


