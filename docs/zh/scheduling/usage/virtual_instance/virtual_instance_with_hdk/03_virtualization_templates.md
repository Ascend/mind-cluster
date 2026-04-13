# 虚拟化模板<a name="ZH-CN_TOPIC_0000002511346345"></a>

当前各产品型号支持的虚拟化实例模板如[表1](#zh-cn_topic_0000002038226813_table140421911260)所示。

**表 1**  虚拟化实例模板

<a name="zh-cn_topic_0000002038226813_table140421911260"></a>
<table>
    <tr>
        <td>产品型号</td>
        <td>虚拟化实例模板</td>
        <td>说明</td>
    </tr>
    <tr>
        <td>Atlas 训练系列产品（30或32个AI Core）</td>
        <td>虚拟化实例模板包括：vir02、vir04、vir08、vir16。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li></ul></td>
    </tr>
    <tr>
        <td>Atlas 推理系列产品（8个AI Core）</td>
        <td>虚拟化实例模板包括：vir01、vir02、vir04、vir02_1c、vir04_3c、vir04_3c_ndvpp、vir04_4c_dvpp。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li><li>c前面的数字表示AI CPU数量。</li><li>dvpp表示虚拟化时包含所有数字视觉预处理模块（即VPC，VDEC，JPEGD，PNGD，VENC，JPEGE）。</li><li>ndvpp表示虚拟化时没有数字视觉预处理硬件资源。</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A2 训练系列产品（20或24或25个AI Core）</td>
        <td>虚拟化实例模板包括：vir05_1c_16g、vir10_3c_32g、vir06_1c_16g、vir12_3c_32g。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li><li>c前面的数字表示AI CPU数量。</li><li>g前面的数字表示内存数量。</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A2 推理系列产品（20个AI Core）</td>
        <td>虚拟化实例模板包括：vir05_1c_8g、vir10_3c_16g_nm、vir10_4c_16g_m、vir10_3c_16g、vir10_3c_32g、vir05_1c_16g。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li><li>c前面的数字表示AI CPU数量。</li><li>m同dvpp表示虚拟化时包含所有数字视觉预处理模块（即VPC，VDEC，JPEGD，PNGD，VENC，JPEGE）。</li><li>nm同ndvpp表示虚拟化时没有数字视觉预处理硬件资源。</li><li>g前面的数字表示内存数量。</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A3 训练系列产品（48个AI Core）</td>
        <td>虚拟化实例模板包括：vir06_1c_16g、vir12_3c_32g。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li><li>c前面的数字表示AI CPU数量。</li><li>g前面的数字表示内存数量。</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A3 推理系列产品（40个AI Core）</td>
        <td>虚拟化实例模板包括：vir05_1c_16g、vir10_3c_32g。</td>
        <td><ul><li>vir后面的数字表示AI Core数量。</li><li>c前面的数字表示AI CPU数量。</li><li>g前面的数字表示内存数量。</li></ul></td>
    </tr>
    <tr>
        <td colspan="3">注：具体服务器支持的模板可通过<strong>npu-smi info -t template-info</strong>命令查询。</td>
    </tr>
</table>

>[!NOTE]  
>昇腾AI处理器包含AI Core、AI CPU、DVPP、内存等硬件资源，主要用途如下：
>
>- AI Core主要用于矩阵乘等计算，适用于卷积模型。
>- AI CPU主要负责执行CPU类算子（包括控制算子、标量和向量等通用计算）。
>- 虚拟化实例（创建指定芯片的vNPU）会使能SRIOV，将data CPU转化为AI CPU，因此会导致NPU信息中的AI CPU个数发生变化。
>- DVPP为数字视觉预处理模块，提供对特定格式的视频和图像进行解码、缩放等预处理操作，以及对处理后的视频、图像进行编码再输出的能力，包含VPC、VDEC、JPEGD、PNGD、VENC、JPEGE模块。
>    - VPC：视觉预处理核心，提供对图像进行缩放、色域转换、降bit数处理、存储格式转换、区块切割转换等能力。
>    - VDEC：视频解码器，提供对特定格式的视频进行解码的能力。
>    - JPEGD：JPEG图像解码器，提供对JPEG格式的图像进行解码的能力。
>    - PNGD：PNG图像解码器，提供对PNG格式的图像进行解码的能力。
>    - VENC：视频编码器，提供对特定格式的视频进行编码的能力。
>    - JPEGE：JPEG图像编码器，提供对图像进行编码输出为JPEG格式的能力。
