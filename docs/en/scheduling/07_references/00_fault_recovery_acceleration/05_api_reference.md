# API Reference

MindIO TFT provides training fault tolerance capabilities, supporting fault detection, fault recovery, and optimizer state backup and recovery. The key APIs are as follows:

|API|Description|
|--|--|
|tft_init_controller|Initializes the MindIO TFT Controller module.|
|tft_start_controller|Starts the Controller module service and binds the IP address and port.|
|tft_destroy_controller|Shuts down the Controller service after training is complete.|
|tft_init_processor|Initializes the MindIO TFT Processor module and configures parameters such as rank, replica count, and TLS.|
|tft_start_processor|Starts the Processor module service and connects to the Controller.|
|tft_destroy_processor|Shuts down the Processor service after training is complete.|
|tft_start_updating_os|Marks the optimizer state as Updating before the optimizer state is updated.|
|tft_start_copy_os|Notifies the Processor to start copying the optimizer state.|
|tft_end_updating_os|Marks the optimizer state as Updated after the optimizer state update is complete.|
|tft_set_optimizer_replica|Sets the replica relationship of the optimizer state data corresponding to the rank.|
|tft_exception_handler|A decorator that captures training state exceptions and reports them for handling.|
|tft_set_step_args|Sets the training framework parameter set for use by callback functions.|
|tft_register_rename_handler|Registers a rename callback to rename the final checkpoint.|
|tft_register_save_ckpt_handler|Registers a dump callback to save the final checkpoint.|
|tft_register_exit_handler|Registers a user-defined exit method (MindSpore framework only).|
|tft_register_stop_handler|Registers a callback function for stopping training.|
|tft_register_clean_handler|Registers a callback function for cleaning up residual operator execution.|
|tft_register_rebuild_group_handler|Registers a callback function for rebuilding the MindIO ARF group.|
|tft_register_repair_handler|Registers a repair callback to complete data repair such as optimizer repair.|
|tft_register_rollback_handler|Registers a rollback callback to complete reset operations such as dataset rollback.|
|tft_register_stream_sync_handler|Registers a synchronization callback to ensure that no residual operators remain in the operator queue after training is paused.|

For complete API parameter descriptions, return values, and usage examples, see [MindIO TFT APIs](../../06_api/09_mindio/00_mindio_tft_apis.md).
