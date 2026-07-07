# (Optional) Masking Fault Logs<a name="ZH-CN_TOPIC_0000001846651017"></a>

You can add a masking rule that contains fault keywords to prevent information containing fault keywords from being recorded in the file after cleaning logs. By default, the system saves the fault keywords to be masked in the `${HOME}/.ascend_faultdiag/custom-blacklist.json` file.

>[!NOTE]
>
>- Currently, only ERROR-level logs of CANN App logs can be masked.
>- If you need to customize the path for saving the fault keywords to be masked, see [Customizing the MindCluster Ascend FaultDiag Home Directory](../common_operations.md#customizing-the-mindcluster-ascend-faultdiag-home-directory).

**Procedure<a name="section71028910168"></a>**

1. Add a masking rule. For example, you can run the following command to filter out logs containing "ERROR" and "FUSION".

    ```shell
    ascend-fd blacklist --add ERROR FUSION
    ```

    >[!NOTE]
    >- A keyword can contain a maximum of 200 characters, including uppercase letters, lowercase letters, digits, and special characters, such as hyphens (-), periods (.), and slashes (/).
    >- A masking rule supports a maximum of 10 keywords. Multiple keywords can be separated by spaces, for example, ERROR FUSION, ERROR "FUSION", or "ERROR" "FUSION".
    >- If a keyword contains backslashes (\), the entire keyword must be enclosed in quotation marks to prevent the backslashes from being ignored, for example, "ERR\\OR".
    >- A maximum of 50 masking rules can be saved. If the number of masking rules exceeds 50, the system discards the earliest masking rule and retains the newest one.

2. Import a masking rule using a JSON file.

    ```shell
    ascend-fd blacklist --file test.json
    ```

    >[!NOTE]
    >If you add a masking rule by importing a JSON file, the imported JSON file will overwrite the original JSON file in the system.

    The JSON file format is as follows:

    ```json
    {
        "blacklist":[["ERROR2","ERROR3","ERROR3"],
            ["ERR1","ERR2","ERR3","ERR4"]
        ]
    }
    ```

3. View existing masking rules.

    ```shell
    ascend-fd blacklist --show
    ```

    Command output:

    ```ColdFusion
    [BLACKLIST]
    0. 'ERROR', 'FUSION'
    1. 'ERROR2', 'ERROR2', 'ERROR3'
    2. 'ERR1', 'ERR2', 'ERR3', 'ERR4'
    ```

4. (Optional) Delete a masking rule, for example, rule 0. Multiple rules can be deleted at a time. Use spaces to separate them.

    ```shell
    ascend-fd blacklist --delete 0
    ```
