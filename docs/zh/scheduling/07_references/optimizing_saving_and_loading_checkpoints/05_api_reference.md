# API接口参考

MindIO ACP提供高性能异步Checkpoint能力，支持将训练过程中的模型状态快速保存与恢复。关键API如下：

|API|功能说明|
|--|--|
|initialize|初始化MindIO ACP Client，可配置内存池大小、并发写线程数等参数。|
|save|将数据保存到指定路径，支持memfs和fopen两种方式。|
|multi_save|将同一个数据保存到多个文件中，适用于多副本场景。|
|register_checker|注册异步回调函数，用于数据完整性校验。|
|load|从文件中加载save/multi_save接口持久化的对象。|
|convert|将MindIO ACP格式的Checkpoint文件转换为Torch原生保存格式。|
|preload|预加载Torch格式数据并将其保存为MindIO ACP的高性能MemFS数据。|
|flush|等待后台异步刷盘任务全部执行完成。|
|open_file|以只读方式打开文件，返回可读文件句柄（仅支持MindSpore框架）。|
|create_file|创建文件并返回可写文件句柄（仅支持MindSpore框架）。|

> 完整的接口参数说明、返回值及使用样例请参见[MindIO ACP 接口](../../06_api/mindio/01_mindio_acp_apis.md)。
