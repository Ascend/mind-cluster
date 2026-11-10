# MindCluster Cluster Scheduling User Guide

- [Introduction](./01_introduction/_menu_introduction.md)
  - [Overview](./01_introduction/00_overview.md)
  - [Component Description](./01_introduction/01_component_description.md)
  - [Feature Description](./01_introduction/02_feature_description.md)
  - [Supported Product Models and OSs](./01_introduction/03_supported_product_models_and_os.md)
- [Quick Start](./02_quick_start/quick_start.md)
- [Installation Guide](./03_installation_guide/menu_installation_guide.md)
  - [Before You Install](./03_installation_guide/00_before_you_start.md)
  - [Environment Dependencies](./03_installation_guide/01_environment_dependencies.md)
  - [Installation and Deployment](03_installation_guide/02_installation/menu_installation.md)
    <!-- npu="950,A3,910b,910,310p" id45 -->
    - [Helm Installation (Recommended)](03_installation_guide/02_installation/00_helm_installation.md)
    <!-- end id45 -->
    - [Manual Installation](03_installation_guide/02_installation/01_manual_installation.md)
    - [Tool-based Installation](03_installation_guide/02_installation/02_installation_using_tool.md)
  - [Confirming Component Status](./03_installation_guide/03_confirming_status.md)
  - [Upgrade](./03_installation_guide/04_upgrade/menu_upgrade.md)
    <!-- npu="950,A3,910b,910,310p" id46 -->
    - [Helm Upgrade (Recommended)](./03_installation_guide/04_upgrade/02_helm_upgrade.md)
    <!-- end id46 -->
    - [Manual Upgrade](03_installation_guide/04_upgrade/01_manual_upgrade.md)
  - [Uninstallation](./03_installation_guide/05_uninstallation/menu_uninstallation.md)
    <!-- npu="950,A3,910b,910,310p" id47 -->
    - [Helm Uninstallation (Recommended)](./03_installation_guide/05_uninstallation/02_helm_uninstallation.md)
    <!-- end id47 -->
    - [Manual Uninstallation](03_installation_guide/05_uninstallation/01_manual_uninstallation.md)
- [Feature Guide](./04_usage/menu_usage.md)
  - [Containerization Feature Guide](./04_usage/00_containerization/menu_containerization.md)
    - [Before You Start](./04_usage/00_containerization/00_before_you_start.md)
    - [(Optional) Configuring Custom Mounted Content](./04_usage/00_containerization/01_configuring_custom_mounted_content.md)
    - [Usage on the Docker Client](./04_usage/00_containerization/02_usage_on_the_docker_client.md)
    - [Usage on the Containerd Client](./04_usage/00_containerization/03_usage_on_the_containerd_client.md)
    - [Usage on the CRI-O Client Using the crictl CLI](./04_usage/00_containerization/04_usage_on_the_crio_client.md)
  <!-- npu="950,A3,910b,910,310p" id1 -->
  - [Resource Monitoring Feature Guide](./04_usage/01_resource_monitoring/menu_resource_monitoring.md)
    - [NPU Resource Monitoring](./04_usage/01_resource_monitoring/00_npu_resource_monitoring/menu_npu_resource_monitoring.md)
      - [Before You Start](./04_usage/01_resource_monitoring/00_npu_resource_monitoring/00_before_you_start.md)
      - [Working Principle](./04_usage/01_resource_monitoring/00_npu_resource_monitoring/01_working_principle.md)
      - [Working with Prometheus](./04_usage/01_resource_monitoring/00_npu_resource_monitoring/02_working_with_prometheus.md)
      - [Working with Telegraf](./04_usage/01_resource_monitoring/00_npu_resource_monitoring/03_working_with_telegraf.md)
    <!-- npu="950" id2 -->
    - [DPU Resource Monitoring](./04_usage/01_resource_monitoring/01_dpu_resource_monitoring/menu_dpu_resource_monitoring.md)
      - [Before You Start](./04_usage/01_resource_monitoring/01_dpu_resource_monitoring/00_before_you_start.md)
      - [Working Principle](./04_usage/01_resource_monitoring/01_dpu_resource_monitoring/01_working_principle.md)
      - [Working with Prometheus](./04_usage/01_resource_monitoring/01_dpu_resource_monitoring/02_working_with_prometheus.md)
    <!-- end id2 -->
  <!-- end id1 -->
  <!-- npu="950,A3,910b,910,310p" id3 -->
  - [Virtual Instance Feature Guide](./04_usage/02_virtual_instance/menu_virtual_instance.md)
    <!-- npu="A3,910b,910,310p" id4 -->
    - [HDK-based Virtual Instance](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/menu_virtual_instance_with_hdk.md)
      - [Feature Description](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/01_description.md)
      - [Application Scenarios and Solutions](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/02_application_scenarios_solutions.md)
      - [Virtualization Templates](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md)
      - [Static vNPU Scheduling](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/menu_static_vnpu_scheduling.md)
        - [Creating vNPU](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md)
        - [Mounting vNPU (Static Virtualization)](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/02_mounting_vnpu_static.md)
        - [Destroying vNPU](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/03_destroying_vnpu.md)
      <!-- npu="A3,910b,310p" id6 -->
      - [Dynamic vNPU Scheduling](./04_usage/02_virtual_instance/00_virtual_instance_with_hdk/05_dynamic_vnpu_scheduling/01_dynamic_vnpu_scheduling_inference.md)
      <!-- end id6 -->
    <!-- end id4 -->
    <!-- npu="950,A3,910b" id5 -->
    - [vCANN-RT-based Virtual Instance](./04_usage/02_virtual_instance/01_virtual_instance_with_vcann_rt/menu_virtual_instance_with_vcann_rt.md)
      - [Feature Description](./04_usage/02_virtual_instance/01_virtual_instance_with_vcann_rt/00_description.md)
      - [Soft Allocation Scheduling (Inference)](./04_usage/02_virtual_instance/01_virtual_instance_with_vcann_rt/01_soft_allocation_scheduling_inference.md)
    <!-- end id5 -->
  <!-- end id3 -->
  <!-- npu="950,A3,910b,910,310p" id7 -->
  - [Scheduling Feature Guide](./04_usage/03_basic_scheduling/menu_basic_scheduling.md)
    - [Feature Description](./04_usage/03_basic_scheduling/00_feature_description.md)
    <!-- npu="950,A3,910b,910" id52 -->
    - [Affinity Scheduling](./04_usage/03_basic_scheduling/01_affinity_scheduling/menu_affinity_scheduling.md)
      - [Solution Description](./04_usage/03_basic_scheduling/01_affinity_scheduling/00_solution_description.md)
      - [Affinity Scheduling Interconnection Description](./04_usage/03_basic_scheduling/01_affinity_scheduling/01_affinity_scheduling_interconnection.md)
      - [Affinity Scheduling Description](./04_usage/03_basic_scheduling/01_affinity_scheduling/02_affinity_scheduling_description.md)
      - [Ascend AI Processor-based Affinity](./04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md)
      - [Node-based Affinity](./04_usage/03_basic_scheduling/01_affinity_scheduling/04_node_based_affinity.md)
      - [Scheduling Algorithm of Ascend AI Processor](./04_usage/03_basic_scheduling/01_affinity_scheduling/05_scheduling_algorithm_of_ascend_ai_processor.md)
      <!-- npu="950,A3" id11 -->
      - [Inference Service Affinity Scheduling](./04_usage/03_basic_scheduling/01_affinity_scheduling/06_infer_affinity_scheduling.md)
      <!-- end id11 -->
    <!-- end id52 -->
    - [Scheduling Process of the Ascend AI Processor](./04_usage/03_basic_scheduling/02_scheduling_process_of_the_ascend_ai_processor.md)
    - [Full NPU Scheduling](./04_usage/03_basic_scheduling/03_full_npu_scheduling.md)
    <!-- npu="A3" id8 -->
    - [Multi-level Scheduling](./04_usage/03_basic_scheduling/04_multi_level_scheduling.md)
    <!-- end id8 -->
    - [Rescheduling](./04_usage/03_basic_scheduling/05_rescheduling.md)
    <!-- npu="950" id9 -->
    - [DPU Job Scheduling](./04_usage/03_basic_scheduling/06_dpu_scheduling.md)
    <!-- end id9 -->
    <!-- npu="950,A3" id10 -->
    - [Intergenerational Upgrade Adaptation Guide](./04_usage/03_basic_scheduling/07_Intergenerational_upgrade.md)
    <!-- end id10 -->
    - [Degraded Scheduling with Fault Backoff Tolerance](./04_usage/03_basic_scheduling/08_fault_tolerance_backoff_scheduling.md)
  <!-- end id7 -->
  <!-- npu="950,A3,910b,910,310p" id12 -->
  - [Fault Detection Feature Guide](./04_usage/04_fault_detection_and_diagnosis/menu_fault_detection_and_diagnosis.md)
    - [Feature Description](./04_usage/04_fault_detection_and_diagnosis/00_feature_description.md)
    - [Working Principle](./04_usage/04_fault_detection_and_diagnosis/01_working_principle.md)
    - [Fault Types](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/menu_fault_types.md)
      - [Node Faults](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/01_node_faults.md)
      - [Chip Faults](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/02_chip_faults.md)
      - [Network Faults](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/03_network_faults.md)
      - [Workload Faults and Task Hangs](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/04_workload_faults.md)
      - [Performance Degradation, Slow Nodes, and Slow Networks](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/05_performance_degradation.md)
      - [Public Faults](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/06_public_faults.md)
      - [Silent Faults](./04_usage/04_fault_detection_and_diagnosis/02_fault_types/07_silent_faults.md)
    - [Configuration](./04_usage/04_fault_detection_and_diagnosis/03_configuration/menu_configuration.md)
      - [Configuring the Fault Level](./04_usage/04_fault_detection_and_diagnosis/03_configuration/01_fault_classification.md)
      - [Configuring Node Faults](./04_usage/04_fault_detection_and_diagnosis/03_configuration/02_node_faults.md)
      - [Configuring Chip Faults](./04_usage/04_fault_detection_and_diagnosis/03_configuration/03_chip_faults.md)
      - [Configuring Network Faults](./04_usage/04_fault_detection_and_diagnosis/03_configuration/04_network_faults.md)
      - [Configuring Workload Fault and Task Hang Detection](./04_usage/04_fault_detection_and_diagnosis/03_configuration/05_task_hang_detection.md)
      - [Configuring Performance Degradation, Slow Node, and Slow Network Diagnosis](./04_usage/04_fault_detection_and_diagnosis/03_configuration/06_performance_diagnosis.md)
      <!-- npu="A3,910b" id23 -->
      - [Configuring Online Stress Testing](./04_usage/04_fault_detection_and_diagnosis/03_configuration/07_online_stress_testing.md)
      <!-- end id23 -->
      - [Configuring Public Faults](./04_usage/04_fault_detection_and_diagnosis/03_configuration/08_public_faults.md)
      - [Configuring Silent Fault Detection](./04_usage/04_fault_detection_and_diagnosis/03_configuration/09_silent_faults.md)
    - [Querying and Verifying Faults](./04_usage/04_fault_detection_and_diagnosis/04_querying_and_verifying_faults.md)
  <!-- end id12 -->
  <!-- npu="950,A3,910b,910,310p" id49 -->
  - [Fault Recovery Feature Guide](./04_usage/05_fault_recovery/menu_fault_recovery.md)
    <!-- npu="950,A3,910b,910" id43 -->
    - [Resumable Training Feature Guide](./04_usage/05_fault_recovery/01_resumable_training/menu_resumable_training.md)
      - [Feature Description](./04_usage/05_fault_recovery/01_resumable_training/00_feature_description.md)
      - [Solutions and Principles](./04_usage/05_fault_recovery/01_resumable_training/01_solutions_principles/menu_solutions_principles.md)
        - [Fault Handling Principles](./04_usage/05_fault_recovery/01_resumable_training/01_solutions_principles/01_fault_handling.md)
        - [Recovery Acceleration Principles](./04_usage/05_fault_recovery/01_resumable_training/01_solutions_principles/02_recovery_acceleration.md)
      - [Configuration](./04_usage/05_fault_recovery/01_resumable_training/02_configuration/menu_configuration.md)
        - [Configuring Fault Handling](./04_usage/05_fault_recovery/01_resumable_training/02_configuration/01_configuring_fault_handling_policies.md)
        - [Configuring Recovery Acceleration](./04_usage/05_fault_recovery/01_resumable_training/02_configuration/02_configuring_training_recovery.md)
      - [Examples and Verification](./04_usage/05_fault_recovery/01_resumable_training/03_examples_and_verification/menu_examples_and_verification.md)
        - [PyTorch Examples and Verification](./04_usage/05_fault_recovery/01_resumable_training/03_examples_and_verification/01_pytorch_examples_and_verification.md)
        - [MindSpore Examples and Verification](./04_usage/05_fault_recovery/01_resumable_training/03_examples_and_verification/02_mindspore_examples_and_verification.md)
    <!-- end id43 -->
    <!-- npu="950,A3,910b" id44 -->
    - [Resumable Inference Feature Guide](./04_usage/05_fault_recovery/02_resumable_inference/menu_resumable_inference.md)
      - [Configuring Pod-Level Rescheduling for Inference Jobs](./04_usage/05_fault_recovery/02_resumable_inference/01_configuring_infer_pod_rescheduling.md)
    <!-- end id44 -->
    - [Chip Fault Recovery](./04_usage/05_fault_recovery/00_chip_fault_recovery.md)
  <!-- end id49 -->
  <!-- npu="950,A3,910b,910,310p" id13 -->
  - [Appliance Feature Guide](./04_usage/06_appliance/menu_appliance.md)
    - [Before You Start](./04_usage/06_appliance/00_before_you_start.md)
    - [NPU Hardware Fault Detection and Recovery](./04_usage/06_appliance/01_npu_hardware_fault_detection_and_rectification.md)
    - [Configuring Inference Service Fault Recovery](./04_usage/06_appliance/02_inference_service_fault_recovery.md)
  <!-- end id13 -->
  <!-- npu="A3,910b" id14 -->
  - [MindIE CMotor Inference Jobs Best Practices (Sunset)](./04_usage/07_mindie_motor_best_practice/menu_mindie_motor_best_practice.md)
    - [Before You Start](./04_usage/07_mindie_motor_best_practice/00_before_you_start.md)
    - [Deploying MindIE CMotor](./04_usage/07_mindie_motor_best_practice/01_deploying_mindie_motor.md)
    - [Configuring Inference Job Rescheduling](./04_usage/07_mindie_motor_best_practice/02_rescheduling_inference_job.md)
    - [Configuring Offline Reset for Inference Job Scenarios](./04_usage/07_mindie_motor_best_practice/03_configuring_offline_reset_for_inference_job.md)
    - [Configuring Elastic Scaling for Inference Jobs](./04_usage/07_mindie_motor_best_practice/04_configuring_elastic_scaling_for_inference_job.md)
    <!-- npu="910b" id48 -->
    - [Configuring Switch Affinity for Inference Jobs](./04_usage/07_mindie_motor_best_practice/05_configuring_switch_affinity_for_inference_job.md)
    <!-- end id48 -->
  <!-- end id14 -->
  <!-- npu="A3,910b" id15 -->
  - [SGLang Inference Jobs Best Practices](./04_usage/08_sglang_best_practice/menu_sglang_best_practice.md)
    - [Before You Start](./04_usage/08_sglang_best_practice/00_before_you_start.md)
    - [Deploying OME-based SGLang Inference Job](./04_usage/08_sglang_best_practice/01_deploying_ome_sglang_inference_job.md)
    - [Configuring Rescheduling of Inference Job Instances](./04_usage/08_sglang_best_practice/02_rescheduling_inference_job_instance.md)
  <!-- end id15 -->
  <!-- npu="A3,910b" id16 -->
  - [vLLM Inference Jobs Best Practices](./04_usage/09_vllm_best_practice/menu_vllm_best_practice.md)
    - [Before You Start](./04_usage/09_vllm_best_practice/00_before_you_start.md)
    - [Deploying vLLM Inference Jobs](./04_usage/09_vllm_best_practice/01_deploying_vllm_inference_job.md)
    - [Configuring Rescheduling of Inference Job Instances](./04_usage/09_vllm_best_practice/02_rescheduling_inference_job_instance.md)
    - [Configuring Fault Isolation](./04_usage/09_vllm_best_practice/03_configuring_fault_isolation.md)
  <!-- end id16 -->
  <!-- npu="950,A3,910b" id17 -->
  - [Infer Operator Inference Jobs Best Practices](./04_usage/10_infer_operator_best_practice/menu_infer_operator_best_practice.md)
    - [Before You Start](./04_usage/10_infer_operator_best_practice/00_before_you_start.md)
    - [Deploying Infer Operator Inference Jobs Based on vLLM Proxy](./04_usage/10_infer_operator_best_practice/01_deploying_infer_operator_inference_job_with_vllm_proxy.md)
    - [Deploying Infer Operator Inference Jobs Based on MindIE PyMotor](./04_usage/10_infer_operator_best_practice/02_deploying_infer_operator_inference_job_with_mindie_pymotor.md)
    - [Configuring Priority Scheduling for Inference Jobs](./04_usage/10_infer_operator_best_practice/03_configuring_priority_scheduling.md)
    - [Configuring Instance-level Rescheduling for Inference Jobs](./04_usage/10_infer_operator_best_practice/04_configuring_rescheduling_inference_job.md)
    - [Configuring Load-based Elastic Scaling](./04_usage/10_infer_operator_best_practice/05_configuring_elastic_scaling.md)
    <!-- npu="A3,910b" id18 -->
    - [Container Snapshot Deployment and Usage](./04_usage/10_infer_operator_best_practice/06_container_snapshot_usage.md)
    <!-- end id18 -->
    <!-- npu="950,A3" id19 -->
    - [Configuring Affinity Scheduling for Inference Service](./04_usage/10_infer_operator_best_practice/07_configuring_infer_service_affinity_scheduling.md)
    <!-- end id19 -->
  <!-- end id17 -->
  <!-- npu="950,A3,910b,910,310p" id20 -->
  - [Tidal Scheduling Best Practices](./04_usage/11_tidal_scheduling/menu_tidal_scheduling.md)
    - [Before You Start](./04_usage/11_tidal_scheduling/00_before_you_start.md)
    - [Preempt-based Tidal Scheduling for Inference/Training Jobs](./04_usage/11_tidal_scheduling/01_preempt_based_alternation.md)
    - [Reclaim-based Tidal Scheduling for Inference/Training Jobs](./04_usage/11_tidal_scheduling/02_reclaim_based_alternation.md)
  <!-- end id20 -->
  <!-- npu="A3,910b" id21 -->
  - [verl Reinforcement Learning Jobs Best Practices](./04_usage/12_verl_best_practice/menu_verl_best_practice.md)
    - [Colocated Training and Inference Scenario](./04_usage/12_verl_best_practice/01_colocated_scenario/menu_colocated_scenario.md)
      - [Before You Start](./04_usage/12_verl_best_practice/01_colocated_scenario/00_before_you_start.md)
      - [Deploying a verl Reinforcement Learning Job](./04_usage/12_verl_best_practice/01_colocated_scenario/01_deploying_verl_reinforcement_learning_job.md)
      - [Configuring Pod Rescheduling for Reinforcement Learning Jobs](./04_usage/12_verl_best_practice/01_colocated_scenario/02_configuring_rescheduling_reinforcement_learning_job.md)
    - [Separated Training and Inference Scenario](./04_usage/12_verl_best_practice/02_separated_scenario/menu_separated_scenario.md)
      - [Configuring verl Elastic Inference Fault Tolerance](./04_usage/12_verl_best_practice/02_separated_scenario/00_configuring_verl_elastic_rollout.md)
  <!-- end id21 -->
  <!-- npu="950,A3,910b" id22 -->
  - [Ascend DRA Component Best Practices](./04_usage/13_ascend_dynamic_resource_allocation_best_practice/menu_ascend_dynamic_resource_allocation_best_practice.md)
    - [Before You Start](./04_usage/13_ascend_dynamic_resource_allocation_best_practice/00_before_you_start.md)
    - [Resource Request and Release with the DRA Mechanism](./04_usage/13_ascend_dynamic_resource_allocation_best_practice/01_deploying_dra_and_running_task.md)
  <!-- end id22 -->
  - [ClusterOps Agent Feature Guide](./04_usage/14_clusterops_agent/menu_fault_diagnosis.md)
    - [Feature Description](./04_usage/14_clusterops_agent/00_before_you_start.md)
    - [Configuring Task Types](./04_usage/14_clusterops_agent/01_configuring_task_crds.md)
    - [Configuring Log Collection](./04_usage/14_clusterops_agent/02_configuring_log_collection.md)
    - [Using Fault Diagnosis](./04_usage/14_clusterops_agent/03_using_fault_diagnosis.md)
    - [Configuring and Clearing the LLM](./04_usage/14_clusterops_agent/04_configuring_llm.md)
- [Developer Guide](./05_developer_guide/menu_developer_guide.md)
  - [Manual Installation and Deployment](./05_developer_guide/00_installation_deployment/menu_installation_deployment.md)
    - [Installation and Deployment](./05_developer_guide/00_installation_deployment/00_manual_installation/menu_manual_installation.md)
      - [Obtaining Software Packages](./05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)
      - [Preparing for Installation](./05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md)
      - [Ascend Docker Runtime](./05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md)
      <!-- npu="950,A3,910b,910,310p" id24 -->
      - [NPU Exporter](./05_developer_guide/00_installation_deployment/00_manual_installation/03_npu_exporter.md)
      - [Ascend Device Plugin](./05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md)
      - [Volcano](./05_developer_guide/00_installation_deployment/00_manual_installation/05_volcano.md)
      - [ClusterD](./05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md)
      <!-- end id24 -->
      <!-- npu="950,A3,910b" id41 -->
      - [Infer Operator](./05_developer_guide/00_installation_deployment/00_manual_installation/07_infer_operator.md)
      <!-- end id41 -->
      <!-- npu="950,A3,910b,910,310p" id42 -->
      - [Ascend Operator](./05_developer_guide/00_installation_deployment/00_manual_installation/08_ascend_operator.md)
      - [NodeD](./05_developer_guide/00_installation_deployment/00_manual_installation/09_noded.md)
      - [Container Manager](./05_developer_guide/00_installation_deployment/00_manual_installation/10_container-manager.md)
      <!-- end id42 -->
      <!-- npu="950" id25 -->
      - [K8s RDMA Shared Dev Plugin](./05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md)
      - [DPU Exporter](./05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md)
      <!-- end id25 -->
      <!-- npu="950,A3,910b" id26 -->
      - [Ascend Dynamic Resource Allocation](./05_developer_guide/00_installation_deployment/00_manual_installation/13_ascend_dynamic_resource_allocation.md)
      <!-- end id26 -->
      - [Kubectl Plugin](./05_developer_guide/00_installation_deployment/00_manual_installation/14_kubectl_plugin.md)
      - [Agent Core](./05_developer_guide/00_installation_deployment/00_manual_installation/15_agent_core.md)
      - [Node Collector](./05_developer_guide/00_installation_deployment/00_manual_installation/16_node_collector.md)
    - [Upgrade](./05_developer_guide/00_installation_deployment/01_upgrade.md)
    - [Uninstallation](./05_developer_guide/00_installation_deployment/02_uninstallation.md)
  <!-- npu="950,A3,910b,910,310p" id27 -->
  - [Custom Metrics Development](./05_developer_guide/01_custom_metrics_development.md)
  - [Chip Hot Reset Plugin Development](./05_developer_guide/02_dp_hot_reset_plugin_development.md)
  - [Public Fault Reporting](./05_developer_guide/03_clusterd_fault_awareness.md)
  <!-- end id27 -->
  - [Setting Up the Development and Compilation Environment](./05_developer_guide/04_development_environment.md)
- [API Reference](./06_api/menu_api.md)
  <!-- npu="950,A3,910b,910,310p" id28 -->
  - [NPU Exporter](./06_api/00_npu_exporter/menu_npu_exporter.md)
    - [NPU Exporter Home Page](./06_api/00_npu_exporter/00_npu_exporter_home_page.md)
    - [Prometheus Metrics API](./06_api/00_npu_exporter/01_prometheus_metrics_api.md)
    - [Telegraf Data Description](./06_api/00_npu_exporter/02_telegraf_data_description.md)
    - [Custom Metrics File](./06_api/00_npu_exporter/03_custom_metrics_file.md)
    - [Health Probe API](./06_api/00_npu_exporter/04_health_probe.md)
  - [Volcano](./06_api/01_volcano.md)
  - [Ascend Device Plugin](./06_api/02_ascend_device_plugin.md)
  - [NodeD](./06_api/03_noded.md)
  - [ClusterD](./06_api/04_clusterd/menu_clusterd.md)
    - [Cluster Resources](./06_api/04_clusterd/00_cluster_resources.md)
    - [Job Information](./06_api/04_clusterd/01_job_information.md)
    - [Process-level Recovery APIs](./06_api/04_clusterd/02_process_level_recovery_apis.md)
    - [Public Fault APIs](./06_api/04_clusterd/03_public_fault_apis.md)
    - [Performance Degradation Fault APIs](./06_api/04_clusterd/04_performance_degradation_apis.md)
    - [Service Configuration APIs](./06_api/04_clusterd/05_service_configuration_apis.md)
    - [Fault Service APIs](./06_api/04_clusterd/06_fault_service_apis.md)
    - [Job Information APIs](./06_api/04_clusterd/07_job_information_apis.md)
    - [Link Failover and Switchback APIs](./06_api/04_clusterd/08_link_failover_and_switchback_apis.md)
    - [Online Stress Testing APIs](./06_api/04_clusterd/09_online_stress_testing_apis.md)
    - [Health Probe API](./06_api/04_clusterd/10_health_probe.md)
    - [Cluster Version Information](./06_api/04_clusterd/11_cluster_version.md)
  - [Ascend Operator](./06_api/05_ascend_operator.md)
  <!-- end id28 -->
  <!-- npu="A3,910b" id29 -->
  - [Elastic-Agent (Resumable Training APIs)](./06_api/06_elastic_agent.md)
  <!-- end id29 -->
  <!-- npu="950,A3,910b" id30 -->
  - [TaskD](./06_api/07_taskd/menu_taskd.md)
    - [taskd.\_\_version_\_](./06_api/07_taskd/00_taskd_version.md)
    - [TaskD Worker APIs](./06_api/07_taskd/01_taskd_worker_apis.md)
    - [TaskD Agent APIs](./06_api/07_taskd/02_taskd_agent_apis.md)
    - [TaskD Proxy APIs](./06_api/07_taskd/03_taskd_proxy_apis.md)
    - [TaskD Manager APIs](./06_api/07_taskd/04_taskd_manager_apis.md)
    - [Resumable Training APIs](./06_api/07_taskd/05_resumable_training_apis.md)
    - [Internal TaskD APIs](./06_api/07_taskd/06_internal_taskd_apis.md)
    - [Return Codes](./06_api/07_taskd/07_return_codes.md)
  <!-- end id30 -->
  <!-- npu="950,A3,910b,910,310p" id32 -->
  - [Ascend Job](./06_api/08_ascend_job.md)
  <!-- end id32 -->
  <!-- npu="950,A3,910b" id31 -->
  - [MindIO](./06_api/09_mindio/menu_mindio.md)
    - [MindIO TFT APIs](./06_api/09_mindio/00_mindio_tft_apis.md)
    - [MindIO ACP APIs](./06_api/09_mindio/01_mindio_acp_apis.md)
  <!-- end id31 -->
  <!-- npu="950,A3,910b" id33 -->
  - [Infer Operator](./06_api/10_infer_operator.md)
  <!-- end id33 -->
  <!-- npu="950" id34 -->
  - [K8s RDMA Shared Dev Plugin](./06_api/11_k8s_rdma_shared_dev_plugin.md)
  <!-- end id34 -->
  <!-- npu="950,A3,910b,910,310p" id35 -->
  - [K8s Native Objects](./06_api/12_k8s.md)
  - [Environment Variable Description](./06_api/13_environment_variable_description.md)
  <!-- end id35 -->
  <!-- npu="950,A3,910b,910" id50 -->
  - [hccl.json File Description](./06_api/14_hccl.json_file_description.md)
  <!-- end id50 -->
  <!-- npu="950,A3,910b,910,310p" id51 -->
  - [Job YAML Configuration](./06_api/15_yaml_configuration.md)
  <!-- end id51 -->
  <!-- npu="950" id36 -->
  - [DPU Exporter](./06_api/16_dpu_exporter.md)
  <!-- end id36 -->
  <!-- npu="950,A3,910b" id37 -->
  - [Ascend Dynamic Resource Allocation](./06_api/17_ascend_dynamic_resource_allocation.md)
  <!-- end id37 -->
  <!-- npu="950,A3,910b,910,310p" id39 -->
  - [Container Manager](./06_api/18_container-manager.md)
  <!-- end id39 -->
  <!-- npu="950" id38 -->
  - [UB Host Device Cni](./06_api/19_ub_host_device_cni.md)
  <!-- end id38 -->
  - [Agent Core, Node Collector, and Kubectl Plugin](./06_api/20_clusterops_agent.md)
- [References](./07_references/menu_references.md)
  <!-- npu="950,A3,910b" id40 -->
  - [Fault Recovery Acceleration](./07_references/00_fault_recovery_acceleration/menu_fault_recovery_acceleration.md)
    - [Product Description](./07_references/00_fault_recovery_acceleration/01_product_description.md)
    - [Installation and Deployment](./07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md)
    - [Usage Guidance](./07_references/00_fault_recovery_acceleration/03_usage_guidance.md)
    - [Security Management and Hardening](./07_references/00_fault_recovery_acceleration/04_security_management_and_hardening.md)
    - [API Reference](./07_references/00_fault_recovery_acceleration/05_api_reference.md)
    - [Appendix](./07_references/00_fault_recovery_acceleration/06_appendixes.md)
  - [Optimizing Checkpoint Saving and Loading](./07_references/01_optimizing_saving_and_loading_checkpoints/menu_optimizing_saving_and_loading_checkpoints.md)
    - [Product Description](./07_references/01_optimizing_saving_and_loading_checkpoints/01_product_description.md)
    - [Installation and Deployment](./07_references/01_optimizing_saving_and_loading_checkpoints/02_installation_and_deployment.md)
    - [Usage Guidance](./07_references/01_optimizing_saving_and_loading_checkpoints/03_usage_guidance.md)
    - [Security Management and Hardening](./07_references/01_optimizing_saving_and_loading_checkpoints/04_security_management_and_hardening.md)
    - [API Reference](./07_references/01_optimizing_saving_and_loading_checkpoints/05_api_reference.md)
    - [Alarm Reference](./07_references/01_optimizing_saving_and_loading_checkpoints/06_alarm_reference.md)
    - [Appendix](./07_references/01_optimizing_saving_and_loading_checkpoints/07_appendixes.md)
  <!-- end id40 -->
  - [Common Operations](07_references/02_common_operations.md)
  - [FAQs](07_references/03_faq.md)
  - [Security Hardening](07_references/04_security_hardening.md)
  - [Appendix](07_references/05_appendix.md)
