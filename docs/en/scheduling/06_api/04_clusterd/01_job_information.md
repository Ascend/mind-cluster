# Job Information<a name="ZH-CN_TOPIC_0000002511426769"></a>

## job-summary-<Job-Name\><a name="section24017282404"></a>

**Table 1**  job-summary-<Job-Name\> ConfigMap

|Parameter|Description|Value|
|--|--|--|
|hccl.json|Chip communication information used by the job. Can be escaped to JSON format, with the following field descriptions: <ul><li>`status`: Whether the job RankTable has been generated.</li><ul><li>`initializing`: Still allocating devices for the job, RankTable not yet generated.</li><li>`complete`: Once the RankTable is generated, the status immediately changes to `complete`, and other fields such as `server_list` appear synchronously.</li></ul><li>`server_list`: Job device allocation status.</li><ul><li>`device`: Records NPU allocation, NPU IP and rank_id information.</li><ul><li>`device_id`: Device ID of the NPU.</li><li>`device_ip`: Device IP of the NPU.</li><li>`rank_id`: Training Rank ID corresponding to the NPU.</li><li>`super_device_id`: Unique identifier of the NPU within the SuperPoD.</li></ul><li>`server_id`: AI Server identifier, globally unique.</li><li>`server_name`: Node name.</li><li>`server_sn`: SN number of the node. The device SN must exist. If it does not, contact Huawei technical support.</li><li>`host_ip`: Host IP.</li><li>`super_pod_id`: SuperPoD ID.</li><li>`pod_name`: Pod name.</li><li>`container_ids`: ID mapping table for all containers in the Pod.</li></ul><li>`server_count`: Number of nodes used by the job.</li><li>`version`: RankTable format version. For jobs on Atlas 950 SuperPoD, Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and similar products that generate a `rank_list` (v2.0 format), the value is "2.0".</li><li>`rank_list`: List of rank information for jobs on Atlas 950 SuperPoD, Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and similar products. Each element records the topology information of one training rank.</li><ul><li>`rank_id`: Training rank ID, globally unique, increasing from 0.</li><li>`local_id`: Associated with the device ID in the topology file, that is, the local sequence number of the device in the Pod.</li><li>`device_id`: Physical device ID (NPU device ID).</li><li>`level_list`: List of network topology level information for the rank.</li><ul><li>`net_layer`: Communication layer. 0 indicates intra-frame communication; 1 indicates intra-super-node communication (this layer does not exist in non-super-node scenarios); 2 indicates UBoE/UBG communication; 3 indicates RoCE communication.</li><li>`net_instance_id`: Network ID, a unique identifier within the same communication layer (`net_layer`).</li><li>`net_type`: Network type. `TOPO_FILE_DESC` indicates that the network type is queried from the topology file; `CLOS` indicates a CLOS network.</li><li>`net_attr`: Network layer. Reserved field, currently left empty.</li><li>`rank_addr_list`: Rank address list.</li><ul><li>`addr_type`: Address type, with the value `IP` or `EID`.</li><li>`addr`: Address value.</li><li>`ports`: NPU port list.</li><li>`plane_id`: Network plane ID.</li></ul></ul></ul><li>`rank_count`: Number of ranks participating in training. Output together with `rank_list`.</li><li>`total`: Number of ConfigMaps.</li></ul>|String|
|job_id|K8s ID information of the job.|String|
|operator|<ul><li>`add`: Status updates to `add` after receiving the job adding command.</li><li>`delete`: Status updates to delete after receiving the job deletion command.</li></ul>|String|
|deleteTime|Time when the job was deleted.|String|
|sharedTorIp|Shared switch information used by the job.|String|
|masterAddr|`MASTER_ADDR` value specified during PyTorch training.|String|
|total|Number of ConfigMaps.|String|
|time|Job start time.|String|
|framework|Framework used by the job.|String|
|job_status|Job status:<ul><li>`pending`</li><li>`running`</li><li>`complete`</li><li>`failed`</li></ul>|String|
|job_name|Job name.|String|
|cm_index|Sequence number of the current ConfigMap.|String|
|sid|User-defined job ID|String|

## current-job-statistic<a name="section39901331194218"></a>

Used to display statistical information of current jobs in the cluster. Detailed information is recorded in the `/var/log/mindx-dl/clusterd/event_job.log` file. Due to the capacity limit of K8s ConfigMap, the maximum number of cluster jobs supported for statistics is approximately 10,000. When the log file reaches 20 MB, automatic dumping is triggered, and a maximum of 5 dump logs are saved. The maximum retention period for dump logs is 40 days.

|Parameter|Description|
|--|--|
|data|-|
|- ID|Job ID assigned by the K8s cluster.|
|- customID|User-defined Job ID. It is not displayed if the content is empty.|
|- cardNum|Number of cards used by the job. It is not displayed if the content is empty.|
|- podFirstRunTime|Time when all Pods of the job first entered the running state. It is not displayed if the content is empty.|
|- stopTime|Time when all Pods of the job were completed or forcibly deleted. It is not displayed if the content is empty.|
|- podLastRunTime|Time when all Pods of the job last recovered to the running state. It is not displayed if the content is empty.|
|- podLastFaultTime|Time when some or all Pods of the job last failed. It is not displayed if the content is empty.|
|- podFaultTimes|Number of Pod rescheduling times caused by job faults. It is not displayed if the count is 0.|
|totalJob|Total number of jobs in the current cluster.|

## scheduling-exception-report<a name="section_scheduling_exception_report"></a>

This ConfigMap is located in the `cluster-system` namespace. It is used to display information about jobs with scheduling exceptions in the cluster, helping users quickly locate the causes of job scheduling failures.

**Table 7**  scheduling-exception-report ConfigMap

|Parameter|Description|Value|
|--|--|--|
|\<jobName\>.\<jobUID\>|Key for job exception information, composed of the job name and job UID.|String|
|- jobName|Job Name.|String|
|- jobType|Job type, for example, vcjob, acjob, etc.|String|
|- nameSpace|Namespace where the job resides.|String|
|- conditions|Details of job exception conditions.|Object|
|-- status|Job status.<ul><li>`JobEmptyStatus`: Job status is empty.</li><li>`JobInitialized`: Job has been initialized.</li><li>`JobFailed`: Job failed.</li><li>`PodGroupCreated`: PodGroup has been created.</li><li>`PodGroupPending`: PodGroup is in Pending status.</li><li>`PodGroupInqueue`: PodGroup is in Inqueue status.</li><li>`PodGroupUnknown:` PodGroup status is unknown.</li><li>`PodGroupRunning`: PodGroup is in Running status.</li></ul>|String|
|-- reason|Exception reason, including `JobEnqueueFailed`, `JobValidateFailed`, `NodePredicateFailed`, `BatchOrderFailed`, `NotEnoughResources`, `PodPending`, `PodFailed`, `PgNotInitialized`, `JobNoInitialized`, etc.|String|
|-- message|Detailed exception information, including fault description and troubleshooting suggestions.|String|
