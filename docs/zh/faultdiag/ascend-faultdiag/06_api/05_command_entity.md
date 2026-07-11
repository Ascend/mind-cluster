# entity命令（自定义故障实体）

## 功能说明

用于管理自定义故障实体，可以新增、查看、删除自定义的故障检测规则。

## 命令格式

```shell
ascend-fd entity [-h] (-u UPDATE | -d DELETE [DELETE ...] | -s [SHOW ...] | -c CHECK)
```

## 参数说明

| 参数         | 类型   | 必选                       | 说明                                         |
|--------------|--------|----------------------------|----------------------------------------------|
| -h, --help   | -      | 否                         | 显示帮助信息                                 |
| -u, --update | string | 必选（与 -d, -s, -c 互斥） | 新增或修改自定义故障实体的 JSON 文件路径     |
| -d, --delete | string | 必选（与 -u, -s, -c 互斥） | 删除指定故障码的自定义故障实体               |
| -s, --show   | string | 必选（与 -u, -d, -c 互斥） | 查看自定义故障实体信息                       |
| -c, --check  | string | 必选（与 -u, -d, -s 互斥） | 校验 custom-ascend-kg-config.json 文件路径   |
| --item       | string | 可选                       | 查看部分信息，可选值：attribute, rule, regex |
| -f, --force  | -      | 可选                       | 删除时跳过确认提示                           |

## 使用示例

### 新增或修改自定义故障实体

1. 通过 JSON 文件新增或修改自定义故障实体。JSON 文件最多支持 1000 条自定义故障信息。

    ```shell
    ascend-fd entity --update updated_entity.json
    ```

    回显 `Updated entity successfully.` 表示操作成功。

    json 文件示例和参数说明请参考 [JSON文件字段说明](#json文件字段说明)。

### 查看自定义故障实体

```shell
ascend-fd entity --show
```

### 删除指定故障码的自定义实体

```shell
ascend-fd entity --delete 故障码1 故障码2
```

### 查看所有自定义故障实体

```shell
ascend-fd entity -s
```

### 按故障码查询

```shell
ascend-fd entity -s 故障码1 故障码2
```

### 查看指定属性信息

```shell
ascend-fd entity -s --item attribute rule regex
```

### 跳过确认提示

```shell
ascend-fd entity -d 故障码1 --force
```

### 校验故障实体文件

```shell
ascend-fd entity -c 自定义故障实体JSON文件
```

## JSON文件字段说明

### JSON文件示例

```json
{
    "41001": {      // 故障码，用户需根据实际情况自定义故障码，不能与MindCluster Ascend FaultDiag已支持的故障码相同
        "attribute.class": "Software",
        "attribute.component": "AI Framework",
        "attribute.module": "Compiler",
        "attribute.cause_zh": "抽象类型合并失败",
        "attribute.description_zh": "对函数输出求梯度时，抽象类型不匹配，导致抽象类型合并失败。",
        "attribute.suggestion_zh": [
               "1. 检查求梯度的函数的输出类型与sens_param的类型是否相同，如果不相同，修改为相同类型；",
               "2. 自动求导报错Type Join Failed"
           ],
        "attribute.cause_en": "Abstract type merge failed",
        "attribute.description_en": "When computing the gradient of a function output, the abstract types do not match, leading to a failure in abstract type merging.",
        "attribute.suggestion_en": [
               "1. Check whether the output type of the gradient calculation function matches the type of sens_param. If they do not match, modify them to be of the same type.",
               "2. Automatic differentiation reports an error: Type Join Failed."
           ],
        "attribute.error_case": [
            "grad = ops.GradOperation(sens_param=True)",
            "# test_net输出类型为tuple(Tensor, Tensor)",
            "def test_net(a, b):",
            "    return a, b"
              ],
        "attribute.fixed_case": [
            "grad = ops.GradOperation(sens_param=True)",
            "# test_net输出类型为tuple(Tensor, Tensor)",
            "def test_net(a, b):",
            "    return a, b"
            ],
        "rule": [
            {
                "dst_code": "20106"
            }
        ],
        "source_file": "TrainLog",
        "regex.in": [
            "Abstract type", "cannot join with"
            ]
    }
}
```

### JSON参数说明

<!-- markdownlint-disable MD033 -->
<table><thead><tr><th><p>参数名称</p>
</th>
<th><p>取值类型</p>
</th>
<th><p>参数说明</p>
</th>
<th><p>是否必选</p>
</th>
<th><p>取值说明</p>
</th>
</tr>
</thead>
<tbody><tr><td><p>attribute.class</p>
</td>
<td><p>String</p>
</td>
<td><p>故障类别</p>
</td>
<td><p>必选</p>
</td>
<td rowspan="3"><p>取值长度为1~50个字符，支持英文字母、数字、英文符号与空格。</p>

</td>
</tr>
<tr><td><p>attribute.component</p>
</td>
<td><p>String</p>
</td>
<td><p>故障组件</p>
</td>
<td><p>必选</p>
</td>
</tr>
<tr><td><p>attribute.module</p>
</td>
<td><p>String</p>
</td>
<td><p>故障模块</p>
</td>
<td><p>必选</p>
</td>
</tr>
<tr><td><p>attribute.cause_zh</p>
</td>
<td><p>String</p>
</td>
<td><p>故障原因（中文）</p>
</td>
<td><p>必选</p>
</td>
<td><p>取值长度为1~200个字符，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</p>
</td>
</tr>
<tr><td><p>attribute.cause_en</p>
</td>
<td><p>String</p>
</td>
<td><p>故障原因（英文）</p>
</td>
<td><p>可选</p>
</td>
<td><p>取值长度为1~200个字符，支持英文字母、数字、英文符号与空格。</p>
</td>
</tr>
<tr><td><p>attribute.description_zh</p>
</td>
<td><p>String</p>
</td>
<td><p>故障描述（中文）</p>
</td>
<td><p>必选</p>
</td>
<td rowspan="6"><div>支持字符串或列表。字符串为整段信息，可换行；列表则每一个元素为一行信息，组合起来为整段信息。<ul><li>字符串：取值长度为1~2000个字符，支持英文字母、数字、英文符号、中文汉字、中文符号、空格与“\n”。</li><li>列表：列表下每个字符串的取值长度为1~200，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</li></ul>
</div>
</td>
</tr>
<tr><td><p>attribute.description_en</p>
</td>
<td><p>String</p>
</td>
<td><p>故障描述（英文）</p>
</td>
<td><p>可选</p>
</td>
</tr>
<tr><td><p>attribute.suggestion_zh</p>
</td>
<td><p>String</p>
</td>
<td><p>建议方案（中文）</p>
</td>
<td><p>必选</p>
</td>
</tr>
<tr><td><p>attribute.suggestion_en</p>
</td>
<td><p>String</p>
</td>
<td><p>建议方案（英文）</p>
</td>
<td><p>可选</p>
</td>
</tr>
<tr><td><p>attribute.error_case</p>
</td>
<td><p>String</p>
</td>
<td><p>错误示例</p>
</td>
<td><p>可选</p>
</td>
</tr>
<tr><td><p>attribute.fixed_case</p>
</td>
<td><p>String</p>
</td>
<td><p>修正示例</p>
</td>
<td><p>可选</p>
</td>
</tr>
<tr><td><p>rule</p>
</td>
<td><p>列表</p>
</td>
<td><p>故障链，存储该故障所有触发的下一级故障实体</p>
</td>
<td><p>可选</p>
</td>
<td><p>列表内包含以下字段。</p>
<ul><li>dst_code：必选，表示本次故障触发的下一级故障实体故障码，该故障码必须为 ascend-fd 已支持的故障码或用户自定义故障码。</li><li>expression：可选，表示故障触发约束，当前为预留字段。取值长度为1~200个字符，支持英文字母、数字、英文符号与空格。</li></ul>
</td>
</tr>
<tr><td><p>source_file</p>
</td>
<td><p>String</p>
</td>
<td><p>故障日志文件</p>
</td>
<td><p>必选</p>
</td>
<td>
<p>各个日志文件类型对应的日志文件名称。</p>
</td>
</tr>
<tr><td><p>regex.in</p>
</td>
<td><p>String</p>
</td>
<td><p>故障关键词</p>
</td>
<td><p>必选</p>
</td>
<td><div>支持一级列表与二级列表。<ul><li>一级列表<ul><li>每个元素为字符串。取值长度为1~200个字符，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</li><li>列表中每个关键词都需要满足存在性判断，且符合前后关系</li></ul>
</li><li>二级列表<ul><li>每个子列表满足一级列表的取值约束。</li><li>每个子列表内的判断规则同一级列表，每个子列表间为或关系，仅需满足一个子列表的关键词即可。</li></ul>
</li></ul>
</div>
</td>
</tr>
<tr><td colspan="5"><ul><li>新增自定义故障实体时，所有必选字段都需要存在JSON文件中，且符合相关取值要求。</li><li>修改自定义故障实体时，只需要符合相关取值要求即可。</li></ul>
</td>
</tr>
</tbody>
</table>
<!-- markdownlint-enable MD033 -->

## 注意事项

- JSON 文件最多支持 1000 条自定义故障信息
- 故障码取值长度需为 1~50 个字符
- 故障码不能与 ascend-fd 已支持的故障码重复
- 用户新增故障实体时，数据保存在`$HOME/.ascend_faultdiag/custom-ascend-kg-config.json`文件中
- 用户可通过修改`ASCEND_FD_HOME_PATH`环境变量来指定自定义故障实体文件路径，请查阅[参考 -> 常用操作 -> 环境变量](../07_references/01_common_operations.md#环境变量)
- ascend-fd 运行错误码请查阅[参考 -> 常用操作 -> 组件错误码](../07_references/04_appendix.md#组件错误码)
