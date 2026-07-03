# API Reference

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:29:19.691Z pushedAt=2026-06-09T07:15:15.754Z -->

## initialize

**Function**

Initializes the MindIO ACP Client.

**Format**

```python
mindio_acp.initialize(server_info: Dict[str, str] = None) -> int
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|server_info|Optional|Configuration required for a self-started server process. If this parameter is not passed, all default values are used.|A valid set of parameters or None.|

**Table 1** server_info parameter description

|Parameter Key|Default Value|Mandatory|Description|Value Range|
|--|--|--|--|--|
|'memfs.data_block_pool_capacity_in_gb'|'128'|Optional|Memory allocation size for the MindIO ACP file system, in GB. Configure based on the server memory size. It is recommended not to exceed 25% of the total system memory.|[1, 1024]|
|'memfs.data_block_size_in_mb'|'128'|Optional|Minimum granularity for file data block allocation, in MB. Determined by the size of most files in the usage scenario. It is recommended that the average data block size per file does not exceed 128 MB.|[1, 1024]|
|'memfs.write.parallel.enabled'|'true'|Optional|Switch for the MindIO ACP concurrent read/write performance optimization. Decide whether to enable this configuration based on the characteristics of the business data model.|<ul><li>false: Disabled</li><li>true: Enabled</li></ul>|
|'memfs.write.parallel.thread_num'|'16'|Optional|Number of concurrent threads for MindIO ACP concurrent read/write performance optimization.|[2, 96]|
|'memfs.write.parallel.slice_in_mb'|'16'|Optional|Data slicing granularity for MindIO ACP concurrent write performance optimization, in MB.|[1, 1024]|
|'background.backup.thread_num'|'32'|Optional|Number of backup threads.|[1, 256]|

> [!NOTE]
> If `server_info` is not passed to m`indio_acp.initialize`, the server starts with the default parameters listed in the table.

**Example 1**

```python
>>> # Initialize with default param
>>> mindio_acp.initialize()
```

**Example 2**

```python
>>> # Initialize with server_info
>>> server_info = {
        'memfs.data_block_pool_capacity_in_gb': '200',
    }
>>> mindio_acp.initialize(server_info=server_info)
```

**Return Value**

- `0`: Success
- `-1`: Failure

## save

**Function**

Saves data to the specified path.

**Format**

```python
mindio_acp.save(obj, path, open_way='memfs')
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|obj|Mandatory|Object to save.|Valid data object.|
|path|Mandatory|Data storage path.|Valid file path.|
|open_way|Optional|Save method.<ul><li>memfs: Uses the high-performance MemFS of MindIO ACP to save data.</li><li>fopen: Calls file operation functions in the C standard library to save data, typically serving as a backup for the memfs method.</li></ul>Default value: memfs.|<ul><li>memfs</li><li>fopen</li></ul>|

**Example**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> mindio_acp.save(x, '/mnt/dpc01/tensor.pt')
```

**Return Value**

- `-1`: Failure.
- `0`: Save is implemented via the native `torch.save` method.
- `1`: Save using the memfs method.
- 2`: Save using the fopen method.

## multi_save

**Function**

Saves the same data to multiple files.

**Format**

```python
mindio_acp.multi_save(obj, path_list)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|obj|Mandatory|Object to save.|Valid data object.|
|path_list|Mandatory|List of data storage paths.|List of valid file paths.|

**Example**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> path_list = ["/mnt/dpc01/dir1/rank_1.pt","/mnt/dpc01/dir2/rank_1.pt"]
>>> mindio_acp.multi_save(x, path_list)
```

**Return Value**

- `None`: Failure.
- `0`: Save using the native `torch.save` method.
- `1`: Save using the memfs method.
- `2`: Save using the fopen method.

## register_checker

**Function**

Registers an asynchronous callback.

**Format**

```python
mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|callback|Mandatory|Callback (the first parameter, result, is the result of the data integrity check, 0 indicates success, and any other value indicates failure; the second parameter is user_context).|Valid function name.|
|check_dict|Mandatory|Data integrity check condition, of type dict, used to verify whether the number of files under the specified path meets the requirement.|<ul><li>key: path, the data path.</li><li>value: the number of files under the path corresponding to the key.</li></ul>|
|user_context|Mandatory|The second parameter of the callback.|-|
|timeout_sec|Mandatory|Callback timeout period, in seconds.<br>If the training client log prompts "watching checkpoint failed", increase this parameter value. The code is in the async_write_tracker_file function under the actual installation path of MindIO ACP (mindio_acp/acc_checkpoint/framework_acp.py).|[1, 3600]|

**Example**

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

- `None`: Failure.
- `1`: Success.

## load

**Function**

Loads objects persisted by the `save` or `multi_save` interface from a file.

**Format**

```python
mindio_acp.load(path, open_way='memfs', map_location=None)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|path|Mandatory|Load Path.|Valid file path.|
|open_way|Optional|Loading method.<ul><li>memfs: Uses the high-performance MemFS of MindIO ACP to save data.</li><li>fopen: Calls file operation functions in the C standard library to save data, typically used as a backup for the memfs method.</li></ul>Default Value: memfs.|<ul><li>memfs</li><li>fopen</li></ul>|
|map_location|Optional|The device to which the loaded data should be mapped. Default Value: None.|<ul><li>None</li><li>cpu</li></ul>|

**Example**

```python
>>> # load from file
>>> mindio_acp.load('/mnt/dpc01/checkpoint/rank-0.pt')
```

**Return Value**

Any

> [!CAUTION] **CAUTION**
> Like the PyTorch load interface, this interface also uses the pickle module internally, which carries the risk of attacks from maliciously crafted data during unpickling. Ensure that the loaded data comes from a secure source and only load trusted data.

## convert

**Function**

Converts a checkpoint file in MindIO ACP format to the Torch native save format.

**Format**

```python
mindio_acp.convert(src, dst)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|src|Mandatory|Source path or source file to be converted. The source path or source file must exist.|Valid file path, cannot contain soft links.|
|dst|Mandatory|Destination path or destination file to be converted. The parent directory of the specified path must exist. If the file already exists, it will be overwritten.|Valid file path, cannot contain soft links.|

**Example**

```python
>>> mindio_acp.convert('/mnt/dpc01/iter_0000050/mp_rank_00/distrib_optim.pt', '/mnt/dpc02/iter_0000050/mp_rank_00/distrib_optim.pt')
```

**Return Value**

- `0`: Conversion success.
- `-1`: Conversion failure.

## preload

**Function**

Preloads data objects saved using torch from a file and saves them as high-performance MemFS data for MindIO ACP.

**Format**

```python
mindio_acp.preload(*path)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|path|Mandatory|Source file to preload. The source file must exist.|Valid file path or set of paths.|

**Example**

```python
>>> # preload from file
>>> mindio_acp.preload('/mnt/dpc01/checkpoint/rank-0.pt')
```

**Return Value**

- `0`: Preloading success.
- `1`: Preloading failure.

## flush

**Function**

Waits for all asynchronous flush tasks at background to complete successfully.

**Format**

```python
mindio_acp.flush()
```

**Parameters**

None

**Example**

```python
>>> # flush all data to disk
>>> mindio_acp.flush()
```

**Return Value**

- `0`: Flush success.
- `1`: Flush failure.

## open_file

This interface supports only the MindSpore framework.

**Function**

Uses `with` to call the `open_file` interface to open a file in read-only mode and returns the corresponding `_ReadableFileWrapper` instance. This instance provides the `read()` and `close()` methods.

- `read()`: Reads the file content.

    ```python
    read(self, offset=0, count=-1)
    ```

    |Parameter|Mandatory|Description|Value Requirement|
    |--|--|--|--|
    |offset|Optional|The offset position for reading the file. Must satisfy "count + offset <= file_size"|[0, file_size)|
    |count|Optional|The size of the file to read. Must satisfy "count + offset <= file_size"|<ul><li>-1: Reads the entire file.</li><li>(0, file_size]</li></ul>|

- `close()`: Closes the file.

    This method is automatically called when the `with` context exits.

    ```python
    close(self)
    ```

**Format**

```python
mindio_acp.open_file(path: str)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|path|Mandatory|Load path.|Valid file path.|

**Example**

```python
>>> with mindio_acp.open_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     read_data = f.read()
```

**Return Value**

`_ReadableFileWrapper` instance.

> **NOTE**
> For details, see [MindSpore documentation](https://www.mindspore.cn/docs/en/master/api_python/mindspore/mindspore.load_checkpoint.html).

## create_file

This interface supports only the MindSpore framework.

**Function Description**

Calls `create_file` using `with` to create a file and returns the corresponding `_WriteableFileWrapper` instance. This instance provides the `write()`, `drop()`, and `close()` methods.

- `write()`: Writes data to the file.

    ```python
    write(self, data: bytes)
    ```

    |Parameter|Mandatory|Description|Value Requirement|
    |--|--|--|--|
    |data|Mandatory|Object to write.|bytes object.|

- `drop()`: Deletes a file.

    ```python
    drop(self)
    ```

- `close()`: Closes a file.

    This method is automatically called when the `with` context exits.

    ```python
    close(self)
    ```

**Format**

```python
mindio_acp.create_file(path: str, mode: int = 0o600)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|path|Yes|Data storage path.|Valid file path.|
|mode|No|File creation permission.|[0o000, 0o777]|

**Example**

```python
>>> x = b'\x00\x01\x02\x03\x04'
>>> with mindio_acp.create_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     write_result = f.write(x)
```

**Return Value**

`_WriteableFileWrapper` instance.

> **NOTE**
> For details, see [MindSpore documentation](https://www.mindspore.cn/docs/en/master/api_python/mindspore/mindspore.save_checkpoint.html).
