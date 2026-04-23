# API接口参考

## initialize接口

**接口功能**

初始化MindIO ACP  Client。

**接口格式**

```python
mindio_acp.initialize(server_info: Dict[str, str] = None) -> int
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|server_info|可选|自启动的Server进程需要配置参数信息。若不传入该参数，则全部使用默认值。|有效参数集合或None。|

**表 1**  server\_info参数说明

|参数key|默认参数value|是否必选|说明|取值范围|
|--|--|--|--|--|
|'memfs.data_block_pool_capacity_in_gb'|'128'|可选|MindIO ACP文件系统内存分配大小，单位：GB，根据服务器内存大小来配置，建议不超过系统总内存的25%。|[1, 1024]|
|'memfs.data_block_size_in_mb'|'128'|可选|文件数据块分配最小粒度，单位：MB，根据使用场景中大多数文件的size决定配置，建议平均每个文件的数据块大小不超过128MB。|[1, 1024]|
|'memfs.write.parallel.enabled'|'true'|可选|MindIO ACP并发读写性能优化开关配置，用户需结合业务数据模型特征决定是否打开本配置。|<ul><li>false：关闭</li><li>true：开启</li></ul>|
|'memfs.write.parallel.thread_num'|'16'|可选|MindIO ACP并发读写性能优化并发数。|[2, 96]|
|'memfs.write.parallel.slice_in_mb'|'16'|可选|MindIO ACP并发写性能优化数据切分粒度，单位：MB。|[1, 1024]|
|'background.backup.thread_num'|'32'|可选|备份线程数量。|[1, 256]|

> [!NOTE]说明
> mindio\_acp.initialize如果不传入server\_info参数，则按照表中默认参数启动Server。

**使用样例1**

```python
>>> # Initialize with default param
>>> mindio_acp.initialize()
```

**使用样例2**

```python
>>> # Initialize with server_info
>>> server_info = {
        'memfs.data_block_pool_capacity_in_gb': '200',
    }
>>> mindio_acp.initialize(server_info=server_info)
```

**返回值**

- 0：成功
- -1：失败

## save接口

**接口功能**

将数据保存到指定的路径下。

**接口格式**

```python
mindio_acp.save(obj, path, open_way='memfs')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|obj|必选|需要保存的对象。|有效数据对象。|
|path|必选|数据保存路径。|有效文件路径。|
|open_way|可选|保存方式。<ul><li>memfs：使用MindIO ACP的高性能MemFS保存数据。</li><li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</li></ul>默认值：memfs。|<ul><li>memfs</li><li>fopen</li></ul>|

**使用样例**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> mindio_acp.save(x, '/mnt/dpc01/tensor.pt')
```

**返回值**

- -1：保存失败。
- 0：通过原生torch.save方式实现保存。
- 1：通过memfs方式实现保存。
- 2：通过fopen方式实现保存。

## multi\_save接口

**接口功能**

将同一个数据保存到多个文件中。

**接口格式**

```python
mindio_acp.multi_save(obj, path_list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|obj|必选|需要保存的对象。|有效数据对象。|
|path_list|必选|数据保存路径列表。|有效文件路径列表。|

**使用样例**

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> path_list = ["/mnt/dpc01/dir1/rank_1.pt","/mnt/dpc01/dir2/rank_1.pt"]
>>> mindio_acp.multi_save(x, path_list)
```

**返回值**

- None：失败。
- 0：通过原生torch.save方式实现保存。
- 1：通过memfs方式实现保存。
- 2：通过fopen方式实现保存。

## register\_checker接口

**接口功能**

注册异步回调函数。

**接口格式**

```python
mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|callback|必选|回调函数（第一个参数result为数据完整性校验的结果，0为成功，其他为失败；第二个参数为user_context）。|有效函数名。|
|check_dict|必选|数据完整性校验条件，类型dict，用来校验指定path下的文件个数是否符合要求。|<ul><li>key：path，数据路径。</li><li>value：对应key路径下的文件个数。</li></ul>|
|user_context|必选|回调函数的第二个参数。|-|
|timeout_sec|必选|回调超时时间，单位：秒。<br>如果训练客户端日志中提示："watching checkpoint failed"，则需要调大该参数。代码在mindio_acp实际安装路径（mindio_acp/acc_checkpoint/framework_acp.py）下的async_write_tracker_file函数中。|[1, 3600]|

**使用样例**

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

**返回值**

- None：失败。
- 1：成功。

## load接口

**接口功能**

从文件中加载save/multi\_save接口持久化的对象。

**接口格式**

```python
mindio_acp.load(path, open_way='memfs', map_location=None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|加载路径。|有效文件路径。|
|open_way|可选|加载方式。<ul><li>memfs：使用MindIO ACP的高性能MemFS保存数据。</li><li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</li></ul>默认值：memfs。|<ul><li>memfs</li><li>fopen</li></ul>|
|map_location|可选|加载时需要映射到的设备。默认值：None。|<ul><li>None</li><li>cpu</li></ul>|

**使用样例**

```python
>>> # load from file
>>> mindio_acp.load('/mnt/dpc01/checkpoint/rank-0.pt')
```

**返回值**

Any

> [!CAUTION]注意
> 如同PyTorch的load接口，本接口内部也使用pickle模块，有被恶意构造的数据在unpickle期间攻击的风险。需要保证被加载的数据来源是安全存储的，仅可以load可信的数据。

## convert接口

**接口功能**

将MindIO ACP格式的Checkpoint文件转换为Torch原生保存的格式。

**接口格式**

```python
mindio_acp.convert(src, dst)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|src|必选|待转换的源路径或源文件，源路径或源文件必须存在。|有效文件路径，不能包含软链接。|
|dst|必选|待转换的目标路径或目标文件。指定路径的父目录必须存在，如果文件已存在，则会被覆盖。|有效文件路径，不能包含软链接。|

**使用样例**

```python
>>> mindio_acp.convert('/mnt/dpc01/iter_0000050/mp_rank_00/distrib_optim.pt', '/mnt/dpc02/iter_0000050/mp_rank_00/distrib_optim.pt')
```

**返回值**

- 0：转换成功。
- -1：转换失败。

## preload接口

**接口功能**

从文件中预加载使用torch保存的数据对象，并将其保存为MindIO ACP的高性能MemFS数据。

**接口格式**

```python
mindio_acp.preload(*path)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|预加载的源文件，源文件必须存在。|有效文件路径或路径集合。|

**使用样例**

```python
>>> # preload from file
>>> mindio_acp.preload('/mnt/dpc01/checkpoint/rank-0.pt')
```

**返回值**

- 0：预加载成功。
- 1：预加载失败。

## flush接口

**接口功能**

等待后台异步刷盘任务全部执行成功。

**接口格式**

```python
mindio_acp.flush()
```

**接口参数**

无

**使用样例**

```python
>>> # flush all data to disk
>>> mindio_acp.flush()
```

**返回值**

- 0：刷盘成功。
- 1：刷盘失败。

## open\_file接口

本接口只支持MindSpore框架。

**接口功能**

使用with调用open\_file接口，以只读的方式打开文件，并返回对应的\_ReadableFileWrapper实例。该实例提供read\(\)和close\(\)方法。

- read：读取文件内容。

    ```python
    read(self, offset=0, count=-1)
    ```

    |参数|是否必选|说明|取值要求|
    |--|--|--|--|
    |offset|可选|读取文件的偏移位置。需满足count + offset <= file_size|[0, file_size)|
    |count|可选|读取文件的大小。需满足count + offset <= file_size|<ul><li>-1：读取整个文件。</li><li>(0, file_size]</li></ul>|

- close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```python
    close(self)
    ```

**接口格式**

```python
mindio_acp.open_file(path: str)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|加载路径。|有效文件路径。|

**使用样例**

```python
>>> with mindio_acp.open_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     read_data = f.read()
```

**返回值**

\_ReadableFileWrapper实例。

> [!NOTE]说明
> 接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.load_checkpoint.html#mindspore.load_checkpoint)。

## create\_file接口

本接口只支持MindSpore框架。

**接口功能**

使用with调用create\_file接口，用于创建文件，并返回对应的\_WriteableFileWrapper实例。该实例提供write\(\)、drop\(\)和close\(\)方法。

- write：向文件中写入数据。

    ```python
    write(self, data: bytes)
    ```

    |参数|是否必选|说明|取值要求|
    |--|--|--|--|
    |data|必选|需要写入的对象。|bytes对象。|

- drop：删除文件。

    ```python
    drop(self)
    ```

- close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```python
    close(self)
    ```

**接口格式**

```python
mindio_acp.create_file(path: str, mode: int = 0o600)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|数据保存路径。|有效文件路径。|
|mode|可选|文件创建权限。|[0o000, 0o777]|

**使用样例**

```python
>>> x = b'\x00\x01\x02\x03\x04'
>>> with mindio_acp.create_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     write_result = f.write(x)
```

**返回值**

\_WriteableFileWrapper实例。

> [!NOTE]说明
> 接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.save_checkpoint.html#mindspore.save_checkpoint)。
