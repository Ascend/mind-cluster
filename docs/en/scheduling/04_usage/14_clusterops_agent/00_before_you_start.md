# Feature Description<a name="ZH-CN_TOPIC_00000026faultdiagnosis01"></a>

When a training or inference job fails during running, the fault root cause must be located quickly. Jobs are usually distributed across multiple nodes, and fault-related logs (system logs, device logs, service logs, plog, and so on) are scattered on the host machine of each node, making manual troubleshooting costly and inefficient. MindCluster provides the ClusterOps Agent feature, which triggers diagnosis for a specific job with the `kubectl ascend_diag` command and automatically completes the whole process of node log collection, log parsing, and centralized diagnosis, helping users quickly locate the fault root cause of a job.

## Feature Highlights<a name="sectionfaultdiagnosisfeatures"></a>

- **One-click diagnosis**: Trigger diagnosis for a specific job with the `kubectl ascend_diag` command, without logging in to each node to manually collect logs.
- **Automatic collection scheduling**: Agent Core acts as the centralized control center for fault diagnosis, automatically maintains the mapping between jobs and Pods, and delivers collection instructions to Node Collector on each compute node.
- **Local parsing and reporting**: Node Collector collects job logs from the host machine according to the collection contract, locally calls ascend-fd parse to parse the logs, and packages and reports the parsing result to Agent Core.
- **Centralized diagnosis**: Agent Core aggregates the parsing results reported by each node, assembles them into a diagnosis input directory, and calls ascend-fd diag to perform centralized diagnosis and generate a diagnosis report.
- **Result caching**: The diagnosis result is cached after each diagnosis. Repeated diagnosis directly returns the cached result, and the `--refresh` parameter can be used to force a refresh.
- **Intelligent summarization**: Optionally connects to an LLM service to intelligently summarize the diagnosis report and output a root cause report. When no LLM is configured, the system automatically falls back to a deterministic diagnosis report.

## Prerequisites<a name="sectionfaultdiagnosisprereq"></a>

- Agent Core, Node Collector, and Kubectl Plugin have been installed and deployed. For details, see [Manual Installation](../../03_installation_guide/02_installation/01_manual_installation.md).
- kubectl has been installed on the local host, and the kubeconfig has the permission to access the cluster.
- socat has been installed on the local host to ensure that `kubectl port-forward` works properly.
- The job to be diagnosed (such as acjob or inferjob) has been created in the cluster. If the job type is not among those supported by Agent Core by default, perform the operations in [Configuring Job Types](./01_configuring_task_crds.md) first.
- If job logs (such as plog) need to be collected, perform the operations in [Configuring Log Collection](./02_configuring_log_collection.md) first.

## Usage<a name="sectionfaultdiagnosisusage"></a>

1. [Configuring Job Types](./01_configuring_task_crds.md): Confirm or configure the job types that Agent Core can detect.
2. [Configuring Log Collection](./02_configuring_log_collection.md): Configure the collection method and collection contract of job logs.
3. [Using Fault Diagnosis](./03_using_fault_diagnosis.md): Run the `kubectl ascend_diag` command to diagnose an abnormal job.
4. [(Optional) Configuring and Clearing the LLM](./04_configuring_llm.md): Connect to an LLM service to intelligently summarize the diagnosis report.
