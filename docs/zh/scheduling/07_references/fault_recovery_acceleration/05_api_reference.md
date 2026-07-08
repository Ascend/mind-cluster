# API接口参考

MindIO TFT提供训练故障容错能力，支持故障检测、故障恢复、优化器状态备份与恢复等功能。关键API如下：

|API|功能说明|
|--|--|
|tft_init_controller|初始化MindIO TFT Controller模块。|
|tft_start_controller|启动Controller模块服务，绑定IP和端口。|
|tft_destroy_controller|训练完成后关闭Controller服务。|
|tft_init_processor|初始化MindIO TFT Processor模块，配置rank、副本数、TLS等参数。|
|tft_start_processor|启动Processor模块服务，连接到Controller。|
|tft_destroy_processor|训练完成后关闭Processor服务。|
|tft_start_updating_os|优化器状态更新前，标记优化器状态为Updating。|
|tft_start_copy_os|通知Processor开始复制优化器状态。|
|tft_end_updating_os|优化器状态更新完成后，标记优化器状态为Updated。|
|tft_set_optimizer_replica|设置rank对应的优化器状态数据副本关系。|
|tft_exception_handler|装饰器，捕获训练状态异常并上报处理。|
|tft_set_step_args|设置训练框架参数集合，供回调函数使用。|
|tft_register_rename_handler|注册rename回调，将临终Checkpoint重命名。|
|tft_register_save_ckpt_handler|注册dump回调，完成临终Checkpoint保存。|
|tft_register_exit_handler|注册用户自定义退出方法（仅MindSpore框架）。|
|tft_register_stop_handler|注册停止训练的回调函数。|
|tft_register_clean_handler|注册清理残留算子执行的回调函数。|
|tft_register_rebuild_group_handler|注册MindIO ARF重新建组的回调函数。|
|tft_register_repair_handler|注册repair回调，完成优化器修复等数据修复。|
|tft_register_rollback_handler|注册rollback回调，完成数据集回滚等重置操作。|
|tft_register_stream_sync_handler|注册同步回调，确保训练暂停后算子队列无残留。|

> 完整的接口参数说明、返回值及使用样例请参见[MindIO TFT 接口](../../06_api/mindio/00_mindio_tft_apis.md)。
