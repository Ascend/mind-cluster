# Kubectl Plugin<a name="ZH-CN_TOPIC_0000002524428760"></a>

Kubectl Plugin is the client command line tool of the ClusterOps Agent feature. It contains two commands, `kubectl ascend-diag` and `kubectl clusterops`, and is installed on the user machine. Kubectl Plugin encapsulates the fault diagnosis capability into `kubectl` commands. You can trigger diagnosis with one command through `kubectl ascend-diag` and manage LLM service configurations through `kubectl clusterops`.

- Users who use the [ClusterOps Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690) feature must install Kubectl Plugin.
- Kubectl Plugin is a client tool implemented with the pure Python standard library and does not require pip installation.

## Procedure<a name="section15023132772914"></a>

1. Download the Ascend ClusterOps Agent package by referring to [Obtaining Software Packages](./00_obtaining_software_packages.md).

2. Upload the software package to the user machine server and decompress it. After decompression, the plugin files are located in the `kubectl-plugin` subdirectory. The directory structure is as follows.

    ```shell
    Ascend-mindxdl-ascend-clusterops-agent_{version}_linux.zip
    └── kubectl-plugin/           # Kubectl Plugin
        ├── install.sh            # One-click installation script
        ├── kubectl-ascend_diag   # kubectl ascend-diag plugin file
        └── kubectl-clusterops    # kubectl clusterops plugin file
    ```

3. Log in to the user machine as the `root` user, go to the `kubectl-plugin` directory, and run the following command to install the plugin.

    ```shell
    cd kubectl-plugin
    bash install.sh
    ```

    The following is the output example, which indicates that the installation is successful.

    ```ColdFusion
    installed. Verification:
    OK: kubectl ascend-diag / kubectl clusterops are available
    ```

4. Verify whether the plugin is available.

    ```shell
    kubectl ascend-diag --help
    kubectl clusterops --help
    ```

    The following is the output example, which indicates that the plugin is available.

    ```ColdFusion
    $ kubectl ascend-diag --help
    usage: kubectl-ascend_diag [-h] [--job JOB] [-n NAMESPACE] [--agent-core-service AGENT_CORE_SERVICE] [--refresh] [--json] [--collect-manifest PATH]

    Ascend fault diagnosis kubectl plugin

    optional arguments:
    -h, --help            show this help message and exit
    --job JOB             job name (job CR name)
    -n NAMESPACE, --namespace NAMESPACE
                            namespace
    --agent-core-service AGENT_CORE_SERVICE
                            Agent Core Service, format svc.ns:9700
    --refresh             ignore the cache, force rerun and refresh the cache
    --json                print the full JSON response
    --collect-manifest PATH
                            write a local collect_manifest.yaml into the cluster ConfigMap and exit

    $ kubectl clusterops --help
    usage: kubectl-clusterops [-h] [--create-llm-config] [--clear-llm-config] [--base-url BASE_URL] [--model MODEL]

    cluster operations kubectl plugin

    optional arguments:
    -h, --help           show this help message and exit
    --create-llm-config  configure the LLM Secret (llm-secret)
    --clear-llm-config   clear all LLM config (delete llm-secret, effective immediately)
    --base-url BASE_URL  LLM Base URL (prompted; or env LLM_BASE_URL)
    --model MODEL        LLM model name (prompted; or env LLM_MODEL)
    ```

5. (Optional) Configure the LLM service. If you need to intelligently summarize the diagnosis report, create an LLM configuration through the `kubectl clusterops` command. If no LLM is configured, the diagnosis report is the original diagnosis result by default. For details about how to configure and clear the LLM, see [Configuring and Clearing the LLM](../../../04_usage/14_clusterops_agent/04_configuring_llm.md).

6. Verify the diagnosis function. Run the following command to trigger the ClusterOps Agent by job dimension and confirm that the diagnosis link is normal.

    ```shell
    kubectl ascend-diag --job {job_name} -n {namespace}
    ```

    If the diagnosis report (including fields such as root cause, fault event, and suggestion) is printed after the command is executed, the link is normal. The following is the output example.

    ```ColdFusion
    The diag job starts. Please wait. Job id: [20260918163512389528_b1b293d4-1c70-4e11-916e-a9023da3660a], run log file is [ascend_faultdiag_36.log].
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                 Ascend-fd Fault-Diag Report                                                                                 |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |     Version Information     |       Type       | Version                                                                                                                                    |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.2.0                                                                                                         |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |   Root Cause Node Analysis  |       Type       | Description                                                                                                                                |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |                             | Root Cause Node  | ['worker0 device-7']                                                                                                                       |
    |                             |     Symptom      | No timeout error logs are recorded in the Plog of any node. Nodes with error logs are suspected root cause nodes. Please investigate.      |
    |                             | First Error Node | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
    |                             | Last Error Node  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |     Fault Event Analysis    |       Type       | Description                                                                                                                                |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    |  Suspected Root Cause Fault |   Status Code    | Comp_OS_Service_Container_04                                                                                                               |
    |                             |  Fault Category  | Category:Software Component:HostOS Module:OS                                                                                               |
    |                             |   Fault Device   | ['worker0']                                                                                                                                |
    |                             |    Fault Name    | Container storage exception                                                                                                                |
    |                             |    Fault Desc    | Docker storage latency.                                                                                                                    |
    |                             |    Suggestion    | 1. Replace the storage disk and restart to create the container;                                                                           |
    |                             |     Key Log      | Sep 18 12:19:24 localhost /usr/sbin/irqbalance[3091]: Cannot change IRQ 1145 affinity: No space left on device                             |
    +-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
    The diag job is complete.

    Tip: to output the full diagnosis JSON, add --json after the command
    ```

    For the complete diagnosis output description and real examples, see [Using Fault Diagnosis](../../../04_usage/14_clusterops_agent/03_using_fault_diagnosis.md).
