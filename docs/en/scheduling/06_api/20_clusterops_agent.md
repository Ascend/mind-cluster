# Agent Core, Node Collector, and Kubectl Plugin<a name="ZH-CN_TOPIC_00000026ascendclusterops01"></a>

The ClusterOps Agent includes the following three components: Agent Core is the centralized control center for diagnosis, Node Collector is responsible for node log collection and parsing, and Kubectl Plugin provides the `kubectl ascend_diag` and `kubectl clusterops` command entries.

- **Agent Core**: Deployed as a Deployment. It maintains the central relationship cache between jobs and Pods, delivers collection instructions to Node Collector on each node, aggregates the parsing output and calls ascend-fd diag to perform centralized diagnosis, and generates a diagnosis report.
- **Node Collector**: Deployed as a DaemonSet, one for each NPU node. It collects job logs from the host according to the collection contract, locally calls ascend-fd parse to complete cleansing, and reports the result.
- **Kubectl Plugin**: Implemented with the pure standard library. It reuses the local kubeconfig for authentication and accesses the Agent Core service through `kubectl port-forward`.

## ConfigMap Description<a name="sectionascendclusteropscm"></a>

After Agent Core is started, the following ConfigMaps are created:

- `clusterops-pathmap`. For details, see [Table 1](#tableclusteropspathmap).
- `agent-core-relcache`. For details, see [Table 2](#tableagentcorerelcache).

**Table 1** clusterops-pathmap

<a name="tableclusteropspathmap"></a>

This ConfigMap resides in the `cluster-system` namespace. Its data key is `pathmap.json`, which records the mount pairs and `env` values of Pods, so that Agent Core can restore relationship data after a restart and deliver them along with the Pod list in the `TriggerCollect` collection instruction. Snapshots are written in sequence by filling shards (`clusterops-pathmap`, `clusterops-pathmap-1`... `clusterops-pathmap-N`; after shard 0 is full, shard 1 is written, and so on; the number of shards is fixed at 100), and each shard does not exceed the 1 MiB data upper limit of a ConfigMap.

|Parameter|Description|
|--|--|
|profiles|Set of mount pairs of Pods. The key is `profile_id` (the hash of the mount pair content), and the value is the mount pair list `host:container`. Pods with the same mounts in the same job share one profile.|
|pods|Pod entries. The key is the pod UID, and only the `profile`, `env`, and `deleted_at` fields are retained.|
|-profile|`profile_id` to which the Pod belongs.|
|-env|Literal `env` values of the Pod (`valueFrom` references are skipped), used to resolve env-type collection entities (such as `ASCEND_PROCESS_LOG_PATH`).|
|-deleted_at|Pod deletion timestamp. The value is `null` for a surviving Pod.|

**Table 2** agent-core-relcache

<a name="tableagentcorerelcache"></a>

This ConfigMap resides in the `cluster-system` namespace. Its data key is `snapshot`, which is used to save the snapshot of the central relationship cache between jobs and Pods, so that Agent Core can restore relationship data upon restart. Snapshots are written in sequence by filling shards (`agent-core-relcache`, `agent-core-relcache-1`... `agent-core-relcache-N`; after shard 0 is full, shard 1 is written, and so on; the number of shards is fixed at 100), and each shard does not exceed the 1 MiB data upper limit of a ConfigMap.

|Parameter|Description|
|--|--|
|pods|List of Pods. Each element contains the following fields.|
|-pod_name|Pod name.|
|-pod_uid|Pod UID.|
|-namespace|Pod namespace.|
|-node|Node where the Pod resides.|
|-rank|Rank of the Pod in the job (from the `ascend/job-rank` or `rank` annotation).|
|-owner_name|Name of the job (CR) to which the Pod belongs.|
|-deleted_at|Pod deletion timestamp. The value is `null` for a surviving Pod.|
|-phase|Current phase of the Pod (`Running`, `Succeeded`, `Failed`, and so on).|

>[!NOTE]
>When Agent Core is started, it reads this snapshot to restore the job-to-Pod relationship data. During diagnosis, it queries this relationship table by job name to obtain the Pods and the nodes where they reside, and then delivers collection instructions to Node Collector on each node. Deleted Pod entries are retained for 7 days (`POD_TTL`). The upper limit of the in-memory table capacity is the same as the total capacity of the sharded snapshots (100 MiB), and the oldest entries are evicted first when the limit is exceeded.

### Capacity and Historical Data Retention<a name="sectionascendclusteropscapacity"></a>

`relcache`/`pathmap` snapshots are stored by filling shards in sequence. After shard 0 is full, shard 1 is written, and so on. The total capacity = number of shards x 1 MiB, and the number of shards is fixed at 100 (100 MiB). The upper limit of the local in-memory table is the same as the snapshot capacity, and the oldest entries are evicted first when the limit is exceeded.

Estimated at 8 cards per job, the amount of data that can be stored in 100 MiB is as follows:

|Pod Form|Size of a Single Record|Records Storable in 100 MiB|Jobs Storable|Corresponding Number of Cards|
|--|--|--|--|--|
|1 Pod per job, 8 cards per Pod|About 264 B|About 397,000|About 397,000 |About 3.17 million |
|8 Pods per job, 1 card per Pod|About 264 B|About 397,000|About 50,000 |About 400,000|

- A single Ascend cluster supports up to 100,000 cards (about 12,500 nodes). With 8 cards per job, a full set of jobs in one cluster is about 12,500 Pods.
- The default 100 MiB capacity can hold about **the total historical jobs of four 100,000-card clusters** (including jobs retained within 7 days after deletion).
- Historical job expiration: Deleted Pod entries are retained for **7 days** by default (`POD_TTL`, adjustable in `agent-core.yaml`, in seconds). After expiration, they are deleted from the in-memory table and sharded snapshots by the GC thread. When the memory usage exceeds the capacity upper limit, the oldest entries are evicted first. Neither operation blocks diagnosis.

> [!NOTE]
>
> - Agent Core holds get/update/patch permissions on the ConfigMaps only in the `cluster-system` namespace (the number of shard CMs varies with the number of shards and cannot be enumerated by name).

## Related ConfigMaps/Secrets<a name="sectionascendclusteropsrelated"></a>

The following ConfigMaps/Secrets are closely related to the ClusterOps Agent:

**Table 3** Related ConfigMaps/Secrets

|Resource|Namespace|Writer|Reader|Purpose|
|--|--|--|--|--|
|agent-core-task-crds|cluster-system|User|Agent Core|Job CR GVK configuration. Agent Core reads it once at startup and tracks only the Pods managed by these job CRs. When a new job type is added, you need to add the GVK in this ConfigMap and add the corresponding get permission in the ClusterRole. For details, see [Configuring Job Types](../04_usage/14_clusterops_agent/01_configuring_task_crds.md).|
|collect-manifest|cluster-system|Kubectl Plugin (`kubectl ascend_diag --collect-manifest`)|Node Collector|Collection contract. Node Collector reads it at each collection. For details, see [Configuring Log Collection](../04_usage/14_clusterops_agent/02_configuring_log_collection.md).|
|llm-secret|mindx-dl|Kubectl Plugin (`kubectl clusterops`)|Agent Core|LLM configuration (api-key, base-url, model). Agent Core monitors it in real time. For details, see [Configuring and Clearing the LLM](../04_usage/14_clusterops_agent/04_configuring_llm.md).|

> In addition to the preceding ConfigMaps, the snapshots described in Table 1 and Table 2 are stored in shards by Pod UID. When Agent Core is started, it creates a corresponding ConfigMap for each shard (`clusterops-pathmap-N` and `agent-core-relcache-N`).

## HTTP Interfaces<a name="sectionascendclusteropshttp"></a>

**Table 4** Diagnosis interface

| Item | Description |
|------|------|
| Path | `/diag` |
| Method | POST |
| Default port | 9700 |
| Protocol | HTTP |

**Table 5** Request parameters

| Parameter | Type | Mandatory | Description |
|------|------|------|------|
| job | string | Yes | Job name, that is, the job CR name. |
| namespace | string | No | Namespace where the job resides. The default value is `default`. |
| refresh | bool | No | Whether to ignore the cache and forcibly refresh. The default value is `false`. |

**Table 6** Response description

| Parameter | Type | Description |
|------|------|------|
| pods | array | Information about the Pods involved in the job. |
| diag_report | object | ascend-fd diagnosis report. |
| final_text | string | Diagnosis report text, which is either an LLM intelligent summary or a deterministic diagnosis report. |
| error | string | Error information. This field exists when diagnosis fails (the job does not exist, collection fails, and so on). |
| cached | bool | Whether the cache is hit. |
| cached_note | string | Prompt information when the cache is hit. |
| llm_error | string | Reason why the LLM summarization failed. This field exists when an LLM is configured and the call fails. |

Request example:

```shell
curl -X POST localhost:9700/diag -H 'Content-Type: application/json' \
     -d '{"job":"job-x","namespace":"default","refresh":false}'
```

## Diagnosis Result Cache<a name="sectionascendclusteropscache"></a>

Agent Core provides two levels of diagnosis result cache: the recent result cache (memory) and the persistent result cache (disk).

**Recent Result Cache (Memory, 10 Seconds)**

When the same job (ns, job) is diagnosed repeatedly within 10 seconds, the latest result is directly returned (marked as cached, with the prompt "This is cached data; add --refresh for real-time data"), and collection is not performed again. The retention period can be adjusted through the environment variable `DIAG_RECENT_TTL`.

**Persistent Result Cache (Disk)**

After each successful diagnosis, the diagnosis result is written to the persistent cache file (the result is cached whether the job is running or stopped). When the same job is diagnosed again, if the cache is hit, the cached result is directly returned (with the prompt "Result from cache; run --refresh for the latest diagnosis").

**Table 7** Cache file description

| Item | Description |
|------|------|
| Cache file | `{cache directory}/cache/<ns>_<job>.json`, named by the job namespace and job name. It can be hit across days. |
| Cache directory | `{AGENT_WORK_ROOT}/cache` by default, that is, `/user/clusterops/agent-core/cache` (host export directory). It can be overridden through the environment variable `AGENT_CACHE_DIR`. |
| Retention period | 7 days (the default value of `POD_TTL`, adjustable through the environment variable `POD_TTL`). Diagnosis can still be executed and the cache can still be hit within 7 days after the job ends or the Pod is deleted. |

**Table 8** Cache data description

| Parameter | Description |
|------|------|
| pods | Information about the Pods involved in the job. |
| diag_report | ascend-fd diagnosis report. |
| final_text | Diagnosis report text (LLM intelligent summary or deterministic diagnosis report). After the LLM summary is generated, it is written back to the cache and directly returned when the cache is hit, without repeatedly calling the LLM. |
| error | Error information. |
| cached_at | Cache write timestamp. |

**Cache Conditions**

- The diagnosis result is cached after each successful diagnosis, independent of the job CR status. When the cache is hit, the cached result is directly returned, and you are prompted to use `--refresh` to forcibly refresh and obtain the latest result.
- When the job CR does not exist (404, the job has been deleted): if there is still a Pod record of the job in `relcache` (within the deletion TTL, and the logs are still retained on the node or shared disk), diagnosis can still be executed and cached. Only when there is no record of the job in `relcache` either is the job directly returned as not existing.

>[!NOTE]
>The cache and diagnosis outputs are stored in the `WORK_ROOT` directory of Agent Core (`/user/clusterops/agent-core` by default, exported by the host). If the cache and outputs need to be retained after Pod recreation (rescheduling), you need to configure a PVC for Agent Core to mount the `WORK_ROOT` directory. hostPath mounting supports retention only across Pod recreation on the same node.

## gRPC Interfaces<a name="sectionascendclusteropsgrpc"></a>

**Table 9** gRPC interface description

| Service | Port | Interface | Direction | Description |
|------|------|------|------|------|
| Agent Core | 9710 | UploadResult | Node Collector -> Agent Core | Receives the collection and cleansing result reported by Node Collector (streamed back as tar.gz). |
| Node Collector | 9720 | TriggerCollect | Agent Core -> Node Collector | Receives the collection instruction delivered by Agent Core, immediately returns the acceptance result, and performs collection asynchronously in the background. |

## Kubectl Plugin<a name="sectionascendclusteropskubectl"></a>

### kubectl ascend_diag<a name="sectionascendclusteropsascenddiag"></a>

Triggers the ClusterOps Agent by job dimension.

**Table 10** Parameter description

| Parameter | Description |
|------|------|
| --job | Job name, that is, the job CR name (mandatory). |
| -n / --namespace | Namespace where the job resides. The default value is `default`. |
| --refresh | Ignores the cache and forcibly re-executes diagnosis and refreshes the cache. |
| --json | Outputs the complete JSON response. |
| --collect-manifest | Writes the local collection contract to the cluster ConfigMap and then exits. |
| --agent-core-service | Specifies the Agent Core service in the format svc.ns:9700. The default value is `agent-core.mindx-dl:9700`. |

Usage example:

```shell
kubectl ascend_diag --job job-x -n training
kubectl ascend_diag --job job-x --refresh
kubectl ascend_diag --collect-manifest ./collect_manifest.yaml
```

### kubectl clusterops<a name="sectionascendclusteropsclusterops"></a>

Configures and clears the LLM service configuration.

**Table 11** Parameter description

| Parameter | Description |
|------|------|
| --create-llm-config | Configures the LLM Secret (`llm-secret`). |
| --clear-llm-config | Clears all LLM configurations (deletes `llm-secret`). |
| --base-url | LLM Base URL. It can be specified through the environment variable `LLM_BASE_UR`L or by interactive input. |
| --model | LLM model name. It can be specified through the environment variable `LLM_MODEL` or by interactive input. |

>[!NOTE]
>`--create-llm-config` and `--clear-llm-config` are mutually exclusive and cannot be used at the same time. The update or clearing takes effect in real time without restarting Agent Core.

Usage example:

```shell
kubectl clusterops --create-llm-config
kubectl clusterops --create-llm-config --base-url https://open.bigmodel.cn/api/paas/v4 --model glm-4-plus
kubectl clusterops --clear-llm-config
```
