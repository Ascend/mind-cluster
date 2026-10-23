# Cluster Version Information

## ConfigMap Description

After ClusterD is started, it creates or updates a ConfigMap named `component-versions`. The following table describes the details.

| Parameter                 | Description                                                       |
|--------------------|----------------------------------------------------------|
| ascend-for-volcano | Detailed version information reported by Ascend-for-Volcano                             |
| - version          | Version number                                                      |
| - gitCommit        | Git commit ID                                               |
| - gitBranch        | Git branch name                                                 |
| - buildOS          | Operating system at build time                                                  |
| - buildArch        | System architecture at build time                                                  |
| - goVersion        | Go compiler version                                                 |
| ascend-operator    | Detailed version information reported by Ascend Operator. For details about the fields, see the preceding information.                       |
| clusterd           | Detailed version information reported by ClusterD. For details about the fields, see the preceding information.                              |
| infer-operator     | Detailed version information reported by Infer Operator. For details about the fields, see the preceding information.                        |
| device-plugin      | Detailed version information of Ascend Device Plugin in the cluster, aggregated by ClusterD               |
| - type             | Component type. The value is DaemonSet                                        |
| - versions        | Statistics on the number of each version                                                 |
| - totalNodes        | Number of cluster nodes                                                   |
| - queryCommand          | Command for querying the detailed version information of the corresponding component in the node annotation. When using this command, remove the escape character (/).       |
| noded              | Detailed version information of NodeD in the cluster, aggregated by ClusterD. For details about the fields, see the preceding information.                     |
| k8s-rdma-shared-dp | Detailed version information of K8s RDMA Shared Dev Plugin in the cluster, aggregated by ClusterD. For details about the fields, see the preceding information. |
| ascend-dra         | Detailed version information of Ascend DRA in the cluster, aggregated by ClusterD. For details about the fields, see the preceding information.                |

>[!NOTE]
>
>- Only MindCluster 26.2.0 and later are supported.
>- For detailed version information about TaskD, see the `version.py` file in its installation directory.
>- For detailed version information about MindIO TFT and MindIO ACP, see the `VERSION` file in their installation directories.
>- If the version information ConfigMap is accidentally deleted, restart Volcano, Ascend Operator, ClusterD, and Infer Operator.
>- The jq command line tool must be installed to use commands in `queryCommand`. Otherwise, view the node annotation to obtain the component version information.
