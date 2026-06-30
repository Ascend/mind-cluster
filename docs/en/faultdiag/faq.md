# FAQ<a name="ZH-CN_TOPIC_0000001681379661"></a>

## Diagnosis Failure: `[Errno 24] Too many open files`<a name="ZH-CN_TOPIC_0000001681140317"></a>

**Symptom<a name="section846317157147"></a>**

In large clusters, the diagnosis feature may fail due to an excessive number of log files in the input directory, resulting in a "Too many open files" error in the log.

![](../figures/faultdiag/zh-cn_image_0000001632860848.png)

**Solution<a name="section18590815191418"></a>**

1. Run the `ulimit -n` command to view the maximum number of file descriptors allowed to be open simultaneously.

    ![](../figures/faultdiag/zh-cn_image_0000001632540996.png)

2. Run the `ulimit -n num` command to adjust the file descriptor limit, for example, `ulimit -n 2048`.

    ![](../figures/faultdiag/zh-cn_image_0000001632700908.png)
