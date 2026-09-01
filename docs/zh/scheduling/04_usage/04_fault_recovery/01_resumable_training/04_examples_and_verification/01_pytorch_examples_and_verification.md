# PyTorch场景示例

## 制作镜像<a name="ZH-CN_TOPIC_0000002511426469"></a>

[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/26.1.0)作为昇腾大模型训练框架，旨在为昇腾芯片提供端到端的大语言模型训练方案，包含分布式预训练、分布式指令微调、分布式偏好对齐以及对应的开发工具链。[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)包括了仓库拉取、环境搭建与大模型训练等章节，制作MindSpeed-LLM训练框架镜像可以结合本章节和[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)。

断点续训可以基于基础训练镜像制作，基础训练镜像的制作可参考[使用Dockerfile构建容器镜像（PyTorch）](../../../../07_references/02_common_operations.md#使用dockerfile构建容器镜像pytorch)章节进行操作。

本章节结合基础训练镜像的制作步骤，展示基于Ubuntu 20.04来构建训练镜像。

>[!NOTE]
>以下示例使用MindSpeed-LLM  26.1.0版本。

**准备软件包<a name="zh-cn_topic_0000002039339945_section18254161612586"></a>**

请按照[表1](#zh-cn_topic_0000002039339945_table1172542119019)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。软件包名称中{version}表示版本号、{arch}表示架构、{chip_type}表示芯片类型。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002039339945_table1172542119019"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339945_row157251121508"><th class="cellrowborder" valign="top" width="24.55%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002039339945_p1441653254"><a name="zh-cn_topic_0000002039339945_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="25.45%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002039339945_p2052053751"><a name="zh-cn_topic_0000002039339945_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002039339945_p657531455"><a name="zh-cn_topic_0000002039339945_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002039339945_p1859531759"><a name="zh-cn_topic_0000002039339945_p1859531759"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339945_row16726192116014"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1754534515"><a name="zh-cn_topic_0000002039339945_p1754534515"></a>taskd-<em id="i2511021165615"><a name="i2511021165615"></a>{version}</em>-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p4321152352612"><a name="p4321152352612"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p205155310512"><a name="zh-cn_topic_0000002039339945_p205155310512"></a>集群调度组件断点续训whl包。</p>
<div class="note" id="zh-cn_topic_0000002039339945_note494818501423"><a name="zh-cn_topic_0000002039339945_note494818501423"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p159489506423"><a name="zh-cn_topic_0000002039339945_p159489506423"></a>安装<span id="ph1670711477256"><a name="ph1670711477256"></a>TaskD</span>组件前需确保<span id="zh-cn_topic_0000002039339945_ph998914174412"><a name="zh-cn_topic_0000002039339945_ph998914174412"></a>PyTorch</span>框架已正确安装，当前支持的<span id="zh-cn_topic_0000002039339945_ph2908133144419"><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a>PyTorch</span>版本为：2.1.0、2.3.0、2.4.0、2.5.0、2.6.0、2.7.1。TaskD运行依赖PyTorch框架，请选择无已知安全漏洞的PyTorch版本或从官方社区获取已修复安全问题的对应版本。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p19595310517"><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1386820525510"><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p12512532515"><a name="zh-cn_topic_0000002039339945_p12512532515"></a>用户通过获取链接得到的是<span id="ph480901420289"><a name="ph480901420289"></a>TaskD</span>压缩包Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a>{arch}</em>.zip，需要通过解压后，获得相应的whl软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row572619211108"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1863531756"><a name="zh-cn_topic_0000002039339945_p1863531756"></a>mindio_ttp-<em id="zh-cn_topic_0000002039339945_i15340181201416"><a name="zh-cn_topic_0000002039339945_i15340181201416"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002039339945_i19614531957"><a name="zh-cn_topic_0000002039339945_i19614531957"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p96553958"><a name="zh-cn_topic_0000002039339945_p96553958"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p66253359"><a name="zh-cn_topic_0000002039339945_p66253359"></a><span id="zh-cn_topic_0000002039339945_ph845710020145"><a name="zh-cn_topic_0000002039339945_ph845710020145"></a>MindIO TFT</span>安装包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1862053354"><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row672652117018"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1468532510"><a name="zh-cn_topic_0000002039339945_p1468532510"></a>apex-0.1+ascend-cp3x-cp3x-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1869531256"><a name="zh-cn_topic_0000002039339945_p1869531256"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p156353454"><a name="zh-cn_topic_0000002039339945_p156353454"></a>MindSpeed-LLM依赖</p>
<p id="zh-cn_topic_0000002039339945_p1861353651"><a name="zh-cn_topic_0000002039339945_p1861353651"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p166105316515"><a name="zh-cn_topic_0000002039339945_p166105316515"></a>混合精度训练是在训练时混合使用单精度（float32）与半精度(float16)数据类型，将两者结合在一起，并使用相同的超参数实现了与float32几乎相同的精度。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a>请参见<span id="zh-cn_topic_0000002039339945_ph156792413596"><a name="zh-cn_topic_0000002039339945_ph156792413596"></a>《Ascend Extension for PyTorch 软件安装指南》中的“<a href="https://gitcode.com/Ascend/apex/blob/master/docs/zh/installing_apex.md">安装APEX模块</a>”章节</span>，根据实际情况编译APEX软件包。</p>
<p id="zh-cn_topic_0000002039339945_p1761531257"><a name="zh-cn_topic_0000002039339945_p1761531257"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row197268213011"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1361953254"><a name="zh-cn_topic_0000002039339945_p1361953254"></a>torch_npu-2.7.1.<em id="zh-cn_topic_0000002039339945_i16204112111321"><a name="zh-cn_topic_0000002039339945_i16204112111321"></a>{version}</em>-cp3x-cp3x-manylinux_2_28_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p56185314510"><a name="zh-cn_topic_0000002039339945_p56185314510"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p186653757"><a name="zh-cn_topic_0000002039339945_p186653757"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p15720535517"><a name="zh-cn_topic_0000002039339945_p15720535517"></a>Ascend Extension for PyTorch插件是基于昇腾的深度学习适配框架，使昇腾NPU可以支持PyTorch框架，为PyTorch框架的使用者提供昇腾AI处理器的超强算力。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p10718533510"><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a href="https://www.hiascend.com/document/detail/zh/Pytorch/720/configandinstg/instg/insg_0004.html" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1165115165020"><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p167047813263"><a name="zh-cn_topic_0000002039339945_p167047813263"></a>如果使用MindSpeed-LLM代码仓上的<span id="zh-cn_topic_0000002039339945_ph1987542822613"><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a>PyTorch</span>模型，需要使用<span id="zh-cn_topic_0000002039339945_ph1412723132619"><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a>Ascend Extension for PyTorch</span> 2.6.0及以上版本。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row1412215399516"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><ul id="zh-cn_topic_0000002039339945_ul104867135415"><li><span id="zh-cn_topic_0000002039339945_ph0853174512272"><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a>x86_64</span>架构：torch-2.7.1+cpu.cxx11.abi-cp3x-cp3x-linux_x86_64.whl</li><li><span id="zh-cn_topic_0000002039339945_ph7852164518272"><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a>ARM</span>架构：torch-2.7.1+cpu-cp3x-cp3x-manylinux_2_28_aarch64.whl</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p871953357"><a name="zh-cn_topic_0000002039339945_p871953357"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p17715531453"><a name="zh-cn_topic_0000002039339945_p17715531453"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a>官方<span id="zh-cn_topic_0000002039339945_ph19355165113512"><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a>PyTorch</span>包。</p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10，具体Python版本以MindSpeed-LLM版本说明为准。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p99745421447"><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a href="https://download.pytorch.org/whl/torch/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<p id="zh-cn_topic_0000002039339945_p483943610920"><a name="zh-cn_topic_0000002039339945_p483943610920"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row151232039750"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p4714531658"><a name="zh-cn_topic_0000002039339945_p4714531658"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1774531250"><a name="zh-cn_topic_0000002039339945_p1774531250"></a>是</p><p>CANN 8.5.0之前版本该包名为Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5720531514"><a name="zh-cn_topic_0000002039339945_p5720531514"></a>CANN算子包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p87153755"><a name="zh-cn_topic_0000002039339945_p87153755"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note13775154104217"><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p2075812144313"><a name="zh-cn_topic_0000002039339945_p2075812144313"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="row1173819266428"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="p7738142674217"><a name="p7738142674217"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p12738626184214"><a name="p12738626184214"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13738226114218"><a name="p13738226114218"></a>CANN Toolkit开发套件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p19271154916428"><a name="p19271154916428"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="note3272104913427"><a name="note3272104913427"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p2272174920421"><a name="p2272174920421"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row121231639952"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1048218391768"><a name="zh-cn_topic_0000002039339945_p1048218391768"></a>MindSpeed</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p3482939367"><a name="zh-cn_topic_0000002039339945_p3482939367"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p048215393610"><a name="zh-cn_topic_0000002039339945_p048215393610"></a>MindSpeed是针对昇腾设备的大模型加速库。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p7482193913619"><a name="zh-cn_topic_0000002039339945_p7482193913619"></a>git clone https://gitcode.com/Ascend/MindSpeed.git</p>
<p id="zh-cn_topic_0000002039339945_p9482139663"><a name="zh-cn_topic_0000002039339945_p9482139663"></a>cd MindSpeed</p>
<p id="zh-cn_topic_0000002039339945_p1948213912618"><a name="zh-cn_topic_0000002039339945_p1948213912618"></a>git checkout e4d5855250f0074670f41c423021286502410bf1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row144125121466"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p848215391768"><a name="zh-cn_topic_0000002039339945_p848215391768"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1548319391465"><a name="zh-cn_topic_0000002039339945_p1548319391465"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p16483239968"><a name="zh-cn_topic_0000002039339945_p16483239968"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p14483939561"><a name="zh-cn_topic_0000002039339945_p14483939561"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p748314394617"><a name="zh-cn_topic_0000002039339945_p748314394617"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row17301171913614"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p148310396616"><a name="zh-cn_topic_0000002039339945_p148310396616"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p348313399610"><a name="zh-cn_topic_0000002039339945_p348313399610"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p1348313391961"><a name="zh-cn_topic_0000002039339945_p1348313391961"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5483239564"><a name="zh-cn_topic_0000002039339945_p5483239564"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p2483339861"><a name="zh-cn_topic_0000002039339945_p2483339861"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row93022191368"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15485173911615"><a name="zh-cn_topic_0000002039339945_p15485173911615"></a>Dllogger代码仓</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p248514391563"><a name="zh-cn_topic_0000002039339945_p248514391563"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164850391565"><a name="zh-cn_topic_0000002039339945_p164850391565"></a>PyTorch日志工具。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1548619391164"><a name="zh-cn_topic_0000002039339945_p1548619391164"></a>git clone https://github.com/NVIDIA/dllogger.git</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row33025197610"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p0486133919618"><a name="zh-cn_topic_0000002039339945_p0486133919618"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1148619393610"><a name="zh-cn_topic_0000002039339945_p1148619393610"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164861439461"><a name="zh-cn_topic_0000002039339945_p164861439461"></a>用于安装pip模块。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p34865396611"><a name="zh-cn_topic_0000002039339945_p34865396611"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row03021191165"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15486163919616"><a name="zh-cn_topic_0000002039339945_p15486163919616"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1448693915619"><a name="zh-cn_topic_0000002039339945_p1448693915619"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p74862391165"><a name="zh-cn_topic_0000002039339945_p74862391165"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p9571628104613"><a name="zh-cn_topic_0000002039339945_p9571628104613"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

taskd和mindio_ttp的校验过程可参考[软件包 SUM 值校验](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md#section51703441649)小节。其余软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE]
>本章节以单台Atlas 800T A2 训练服务器、Ubuntu 20.04 Arm、配套Python 3.10为例来介绍训练镜像的制作，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000002039339945_section20489630477"></a>**

1. 参照[表1](#zh-cn_topic_0000002039339945_table1172542119019)，在宿主机上完成软件包的准备工作。
2. 编写如下Dockerfile。

    ```text
    FROM ubuntu:20.04
    WORKDIR /root
    COPY . .

    ARG PYTORCH_WHL=torch-2.7.1+cpu-cp310-cp310-manylinux_2_28_aarch64.whl
    ARG PYTORCH_NPU_WHL=torch_npu-2.7.1.{version}-cp310-cp310-manylinux_2_28_aarch64.whl
    ARG APEX_WHL=apex-0.1+ascend-cp310-cp310-linux_aarch64.whl
    ARG HOST_ASCEND_BASE=/usr/local/Ascend
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann
    # 示例使用的CANN版本为8.5.0,使用过程中请根据实际情况修改
    ARG TOOLKIT=Ascend-cann-toolkit_8.5.0_linux-aarch64.run
    ARG OPS=Ascend-cann-910b-ops_8.5.0_linux-aarch64.run
    ARG TASKD_WHL=taskd-7.3.0-py3-none-linux_aarch64.whl
    ARG MINDIO_TTP_WHL=mindio_ttp-1.0.0-py3-none-linux_aarch64.whl
    ARG MINDSPEED=MindSpeed
    ARG DLLOGGER=dllogger

    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf

    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list

    ARG DEBIAN_FRONTEND=noninteractive

    # 系统包
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
    # 建立Python软链
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
    # 系统
    RUN umask 0022 && apt update && \
            apt-get install -y --no-install-recommends \
            gcc g++ make cmake vim \
            zlib1g zlib1g-dev \
            openssl libsqlite3-dev libssl-dev \
            libffi-dev unzip pciutils \
            net-tools libblas-dev \
            gfortran libblas3 libopenblas-dev \
            curl unzip liblapack3 liblapack-dev \
            libhdf5-dev libxml2 patch
    # 时区
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
    # 配置pip源
    RUN mkdir -p ~/.pip \
    && echo '[global] \n\
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf
    # pip3.10
    RUN cd /tmp && \
        apt-get download python3-distutils && \
        dpkg-deb -x python3-distutils_*.deb / && \
        rm python3-distutils_*.deb && \
        cd - && \
        python get-pip.py && \
        rm get-pip.py
    RUN umask 0022 && \
        pip install sympy==1.4 && \
        pip install cffi && \
        pip install pathlib2 && \
        pip install grpcio && \
        pip install grpcio-tools && \
        pip install torchvision==0.22.1 && \
        pip install transformers==4.51.0 && \
        pip install absl-py && \
        pip install datasets && \
        pip install tokenizers==0.20.1
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
    # 安装torch、TorchNPU、apex包
    RUN umask 0022 && pip install $PYTORCH_WHL && \
        pip install $PYTORCH_NPU_WHL && \
        pip install $APEX_WHL

    # Ascend包
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录
    RUN umask 0022 &&  \
        cp ascend_install.info /etc/ && \
        mkdir -p /usr/local/Ascend/driver/ && \
        cp version.info /usr/local/Ascend/driver/ && \
        chmod +x $TOOLKIT && \
        chmod +x $OPS

    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc
    RUN umask 0022 && ./$OPS --install --quiet

    # 只为了安装toolkit包，所以需要清理，容器启动时通过Ascend Docker Runtime挂载进来
    RUN rm -f version.info && rm -f ascend_install.info && \
        rm -rf /usr/local/Ascend/driver/

    RUN umask 0022 && cd $MINDSPEED && \
        pip install -r requirements.txt && \
        pip install -e . && \
        echo "export PYTHONPATH=/root/MindSpeed:\$PYTHONPATH" >> ~/.bashrc

    RUN umask 0022 && cd $DLLOGGER && \
        python setup.py build && \
        python setup.py install

    # 导入环境变量
    ENV HCCL_WHITELIST_DISABLE=1

    # 创建/lib64/ld-linux-aarch64.so.1
    RUN umask 0022 && \
        if [ ! -d "/lib64" ]; \
        then \
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
        fi

    # MindCluster断点续训适配脚本。
    RUN umask 0022 && \
        pip install $TASKD_WHL && \
        pip install $MINDIO_TTP_WHL

    # 可选，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令。
    RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

    # 增加安装任务调度依赖库
    RUN pip install apscheduler

    RUN rm -rf tmp && \
        rm -f $PYTORCH_WHL && \
        rm -f $PYTORCH_NPU_WHL && \
        rm -f $APEX_WHL && \
        rm -f $TOOLKIT && \
        rm -f $OPS && \
        rm -f $TASKD_WHL && \
        rm -f $MINDIO_TTP_WHL && \
        rm -rf $DLLOGGER && \
        rm -rf Dockerfile
    ## 最后打包成镜像mindspeed-dl:v1
    ```

    >[!NOTE]
    >Python 3.10若无法通过PPA直接安装成功，或者deadsnakes PPA不提供Python 3.10版本的镜像源，则可下载源码手动编译安装。

3. 构建镜像。执行以下命令生成镜像。为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t mindspeed-dl:v1 .
    ```

4. 除按上述流程手动构建镜像外，用户也可直接使用开箱即用的示例镜像`torch-npu-sample-dl:v26.2.0-cann8.5.0-torch_npu2.7.1-910b-ubuntu22.04-py3.10-aarch64`。示例镜像可通过[AscendHub](https://www.hiascend.com/developer/ascendhub)获取，用于快速验证。

## 脚本适配<a name="ZH-CN_TOPIC_0000002511426481"></a>

### 流程说明<a name="ZH-CN_TOPIC_0000002511346469"></a>

模型脚本需要适配CKPT之后才可以使用断点续训功能，脚本适配大致流程和逻辑如[图1](#fig88341718121515)所示。

**图 1**  脚本适配流程<a name="fig88341718121515"></a>

![](../../../../../figures/scheduling/脚本适配流程.png "脚本适配流程")

### 适配示例<a name="ZH-CN_TOPIC_0000002511346445"></a>

本章节将指导用户step by step地完成断点续训的适配步骤。

>[!NOTE]
>
>- 为保证优雅容错与进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>- 断点续训展示的组件代码为开源代码，其中涉及到相关安全说明请参见[安全说明](../../../../07_references/05_appendix.md#安全说明)。
>- 下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
>- 模型的参数配置，根据模型仓的模型配置以实际情况来写。若修改不当，可能会引发不可预知的问题。
>- 若训练过程中出现“Failed to bind the IP port. Reason: The IP address and port have been bound already”报错，可以按照如下进行配置，详情请参见《CANN HCCL集合通信库》中的“[HCCL_HOST_SOCKET_PORT_RANGE](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/commlib/hcclug/docs/zh/user_guide/hccl_env/HCCL_HOST_SOCKET_PORT_RANGE.md)”章节。
>
>   ```shell
>   export HCCL_HOST_SOCKET_PORT_RANGE="60000-60050"
>   export HCCL_NPU_SOCKET_PORT_RANGE="61000-61050"
>   ```
>
>- 若使用TaskD组件且训练容器使用Host网络，则先通过`sysctl net.ipv4.ip_local_reserved_ports`查询当前预留端口配置后，通过`sysctl -w net.ipv4.ip_local_reserved_ports="xxx,9601,9602"`新增预留端口9601、9602（其中xxx指的是前面查出来已配置的端口，若无则省略）。

训练代码与数据集准备，可以参考[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/26.1.0/docs/zh/pytorch/training/pretrain/mcore/pretrain.md)。下面以两台Atlas 800T A2 训练服务器为例，说明具体操作步骤。

1. 拉取训练代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/MindSpeed-LLM.git
    git clone https://github.com/NVIDIA/Megatron-LM.git
    cd MindSpeed-LLM
    git checkout 2af5bab9785ce64af3b7d1fe12972a27152cd7af
    cd ..
    cd Megatron-LM
    git checkout core_v0.12.1
    cp -r megatron ../MindSpeed-LLM #此处目的是将Megatron-LM项目下的Megatron目录复制到MindSpeed-LLM项目下
    ## 重命名MindSpeed-LLM为QWEN3_for_PyTorch_2.7_code
    cd ..
    mv MindSpeed-LLM QWEN3_for_PyTorch_2.7_code
    ```

2. 获取模型权重。

    请用户自行从[Qwen3](https://huggingface.co/Qwen/Qwen3-8B/tree/main)下载模型权重放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-8b-hf”。

3. 获取数据集。

    请用户自行从[Alpaca](https://huggingface.co/datasets/tatsu-lab/alpaca/blob/main/data/train-00000-of-00001-a09b74b3ef9c3b56.parquet)下载数据集（以Alpaca数据集为例）放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-alpaca”。

4. 处理数据集。
    1. 启动容器。

        ```shell
        docker run -it -v /data/atlas_dls/public/:/data/atlas_dls/public/ -e ASCEND_VISIBLE_DEVICES=0-7 mindspeed-dl:v1 bash
        ```

    2. 在容器中执行如下操作。

        ```shell
        export TORCH_DEVICE_BACKEND_AUTOLOAD=0
        source /usr/local/Ascend/cann/set_env.sh
        cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code
        # 可选，如下为安装MindSpeed加速库操作，可在任意目录下执行。若制作镜像时已安装，则跳过该操作
        git clone https://gitcode.com/ascend/MindSpeed.git
        cd MindSpeed
        git checkout e4d5855250f0074670f41c423021286502410bf1
        pip install -r requirements.txt
        pip install -e .
        export PYTHONPATH=/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/MindSpeed:$PYTHONPATH
        cd ..
        ```

    3. 处理数据集。

        Qwen3要求使用Transformers\>=4.51.0，因此Python需使用3.9及以上版本且需要安装4.51.0及以上的Transformers。

        ```shell
        python preprocess_data.py \
            --input /data/atlas_dls/public/dataset/qwen3-alpaca/train-00000-of-00001-a09b74b3ef9c3b56.parquet \ # 数据集文件路径
            --tokenizer-name-or-path /data/atlas_dls/public/dataset/qwen3-8b-hf \ # 开源模型权重文件目录
            --tokenizer-type PretrainedFromHF \
            --handler-name GeneralPretrainHandler \
            --output-prefix /data/atlas_dls/public/dataset/qwen3-alpaca/alpaca \ # 会生成alpaca_text_document.bin和.idx文件
            --json-keys text \
            --workers 4 \
            --log-interval 1000
        ```

        >[!NOTE]
        >若出现报错：/usr/local/lib/python3.10/dist-packages/sklearn/utils/../../scikit\_learn.libs/libgomp-947d5fa1.so.1.0.0: cannot allocate memory in static TLS block，可执行以下命令预加载libgomp库。
        >
        >```shell
        >export LD_PRELOAD=/usr/local/lib/python3.10/dist-packages/scikit_learn.libs/libgomp-947d5fa1.so.1.0.0:$LD_PRELOAD
        >```

5. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindcluster-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../../../07_references/05_appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3”目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

6. 获取[训练任务YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml)。该YAML中已经配置了Pod级别重调度、进程级别重调度、进程级在线恢复、弹性训练等。根据实际情况配置挂载卷的服务器IP地址、各种重调度级别等。

    进程级别重调度、进程级在线恢复、弹性训练等训练进程级别的恢复与优雅容错不可同时存在。优雅容错的配置步骤请参见[优雅容错模式](./menu_examples_and_verification.md#ZH-CN_TOPIC_0000002511346449)。

7. 配置训练启动脚本train\_start.sh和训练任务YAML，请根据实际情况进行修改。
    1. 修改启动脚本基础参数。

        ```shell
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/ttplogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/trainlogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/demo/
        # 日志保存路径，可根据实际情况修改
        export ASCEND_PROCESS_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP       # 设置plog保存路径，其中$MINDX_TASK_ID为Ascend Operator注入的任务UID环境变量，$XDL_IP为任务YAML中写入的环境变量status.hostIP
        export TTP_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/ttplogs/ttplog$XDL_IP-$RANK    # 设置TTP日志保存路径，其中$RANK为Ascend Operator为PyTorch框架注入的环境变量
        export TRAIN_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/trainlogs/$XDL_IP-$RANK      # 设置训练日志保存路径
        export GLOO_SOCKET_IFNAME=enp189s0f0               # 物理机上可以通信的网口，根据主节点高速网卡实际情况进行配置，如任务YAML中配置hostNetwork为false，则设置为eth0
        export HCCL_SOCKET_IFNAME=enp189s0f0               # 如任务YAML中配置hostNetwork为false，则设置为eth0

        CKPT_SAVE_DIR="/job/code/output/ckpt" # 训练完成后的权重保存路径
        DATA_PATH="/job/data/alpaca_text_document" # 数据集路径，填入数据预处理时保存的数据路径
        TOKENIZER_PATH="/job/data/qwen3-8b-hf" # 词表路径，填入下载的开源权重词表路径
        CKPT_LOAD_DIR="/job/code/output/ckpt" # 权重加载路径
        ```

    2. 使用TaskD完成进程级别重调度、进程级在线恢复、进程级别原地恢复或弹性训练，还需拉起TaskD Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

            ```python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码，拉起TaskD Manager。

            在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi

            torchrun $DISTRIBUTED_ARGS ...
            ```

        3. 修改训练任务YAML，新增容器端口，在所有的Pod下增加TaskD通信使用的端口9601（如已有则跳过）。

            ```yaml
            ...
                    spec:
            ...
                      containers:
            ...
                        ports:
                         - containerPort: 9601
                           name: taskd-port
            ...
            ```

## 准备任务YAML<a name="ZH-CN_TOPIC_0000002511426415"></a>

集群调度组件为用户提供YAML示例，用户需要根据使用的功能、模型类型和任务类型等，并根据使用的故障处理模式，选择相应的YAML示例并根据需求进行相应修改后才可使用。

**表 2**  训练任务YAML示例

<a name="table350244433714"></a>
<table><thead align="left"><tr id="row135031644183710"><th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.1"><p id="p8503244173715"><a name="p8503244173715"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="16.173234646929384%" id="mcps1.2.8.1.2"><p id="p145038448375"><a name="p145038448375"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="8.521704340868173%" id="mcps1.2.8.1.3"><p id="p919210345266"><a name="p919210345266"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="13.672734546909378%" id="mcps1.2.8.1.4"><p id="p5503544193713"><a name="p5503544193713"></a>模型</p>
</th>
<th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.5"><p id="p19672186404"><a name="p19672186404"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="15.433086617323463%" id="mcps1.2.8.1.6"><p id="p1096741894013"><a name="p1096741894013"></a>获取链接</p>
</th>
<th class="cellrowborder" valign="top" width="15.413082616523303%" id="mcps1.2.8.1.7"><p id="p2967518174012"><a name="p2967518174012"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row4503174412371"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p09365292408"><a name="p09365292408"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul129364297402"></a><ul id="ul129364297402"><li><span id="ph157633217501"><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p319343422611"><a name="p319343422611"></a><span id="ph310231710274"><a name="ph310231710274"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p493616294406"><a name="p493616294406"></a><span id="ph22631282914"><a name="ph22631282914"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p893610293406"><a name="p893610293406"></a>pytorch_multinodes_acjob_910b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1987716427402"><a name="p1987716427402"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml" target="_blank" rel="noopener noreferrer">pytorch_multinodes_acjob_910b.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p8936152964011"><a name="p8936152964011"></a>示例默认使用2*8卡任务</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>当前部分训练框架未提供Atlas 900 A3 SuperPoD 超节点的断点续训示例YAML，用户可以在示例YAML中的labels下新增annotations字段即可。示例如下：
>
>```yaml
>...
>  labels:
>...
>  annotations:
>    sp-block: "32"   # 逻辑超节点芯片数量，sp-block字段的详细说明，可以参见YAML参数说明。
>    huawei.com/schedule_policy: "chip2-node16-sp"    # 根据硬件形态设置调度策略
>...
>```

## 下发任务<a name="ZH-CN_TOPIC_0000002479226548"></a>

示例YAML中，任务部署在default命名空间下。本章节以Pytorch框架为例，下发训练任务。

1. 登录管理节点，进入YAML文件所在路径。
2. 在管理节点执行以下命令，使用YAML下发训练任务。

    ```shell
    kubectl apply -f XXX.yaml
    ```

    例如：

    ```shell
    kubectl apply -f pytorch_multinodes_acjob_910b.yaml
    ```

    回显如下：

    ```ColdFusion
    configmap/reset-config-default-test-pytorch created
    ascendjob.mindxdl.gitee.com/default-test-pytorch created
    ```

## 查看任务进程<a name="ZH-CN_TOPIC_0000002511426461"></a>

训练任务下发成功后，训练任务就可正常运行。可通过如下内容查看训练任务运行情况。

**查看所有训练任务<a name="section16792164211375"></a>**

查看当前节点上运行的所有训练任务，操作步骤如下。

1. 登录管理节点，进入YAML文件所在路径。
2. 执行以下命令，查看训练任务运行情况。

    ```shell
    kubectl get pods -A -o wide
    ```

    回显示例如下。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node1          <none>           <none>
    default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
    ……
    ```

**查看单个Pod的训练任务<a name="zh-cn_topic_0000001621551937_section1141119143319"></a>**

查看其中一个Pod上运行的训练任务，操作步骤如下。

执行以下命令，查看训练任务运行情况。

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显示例如下，出现loss即表示任务正常运行。

![](../../../../../figures/scheduling/unnaming-71.png)

**查看是否存在CKPT文件<a name="section979416428371"></a>**

故障恢复功能是通过参考CKPT文件实现的，用户需要查看存储节点上是否存在CKPT文件。

用户可以等待训练任务运行时间超过用户设置的保存CKPT文件的时间后，查看设置的保存CKPT文件的路径下是否存在周期性CKPT文件，操作步骤如下。

1. 登录存储节点，执行以下命令，进入CKPT文件路径。

    ```shell
    cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/output/ckpt
    ```

2. 执行以下命令，查看当前目录是否存在周期性CKPT文件。

    ```shell
    ll ./
    ```

    回显示例如下，说明存在周期性CKPT文件。

    ```ColdFusion
    total 8
    drwxr-xr-x  18 root root   8192 Jun 22 18:39 iter_0000100
    -rw-r--r--  1 root root    2    Jun 22 18:39 latest_checkpointed_iteration.txt
    ```

3. （可选）如果使用临终遗言，可以在保存CKPT的路径下，执行以下命令，查看当前目录是否存在临终CKPT文件。

    ```shell
    ll ./
    ```

    回显示例如下，说明存在临终CKPT文件。

    ```ColdFusion
    total 8
    drwxr-xr-x  18 root root   8192 Jun 22 15:39 iter_0000009
    -rw-r--r--  1 root root    2    Jun 22 15:39 latest_checkpointed_iteration.txt
    ```

## 查看训练结果<a name="ZH-CN_TOPIC_0000002479386554"></a>

### （可选）构造故障<a name="ZH-CN_TOPIC_0000002511426449"></a>

本章节将指导用户构造简单的故障，包括节点故障、参数面网络故障和业务面故障。

>[!NOTE]
>构造芯片故障存在安全风险，如需构造请联系华为技术支持工程师处理。

**构造节点故障<a name="section173881558133914"></a>**

通过重启训练节点，模拟节点下电导致节点状态丢失。该故障在节点重启完成后可自动恢复。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，重启该训练节点，模拟节点状态丢失故障。

    ```shell
    reboot
    ```

3. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

4. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

**构造参数面网络故障<a name="section22113033919"></a>**

通过断开NPU网络链路模拟参数面网络故障。NPU网络故障不影响单机训练任务。用户在断开链路后需手动恢复，否则该故障会一直存在。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，构造NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -link -s down
    ```

    >[!NOTE]
    >device\_id为NPU的ID，可以通过<b>npu-smi info</b>命令查看NPU的ID。

3. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障构造成功。

    ```ColdFusion
    net health status: Fault
    ```

4. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

6. 执行以下命令，恢复NPU网络链路故障。

    ```shell
    hccn_tool -i {device_id} -cfg recovery
    ```

7. 执行以下命令，查看NPU链路状态。

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障已经恢复。

    ```ColdFusion
    net health status: Success
    ```

**构造业务面故障<a name="section9891038124213"></a>**

通过删除训练进程，模拟业务面故障。

1. 在训练任务正常训练出iteration后，登录正在训练的节点。
2. 执行以下命令，使用训练启动脚本，查询训练进程信息。

    ```shell
    ps -ef | grep python| grep 训练启动脚本.py
    ```

3. 执行以下命令，手动删除PID最小的训练进程。

    ```shell
    kill -9 pid
    ```

4. 在Master节点多次执行以下命令，查看Pod状态。

    ```shell
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5. 在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```shell
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

### 重调度模式<a name="ZH-CN_TOPIC_0000002479386534"></a>

**重调度情况<a name="section87441013105513"></a>**

>[!NOTE]
>当节点发生故障时，Volcano会将该训练任务调度到其他满足条件的节点上继续运行。

登录管理节点，执行以下命令查看训练任务运行情况。

```shell
kubectl get pods -A -o wide
```

故障前，若训练任务调度到了node1和node2上面，当node1节点上发生故障，此时Volcano组件会将node1和node2上训练任务重调度到node2和node3节点上，重调度后回显示例如下。

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node3          <none>           <none>
……
```

**查看其中一个Pod运行情况<a name="section28985295314"></a>**

执行以下命令，查看单个Pod的训练任务运行情况。

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显如下表示发生故障时，使用最近保存的第9步的Checkpoint文件恢复，实现训练任务第10个iteration开始继续训练。

```ColdFusion
2025-09-08 11:34:00.400331 warn 1900637 [77840][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.401841 warn 1900631 [28432][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.402489 warn 1900639 [10928][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.426989 warn 1900627 [98608][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.429141 warn 1900634 [24592][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
(min, max) time across ranks (ms):
    load-checkpoint ................................: (32107.12, 32108.53)
(min, max) time across ranks (ms):
    model-and-optimizer-setup ......................: (32528.79, 32544.35)
    train/valid/test-data-iterators-setup ..........: (72.68, 656.79)
[rank16]:[W908 11:34:01.252908110 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank24]:[W908 11:34:01.254614170 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank17]:[W908 11:34:01.421349990 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank20]:[W908 11:34:01.431165020 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank19]:[W908 11:34:01.431240250 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank30]:[W908 11:34:01.431707980 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
...
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
 [2025-09-08 11:37:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 6932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
 [2025-09-08 11:37:03] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
...
```

**查看任务重调度记录<a name="section97707231547"></a>**

执行如下命令查看任务重调度记录。

```shell
kubectl describe cm -n mindx-dl job-reschedule-reason
```

回显示例如下。

```ColdFusion
Name:         job-reschedule-reason
Namespace:    mindx-dl
Labels:       <none>
Annotations:  <none>
Data
====
recent-reschedule-records:
----
{"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b":{"JobID":"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0908 11:36:10","RescheduleTimeStamp":1759683370,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"default-test-pytorch-worker-0","NodeName":"node2","NodeRankIndex":"1"}]}]}}
Events:  <none>
```

### 优雅容错模式（本功能已日落）<a name="ZH-CN_TOPIC_0000002511346479"></a>

本章节指导用户查看使用故障处理的优雅容错模式的训练信息。当芯片发生故障时，进程退出后进行优雅容错处理，恢复后重新拉起进程。

**日志说明<a name="section83075820188"></a>**

重新拉起的训练进程的训练日志在“_训练脚本路径_/newlog”中，具体说明如下。

- QWEN3（PyTorch）训练日志：“/data/atlas\_dls/public/code/QWEN3\_for\_PyTorch\_2.7\_code/alllogs”。

**操作步骤<a name="section25042117188"></a>**

1. 登录管理节点，执行以下命令查看芯片情况。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程占用片上内存，正常训练中。

    ![](../../../../../figures/scheduling/1-13.png)

2. 故障发生后，执行以下命令查看芯片信息。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已退出，释放片上内存。

    ![](../../../../../figures/scheduling/2.png)

3. 故障恢复后，执行以下命令查看芯片信息。

    ```shell
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已重新拉起占用片上内存，正常训练中。

    ![](../../../../../figures/scheduling/3.png)

## 构造故障并验证故障处理

本文档提供多种故障处理策略的验证方法，用户可根据实际配置的故障处理策略，快速跳转到对应验证章节。

**快速导航**

- [验证Job级别重调度](#验证job级别重调度)
- [验证Pod级别重调度](#验证pod级别重调度)
- [验证进程级别重调度](#验证进程级别重调度)
- [验证进程级别在线恢复](#验证进程级别在线恢复)

### 验证Job级别重调度<a name="验证job级别重调度"></a>

**测试准备**

在基础调度的任务YAML中，添加Job级别重调度的配置，配置说明可参考[配置Job级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置job级别重调度)，原理可参考[Job级别重调度](../01_solutions_principles/01_fault_handling.md#job级别重调度)。

**测试操作**

1. 下发任务。

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`taskmgr-npu-020-default-test-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态和UID。

   1. 查看任务状态。

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，STATUS字段为Running表示任务正常运行。

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>    0          2s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>    0          3s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. 查看2个Pod的UID：

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下：

      ```bash
      7286faf8-f029-450a-b302-5e6e94d4346c
      997add9e-6115-456c-9e8e-e05e4b70bb12
      ```

3. 构造故障。

   1. 查询任务进程。

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      回显示例如下：

      ```ColdFusion
      2398104
      2398105
      2398107
      ```

   2. 终止进程模拟故障。

      ```bash
      kill -9 2398104
      ```

4. 观察重调度过程。

   监控该Job的2个Pod状态变化。

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   该Job的2个Pod历史状态如下，观察加粗字段的变化可以发现该Job的2个Pod会经历Terminating→Pending→ContainerCreating→Running阶段，然后正常运行，表示Job重调度成功：

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running             0          3s      xx.xx.xx.xx       localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== 注入故障 ======================
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          73s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;                 localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;                 localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Running</strong>             0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0          2s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   </pre>

**预期结果**

1. 查看任务状态和UID。

   1. 查看任务状态。

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，STATUS字段为Running表示任务正常运行。

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          2s      xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          33s     xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. 查看2个Pod的UID：

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下，该Job的2个Pod的UID均发生变化，说明2个Pod都经历了重调度，即触发Job级别重调度：

      ```ColdFusion
      2a24eee8-88f1-4107-bc9d-dabcfb09dea9
      074f9f9c-35f1-4b9e-9298-5b2bcf3759e7
      ```

### 验证Pod级别重调度<a name="验证pod级别重调度"></a>

**测试准备**

在基础调度的任务YAML中，添加Pod级别重调度的配置，配置说明可参考[配置Pod级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置pod级别重调度)，原理可参考[Pod级别重调度](../01_solutions_principles/01_fault_handling.md#pod级别重调度)。

**测试操作**

1. 下发任务。

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`taskmgr-npu-020-default-test-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态和UID。

   1. 查看任务状态。

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行。

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. 查看2个Pod的UID。

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下：

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      47291595-85b0-47ff-8393-c922d0e2dfb2
      ```

3. 构造故障。

   1. 查询任务进程。

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      回显示例如下：

      ```ColdFusion
      2398132
      2398144
      2398158
      ```

   2. 终止进程模拟故障。

      ```bash
      kill -9 2398144
      ```

4. 观察重调度过程。

   监控该Job的2个Pod状态变化。

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   该Job的2个Pod历史状态如下，观察加粗字段的变化可以发现故障Pod（taskmgr-npu-020-default-test-1）会经历Error→Terminating→Pending→ContainerCreating→Running阶段，然后正常运行，表示Pod重调度成功：

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running              0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running              0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== 注入故障 ======================
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Error</strong>               0          34s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      &lt;none&gt;                localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      &lt;none&gt;                localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0           2s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   </pre>

**预期结果**

1. 查看任务状态和UID。

   1. 查看任务状态。

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行。

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          66s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          31s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. 再次查看2个Pod的UID。

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下，taskmgr-npu-020-default-test-0 Pod的UID未发生变化，taskmgr-npu-020-default-test-1 Pod的UID发生变化，说明只有发生故障的Pod（taskmgr-npu-020-default-test-1）经历了重调度，即触发Pod级别重调度：

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      6eb3c217-3b63-457a-9010-9d236d281634
      ```

### 验证进程级别重调度<a name="验证进程级别重调度"></a>

**测试准备**

在基础调度的任务YAML中，添加进程级别重调度的配置，配置说明可参考[配置进程级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级别重调度)，原理可参考[进程级别重调度](../01_solutions_principles/01_fault_handling.md#进程级别重调度)。

**测试操作**

1. 下发任务。

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`process-reschedule-function-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态。

   ```bash
   kubectl get pod -A -o wide
   ```

   回显示例如下，出现Running表示任务正常运行：
   <pre codetype="ColdFusion">
   trjob            process-reschedule-function-master-0   1/1     Running   0               14s   xx.xx.xx.xx     master-69-117   &lt;none&gt;           &lt;none&gt;
   trjob            process-reschedule-function-worker-0   1/1     Running   0               14s   xx.xx.xx.xx     work-69-115     &lt;none&gt;           &lt;none&gt;
   </pre>

3. 查看训练日志迭代步数，确认训练已正常迭代。

   ```bash
   kubectl logs -n trjob process-reschedule-function-worker-0|grep -Po '] iteration [[:space:]]*'|wc -l
   ```

   回显示例如下：

   ```bash
   50
   ```

4. 查看进程ID，并构造故障。

   1. 查看进程ID。

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      回显示例如下：

      ```ColdFusion
      635755
      635756
      635760
      635770
      635777
      635784
      635791
      635795
      ```

   2. 终止其中一个进程模拟故障。

      ```bash
      kill -9 635777
      ```

**预期结果**

1. 观察训练日志。

   ```bash
   kubectl logs -n trjob process-reschedule-function-master-0
   ```

   回显示例如下：

   ```ColdFusion
   # 出现以下信息说明开始触发ARF流程
   Mindx calling notify do ARF repair

   # 出现以下信息说明ARF成功
   ... Mindio do repair operation ok ...
   ```

2. 查看ConfigMap job-reschedule-reason中是否有任务信息。

   ```bash
   kubectl describe cm -n mindx-dl job-reschedule-reason |grep process-reschedule-function
   ```

   回显示例如下，其中包含重调度的时间，触发重调度的pod、node、rank，本任务当前重调度次数等信息：

   ```ColdFusion
   {"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07":{"JobID":"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0603 05:16:52","RescheduleTimeStamp":1780435012,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"process-reschedule-function-worker-0","NodeName":"work-69-115","NodeRankIndex":"1"}]}]}}
   ```

### 验证进程级别在线恢复<a name="验证进程级别在线恢复"></a>

本章节通过在训练代码中打桩构造片上内存的UCE故障，指导用户完成进程级在线恢复验证的适配步骤。

>[!NOTE]
>
>- 本章节相关修改仅用于指导用户在测试环境下验证进程级在线恢复功能，切勿将此打桩版本上线到生产环境。
>- 配置本章节步骤前，请确保训练能正常拉起并已配置进程级在线恢复。
>- 为保证进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>- 下文中代码可能与实际版本存在差异，请以实际版本代码为准。

#### MindCluster适配<a name="ZH-CN_TOPIC_0000002479386410"></a>

1. <a name="li977718409381"></a>拉取MindCluster代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/mind-cluster.git
    cd ./mind-cluster/component/clusterd
    git checkout branch_v26.1.0   # branch_v26.1.0是代码仓版本分支，请自行切换到目标分支
    ```

2. 修改ClusterD代码。
   1. 打开“pkg/application/faultmanager/jobprocess/faultrank/job\_fault\_rank\_processor.go”文件。

      ```shell
      vi pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go
      ```

   2. 按“i”进入编辑模式，添加如下加粗代码。

      <pre codetype="go">
         package faultrank

         import (
         …
            <strong>"clusterd/pkg/domain/faultdomain/collector"</strong>
         …
         )
         …
         func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(
         …
               if deviceDetail, ok := processor.retryInBusinessPlane(podInfo.jobId, nodeName, deviceName); ok {
                  faultRankList = append(faultRankList, constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid,
                     PodRank: podRankStr, FaultCode: faultdomain.GetRetryCodeByFaultType(deviceDetail.FaultType),
                     FaultLevel:  constant.RestartBusiness,
                     DoStepRetry: processor.canDoStepRetry(podInfo.jobId, nodeName, deviceName),
                     DeviceId:    deviceInfo.DeviceID,
               })
               <strong>collector.ReportInfoCollector.ReportRetryInfo(podInfo.jobId, deviceInfo.RankID, constant.JobNotRecover, constant.UceFaultType)   // 业务面故障时间设置为无效时间，避免单次故障重复触发进程级在线恢复</strong>
            }
        …
      </pre>

   3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

3. <a name="li114977117517"></a>编译ClusterD。

   ```shell
   cd ./build/
   chmod +x build.sh && dos2unix build.sh
   sed -i 's|build_version="v[^"]\+"|build_version="xxx"|g' build.sh  # xxx替换为版本号，如v26.1.0
   sed -i 's|export CGO_ENABLED=0|export CGO_ENABLED=1|g' build.sh  # 开启CGO功能
   ./build.sh # 编译ClusterD，需要提前安装好Go sdk，具体版本以ClusterD组件代码的go.mod文件内容为准
   ```

   编译成功后，会在“../output/”目录下生成相关文件，可执行如下命令进行查看：

   ```shell
   ll ../output/
   ```

   回显示例如下：

   ```bash
   -r-x------. 1 root root 45891128 Aug 13 10:52 clusterd
   -r--------. 1 root root     4021 Aug 13 10:52 clusterd-v26.1.0.yaml
   -r--------. 1 root root      946 Aug 13 10:52 Dockerfile
   -r--------. 1 root root      209 Aug 13 10:52 faultDuration.json
   -r--------. 1 root root      207 Aug 13 10:52 fdConfig.yaml
   -r--------. 1 root root      467 Aug 13 10:52 publicFaultConfiguration.json
   -r--------. 1 root root      756 Aug 13 10:52 relationFaultCustomization.json
   ```

4. <a name="li89701053589"></a>进入output目录，制作ClusterD镜像。

   ```shell
   cd ../output/
   docker build --no-cache -t clusterd:{tag} ./  # {tag}与步骤3中build_version="xxx"的取值保持一致
   ```

5. （可选）保存镜像，并将保存后的镜像文件和clusterd-\{tag\}.yaml文件上传到主节点。若[步骤1](#li977718409381)到[步骤4](#li89701053589)在主节点执行，可跳过该步骤。

   ```shell
   docker save -o clusterd.tar clusterd:{tag}  # 保存镜像
   docker load -i clusterd.tar  # 在主节点导入镜像
   ```

6. 在主节点重新拉起ClusterD。

   ```shell
   kubectl delete -f  clusterd-{tag}.yaml  # 删除旧ClusterD容器
   kubectl apply -f  clusterd-{tag}.yaml  # 拉起新容器
   ```

#### 脚本适配<a name="ZH-CN_TOPIC_0000002479226412"></a>

##### PyTorch场景适配示例（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002511426361"></a>

1. 搭建训练环境，拉起训练，详细请参见[PyTorch场景适配示例（基于MindSpeed-LLM）](#ZH-CN_TOPIC_0000002511346445)。
2. 开启进程级在线恢复，详细请参见[配置进程级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级在线恢复)。
3. 在“QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py”代码中增加如下加粗内容，打桩注入故障，新增代码根据环境变量“RAISE\_UCE\_ERROR\_STEP\_AND\_RANK”获取注入故障迭代位置和故障rank信息。

   <pre codetype="Python">
      <strong>import os</strong>
      <strong>import ast</strong>
      <strong>…</strong>
      <strong>GLB_CNT = 0</strong>
      def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
         """Train the model function."""
         args = get_args()
         timers = get_timers()
         …
         while iteration < args.train_iters:
            …
            num_microbatches = get_num_microbatches()
            update_num_microbatches(args.consumed_train_samples, consistency_check=True)
            <strong>global GLB_CNT</strong>
            <strong>cur_rank = torch.distributed.get_rank()</strong>
            <strong>uce_env = os.getenv("RAISE_UCE_ERROR_STEP_AND_RANK", "{}")</strong>
            <strong>uce_step_rank = ast.literal_eval(uce_env)</strong>
            <strong>if iteration in uce_step_rank and cur_rank == uce_step_rank[iteration] and GLB_CNT < iteration:</strong>
               <strong>GLB_CNT = iteration</strong>
               <strong>print(f"############# rank:{cur_rank} start UCE error #############")</strong>
               <strong>raise RuntimeError('UCE ERROR')</strong>
            args.curr_iteration = iteration
            …
   </pre>

4. 修改启动脚本“QWEN3\_for\_PyTorch\_2.7\_code/scripts/train\_start.sh”。

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # 配置故障注入的迭代和卡号，在第3个迭代的rank 8卡和第10个迭代的rank 9卡上注入UCE故障
   sed -i 's/check_memory_result = torch_npu.npu.check_uce_in_memory(device)/check_memory_result = ha_constant.UCE_HIGH_LEVEL/g' /job/code/mindspeed_llm/core/high_availability/tft_stop_clean.py #修改TorchNPU接口返回值，将训练代码抛出的异常识别为UCE故障
   …
   ```

#### 验证流程

以下示例基于**双机16卡**（单机8卡，Master rank 0–7、Worker rank 8–15）环境，与[脚本适配](#ZH-CN_TOPIC_0000002479226412)中 `RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"` 的配置一致。若使用单机或其他拓扑，请同步调整环境变量与下文 `grep` 中的rank、Pod名称。

**测试准备**

- 在基础调度的任务 YAML 中，添加进程级在线恢复的配置，配置说明可参考[配置进程级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级在线恢复)，原理可参考[进程级在线恢复](../01_solutions_principles/01_fault_handling.md#进程级在线恢复)。
- 已完成 MindCluster 适配和脚本适配；启动脚本中的 `RAISE_UCE_ERROR_STEP_AND_RANK` 与下文验证命令中的 rank、迭代步保持一致。

**测试操作**

1. 下发任务。

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - 请将 `trjob.yaml` 替换为实际的任务YAML文件；若按上文QWEN3脚本适配，请使用对应的任务YAML与Pod名称。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的 `process-online-recovery-` 和 `trjob` 均为示例值。

2. 查看任务状态。

   ```bash
   kubectl get pod -A -o wide
   ```

   回显示例如下，出现Running表示任务正常运行。

   ```ColdFusion
   trjob            process-online-recovery-master-0                   1/1     Running   0                 14s     xx.xx.xx.xx     master-x   <none>           <none>
   trjob            process-online-recovery-worker-0                   1/1     Running   0                 14s     xx.xx.xx.xx     worker-x   <none>           <none>
   ```

3. 监控训练日志
   1. 监控训练日志检查是否触发UCE故障。

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "status error, rank:8"
      ```

      >[!NOTE]
      > 本示例在第3步将故障注入rank 8。`grep` 关键字中的rank需与环境变量中配置的全局rank保持一致。

      回显示例如下，触发UCE故障。

      ```ColdFusion
      2026-06-04 09:24:31.767278 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 0
      2026-06-04 09:24:33.767422 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 1417
      ```

      >[!NOTE]
      >日志中的 `step: 3` 表示故障在第3个训练迭代步触发。`npu_status: 2` 表示MindIO/TTP侧已进入UCE处理状态；在本打桩场景下由软件模拟路径触发，不代表真实硬件片上内存故障。

**预期结果**

   1. 检查第3步故障的恢复结果。在Master或Worker任一Pod上输出大于等于1，即说明修复成功。

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "(0, 'Mindio do repair operation ok', {}, 'retry')"|wc -l
      ```

   2. 检查迭代是否正常。

      1. 查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下：

      ```ColdFusion
      trjob            process-online-recovery-master-0                   1/1     Running   0                 110s    xx.xx.xx.xx     master-x   <none>           <none>
      trjob            process-online-recovery-worker-0                   1/1     Running   0                 110s    xx.xx.xx.xx     worker-x   <none>           <none>
      ```

      >[!NOTE]
      > 此时请检查RESTARTS列，该数值必须保持为0。证明在整个UCE故障及修复过程中，Pod容器从未发生过重启。

   3. 查看训练迭代步数。

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # 返回：0
      kubectl logs -n trjob process-online-recovery-worker-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # 返回：11
      ```

      >[!NOTE]
      > - 以上命令中 `grep` 的迭代步数（如 `iteration 4`）需根据实际注入故障的步数调整。若故障注入在第 `N` 步，恢复后应从第 `N+1` 步继续训练，因此应为 `grep iteration [[:space:]]*{N+1}`。本示例中第3步故障对应 `iteration 4`，第10步故障对应 `iteration 11`。
      > - 在分布式多机训练中，受训练框架的日志重定向机制影响，各Rank的迭代日志可能仅输出在部分节点的stdout中，或被重定向至本地物理日志文件。
      > - 本示例中Master节点返回0、Worker节点返回11，只要任一节点存在大于0的计数，即证明热修复后训练已跨越对应故障步数并继续。

## 删除任务<a name="ZH-CN_TOPIC_0000002479386566"></a>

**操作步骤<a name="section324819211118"></a>**

在下发任务的YAML目录执行以下命令，删除对应的训练任务。

```shell
kubectl delete -f XXX.yaml
```

示例如下：

```shell
kubectl delete -f pytorch_multinodes_acjob_910b.yaml
```

回显示例如下：

```ColdFusion
configmap "reset-config-default-test-pytorch" deleted
ascendjob.mindxdl.gitee.com "default-test-pytorch" deleted
```

## 运行维护<a name="ZH-CN_TOPIC_0000002479386520"></a>

**前提条件<a name="section18751194535314"></a>**

此功能只适用于特定场景下，用户需要使用重调度功能，且Ascend Device Plugin的启动YAML中已设置autoStowing参数（该字段已日落）为false。

**操作方法<a name="section8557331115714"></a>**

- 用户可以使用以下命令，将健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-Recover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-Recover**”标签，该标签中的芯片会重新放入资源池中供程序调度。

    >[!NOTE]
    >该命令仅作清除Recover标签信息使用，请不要用于添加标签。

- 用户可以使用以下命令，将参数面网络健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-NetworkRecover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-NetworkRecover**”标签，同时也会清除“**huawei.com/Ascend910-NetworkUnhealthy**”中对应的芯片。

    >[!NOTE]
    >该命令仅作清除NetworkRecover标签信息使用，请不要用于添加标签。
