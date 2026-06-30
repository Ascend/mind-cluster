# (Optional) Customizing Fault Entities<a name="ZH-CN_TOPIC_0000001849571601"></a>

You can customize fault entities to extend the fault types supported by MindCluster Ascend FaultDiag by adding, querying, or deleting fault entities. User-added faults are saved in the `${HOME}/.ascend_faultdiag/custom-ascend-kg-config.json` file. When you perform log cleaning and dumping and fault diagnosis, MindCluster Ascend FaultDiag automatically loads the user-defined fault files and the fault files already supported by MindCluster Ascend FaultDiag from the corresponding paths.

>[!NOTE]
>If you need to customize the save path of fault files, see the [Customizing the MindCluster Ascend FaultDiag Home Directory](../common_operations.md#customizing-the-mindcluster-ascend-faultdiag-home-directory) section for operations.

**Procedure<a name="section620272594310"></a>**

1. Add or modify a custom fault entity through a JSON file.

    ```shell
    ascend-fd entity --update updated_entity.json
    ```

    The following is an example of the command output, indicating that the operation is successful.

    ```ColdFusion
    Updated entity successfully.
    ```

    The following is an example of a JSON file. This example cannot be used directly. You need to modify the information about the custom fault based on the actual situation. A JSON file can contain a maximum of 1,000 custom fault information entries. Any excess entries will not be saved to the system. For details about the parameters in the file, see [Table 1](#table1225010132553).

    ```json
    {
        "41001": {      #Fault code. You need to customize the fault code based on the actual situation. It cannot be the same as the fault codes already supported by MindCluster Ascend FaultDiag.
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
                "# The output type of test_net is tuple(Tensor, Tensor)",
                "def test_net(a, b):",
                "    return a, b"
                  ],
            "attribute.fixed_case": [
                "grad = ops.GradOperation(sens_param=True)",
                "# The output type of test_net is tuple(Tensor, Tensor)",
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
        },
        "41002": {                #Fault code. Users need to customize the fault code based on actual conditions. It must not be the same as the fault codes already supported by MindCluster Ascend FaultDiag.
            "attribute.class": "",
            "attribute.component": "",
            "attribute.module": "",
            "attribute.cause_zh": "",
            "attribute.description_zh": "",
            "attribute.suggestion_zh": "",
            "attribute.cause_en": "",
            "attribute.description_en": "",
            "attribute.suggestion_en": "",
            "attribute.error_case": "",
            "attribute.fixed_case": "",
            "rule": [
                {
                    "dst_code": "20107"
                }
            ],
            "source_file": "CANN_Plog",
            "regex.in": [
                    "tsd client wait response fail"
                ]
        }
    ...
    }
    ```

    >[!NOTE]
    >In the JSON file example, `41001` and `41002` are user-defined fault codes. The value length must be 1 to 50 characters. English letters, digits, English symbols, underscores (_), and hyphens (-) are supported. The fault code cannot be the same as the fault codes already supported by MindCluster Ascend FaultDiag.

    **Table 1**  Parameter description

    <a name="table1225010132553"></a>
    <table><thead align="left"><tr id="row1125021311552"><th class="cellrowborder" valign="top" width="19.998000199980005%" id="mcps1.2.6.1.1"><p id="p16250131395516"><a name="p16250131395516"></a><a name="p16250131395516"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.989001099890011%" id="mcps1.2.6.1.2"><p id="p1625021315515"><a name="p1625021315515"></a><a name="p1625021315515"></a>Value Type</p>
    </th>
    <th class="cellrowborder" valign="top" width="10.00899910008999%" id="mcps1.2.6.1.3"><p id="p96201150175710"><a name="p96201150175710"></a><a name="p96201150175710"></a>Parameter Description</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.999000099990003%" id="mcps1.2.6.1.4"><p id="p10250191395510"><a name="p10250191395510"></a><a name="p10250191395510"></a>Mandatory</p>
    </th>
    <th class="cellrowborder" valign="top" width="50.00499950005%" id="mcps1.2.6.1.5"><p id="p325031375510"><a name="p325031375510"></a><a name="p325031375510"></a>Value Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row22501213145518"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p1725071317558"><a name="p1725071317558"></a><a name="p1725071317558"></a>attribute.class</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p12456122244916"><a name="p12456122244916"></a><a name="p12456122244916"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p1620050115712"><a name="p1620050115712"></a><a name="p1620050115712"></a>Fault category</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p725061310558"><a name="p725061310558"></a><a name="p725061310558"></a>Mandatory</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p625051385513"><a name="p625051385513"></a><a name="p625051385513"></a>The value length ranges from 1 to 50 characters. English letters, digits, English symbols, and spaces are supported.</p>
    <p id="p11337187471"><a name="p11337187471"></a><a name="p11337187471"></a></p>
    <p id="p3251171365519"><a name="p3251171365519"></a><a name="p3251171365519"></a></p>
    </td>
    </tr>
    <tr id="row725021316558"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7250313165513"><a name="p7250313165513"></a><a name="p7250313165513"></a>attribute.component</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p4250141320556"><a name="p4250141320556"></a><a name="p4250141320556"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7620135055718"><a name="p7620135055718"></a><a name="p7620135055718"></a>Fault component</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p32501713105520"><a name="p32501713105520"></a><a name="p32501713105520"></a>Mandatory</p>
    </td>
    </tr>
    <tr id="row14250813135517"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3251201355515"><a name="p3251201355515"></a><a name="p3251201355515"></a>attribute.module</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p525141310559"><a name="p525141310559"></a><a name="p525141310559"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p4620105065718"><a name="p4620105065718"></a><a name="p4620105065718"></a>Fault module</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p125161315555"><a name="p125161315555"></a><a name="p125161315555"></a>Mandatory</p>
    </td>
    </tr>
    <tr id="row0251131313553"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p182511133552"><a name="p182511133552"></a><a name="p182511133552"></a>attribute.cause_zh</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p14251181385515"><a name="p14251181385515"></a><a name="p14251181385515"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p0620650105712"><a name="p0620650105712"></a><a name="p0620650105712"></a>Fault cause (Chinese)</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p1925191313553"><a name="p1925191313553"></a><a name="p1925191313553"></a>Mandatory</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p1925120133551"><a name="p1925120133551"></a><a name="p1925120133551"></a>The value length ranges from 1 to 200 characters. English letters, digits, English symbols, Chinese characters, Chinese symbols, and spaces are supported.</p>
    </td>
    </tr>
    <tr id="row1181112503394"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p11811550113914"><a name="p11811550113914"></a><a name="p11811550113914"></a>attribute.cause_en</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p19811105063919"><a name="p19811105063919"></a><a name="p19811105063919"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p58111350193910"><a name="p58111350193910"></a><a name="p58111350193910"></a>Fault cause (English)</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p1781125083919"><a name="p1781125083919"></a><a name="p1781125083919"></a>Optional</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p9811145093915"><a name="p9811145093915"></a><a name="p9811145093915"></a>The value length ranges from 1 to 200 characters. English letters, digits, English symbols, and spaces are supported.</p>
    </td>
    </tr>
    <tr id="row1525121315513"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p1251813185519"><a name="p1251813185519"></a><a name="p1251813185519"></a>attribute.description_zh</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p725141385516"><a name="p725141385516"></a><a name="p725141385516"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p162015506571"><a name="p162015506571"></a><a name="p162015506571"></a>Fault description (Chinese)</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p18251171315559"><a name="p18251171315559"></a><a name="p18251171315559"></a>Mandatory</p>
    </td>
    <td class="cellrowborder" rowspan="6" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><div class="p" id="p112601665108"><a name="p112601665108"></a><a name="p112601665108"></a>Supports strings or lists. A string represents the entire message and can contain line breaks. A list represents the entire message with each element as one line of information.<a name="ul185401024389"></a><a name="ul185401024389"></a><ul id="ul185401024389"><li>String: The value length ranges from 1 to 2,000 characters. English letters, digits, English symbols, Chinese characters, Chinese symbols, spaces, and "\n" are supported.</li><li>List: The value length of each string in the list ranges from 1 to 200 characters. English letters, digits, English symbols, Chinese characters, Chinese symbols, and spaces are supported.</li></ul>
    </div>
    </td>
    </tr>
    <tr id="row1513114714816"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p61681719154820"><a name="p61681719154820"></a><a name="p61681719154820"></a>attribute.description_en</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p131688193481"><a name="p131688193481"></a><a name="p131688193481"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11168919134814"><a name="p11168919134814"></a><a name="p11168919134814"></a>Fault description (English)</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p131681319174812"><a name="p131681319174812"></a><a name="p131681319174812"></a>Optional</p>
    </td>
    </tr>
    <tr id="row16251101312557"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16251101311559"><a name="p16251101311559"></a><a name="p16251101311559"></a>attribute.suggestion_zh</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1825111137553"><a name="p1825111137553"></a><a name="p1825111137553"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p662075005716"><a name="p662075005716"></a><a name="p662075005716"></a>Suggestion (Chinese)</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1925111320555"><a name="p1925111320555"></a><a name="p1925111320555"></a>Mandatory</p>
    </td>
    </tr>
    <tr id="row1036382211475"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1136318222471"><a name="p1136318222471"></a><a name="p1136318222471"></a>attribute.suggestion_en</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p193633224477"><a name="p193633224477"></a><a name="p193633224477"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p036320221476"><a name="p036320221476"></a><a name="p036320221476"></a>Suggestion (English)</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p153633227471"><a name="p153633227471"></a><a name="p153633227471"></a>Optional</p>
    </td>
    </tr>
    <tr id="row1125121316556"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1225117133557"><a name="p1225117133557"></a><a name="p1225117133557"></a>attribute.error_case</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p038102551"><a name="p038102551"></a><a name="p038102551"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7620185017574"><a name="p7620185017574"></a><a name="p7620185017574"></a>Error example</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p32513132555"><a name="p32513132555"></a><a name="p32513132555"></a>Optional</p>
    </td>
    </tr>
    <tr id="row1564416529419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p176453528410"><a name="p176453528410"></a><a name="p176453528410"></a>attribute.fixed_case</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p138112216511"><a name="p138112216511"></a><a name="p138112216511"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p564510525419"><a name="p564510525419"></a><a name="p564510525419"></a>Fixed example</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1264517528410"><a name="p1264517528410"></a><a name="p1264517528410"></a>Optional</p>
    </td>
    </tr>
    <tr id="row739419379132"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p53951937161317"><a name="p53951937161317"></a><a name="p53951937161317"></a>rule</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p1439573711137"><a name="p1439573711137"></a><a name="p1439573711137"></a>List</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p8395173701315"><a name="p8395173701315"></a><a name="p8395173701315"></a>Fault chain, storing all next-level fault entities triggered by this fault</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p17395123716136"><a name="p17395123716136"></a><a name="p17395123716136"></a>Optional</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p163951037151315"><a name="p163951037151315"></a><a name="p163951037151315"></a>The list contains the following fields.</p>
    <a name="ul10875010112117"></a><a name="ul10875010112117"></a><ul id="ul10875010112117"><li>dst_code: Mandatory, indicating the fault code of the next-level fault entity triggered by this fault. This fault code must be a fault code supported by <span id="ph1645992316225"><a name="ph1645992316225"></a><a name="ph1645992316225"></a>MindCluster Ascend FaultDiag</span> or a user-defined fault code.</li><li>expression: Optional, indicating the fault trigger constraint. This is a reserved field. The value length ranges from 1 to 200 characters. English letters, digits, English symbols, and spaces are supported.</li></ul>
    </td>
    </tr>
    <tr id="row12754174517195"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p20755154511193"><a name="p20755154511193"></a><a name="p20755154511193"></a>source_file</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p18755134521919"><a name="p18755134521919"></a><a name="p18755134521919"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p17755104515198"><a name="p17755104515198"></a><a name="p17755104515198"></a>Fault log file</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p175524551912"><a name="p175524551912"></a><a name="p175524551912"></a>Mandatory</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p171364919208"><a name="p171364919208"></a><a name="p171364919208"></a>Log file name corresponding to each log file type.</p>
    <p id="p1711214428203"><a name="p1711214428203"></a><a name="p1711214428203"></a>You can customize the file type or use the default supported file types. Multiple configurations are supported, separated by "|" (for example, "TrainLog|CANN_Plog"). A maximum of 10 can be configured. The value length of each string ranges from 1 to 50 characters. English letters, digits, English symbols, and spaces are supported.</p>
    <p id="p18381715259"><a name="p18381715259"></a><a name="p18381715259"></a>The default supported log file types are as follows (for the storage directory corresponding to the file name, see <a href="./03_collecting_logs.md#log-collection-directory-structure">Table 1 Log file list</a>).</p>
    <a name="ul111463916256"></a><a name="ul111463916256"></a><ul id="ul111463916256"><li>TrainLog: Training and inference console log.</li><li>CANN_Plog: Host-side application log.</li><li>CANN_Device: Device-side application log.</li><li>NPU_OS: System log on the Device-side Control CPU and EVENT-level system log on the Device-side Control CPU.</li><li>NPU_Device: System log on the Device-side non-Control CPU.</li><li>NPU_History: Black box log, NPU chip kernel log, Device-side OS basic information, and Device-side on-chip memory log.</li><li>OS: Host-side operating system log file.</li><li>OS-dmesg: Host-side kernel message file.</li><li>OS-vmcore-dmesg: Host-side kernel message log file saved during a system crash.</li><li>OS-sysmon: Host-side system monitoring file.</li><li>NodeDLog: AI server log.</li><li>DL_DevicePlugin: Super node device log, <span id="ph1201047181913"><a name="ph1201047181913"></a><a name="ph1201047181913"></a>Ascend Device Plugin</span> component log.</li><li>DL_Volcano_Scheduler: <span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span> volcano-scheduler component log.</li><li>DL_Volcano_Controller: <span id="ph216472421712"><a name="ph216472421712"></a><a name="ph216472421712"></a>Volcano</span> volcano-controller component log.</li><li>DL_Docker_Runtime: <span id="ph622819010286"><a name="ph622819010286"></a><a name="ph622819010286"></a>Ascend Docker Runtime</span> component log.</li><li>DL_Npu_Exporter: <span id="ph14925450192719"><a name="ph14925450192719"></a><a name="ph14925450192719"></a>NPU Exporter</span> component log.</li><li>MindIE: <span id="ph8749745104719"><a name="ph8749745104719"></a><a name="ph8749745104719"></a>MindIE</span> component log.</li><li>CANN_Amct: <span id="ph1416210515117"><a name="ph1416210515117"></a><a name="ph1416210515117"></a>AMCT</span> component log.</li></ul>
    </td>
    </tr>
    <tr id="row12399114716191"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p11399124721917"><a name="p11399124721917"></a><a name="p11399124721917"></a>regex.in</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p1439919474199"><a name="p1439919474199"></a><a name="p1439919474199"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p203991347121917"><a name="p203991347121917"></a><a name="p203991347121917"></a>Fault keyword</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p18399194712194"><a name="p18399194712194"></a><a name="p18399194712194"></a>Mandatory</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><div class="p" id="p8603142020279"><a name="p8603142020279"></a><a name="p8603142020279"></a>Supports first-level lists and second-level lists.<a name="ul7951196182711"></a><a name="ul7951196182711"></a><ul id="ul7951196182711"><li>First-level list<a name="ul186081634154310"></a><a name="ul186081634154310"></a><ul id="ul186081634154310"><li>Each element is a string. The value length ranges from 1 to 200 characters. English letters, digits, English symbols, Chinese characters, Chinese symbols, and spaces are supported.</li><li>Each keyword in the list must satisfy the existence check and conform to the sequential relationship.</li></ul>
    </li><li>Second-level list<a name="ul416995414319"></a><a name="ul416995414319"></a><ul id="ul416995414319"><li>Each sub-list satisfies the value constraints of the first-level list.</li><li>The judgment rule within each sub-list is the same as that of the first-level list. The relationship between sub-lists is OR, meaning that only the keywords of one sub-list need to be satisfied.</li></ul>
    </li></ul>
    </div>
    </td>
    </tr>
    <tr id="row4132132882810"><td class="cellrowborder" colspan="5" valign="top" headers="mcps1.2.6.1.1 mcps1.2.6.1.2 mcps1.2.6.1.3 mcps1.2.6.1.4 mcps1.2.6.1.5 "><a name="ul8287181216559"></a><a name="ul8287181216559"></a><ul id="ul8287181216559"><li>When adding a custom fault entity, all mandatory fields must exist in the JSON file and comply with the relevant value requirements.</li><li>When modifying a custom fault entity, only the relevant value requirements need to be met.</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

2. View user-defined custom fault entity information. You can query fault information by fault code. If no fault code is specified, all custom fault entity information will be queried.

    ```shell
    ascend-fd entity --show entity_code_1 entity_code_2
    ```

3. (Optional) Delete the custom fault entity information corresponding to the specified fault code.

    ```shell
    ascend-fd entity --delete entity_code_1 entity_code_2
    ```

4. (Optional) Verify the `custom-ascend-kg-config.json` file. If you have directly modified the custom fault entity information in this file, you can run the following command to verify the integrity and availability of the modified file.

    >[!NOTE]
    >Directly modifying the `custom-ascend-kg-config.json` file is not recommended, as it may cause the MindCluster Ascend FaultDiag component to malfunction.

    ```shell
    ascend-fd entity --check custom-ascend-kg-config.json
    ```

    The following is an example of the output, indicating that the file verification is successful.

    ```ColdFusion
    Custom entity verification passed.
    ```
