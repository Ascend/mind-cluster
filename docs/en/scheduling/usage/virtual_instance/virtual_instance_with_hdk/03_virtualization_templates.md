# Virtualization Templates<a name="ZH-CN_TOPIC_0000002511346345"></a>

[Table 1](#zh-cn_topic_0000002038226813_table140421911260) shows virtual instance templates currently supported by each product.

**Table 1** Virtual instance templates

<a name="zh-cn_topic_0000002038226813_table140421911260"></a>
<table>
    <tr>
        <td>Product Model</td>
        <td>Virtual Instance Template</td>
        <td>Description</td>
    </tr>
    <tr>
        <td>Atlas training series products (30 or 32 AICores)</td>
        <td>Virtual instance templates include: vir02, vir04, vir08, vir16.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li></ul></td>
    </tr>
    <tr>
        <td>Atlas inference series products (8 AICores)</td>
        <td>Virtual instance templates include: vir01, vir02, vir04, vir02_1c, vir04_3c, vir04_3c_ndvpp, vir04_4c_dvpp.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li><li>The number before c indicates the number of AICPUs.</li><li>dvpp indicates that all digital visual preprocessing modules (i.e., VPC, VDEC, JPEGD, PNGD, VENC, JPEGE) are included during virtualization.</li><li>ndvpp indicates that there are no digital visual preprocessing hardware resources during virtualization.</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A2 training series products (24 AICores)</td>
        <td>Virtual instance templates include: vir06_1c_16g, vir12_3c_32g.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li><li>The number before c indicates the number of AICPUs.</li><li>The number before g indicates the memory size.</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A2 inference series products (20 AICores)</td>
        <td>Virtual instance templates include: vir05_1c_8g, vir10_3c_16g_nm, vir10_4c_16g_m, vir10_3c_16g, vir10_3c_32g, vir05_1c_16g.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li><li>The number before c indicates the number of AICPUs.</li><li>m, same as dvpp, indicates that all digital visual preprocessing modules (i.e., VPC, VDEC, JPEGD, PNGD, VENC, JPEGE) are included during virtualization.</li><li>nm, same as ndvpp, indicates that there are no digital visual preprocessing hardware resources during virtualization.</li><li>The number before g indicates the memory size.</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A3 training series products (48 AICores)</td>
        <td>Virtual instance templates include: vir06_1c_16g, vir12_3c_32g.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li><li>The number before c indicates the number of AICPUs.</li><li>The number before g indicates the memory size.</li></ul></td>
    </tr>
    <tr>
        <td>Atlas A3 inference series products (40 AICores)</td>
        <td>Virtual instance templates include: vir05_1c_16g, vir10_3c_32g.</td>
        <td><ul><li>The number after vir indicates the number of AICores.</li><li>The number before c indicates the number of AICPUs.</li><li>The number before g indicates the memory size.</li></ul></td>
    </tr>
    <tr>
        <td colspan="3">NOTE: The templates supported by a specific server can be queried using the <strong>npu-smi info -t template-info</strong> command.</td>
    </tr>
</table>

> [!NOTE]
> The Ascend AI processor includes hardware resources such as AICore, AICPU, DVPP, and memory. Their main purposes are as follows:
>
> - `AICore` is mainly used for matrix multiplication and other computations, and is suitable for convolutional models.
> - `AICPU` is primarily responsible for executing CPU-type operators (including control operators, scalars, vectors, and other general-purpose computations).
> - Virtual instances (creating vNPUs for specified chips) enable SRIOV, converting data CPUs into AICPUs. As a result, the number of AICPUs displayed in the NPU information changes.
> - `DVPP` (Digital Vision Pre‑Processing) is a module that provides pre‑processing capabilities for video and image data in specific formats, including decoding, scaling, and encoding of processed video and images. It comprises the following modules.
>     - `VPC` (Vision Pre‑processing Core): Provides capabilities such as image scaling, color space conversion, bit‑depth reduction, format conversion, and block‑based cropping/transformation.
>     - `VDEC` (Video Decoder): Provides decoding capabilities for video in specific formats
>     - `JPEGD` (JPEG Decoder): Provides decoding capabilities for images in JPEG format.
>     - `PNGD` (PNG Decoder): Provides decoding capabilities for images in PNG format.
>     - `VENC` (Video Encoder): Provides encoding capabilities for video in specific formats.
>     - `JPEGE` (JPEG Encoder): Provides the ability to encode images and output them in JPEG format.
