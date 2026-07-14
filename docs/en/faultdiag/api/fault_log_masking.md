# Fault Log Masking Interface<a name="ZH-CN_TOPIC_0000001799772104"></a>

## Prototype<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>

```shell
ascend-fd blacklist <Subcommand>
```

## Function<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>

Adds a masking rule containing fault keywords, so that information containing fault keywords will not be recorded in the file after log cleaning.

>[!NOTE]
>
>- Currently, only ERROR logs of CANN App logs can be masked.
>- If you need to customize the save path of the fault keyword masking file, see [Customizing the MindCluster Ascend FaultDiag Home Directory](../common_operations.md#customizing-the-mindcluster-ascend-faultdiag-home-directory).

## Parameters<a name="zh-cn_topic_0000001511538701_section122149111390"></a>

**Table 1**  Subcommand parameters

|Parameter|Abbreviation|Required|Value Type|Description|
|--|--|--|--|--|
|`--add`|`-a`|Required. The `--add`, `--file`, `--delete`, and `--show` parameters are mutually exclusive, meaning only one can and must be specified.|String|Adds a masking rule that contains keywords.<ul><li>The maximum length of a keyword is 200 characters, supporting uppercase and lowercase letters, digits, and other characters (such as -./).</li><li>A single masking rule supports a maximum of 10 keywords, and multiple keywords can be separated by spaces, for example, `ERROR FUSION`, `ERROR "FUSION"`, or `"ERROR" "FUSION"`.</li><li>If a keyword contains a backslash (\), the entire keyword must be enclosed in quotation marks to prevent the backslash from being ignored, for example, `ERR\OR`.</li><li>A maximum of 50 masking rules can be saved. If the number exceeds 50, the system discards the earliest masking rules and retains the newly added ones.</li></ul>|
|`--file`|`-f`|Required. The `--add`, `--file`, `--delete`, and `--show` parameters are mutually exclusive, meaning only one can and must be specified.|String|Imports a masking rule using a JSON file. The imported JSON file overwrites the content of the original JSON file in the system.|
|`--delete`|`-d`|Required. The `--add`, `--file`, `--delete`, and `--show` parameters are mutually exclusive, meaning only one can and must be specified.|Int|Deletes masking rules. Multiple rules can be deleted simultaneously by separating them with spaces.|
|`--show`|`-s`|Required. The `--add`, `--file`, `--delete`, and `--show` parameters are mutually exclusive, meaning only one can and must be specified.|String|Displays the currently existing masking rules.|
|`--force`|None|Optional|Bool|When this parameter is specified for deleting rules or replacing files, no confirmation prompt will appear on the interface.<p>Must be used together with `--delete` or `--file`.</p>|
|`--help`|`-h`|Optional|-|Queries the usage instructions.|

## Returns<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>

Example: Return the currently existing masking rules.

```ColdFusion
[BLACKLIST]
0. ERROR2, ERROR3, ERROR3
1. ERR1, ERR2, ERR3, ERR4
```
