# Local run modelarts log diag
 
**1、基本说明**:
 
本脚本用于本地将modelarts获取到的多机日志切分后分别进行单机清洗，清洗后转储并进行诊断。
 
**2、运行说明**
 
`python local_run_modelarts_log_diag.py -i {INPUT_PATH} -o {OUTPUT_PATH}`
 
示例：`python local_run_modelarts_log_diag.py -i modelarts-log-dir/ -o output/`
 
**3、参数说明**
 
`-i {INPUT_PATH}`，输入目录，指定到 Modelarts Log Path 文件路径，支持相对路径与绝对路径，路径下日志文件必须符合规范，否则会导致诊断失败
 
`-o {OUTPUT_PATH}`，输出目录，指定到诊断完毕的报告输出目录，支持相对路径与绝对路径，支持不存在的路径

**4、使用说明**

诊断结果文件存放在`{OUTPUT_PATH}/fault_diag_result/`下
 
**5、注意事项**
 
1、如果输出路径不存在，将会创建。

2、如果输出路径存在，则必须为空，否则会抛出异常。
 