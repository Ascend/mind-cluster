# API Reference

MindIO ACP provides high-performance asynchronous checkpoint capabilities, enabling fast saving and restoration of model states during training. The key APIs are as follows:

|API|Description|
|--|--|
|initialize|Initializes the MindIO ACP Client and configures parameters such as memory pool size and the number of concurrent write threads.|
|save|Saves data to a specified path, supporting both memfs and fopen modes.|
|multi_save|Saves the same data to multiple files, suitable for multi-replica scenarios.|
|register_checker|Registers an asynchronous callback function for data integrity verification.|
|load|Loads objects persisted by the save/multi_save APIs from files.|
|convert|Converts Checkpoint files in MindIO ACP format to the native Torch save format.|
|preload|Preloads Torch-format data and saves it as high-performance MemFS data in MindIO ACP.|
|flush|Waits until all background asynchronous flushing tasks are completed.|
|open_file|Opens a file in read-only mode and returns a readable file handle (supported only by the MindSpore framework).|
|create_file|Creates a file and returns a writable file handle (supported only by the MindSpore framework).|

For complete API parameter descriptions, return values, and usage examples, see [MindIO ACP APIs](../../06_api/09_mindio/01_mindio_acp_apis.md).
