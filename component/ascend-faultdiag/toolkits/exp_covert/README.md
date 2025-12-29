# Exp Convertor

**1、基本说明**:

本脚本用于将mindspore troublshooter工具中的traceback专家经验转换为ascend-fd经验格式。

**2、运行说明**

`python convertor.py -s {Suggestion Json File Path} -r {Regex Json File Path} -f {Fault Code Json File Path}`

示例：`ascend-fd parse -s sug.json -r reg.json -f fc.json`

**3、参数说明**

`-s {Suggestion Json File Path}`，Suggestion json 文件路径，支持相对路径与绝对路径，支持不存在的路径

`-r {Regex Json File Path}`，Regex json 文件路径，支持相对路径与绝对路径，支持不存在的路径

`-f {Fault Code Json File Path}`，Fault Code json 文件路径，支持相对路径与绝对路径，支持不存在的路径

**4、使用说明**

1、原始ms troubleshooter经验文件（.py文件）复制到exp_lib_dir中，并在`convertor.py`里导入所有py经验模块，并把这些模块组合为一个list`exp_lib_list`；

```
from exp_lib_dir import common_exp_lib, compiler_exp_lib, dataset_exp_lib, front_exp_lib, operators_exp_lib, vm_exp_lib
exp_lib_list = [common_exp_lib, compiler_exp_lib, dataset_exp_lib, front_exp_lib, operators_exp_lib, vm_exp_lib]
```

2、运行上述命令`python convertor.py -s {Suggestion Json File Path} -r {Regex Json File Path} -f {Fault Code Json File Path}`

**5、注意事项**

如果路径文件不存在，将会创建空文件并以空原始经验进行新经验的更新与添加。

如果存在以前已转换过的经验，那么再次转换时会按照ID对以前转换过的经验保留相同fault code并更新，未存在与fault code 文件中的ID（新错误类型）则按最大code值往后追加并更新经验（包括建议与正则规则）。

注意，转换后suggestion库内部分内容需要人工调整，主要涉及sugesstion_zh和reference。主要是单行过长，需要人工手动添加"\n"。

注意，后续增加时每次建议只转换增加的部分，若全量转换会使得历史已转换过的经验被更新（人工修改添加的"\n"会被覆盖成以前的经验）。
