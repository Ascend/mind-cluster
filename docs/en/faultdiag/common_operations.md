# Common Operations<a name="ZH-CN_TOPIC_0000001829748104"></a>

## Customizing the MindCluster Ascend FaultDiag Home Directory<a name="ZH-CN_TOPIC_0000001876627541"></a>

The home directory of MindCluster Ascend FaultDiag can be set via the environment variable `ASCEND_FD_HOME_PATH`. This directory is used to store the runtime logs and operation log files of MindCluster Ascend FaultDiag. The files of fault keywords to be masked and user-defined faults are also stored in this directory.

**Procedure<a name="section78581144191111"></a>**

Run the following command to set the home directory of MindCluster Ascend FaultDiag.

```shell
export ASCEND_FD_HOME_PATH=/Custom path
```

> **NOTE**
>
> - When this environment variable is not set, the home directory of MindCluster Ascend FaultDiag defaults to `$HOME/.ascend_faultdiag/`.
> - When this environment variable is set, the specified path must exist and be a directory. The `/tmp` path is not supported as the home directory. The directory owner must be `root` or the program executor, and the program executor must have permissions to create, read, and write files in this directory.
