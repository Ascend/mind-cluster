# entity Command (Custom Fault Entities)

<!-- md-trans-meta sourceCommit=0bdfcd26262ca7cc1165bb55c8f28c6a29b377ef translatedAt=2026-08-24T02:29:17.798Z pushedAt=2026-08-24T02:51:43.290Z -->

## Function Description

Used to manage custom fault entities. You can add, view, and delete custom fault detection rules.

## Command Format

```shell
ascend-fd entity [-h] (-u JSON_PATH | -d CODE [CODE...] | -s [CODE [CODE...]] | -c JSON_PATH) [--item ITEM [ITEM...]] [-f]
```

## Parameter Description

| Parameter         | Type                | Mandatory                          | Description                                                          |
|--------------|---------------------|------------------------------------|----------------------------------------------------------------------|
| `-h`, `--help`   | -                   | No                                 | Displays help information.                                           |
| `-u`, `--update` | String              | Mandatory (mutually exclusive with `-d`, `-s`, `-c`) | Adds or modifies the JSON file of a custom fault entity.             |
| `-d`, `--delete` | String/List[String] | Mandatory (mutually exclusive with `-u`, `-s`, `-c`) | Deletes the custom fault entity of the specified fault code. Supports deleting multiple fault codes. |
| `-s`, `--show`   | String/List[String] | Mandatory (mutually exclusive with `-u`, `-d`, `-c`) | Views custom fault entity information. Supports viewing by fault code. |
| `-c`, `--check`  | String              | Mandatory (mutually exclusive with `-u`, `-d`, `-s`) | Validates the custom fault entity JSON file.                         |
| `--item`       | String/List[String] | Optional (only used with `-s`)      | Views partial information. Optional values: `attribute`, `rule`, `regex`.  |
| `-f`, `--force`  | -                   | Optional (only used with `-d`)      | Skips the confirmation prompt when deleting.                         |

## Usage Examples

### Add or Modify a Custom Fault Entity

Add or modify a custom fault entity through a JSON file. A JSON file supports up to 1,000 custom fault records.

```shell
ascend-fd entity --update <updated_entity.json>
```

The output `Updated entity successfully.` indicates that the operation is successful.

For the JSON file example and parameter description, see [JSON File Field Description](#json-file-field-description).

### View Custom Fault Entities

```shell
ascend-fd entity -s
```

### Query by Fault Code

```shell
ascend-fd entity -s <fault_code_1> <fault_code_2>
```

### View Specified Attribute Information

```shell
ascend-fd entity -s --item attribute rule regex
```

### Validate Fault Entity File

```shell
ascend-fd entity -c <custom_entity.json>
```

### Delete Custom Fault Entities of Specified Fault Codes

```shell
ascend-fd entity -d <fault_code_1> <fault_code_2>
```

### Skip Confirmation Prompt When Deleting

```shell
ascend-fd entity -d <fault_code_1> --force
```

## JSON File Field Description

### JSON File Example

```jsonc
{
    "41001": {      // Fault code. Users must customize the fault code based on actual conditions. It must not be the same as the fault codes already supported by MindCluster Ascend FaultDiag.
        "attribute.class": "Software",
        "attribute.component": "AI Framework",
        "attribute.module": "Compiler",
        "attribute.cause_zh": "抽象类型合并失败",
        "attribute.description_zh": "对函数输出求梯度时，抽象类型不匹配，导致抽象类型合并失败。",
        "attribute.suggestion_zh": [
               "1. 检查求梯度的函数的输出类型与sens_param的类型是否相同，如果不相同，修改为相同类型；",
               "2. 自动求导报错Type Join Failed"
           ],
        "attribute.cause_en": "Abstract type merging failed",
        "attribute.description_en": "When computing the gradient of a function output, the abstract types do not match, leading to a failure in abstract type merging.",
        "attribute.suggestion_en": [
               "1. Check whether the output type of the gradient calculation function matches the type of sens_param. If they do not match, modify them to be of the same type.",
               "2. Automatic differentiation reports an error: Type Join Failed."
           ],
        "attribute.error_case": [
            "grad = ops.GradOperation(sens_param=True)",
            "# The output type of test_net is tuple(Tensor, Tensor).",
            "def test_net(a, b):",
            "    return a, b"
              ],
        "attribute.fixed_case": [
            "grad = ops.GradOperation(sens_param=True)",
            "# The output type of test_net is tuple(Tensor, Tensor).",
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

### JSON Parameter Description

<!-- markdownlint-disable MD033 -->
<table><thead><tr><th><p>Parameter Name</p>
</th>
<th><p>Value Type</p>
</th>
<th><p>Parameter Description</p>
</th>
<th><p>Mandatory</p>
</th>
<th><p>Value Description</p>
</th>
</tr>
</thead>
<tbody><tr><td><p>attribute.class</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault category</p>
</td>
<td><p>Mandatory</p>
</td>
<td rowspan="3"><p>The value has 1 to 50 characters. It supports English letters, digits, English punctuation, and spaces.</p>

</td>
</tr>
<tr><td><p>attribute.component</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault component</p>
</td>
<td><p>Mandatory</p>
</td>
</tr>
<tr><td><p>attribute.module</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault module</p>
</td>
<td><p>Mandatory</p>
</td>
</tr>
<tr><td><p>attribute.cause_zh</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault cause (Chinese)</p>
</td>
<td><p>Mandatory</p>
</td>
<td><p>Value length ranges from 1 to 200 characters. Supports English letters, digits, English punctuation, Chinese characters, Chinese punctuation, and spaces.</p>
</td>
</tr>
<tr><td><p>attribute.cause_en</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault cause (English)</p>
</td>
<td><p>Optional</p>
</td>
<td><p>The value length ranges from 1 to 200 characters, and supports English letters, digits, English punctuation, and spaces.</p>
</td>
</tr>
<tr><td><p>attribute.description_zh</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Fault description (Chinese)</p>
</td>
<td><p>Mandatory</p>
</td>
<td rowspan="6"><div>Supports a string or a list. A string is a whole piece of information and can contain line breaks; in a list, each element is one line of information, and the elements together form the whole piece of information.<ul><li>String: the value length is 1 to 2,000 characters, supporting English letters, digits, English punctuation, Chinese characters, Chinese punctuation, spaces, and "\n".</li><li>List: the value length of each string in the list is 1 to 200 characters, supporting English letters, digits, English punctuation, Chinese characters, Chinese punctuation, and spaces.</li></ul>
</div>
</td>
</tr>
<tr><td><p>attribute.description_en</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Fault description (English)</p>
</td>
<td><p>Optional</p>
</td>
</tr>
<tr><td><p>attribute.suggestion_zh</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Suggestion (Chinese)</p>
</td>
<td><p>Required</p>
</td>
</tr>
<tr><td><p>attribute.suggestion_en</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Suggestion (English)</p>
</td>
<td><p>Optional</p>
</td>
</tr>
<tr><td><p>attribute.error_case</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Error example</p>
</td>
<td><p>Optional</p>
</td>
</tr>
<tr><td><p>attribute.fixed_case</p>
</td>
<td><p>String/List[String]</p>
</td>
<td><p>Correction example</p>
</td>
<td><p>Optional</p>
</td>
</tr>
<tr><td><p>rule</p>
</td>
<td><p>List</p>
</td>
<td><p>Fault chain, which stores all next-level fault entities triggered by this fault</p>
</td>
<td><p>Optional</p>
</td>
<td><p>The list contains the following fields.</p>
<ul><li>dst_code: Mandatory. Indicates the fault code of the next-level fault entity triggered by this fault. The fault code must be a fault code supported by ascend-fd or a user-defined fault code.</li><li>expression: Optional. Indicates the fault trigger constraint. This is currently a reserved field. The value length is 1 to 200 characters. Supports English letters, digits, English punctuation, and spaces.</li></ul>
</td>
</tr>
<tr><td><p>source_file</p>
</td>
<td><p>String</p>
</td>
<td><p>Fault log file</p>
</td>
<td><p>Mandatory</p>
</td>
<td>
<p>Name of the log file corresponding to each log file type.</p>
</td>
</tr>
<tr><td><p>regex.in</p>
</td>
<td><p>List</p>
</td>
<td><p>Fault keyword</p>
</td>
<td><p>Mandatory</p>
</td>
<td><div>Supports first-level lists and second-level lists.<ul><li>First-level list<ul><li>Each element is a string. The value length is 1 to 200 characters, supporting English letters, digits, English punctuation, Chinese characters, Chinese punctuation, and spaces.</li><li>Each keyword in the list must satisfy the existence check and conform to the order relationship.</li></ul>
</li><li>Second-level list<ul><li>Each sublist satisfies the value constraints of the first-level list.</li><li>The judgment rules within each sublist are the same as those of the first-level list. Sublists are in an OR relationship, and only the keywords of one sublist need to be satisfied.</li></ul>
</li></ul>
</div>
</td>
</tr>
<tr><td colspan="5"><ul><li>To add a custom fault entity, all required fields must exist in the JSON file and conform to the relevant value requirements.</li><li>To modify a custom fault entity, only the relevant value requirements need to be satisfied.</li></ul>
</td>
</tr>
</tbody>
</table>
<!-- markdownlint-enable MD033 -->

## Notes

- A JSON file supports up to 1,000 custom fault information entries.
- The fault code value length must be 1 to 50 characters.
- The fault code must not conflict with any fault code already supported by ascend-fd.
- When you add a fault entity, the data is saved in the `$HOME/.ascend_faultdiag/custom-ascend-kg-config.json` file.
- You can specify the custom fault entity file path by modifying the `ASCEND_FD_HOME_PATH` environment variable. For details, see [Environment Variable](../07_references/01_common_operations.md#environment-variable-description).
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
