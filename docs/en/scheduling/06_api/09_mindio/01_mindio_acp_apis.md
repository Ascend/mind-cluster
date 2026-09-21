# MindIO ACP APIs

## initialize

**Function**

Initializes the MindIO ACP Client.

**Format**

```python
mindio_acp.initialize(server_info: Dict[str, str] = None) -> int
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|server_info|Optional|The self-started Server process requires configuration parameter information. If this parameter is not passed in, all default values are used.|A valid parameter set or None.|

**Table 1**  server\_info parameter description

|Parameter key|Default parameter value|Mandatory|Description|Value Range|
|--|--|--|--|--|
|'memfs.data_block_pool_capacity_in_gb'|'128'|Optional|Memory allocation size of the MindIO ACP file system, in GB. Configure it based on the server memory size. It is recommended not to exceed 25% of the total system memory.|[1, 1024]|
|'memfs.data_block_size_in_mb'|'128'|Optional|Minimum granularity for file data block allocation, in MB. Configure it based on the size of most files in the usage scenario. It is recommended that the average data block size of each file does not exceed 128 MB.|[1, 1024]|
|'memfs.write.parallel.enabled'|'true'|Optional|Switch configuration for MindIO ACP concurrent read/write performance optimization. Users need to decide whether to enable this configuration based on the characteristics of the business data model.|<ul><li>false: Disabled</li><li>true: Enabled</li></ul>|
|'memfs.write.parallel.thread_num'|'16'|Optional|Concurrency for MindIO ACP concurrent read/write performance optimization.|[2, 96]|
|'memfs.write.parallel.slice_in_mb'|'16'|Optional|Data slicing granularity for MindIO ACP concurrent write performance optimization, in MB.|[1, 1024]|
|'background.backup.thread_num'|'32'|Optional|Number of backup threads.|[1, 256]|

> [!NOTE]
> If mindio\_acp.initialize does not pass in the server\_info parameter, the Server is started with the default parameters in the table.

**Usage Example 1**

```python
>>> # Initialize with default param
>>> mindio_acp.initialize()
```

**Usage Example 2**

```python
>>> # Initialize with server_info
>>> server_info = {
        'memfs.data_block_pool_capacity_in_gb': '200',
    }
>>> mindio_acp.initialize(server_info=server_info)
```

**Return Value**

- 0: Success
- -1: Failure

## save

**Function**

Saves data to the specified path.

**Format**

```python
mindio_acp.save(obj, path, open_way='memfs')
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|obj|Yes|Object to be saved.|Valid data object.|
|path|Yes|Data saving path.|Valid file path.|
|open_way|No|Saving method.<ul><li>memfs: Saves data using the high-performance MemFS of MindIO ACP.</li><li>fopen: Saves data by calling file operation functions in the C standard library, usually serving as a backup for the memfs method.</li></ul>Default value: memfs.|<ul><li>memfs</li><li>fopen</li></ul>|

**Usage Example**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> mindio_acp.save(x, '/mnt/dpc01/tensor.pt')
```

**Return Value**

- -1: Saving failed.
- 0: Saving is implemented through the native torch.save method.
- 1: Saves data through the memfs method.
- 2: Saves data through the fopen method.

## multi_save

**Function**

Saves the same data to multiple files.

**Format**

```python
mindio_acp.multi_save(obj, path_list)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|obj|Mandatory|Object to be saved.|Valid data object.|
|path_list|Mandatory|List of data saving paths.|List of valid file paths.|

**Usage Example**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> path_list = ["/mnt/dpc01/dir1/rank_1.pt","/mnt/dpc01/dir2/rank_1.pt"]
>>> mindio_acp.multi_save(x, path_list)
```

**Return Value**

- None: Failure.
- 0: Saving is implemented through the native torch.save method.
- 1: Saves data using the memfs implementation method.
- 2: Saves data using the fopen implementation method.

## register_checker

**Function**

Registers an asynchronous callback function.

**Format**

```python
mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|callback|Yes|Callback function (the first parameter result is the result of data integrity check, 0 indicates success, and other values indicate failure; the second parameter is user_context).|Valid function name.|
|check_dict|Yes|Data integrity check condition, of type dict, used to check whether the number of files under the specified path meets the requirement.|<ul><li>key: path, the data path.</li><li>value: the number of files under the path corresponding to the key.</li></ul>|
|user_context|Yes|The second parameter of the callback function.|-|
|timeout_sec|Yes|Callback timeout, in seconds.<br>If the training client log shows "watching checkpoint failed", increase this parameter. The code is in the async_write_tracker_file function under the actual installation path of mindio_acp (mindio_acp/acc_checkpoint/framework_acp.py).|[1, 3600]|

**Usage Example**

```python
>>> def callback(result, user_context):
>>>     if result == 0:
>>>         print("success")
>>>     else:
>>>         print("fail")
>>> context_obj = None
>>> check_dict = {'/mnt/dpc01/checkpoint-last': 4}
>>> mindio_acp.register_checker(callback, check_dict, context_obj, 1000)
```

**Return Value**

- None: Failure.
- 1: Success.

## load

**Function**

Loads objects persisted by the save/multi_save APIs from a file.

**Format**

```python
mindio_acp.load(path, open_way='memfs', map_location=None)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|path|Yes|Loading path.|Valid file path.|
|open_way|No|Loading method.<ul><li>memfs: Uses MindIO ACP's high-performance MemFS to load data.</li><li>fopen: Calls file operation functions in the C standard library to load data, usually serving as a backup for the memfs method.</li></ul>Default value: memfs.|<ul><li>memfs</li><li>fopen</li></ul>|
|map_location|No|Device to which data is mapped during loading. Default value: None.|<ul><li>None</li><li>cpu</li></ul>|

**Usage Example**

```python
>>> # load from file
>>> mindio_acp.load('/mnt/dpc01/checkpoint/rank-0.pt')
```

**Return Value**

Any

> [!CAUTION]
> Like PyTorch's load API, this API also uses the pickle module internally, which carries the risk of being attacked by maliciously crafted data during unpickling. Ensure that the source of the loaded data is securely stored, and load only trusted data.

## convert

**Function**

Converts a Checkpoint file in MindIO ACP format to the format saved natively by Torch.

**Format**

```python
mindio_acp.convert(src, dst)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|src|Yes|Source path or source file to be converted. The source path or source file must exist.|Valid file path, which cannot contain symbolic links.|
|dst|Yes|Target path or target file to be converted. The parent directory of the specified path must exist. If the file already exists, it will be overwritten.|Valid file path, which cannot contain symbolic links.|

**Usage Example**

```python
>>> mindio_acp.convert('/mnt/dpc01/iter_0000050/mp_rank_00/distrib_optim.pt', '/mnt/dpc02/iter_0000050/mp_rank_00/distrib_optim.pt')
```

**Return Value**

- 0: Conversion succeeded.
- -1: Conversion failed.

## preload

**Function**

Preloads data objects saved with torch from a file and saves them as high-performance MemFS data for MindIO ACP.

**Format**

```python
mindio_acp.preload(*path)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|path|Yes|Source file to be preloaded. The source file must exist.|A valid file path or a collection of paths.|

**Usage Example**

```python
>>> # preload from file
>>> mindio_acp.preload('/mnt/dpc01/checkpoint/rank-0.pt')
```

**Return Value**

- 0: Preloading succeeded.
- 1: Preloading failed.

## flush

**Function**

Waits for all background asynchronous flush tasks to complete successfully.

**Format**

```python
mindio_acp.flush()
```

**Parameters**

None

**Usage Example**

```python
>>> # flush all data to disk
>>> mindio_acp.flush()
```

**Return Value**

- 0: Flush succeeded.
- 1: Flush failed.

## open_file

This API is supported only by the MindSpore framework.

**Function**

Calls the open\_file API using with to open a file in read-only mode and returns the corresponding\_ReadableFileWrapper instance. This instance provides the read() and close() methods.

- read: Reads the file content.

    ```python
    read(self, offset=0, count=-1)
    ```

  |Name|Mandatory|Description|Value Requirements|
  |--|--|--|--|
  |offset|Optional|Offset position for file reading. Must satisfy count + offset <= file_size.|[0, file_size)|
  |count|Optional|Size of the file to read. Must satisfy count + offset <= file_size.|<ul><li>-1: Read the entire file.</li><li>(0, file_size]</li></ul>|

- close: Closes the file.

  This method is automatically called when the with context exits.

    ```python
    close(self)
    ```

**Format**

```python
mindio_acp.open_file(path: str)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|path|Mandatory|Loading path.|Valid file path.|

**Usage Example**

```python
>>> with mindio_acp.open_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     read_data = f.read()
```

**Return Value**

\_ReadableFileWrapper instance.

> [!NOTE]Description
> For details, see the [MindSpore documentation](https://www.mindspore.cn/docs/en/master/api_python/mindspore/mindspore.load_checkpoint.html#mindspore.load_checkpoint).

## create_file

This API supports only the MindSpore framework.

**Function**

Calls the create_file API with the `with` statement to create a file and return the corresponding_WriteableFileWrapper instance. This instance provides the write(), drop(), and close() methods.

- write: Writes data to the file.

    ```python
    write(self, data: bytes)
    ```

  |Name|Mandatory|Description|Value Requirements|
  |--|--|--|--|
  |data|Mandatory|Object to be written.|bytes object.|

- drop: Deletes the file.

    ```python
    drop(self)
    ```

- close: Closes the file.

  This method is automatically called when the with context exits.

    ```python
    close(self)
    ```

**Format**

```python
mindio_acp.create_file(path: str, mode: int = 0o600)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|path|Yes|Data saving path.|Valid file path.|
|mode|No|File creation permission.|[0o000, 0o777]|

**Usage Example**

```python
>>> x = b'\x00\x01\x02\x03\x04'
>>> with mindio_acp.create_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     write_result = f.write(x)
```

**Return Value**

\_WriteableFileWrapper instance.

> [!NOTE]
> For details about the API, see [MindSpore documentation](https://www.mindspore.cn/docs/en/master/api_python/mindspore/mindspore.save_checkpoint.html#mindspore.save_checkpoint).
>
