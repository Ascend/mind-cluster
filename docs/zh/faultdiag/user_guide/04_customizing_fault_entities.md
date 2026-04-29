# （可选）自定义故障实体<a name="ZH-CN_TOPIC_0000001849571601"></a>

支持用户自定义故障实体，通过新增、查询或者删除自定义故障实体，扩展MindCluster Ascend FaultDiag组件支持的故障类型。用户新增的故障保存在“$\{HOME\}/.ascend\_faultdiag/custom-ascend-kg-config.json”文件中。用户在执行日志清洗与转储和故障诊断功能时，MindCluster Ascend FaultDiag会自动在相应路径下加载用户自定义故障文件和MindCluster Ascend FaultDiag组件已经支持的故障文件。

>[!NOTE] 
>若用户需要自定义故障文件的保存路径，可以参考[自定义MindCluster Ascend FaultDiag家目录](../common_operations.md#自定义mindcluster-ascend-faultdiag家目录)章节进行操作。

**操作步骤<a name="section620272594310"></a>**

1. 通过JSON文件，新增或修改自定义故障实体。

    ```shell
    ascend-fd entity --update updated_entity.json
    ```

    回显示例如下，表示操作成功。

    ```ColdFusion
    Updated entity successfully.
    ```

    JSON文件示例如下，该示例不可直接使用，用户需根据实际情况修改自定义故障的相关信息。JSON文件最多支持1000条自定义故障信息，超出部分将不会保存到系统。文件中的参数说明请参见[表1](#table1225010132553)。

    ```json
    {
        "41001": {      #故障码，用户需根据实际情况自定义故障码，不能与MindCluster Ascend FaultDiag已支持的故障码相同
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
        },
        "41002": {                #故障码，用户需根据实际情况自定义故障码，不能与MindCluster Ascend FaultDiag已支持的故障码相同
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
    >JSON文件示例中的41001和41002为用户自定义故障码，取值长度需为1\~50个字符串，支持英文字母、数字、英文符号、下划线（\_）和中划线（-），不能与MindCluster Ascend FaultDiag已支持的故障码相同。

    **表 1**  参数说明

    <a name="table1225010132553"></a>
    <table><thead align="left"><tr id="row1125021311552"><th class="cellrowborder" valign="top" width="19.998000199980005%" id="mcps1.2.6.1.1"><p id="p16250131395516"><a name="p16250131395516"></a><a name="p16250131395516"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.989001099890011%" id="mcps1.2.6.1.2"><p id="p1625021315515"><a name="p1625021315515"></a><a name="p1625021315515"></a>取值类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="10.00899910008999%" id="mcps1.2.6.1.3"><p id="p96201150175710"><a name="p96201150175710"></a><a name="p96201150175710"></a>参数说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.999000099990003%" id="mcps1.2.6.1.4"><p id="p10250191395510"><a name="p10250191395510"></a><a name="p10250191395510"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="50.00499950005%" id="mcps1.2.6.1.5"><p id="p325031375510"><a name="p325031375510"></a><a name="p325031375510"></a>取值说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row22501213145518"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p1725071317558"><a name="p1725071317558"></a><a name="p1725071317558"></a>attribute.class</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p12456122244916"><a name="p12456122244916"></a><a name="p12456122244916"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p1620050115712"><a name="p1620050115712"></a><a name="p1620050115712"></a>故障类别</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p725061310558"><a name="p725061310558"></a><a name="p725061310558"></a>必选</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p625051385513"><a name="p625051385513"></a><a name="p625051385513"></a>取值长度为1~50个字符，支持英文字母、数字、英文符号与空格。</p>
    <p id="p11337187471"><a name="p11337187471"></a><a name="p11337187471"></a></p>
    <p id="p3251171365519"><a name="p3251171365519"></a><a name="p3251171365519"></a></p>
    </td>
    </tr>
    <tr id="row725021316558"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7250313165513"><a name="p7250313165513"></a><a name="p7250313165513"></a>attribute.component</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p4250141320556"><a name="p4250141320556"></a><a name="p4250141320556"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7620135055718"><a name="p7620135055718"></a><a name="p7620135055718"></a>故障组件</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p32501713105520"><a name="p32501713105520"></a><a name="p32501713105520"></a>必选</p>
    </td>
    </tr>
    <tr id="row14250813135517"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3251201355515"><a name="p3251201355515"></a><a name="p3251201355515"></a>attribute.module</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p525141310559"><a name="p525141310559"></a><a name="p525141310559"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p4620105065718"><a name="p4620105065718"></a><a name="p4620105065718"></a>故障模块</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p125161315555"><a name="p125161315555"></a><a name="p125161315555"></a>必选</p>
    </td>
    </tr>
    <tr id="row0251131313553"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p182511133552"><a name="p182511133552"></a><a name="p182511133552"></a>attribute.cause_zh</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p14251181385515"><a name="p14251181385515"></a><a name="p14251181385515"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p0620650105712"><a name="p0620650105712"></a><a name="p0620650105712"></a>故障原因（中文）</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p1925191313553"><a name="p1925191313553"></a><a name="p1925191313553"></a>必选</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p1925120133551"><a name="p1925120133551"></a><a name="p1925120133551"></a>取值长度为1~200个字符，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</p>
    </td>
    </tr>
    <tr id="row1181112503394"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p11811550113914"><a name="p11811550113914"></a><a name="p11811550113914"></a>attribute.cause_en</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p19811105063919"><a name="p19811105063919"></a><a name="p19811105063919"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p58111350193910"><a name="p58111350193910"></a><a name="p58111350193910"></a>故障原因（英文）</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p1781125083919"><a name="p1781125083919"></a><a name="p1781125083919"></a>可选</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p9811145093915"><a name="p9811145093915"></a><a name="p9811145093915"></a>取值长度为1~200个字符，支持英文字母、数字、英文符号与空格。</p>
    </td>
    </tr>
    <tr id="row1525121315513"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p1251813185519"><a name="p1251813185519"></a><a name="p1251813185519"></a>attribute.description_zh</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p725141385516"><a name="p725141385516"></a><a name="p725141385516"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p162015506571"><a name="p162015506571"></a><a name="p162015506571"></a>故障描述（中文）</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p18251171315559"><a name="p18251171315559"></a><a name="p18251171315559"></a>必选</p>
    </td>
    <td class="cellrowborder" rowspan="6" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><div class="p" id="p112601665108"><a name="p112601665108"></a><a name="p112601665108"></a>支持字符串或列表。字符串为整段信息，可换行；列表则每一个元素为一行信息，组合起来为整段信息。<a name="ul185401024389"></a><a name="ul185401024389"></a><ul id="ul185401024389"><li>字符串：取值长度为1~2000个字符，支持英文字母、数字、英文符号、中文汉字、中文符号、空格与“\n”。</li><li>列表：列表下每个字符串的取值长度为1~200，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</li></ul>
    </div>
    </td>
    </tr>
    <tr id="row1513114714816"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p61681719154820"><a name="p61681719154820"></a><a name="p61681719154820"></a>attribute.description_en</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p131688193481"><a name="p131688193481"></a><a name="p131688193481"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11168919134814"><a name="p11168919134814"></a><a name="p11168919134814"></a>故障描述（英文）</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p131681319174812"><a name="p131681319174812"></a><a name="p131681319174812"></a>可选</p>
    </td>
    </tr>
    <tr id="row16251101312557"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16251101311559"><a name="p16251101311559"></a><a name="p16251101311559"></a>attribute.suggestion_zh</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1825111137553"><a name="p1825111137553"></a><a name="p1825111137553"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p662075005716"><a name="p662075005716"></a><a name="p662075005716"></a>建议方案（中文）</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1925111320555"><a name="p1925111320555"></a><a name="p1925111320555"></a>必选</p>
    </td>
    </tr>
    <tr id="row1036382211475"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1136318222471"><a name="p1136318222471"></a><a name="p1136318222471"></a>attribute.suggestion_en</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p193633224477"><a name="p193633224477"></a><a name="p193633224477"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p036320221476"><a name="p036320221476"></a><a name="p036320221476"></a>建议方案（英文）</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p153633227471"><a name="p153633227471"></a><a name="p153633227471"></a>可选</p>
    </td>
    </tr>
    <tr id="row1125121316556"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1225117133557"><a name="p1225117133557"></a><a name="p1225117133557"></a>attribute.error_case</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p038102551"><a name="p038102551"></a><a name="p038102551"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7620185017574"><a name="p7620185017574"></a><a name="p7620185017574"></a>错误示例</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p32513132555"><a name="p32513132555"></a><a name="p32513132555"></a>可选</p>
    </td>
    </tr>
    <tr id="row1564416529419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p176453528410"><a name="p176453528410"></a><a name="p176453528410"></a>attribute.fixed_case</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p138112216511"><a name="p138112216511"></a><a name="p138112216511"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p564510525419"><a name="p564510525419"></a><a name="p564510525419"></a>修正示例</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1264517528410"><a name="p1264517528410"></a><a name="p1264517528410"></a>可选</p>
    </td>
    </tr>
    <tr id="row739419379132"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p53951937161317"><a name="p53951937161317"></a><a name="p53951937161317"></a>rule</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p1439573711137"><a name="p1439573711137"></a><a name="p1439573711137"></a>列表</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p8395173701315"><a name="p8395173701315"></a><a name="p8395173701315"></a>故障链，存储该故障所有触发的下一级故障实体</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p17395123716136"><a name="p17395123716136"></a><a name="p17395123716136"></a>可选</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p163951037151315"><a name="p163951037151315"></a><a name="p163951037151315"></a>列表内包含以下字段。</p>
    <a name="ul10875010112117"></a><a name="ul10875010112117"></a><ul id="ul10875010112117"><li>dst_code：必选，表示本次故障触发的下一级故障实体故障码，该故障码必须为<span id="ph1645992316225"><a name="ph1645992316225"></a><a name="ph1645992316225"></a>MindCluster Ascend FaultDiag</span>已支持的故障码或用户自定义故障码。</li><li>expression：可选，表示故障触发约束，当前为预留字段。取值长度为1~200个字符，支持英文字母、数字、英文符号与空格。</li></ul>
    </td>
    </tr>
    <tr id="row12754174517195"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p20755154511193"><a name="p20755154511193"></a><a name="p20755154511193"></a>source_file</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p18755134521919"><a name="p18755134521919"></a><a name="p18755134521919"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p17755104515198"><a name="p17755104515198"></a><a name="p17755104515198"></a>故障日志文件</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p175524551912"><a name="p175524551912"></a><a name="p175524551912"></a>必选</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><p id="p171364919208"><a name="p171364919208"></a><a name="p171364919208"></a>各个日志文件类型对应的日志文件名称。</p>
    <p id="p1711214428203"><a name="p1711214428203"></a><a name="p1711214428203"></a>可自定义配置文件类型，也可使用默认支持的文件类型。支持配置多个，以“|”分割（例如："TrainLog|CANN_Plog"），最多配置10个，每个字符串取值长度为1~50个字符，支持英文字母、数字、英文符号与空格。</p>
    <p id="p18381715259"><a name="p18381715259"></a><a name="p18381715259"></a>默认支持的日志文件类型如下（说明：文件名称对应存储目录请参考<a href="./03_collecting_logs.md#日志采集目录结构">表1 日志文件列表</a>）。</p>
    <a name="ul111463916256"></a><a name="ul111463916256"></a><ul id="ul111463916256"><li>TrainLog：训练及推理控制台日志。</li><li>CANN_Plog：Host侧应用类日志。</li><li>CANN_Device：Device侧应用类日志。</li><li>NPU_OS：Device侧Control CPU上的系统类日志和Device侧Control CPU上的EVENT级别系统日志。</li><li>NPU_Device：Device侧非Control CPU上的系统类日志。</li><li>NPU_History：黑匣子日志、NPU芯片内核日志、Device侧OS基本信息和Device侧片上内存日志。</li><li>OS：主机侧操作系统日志文件。</li><li>OS-dmesg：主机侧内核消息类文件。</li><li>OS-vmcore-dmesg：系统崩溃时保存的Host侧内核消息日志文件。</li><li>OS-sysmon：主机侧系统监测类文件。</li><li>NodeDLog：AI服务器日志。</li><li>DL_DevicePlugin：超节点设备日志、<span id="ph1201047181913"><a name="ph1201047181913"></a><a name="ph1201047181913"></a>Ascend Device Plugin</span>组件日志。</li><li>DL_Volcano_Scheduler：<span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span>中的volcano-scheduler组件日志。</li><li>DL_Volcano_Controller：<span id="ph216472421712"><a name="ph216472421712"></a><a name="ph216472421712"></a>Volcano</span>中的volcano-controller组件日志。</li><li>DL_Docker_Runtime：<span id="ph622819010286"><a name="ph622819010286"></a><a name="ph622819010286"></a>Ascend Docker Runtime</span>组件日志。</li><li>DL_Npu_Exporter：<span id="ph14925450192719"><a name="ph14925450192719"></a><a name="ph14925450192719"></a>NPU Exporter</span>组件日志。</li><li>MindIE：<span id="ph8749745104719"><a name="ph8749745104719"></a><a name="ph8749745104719"></a>MindIE</span>组件日志。</li><li>CANN_Amct：<span id="ph1416210515117"><a name="ph1416210515117"></a><a name="ph1416210515117"></a>AMCT</span>组件日志。</li></ul>
    </td>
    </tr>
    <tr id="row12399114716191"><td class="cellrowborder" valign="top" width="19.998000199980005%" headers="mcps1.2.6.1.1 "><p id="p11399124721917"><a name="p11399124721917"></a><a name="p11399124721917"></a>regex.in</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.989001099890011%" headers="mcps1.2.6.1.2 "><p id="p1439919474199"><a name="p1439919474199"></a><a name="p1439919474199"></a>String</p>
    </td>
    <td class="cellrowborder" valign="top" width="10.00899910008999%" headers="mcps1.2.6.1.3 "><p id="p203991347121917"><a name="p203991347121917"></a><a name="p203991347121917"></a>故障关键词</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.999000099990003%" headers="mcps1.2.6.1.4 "><p id="p18399194712194"><a name="p18399194712194"></a><a name="p18399194712194"></a>必选</p>
    </td>
    <td class="cellrowborder" valign="top" width="50.00499950005%" headers="mcps1.2.6.1.5 "><div class="p" id="p8603142020279"><a name="p8603142020279"></a><a name="p8603142020279"></a>支持一级列表与二级列表。<a name="ul7951196182711"></a><a name="ul7951196182711"></a><ul id="ul7951196182711"><li>一级列表<a name="ul186081634154310"></a><a name="ul186081634154310"></a><ul id="ul186081634154310"><li>每个元素为字符串。取值长度为1~200个字符，支持英文字母、数字、英文符号、中文汉字、中文符号与空格。</li><li>列表中每个关键词都需要满足存在性判断，且符合前后关系</li></ul>
    </li><li>二级列表<a name="ul416995414319"></a><a name="ul416995414319"></a><ul id="ul416995414319"><li>每个子列表满足一级列表的取值约束。</li><li>每个子列表内的判断规则同一级列表，每个子列表间为或关系，仅需满足一个子列表的关键词即可。</li></ul>
    </li></ul>
    </div>
    </td>
    </tr>
    <tr id="row4132132882810"><td class="cellrowborder" colspan="5" valign="top" headers="mcps1.2.6.1.1 mcps1.2.6.1.2 mcps1.2.6.1.3 mcps1.2.6.1.4 mcps1.2.6.1.5 "><a name="ul8287181216559"></a><a name="ul8287181216559"></a><ul id="ul8287181216559"><li>新增自定义故障实体时，所有必选字段都需要存在JSON文件中，且符合相关取值要求。</li><li>修改自定义故障实体时，只需要符合相关取值要求即可。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

2. 查看用户自定义的故障实体信息。支持通过故障码查看故障信息，不指定故障码时将查询所有自定义故障实体信息。

    ```shell
    ascend-fd entity --show entity_code_1 entity_code_2
    ```

3. （可选）删除指定对应故障码的自定义故障实体信息。

    ```shell
    ascend-fd entity --delete entity_code_1 entity_code_2
    ```

4. （可选）校验custom-ascend-kg-config.json文件。若用户直接修改custom-ascend-kg-config.json文件的相关自定义故障实体信息，可以执行以下命令，校验修改后文件的完整性和可用性。

    >[!NOTE] 
    >不建议用户直接更改custom-ascend-kg-config.json文件信息，可能造成MindCluster Ascend FaultDiag组件功能异常。

    ```shell
    ascend-fd entity --check custom-ascend-kg-config.json
    ```

    回显示例如下，表示文件校验通过。

    ```ColdFusion
    Custom entity verification passed.
    ```
