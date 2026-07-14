# Custom Metric File<a name="ZH-CN_TOPIC_0000002501343480"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:42.004Z pushedAt=2026-06-09T02:05:50.669Z -->

## Field Description<a name="section42791696421"></a>

The fields in the custom metric file are shown in [Table 1](#ZH-CN_TOPIC_0000002501343480_table5395205714441). Custom metrics can be developed via a custom metric file. For details, see [Custom Metric Development](../../appendix.md#custom-metric-development).

**Table 1**  Fields in custom metric file

<a name="ZH-CN_TOPIC_0000002501343480_table5395205714441"></a>

|Field |Type|Description|
|--|--|--|
|version|string|Fixed value: 1.0.|
|name|string|Metric name. Cannot be empty. Length cannot exceed 128.|
|desc|string|Detailed description of the metric. Cannot be empty. Length cannot exceed 1024.|
|timestamp|timestamp|Update timestamp of the metric, in us.|
|data_list|list|Non-empty array. Length cannot exceed 128.|
|-value|float|Value of the metric.|
|-label|json|Label of the metric. Both keys and values in the JSON must be of string type. The number of child elements in the JSON cannot exceed 10.|

## Constraints<a name="section342514924413"></a>

- Symbolic links are not supported.
- Specifying a directory is not supported.
- The number of specified file paths cannot exceed 10. Multiple files are separated by commas.
- When multiple files are specified, if some of them do not exist or are empty, and the conditions are still not met after 1 minute, metric collection for the corresponding files will be canceled.
- The format of fields in the file must meet the requirements in [Table 1](#ZH-CN_TOPIC_0000002501343480_table5395205714441). Incorrect format will cancel metric collection for the corresponding file.
- When automatically obtaining metric labels, the first data entry in `data_list` shall prevail.
- The group of the custom metric file must be the same as the group of the npu-exporter process, and it must have read permission without any execute permission.
- Modifying the name, desc, and version of a file is not supported during program execution.
- If the label name is modified while the program is running, the program will not detect the change and will continue reporting using the label name from initialization.
- The size of the custom metric file is limited to 100KB.
- When deploying in a container scenario, ensure that the metric file is correctly mounted into the container.
- During metric collection, if a batch of data does not meet the constraints, that batch of data will be ignored, and the cached data will be reported instead.

## Metric File Example

```json
{
  "version": "1.0",
  "name": "hccs_bandwidth",
  "desc": "hccs bandwidth info, unit is 'MB/s'.",
  "timestamp": 1766456419845127,
  "data_list": [
    {
      "value": 190.02,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "0",
        "direction": "in",
        "path": "P0->P1"
      }
    },
    {
      "value": 143.09,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "1",
        "direction": "in",
        "path": "P2->P1"
      }
    }
  ]
}
```

## Example of Custom Metric Reporting in Telegraf

```ColdFusion
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=0,numa=2,path=P0->P1 hccs_bandwidth=190.02 1766456419845127000
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=1,numa=2,path=P2->P1 hccs_bandwidth=143.09 1766456419845127000
```
