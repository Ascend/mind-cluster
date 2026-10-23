# Examples and Verification in MindSpore Scenario

## Building an Image<a name="ZH-CN_TOPIC_0000002511426469"></a>

[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/26.2.0), as the Ascend large model training framework, aims to provide an end-to-end large language model training solution for Ascend chips, including distributed pre-training, distributed instruction fine-tuning, distributed preference alignment, and the corresponding development toolchain. The [MindSpeed-LLM User Guide](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md) includes chapters on repository cloning, environment setup, and large model training. To build a MindSpeed-LLM training framework image, you can refer to this section and the [MindSpeed-LLM User Guide](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md).

Resumable training can be built based on the base training image. For details on building the base training image, see the [Building a Container Image Using Dockerfile (PyTorch)](../../../../07_references/02_common_operations.md#building-a-container-image-using-dockerfile-pytorch) section.

This chapter shows how to build a training image based on Ubuntu 20.04 by following the steps for building a training base image.

>[!NOTE]
>
>- The following example uses MindSpeed-LLM 26.2.0.
>- For product forms with UB, such as Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and Atlas 950 SuperPoD, UB driver related files need to be added to the training image. The following two methods are supported:
>   - Download from the archive directory of the <a href="https://mirrors.huaweicloud.com/ascend/">Huawei Cloud image repository</a> and install it in the training image.
>   - Install Ascend Docker Runtime, which mounts the UB driver related content in `ub_driver.list` by default. The files in `ub_driver.list` depend on the glibc version of the host. The currently supported OS versions and glibc versions are as follows:
>     - Host: openEuler24.03 (glibc2.38)
>     - Container: openEuler24.03 (glibc2.38), Ubuntu24.04 (glibc2.39).

**Preparing Software Packages<a name="zh-cn_topic_0000002039339945_section18254161612586"></a>**

As shown in [Table 1](#zh-cn_topic_0000002039339945_table1172542119019),obtain the software packages of the corresponding OS and prepare the Dockerfile and script file required by the image. In the package names, `{version}` indicates the version number, `{arch}` indicates the architecture, and `{chip_type}` indicates the chip type.

**Table 1** Preparing software packages

<a name="zh-cn_topic_0000002039339945_table1172542119019"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339945_row157251121508"><th class="cellrowborder" valign="top" width="24.55%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002039339945_p1441653254"><a name="zh-cn_topic_0000002039339945_p1441653254"></a>Software Package</p>
</th>
<th class="cellrowborder" valign="top" width="25.45%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002039339945_p2052053751"><a name="zh-cn_topic_0000002039339945_p2052053751"></a>Required</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002039339945_p657531455"><a name="zh-cn_topic_0000002039339945_p657531455"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002039339945_p1859531759"><a name="zh-cn_topic_0000002039339945_p1859531759"></a>Obtaining Method</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339945_row16726192116014"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1754534515"><a name="zh-cn_topic_0000002039339945_p1754534515"></a>taskd-<em id="i2511021165615"><a name="i2511021165615"></a>{version}</em>-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p4321152352612"><a name="p4321152352612"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p205155310512"><a name="zh-cn_topic_0000002039339945_p205155310512"></a>Resumable training whl package.</p>
<div class="note" id="zh-cn_topic_0000002039339945_note494818501423"><a name="zh-cn_topic_0000002039339945_note494818501423"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p159489506423"><a name="zh-cn_topic_0000002039339945_p159489506423"></a>Before installing <span id="ph1670711477256"><a name="ph1670711477256"></a>TaskD</span>, ensure that the <span id="zh-cn_topic_0000002039339945_ph998914174412"><a name="zh-cn_topic_0000002039339945_ph998914174412"></a>PyTorch</span> framework is correctly installed. The currently supported <span id="zh-cn_topic_0000002039339945_ph2908133144419"><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a>PyTorch</span> versions are 2.1.0, 2.3.0, 2.4.0, 2.5.0, 2.6.0, 2.7.1. TaskD depends on the PyTorch framework. Select a PyTorch version without known security vulnerabilities or obtain the corresponding version with security issues fixed from the official community.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p19595310517"><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a href="https://www.hiascend.com/en/developer/software/mindcluster/download?versionId=460&ids=53%2C154%2C%2C58%2C60%2C63%2C" target="_blank" rel="noopener noreferrer">Download Link</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1386820525510"><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p12512532515"><a name="zh-cn_topic_0000002039339945_p12512532515"></a>The link points to the download page of the <span id="ph480901420289"><a name="ph480901420289"></a>TaskD's</span> compressed package Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a>{arch}</em>.zip. You need to decompress it to obtain the corresponding whl software package.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row572619211108"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1863531756"><a name="zh-cn_topic_0000002039339945_p1863531756"></a>mindio_ttp-<em id="zh-cn_topic_0000002039339945_i15340181201416"><a name="zh-cn_topic_0000002039339945_i15340181201416"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002039339945_i19614531957"><a name="zh-cn_topic_0000002039339945_i19614531957"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p96553958"><a name="zh-cn_topic_0000002039339945_p96553958"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p66253359"><a name="zh-cn_topic_0000002039339945_p66253359"></a><span id="zh-cn_topic_0000002039339945_ph845710020145"><a name="zh-cn_topic_0000002039339945_ph845710020145"></a>MindIO TFT</span> installation package.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1862053354"><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a href="https://www.hiascend.com/en/developer/software/mindcluster/download?versionId=460&ids=53%2C154%2C%2C58%2C60%2C63%2C" target="_blank" rel="noopener noreferrer">Download Link</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row672652117018"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1468532510"><a name="zh-cn_topic_0000002039339945_p1468532510"></a>apex-0.1+ascend-cp3x-cp3x-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1869531256"><a name="zh-cn_topic_0000002039339945_p1869531256"></a>Yes</p>
<p id="zh-cn_topic_0000002039339945_p156353454"><a name="zh-cn_topic_0000002039339945_p156353454"></a>MindSpeed-LLM dependency</p>
<p id="zh-cn_topic_0000002039339945_p1861353651"><a name="zh-cn_topic_0000002039339945_p1861353651"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p166105316515"><a name="zh-cn_topic_0000002039339945_p166105316515"></a>Mixed precision training uses a mix of single-precision (float32) and half-precision (float16) data types during training, combining both and using the same hyperparameters to achieve almost the same precision as float32.</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a>The cp3x in the software package indicates the Python version number. For example, x being 10 indicates Python 3.10. The specific Python version is subject to the MindSpeed-LLM version.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a>For details, see the "<a href="https://gitcode.com/Ascend/apex/blob/master/docs/en/installing_apex.md">Installing the APEX Module</a>" section in <span id="zh-cn_topic_0000002039339945_ph156792413596"><a name="zh-cn_topic_0000002039339945_ph156792413596"></a>Ascend Extension for PyTorch Installation Guide</span>, and compile the APEX package based on the actual situation.</p>
<p id="zh-cn_topic_0000002039339945_p1761531257"><a name="zh-cn_topic_0000002039339945_p1761531257"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row197268213011"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1361953254"><a name="zh-cn_topic_0000002039339945_p1361953254"></a>torch_npu-2.7.1.<em id="zh-cn_topic_0000002039339945_i16204112111321"><a name="zh-cn_topic_0000002039339945_i16204112111321"></a>{version}</em>-cp3x-cp3x-manylinux_2_28_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p56185314510"><a name="zh-cn_topic_0000002039339945_p56185314510"></a>Yes</p>
<p id="zh-cn_topic_0000002039339945_p186653757"><a name="zh-cn_topic_0000002039339945_p186653757"></a>MindSpeed-LLM dependency</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p15720535517"><a name="zh-cn_topic_0000002039339945_p15720535517"></a>Ascend Extension for PyTorch is a deep learning adaptation framework based on Ascend, enabling Ascend NPUs to support the PyTorch framework, providing PyTorch users with the superior computing power of Ascend AI processors.</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a>The cp3x in the software package indicates the Python version number. For example, x being 10 indicates Python 3.10. The specific Python version is subject to the MindSpeed-LLM version.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p10718533510"><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a href="https://www.hiascend.com/document/detail/en/Pytorch/720/configandinstg/instg/insg_0004.html" target="_blank" rel="noopener noreferrer">Download Link</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1165115165020"><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p167047813263"><a name="zh-cn_topic_0000002039339945_p167047813263"></a>To use the <span id="zh-cn_topic_0000002039339945_ph1987542822613"><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a>PyTorch</span> model from the MindSpeed-LLM code repository, <span id="zh-cn_topic_0000002039339945_ph1412723132619"><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a>Ascend Extension for PyTorch</span> 2.6.0 or later is required.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row1412215399516"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><ul id="zh-cn_topic_0000002039339945_ul104867135415"><li><span id="zh-cn_topic_0000002039339945_ph0853174512272"><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a>x86_64</span> architecture: torch-2.7.1+cpu.cxx11.abi-cp3x-cp3x-linux_x86_64.whl</li><li><span id="zh-cn_topic_0000002039339945_ph7852164518272"><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a>ARM</span> architecture: torch-2.7.1+cpu-cp3x-cp3x-manylinux_2_28_aarch64.whl</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p871953357"><a name="zh-cn_topic_0000002039339945_p871953357"></a>Yes</p>
<p id="zh-cn_topic_0000002039339945_p17715531453"><a name="zh-cn_topic_0000002039339945_p17715531453"></a>MindSpeed-LLM dependency</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a>Official <span id="zh-cn_topic_0000002039339945_ph19355165113512"><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a>PyTorch</span> package.</p><p>The cp3x in the software package indicates the Python version number. For example, x being 10 indicates Python 3.10. The specific Python version is subject to the MindSpeed-LLM version.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p99745421447"><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a href="https://download.pytorch.org/whl/torch/" target="_blank" rel="noopener noreferrer">Download Link</a></p>
<p id="zh-cn_topic_0000002039339945_p483943610920"><a name="zh-cn_topic_0000002039339945_p483943610920"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row151232039750"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p4714531658"><a name="zh-cn_topic_0000002039339945_p4714531658"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1774531250"><a name="zh-cn_topic_0000002039339945_p1774531250"></a>Yes</p><p>For versions before CANN 8.5.0, this package is named Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5720531514"><a name="zh-cn_topic_0000002039339945_p5720531514"></a>CANN operator package.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p87153755"><a name="zh-cn_topic_0000002039339945_p87153755"></a><a href="https://www.hiascend.com/en/cann/download?versionId=791&ids=d806%2Ch0501%2Ch0601%2Ch0703" target="_blank" rel="noopener noreferrer">Download Link</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note13775154104217"><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p2075812144313"><a name="zh-cn_topic_0000002039339945_p2075812144313"></a>Obtain the software package that matches the server.</p>
</div></div>
</td>
</tr>
<tr id="row1173819266428"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="p7738142674217"><a name="p7738142674217"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p12738626184214"><a name="p12738626184214"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13738226114218"><a name="p13738226114218"></a>CANN Toolkit package.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p19271154916428"><a name="p19271154916428"></a><a href="https://www.hiascend.com/en/cann/download?versionId=791&ids=d806%2Ch0501%2Ch0601%2Ch0703" target="_blank" rel="noopener noreferrer">Download Link</a></p>
<div class="note" id="note3272104913427"><a name="note3272104913427"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p2272174920421"><a name="p2272174920421"></a>Obtain the software package that matches the server.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row121231639952"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1048218391768"><a name="zh-cn_topic_0000002039339945_p1048218391768"></a>MindSpeed</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p3482939367"><a name="zh-cn_topic_0000002039339945_p3482939367"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p048215393610"><a name="zh-cn_topic_0000002039339945_p048215393610"></a>MindSpeed is a large model acceleration library for Ascend devices.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p7482193913619"><a name="zh-cn_topic_0000002039339945_p7482193913619"></a>git clone https://gitcode.com/Ascend/MindSpeed.git</p>
<p id="zh-cn_topic_0000002039339945_p9482139663"><a name="zh-cn_topic_0000002039339945_p9482139663"></a>cd MindSpeed</p>
<p id="zh-cn_topic_0000002039339945_p1948213912618"><a name="zh-cn_topic_0000002039339945_p1948213912618"></a>git checkout e4d5855250f0074670f41c423021286502410bf1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row144125121466"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p848215391768"><a name="zh-cn_topic_0000002039339945_p848215391768"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1548319391465"><a name="zh-cn_topic_0000002039339945_p1548319391465"></a>Yes</p>
<p id="zh-cn_topic_0000002039339945_p16483239968"><a name="zh-cn_topic_0000002039339945_p16483239968"></a>Dependency file for installing CANN.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p14483939561"><a name="zh-cn_topic_0000002039339945_p14483939561"></a>Driver version information file.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p748314394617"><a name="zh-cn_topic_0000002039339945_p748314394617"></a>Copy /usr/local/Ascend/driver/version.info from the host.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row17301171913614"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p148310396616"><a name="zh-cn_topic_0000002039339945_p148310396616"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p348313399610"><a name="zh-cn_topic_0000002039339945_p348313399610"></a>Yes</p>
<p id="zh-cn_topic_0000002039339945_p1348313391961"><a name="zh-cn_topic_0000002039339945_p1348313391961"></a>Dependency file for installing CANN.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5483239564"><a name="zh-cn_topic_0000002039339945_p5483239564"></a>Driver installation information file.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p2483339861"><a name="zh-cn_topic_0000002039339945_p2483339861"></a>Copy /etc/ascend_install.info from the host.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row93022191368"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15485173911615"><a name="zh-cn_topic_0000002039339945_p15485173911615"></a>Dllogger code repository</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p248514391563"><a name="zh-cn_topic_0000002039339945_p248514391563"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164850391565"><a name="zh-cn_topic_0000002039339945_p164850391565"></a>PyTorch log tool.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1548619391164"><a name="zh-cn_topic_0000002039339945_p1548619391164"></a>git clone https://github.com/NVIDIA/dllogger.git</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row33025197610"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p0486133919618"><a name="zh-cn_topic_0000002039339945_p0486133919618"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1148619393610"><a name="zh-cn_topic_0000002039339945_p1148619393610"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164861439461"><a name="zh-cn_topic_0000002039339945_p164861439461"></a>Used to install the pip module.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p34865396611"><a name="zh-cn_topic_0000002039339945_p34865396611"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row03021191165"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15486163919616"><a name="zh-cn_topic_0000002039339945_p15486163919616"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1448693915619"><a name="zh-cn_topic_0000002039339945_p1448693915619"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p74862391165"><a name="zh-cn_topic_0000002039339945_p74862391165"></a>Used to image creation.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p9571628104613"><a name="zh-cn_topic_0000002039339945_p9571628104613"></a>-</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
> This section uses a single Atlas 800T A2, Ubuntu 20.04 Arm, and Python 3.10 as an example to describe how to build a training image. Modify the relevant steps based on the actual situation during use.

**Procedure<a name="zh-cn_topic_0000002039339945_section20489630477"></a>**

1. Prepare the software packages on the host by referring to [Table 1](#zh-cn_topic_0000002039339945_table1172542119019).
2. Write the following Dockerfile.

    ```text
    FROM ubuntu:20.04
    WORKDIR /root
    COPY . .

    ARG PYTORCH_WHL=torch-2.7.1+cpu-cp310-cp310-manylinux_2_28_aarch64.whl
    ARG PYTORCH_NPU_WHL=torch_npu-2.7.1.{version}-cp310-cp310-manylinux_2_28_aarch64.whl
    ARG APEX_WHL=apex-0.1+ascend-cp310-cp310-linux_aarch64.whl
    ARG HOST_ASCEND_BASE=/usr/local/Ascend
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann
    # The CANN version used in the example is 8.5.0. Modify it based on the actual situation during use.
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

    # System packages
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
    # Set up Python symbolic links
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
    # System packages
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
    # Time zone
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
    # Configure the pip source
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
    # Install torch, TorchNPU, and apex packages
    RUN umask 0022 && pip install $PYTORCH_WHL && \
        pip install $PYTORCH_NPU_WHL && \
        pip install $APEX_WHL

    # Ascend package
    # Before building, copy /usr/local/Ascend/driver/version.info from the host to the current directory
    RUN umask 0022 &&  \
        cp ascend_install.info /etc/ && \
        mkdir -p /usr/local/Ascend/driver/ && \
        cp version.info /usr/local/Ascend/driver/ && \
        chmod +x $TOOLKIT && \
        chmod +x $OPS

    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc
    RUN umask 0022 && ./$OPS --install --quiet

    # After the toolkit package is installed, clear the following files. During container startup, the toolkit package is mounted by Ascend Docker Runtime.
    RUN rm -f version.info && rm -f ascend_install.info && \
        rm -rf /usr/local/Ascend/driver/

    RUN umask 0022 && cd $MINDSPEED && \
        pip install -r requirements.txt && \
        pip install -e . && \
        echo "export PYTHONPATH=/root/MindSpeed:\$PYTHONPATH" >> ~/.bashrc

    RUN umask 0022 && cd $DLLOGGER && \
        python setup.py build && \
        python setup.py install

    # Import environment variables
    ENV HCCL_WHITELIST_DISABLE=1

    # Create /lib64/ld-linux-aarch64.so.1
    RUN umask 0022 && \
        if [ ! -d "/lib64" ]; \
        then \
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
        fi

    # MindCluster resumable training adaptation script.
    RUN umask 0022 && \
        pip install $TASKD_WHL && \
        pip install $MINDIO_TTP_WHL

    # Optional. The following commands must be configured when using graceful fault tolerance, pod-level rescheduling, or process-level rescheduling.
    RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

    # Add the dependency library for job scheduling installation
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
    # Pack the preceding content into the image mindspeed-dl:v1.
    ```

    >[!NOTE]
    >If Python 3.10 cannot be installed successfully through PPA, or the deadsnakes PPA does not provide an image source for Python 3.10, you can download the source code and manually compile and install it.

3. Build the image. Run the following command to generate the image. To make the Dockerfile more secure, you can define a `HEALTHCHECK` in it based on your service. Check the container running status by running the **HEALTHCHECK** _\[OPTIONS\]_ **CMD** command inside the container. **Do not omit the dot (".") at the end of the command.**

    ```shell
    docker build -t mindspeed-dl:v1 .
    ```

4. In addition to manually building an image by following the preceding process, you can also directly use the out-of-the-box sample image `torch-npu-sample-dl:v26.2.0-cann8.5.0-torch_npu2.7.1-910b-ubuntu22.04-py3.10-aarch64`. The sample image can be obtained from [AscendHub](https://www.hiascend.com/en/developer/ascendhub) for quick verification.

## Script Adaptation<a name="ZH-CN_TOPIC_0000002511426481"></a>

### Process Description<a name="ZH-CN_TOPIC_0000002511346469"></a>

Model scripts must be adapted for CKPT before the resumable training function can be used. The process and logic of script adaptation are shown in [Figure 1](#fig88341718121515).

**Figure 1** Script adaptation process<a name="fig88341718121515"></a>

![](../../../../../figures/scheduling/script-adaptation-process.png)

### Adaptation Example<a name="ZH-CN_TOPIC_0000002511346445"></a>

This chapter guides users through the adaptation steps for resumable training, step by step.

>[!NOTE]
>
>- To ensure that graceful fault tolerance and process-level online recovery work properly, keep the clocks of the master and worker nodes of the K8s cluster consistent.
>- The component code shown for resumable training is open-source code. For related security statements, see [Security Statement](../../../../07_references/05_appendix.md#ZH-CN_TOPIC_0000002479386374).
>- The model example code below may differ from the actual version. The actual version code prevails.
>- Write the model parameter configuration based on the model configuration in the model repository and the actual situation. Improper modification may cause unpredictable problems.
>- If the error "Failed to bind the IP port. Reason: The IP address and port have been bound already" occurs during training, configure it as follows. For details, see the "[HCCL_HOST_SOCKET_PORT_RANGE](https://www.hiascend.com/document/detail/en/CANNCommunityEdition/910/commlib/hcclug/docs/en/user_guide/hccl_env/HCCL_HOST_SOCKET_PORT_RANGE.md)" section.
>
>   ```shell
>   export HCCL_HOST_SOCKET_PORT_RANGE="60000-60050"
>   export HCCL_NPU_SOCKET_PORT_RANGE="61000-61050"
>   ```
>
>- If TaskD is used and the training container uses the host network, first query the current reserved port configuration by running `sysctl net.ipv4.ip_local_reserved_ports`, and then add reserved ports 9601 and 9602 by running `sysctl -w net.ipv4.ip_local_reserved_ports="xxx,9601,9602"` (where *xxx* refers to the ports already configured in the preceding query; omit it if there are none).

For training code and dataset preparation, see [MindSpeed-LLM User Guide](https://gitcode.com/Ascend/MindSpeed-LLM/blob/26.2.0/docs/en/pytorch/training/pretrain/mcore/pretrain.md). The following uses two Atlas 800T A2 as an example to describe the specific operation steps.

1. Pull the training code.

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
    cp -r megatron ../MindSpeed-LLM # Copy the Megatron directory under the Megatron-LM project to the MindSpeed-LLM project.
    ## Rename MindSpeed-LLM to QWEN3_for_PyTorch_2.7_code.
    cd ..
    mv MindSpeed-LLM QWEN3_for_PyTorch_2.7_code
    ```

2. Obtain model weights.

    Download model weights from [Qwen3](https://huggingface.co/Qwen/Qwen3-8B/tree/main) and place them in a directory on the server, such as `/data/atlas_dls/public/dataset/qwen3-8b-hf`.

3. Prepare a dataset.

    Download the detailed dataset, for example, [Alpaca](https://huggingface.co/datasets/tatsu-lab/alpaca/blob/main/data/train-00000-of-00001-a09b74b3ef9c3b56.parquet), and place it in a directory on the server, such as `/data/atlas_dls/public/dataset/qwen3-alpaca`.

4. Process the dataset.
    1. Start the container.

        ```shell
        docker run -it -v /data/atlas_dls/public/:/data/atlas_dls/public/ -e ASCEND_VISIBLE_DEVICES=0-7 mindspeed-dl:v1 bash
        ```

    2. Perform the following operations in the container.

        ```shell
        export TORCH_DEVICE_BACKEND_AUTOLOAD=0
        source /usr/local/Ascend/cann/set_env.sh
        cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code
        # Optional. The following describes how to install the MindSpeed acceleration library. This operation can be performed in any directory. If it has been installed during image creation, skip this operation.
        git clone https://gitcode.com/ascend/MindSpeed.git
        cd MindSpeed
        git checkout e4d5855250f0074670f41c423021286502410bf1
        pip install -r requirements.txt
        pip install -e .
        export PYTHONPATH=/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/MindSpeed:$PYTHONPATH
        cd ..
        ```

    3. Process the dataset.

        Qwen3 requires that Transformers version be 4.51.0 or later. Therefore, Python 3.9 or later and Transformers 4.51.0 or later must be installed.

        ```shell
        python preprocess_data.py \
            --input /data/atlas_dls/public/dataset/qwen3-alpaca/train-00000-of-00001-a09b74b3ef9c3b56.parquet \ # Dataset file path
            --tokenizer-name-or-path /data/atlas_dls/public/dataset/qwen3-8b-hf \ # Open-source model weight file directory
            --tokenizer-type PretrainedFromHF \
            --handler-name GeneralPretrainHandler \
            --output-prefix /data/atlas_dls/public/dataset/qwen3-alpaca/alpaca \ # Generates alpaca_text_document.bin and .idx files
            --json-keys text \
            --workers 4 \
            --log-interval 1000
        ```

        >[!NOTE]
        >If the error "/usr/local/lib/python3.10/dist-packages/sklearn/utils/../../scikit_learn.libs/libgomp-947d5fa1.so.1.0.0: cannot allocate memory in static TLS block" is displayed, run the following command to preload the libgomp library.
        >
        >```shell
        >export LD_PRELOAD=/usr/local/lib/python3.10/dist-packages/scikit_learn.libs/libgomp-947d5fa1.so.1.0.0:$LD_PRELOAD
        >```

5. Go to the "[mindcluster-deploy](https://gitcode.com/Ascend/mindcluster-deploy)" repository, switch to the branch corresponding to the version according to the [mindcluster-deploy open-source repository version description](../../../../07_references/05_appendix.md#mindcluster-deploy-open-source-repository-version-description), obtain the `train_start.sh` file from the `samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3` directory, and construct the following directory structure on the management node.

    ```text
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

6. Obtain the [training job YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml). This YAML has already configured pod-level rescheduling, process-level rescheduling, process-level online recovery, elastic training, and more. Set the IP address of the server to which the volume is mounted and rescheduling levels as required.

    Training process recovery functions such as process-level rescheduling, process-level online recovery, and elastic training cannot coexist with graceful fault tolerance. For details about how to configure graceful fault tolerance, see [Graceful Fault Tolerance Mode](./menu_examples_and_verification.md#ZH-CN_TOPIC_0000002511346449).

7. Configure `train_start.sh` and the training job YAML based on the actual situation.
    1. Modify the basic parameters of the startup script.

        ```shell
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/ttplogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/trainlogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/demo/
        # Log save path, which can be modified based on the actual situation
        export ASCEND_PROCESS_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP       # Set the plog save path, where $MINDX_TASK_ID is the task UID environment variable injected by Ascend Operator, and $XDL_IP is the environment variable status.hostIP written in the YAML
        export TTP_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/ttplogs/ttplog$XDL_IP-$RANK    # Set the TTP log save path, where $RANK is the environment variable injected by Ascend Operator for the PyTorch framework
        export TRAIN_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/trainlogs/$XDL_IP-$RANK      # Set the training log save path
        export GLOO_SOCKET_IFNAME=enp189s0f0               # The network port on the physical machine that can communicate. Configure it based on the actual high-speed NIC of the master node. If hostNetwork is set to false in the YAML, set it to eth0.
        export HCCL_SOCKET_IFNAME=enp189s0f0               # If hostNetwork is set to false in the YAML, set it to eth0.

        CKPT_SAVE_DIR="/job/code/output/ckpt" # The weight save path after training is completed
        DATA_PATH="/job/data/alpaca_text_document" # Dataset path. Enter the data path saved during data preprocessing.
        TOKENIZER_PATH="/job/data/qwen3-8b-hf" # Tokenizer path. Enter the downloaded open-source weight tokenizer path
        CKPT_LOAD_DIR="/job/code/output/ckpt" # Weight loading path
        ```

    2. To use TaskD for process-level rescheduling, process-level online recovery, process-level in-place recovery, or elastic training, you also need to start TaskD Manager.
        1. Create a `manager.py` file in the current directory when calling the training script. The content of the `manager.py` file is as follows.

            ```python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # Total number of nodes
            proc_per_node=XX     # Number of training processes per node

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >For detailed parameter descriptions in the manager.py file, see [def init\_taskd\_manager\(config:dict\) -\> bool:](../../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            In the following code, the two statements `TASKD_SO_PATH` and `export LD_PRELOAD` are used to configure the path of `libtaskd.so` after installing TaskD into the environment variable `LD_PRELOAD`. If these two statements fail to configure successfully, you can manually run the `pip show taskd` command to obtain the value of Location, concatenate it with `/taskd/python/cython_api/libs/libtaskd.so`, and then set it via `export`.

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path, and the error.log path must be created in advance.
            fi

            torchrun $DISTRIBUTED_ARGS ...
            ```

        3. Modify the training job YAML to add container ports, and add port 9601 used for TaskD communication under all Pods (skip this step if it already exists).

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

## Preparing a Job YAML<a name="ZH-CN_TOPIC_0000002511426415"></a>

Cluster scheduling components provide YAML examples. Select the corresponding YAML example based on the functions, model type, and job type in use as well as the fault handling mode, and modify it as required before use.

**Table 2** Training job YAML example

<a name="table350244433714"></a>
<table><thead align="left"><tr id="row135031644183710"><th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.1"><p id="p8503244173715"><a name="p8503244173715"></a>Job Type</p>
</th>
<th class="cellrowborder" valign="top" width="16.173234646929384%" id="mcps1.2.8.1.2"><p id="p145038448375"><a name="p145038448375"></a>Hardware Model</p>
</th>
<th class="cellrowborder" valign="top" width="8.521704340868173%" id="mcps1.2.8.1.3"><p id="p919210345266"><a name="p919210345266"></a>Training Framework</p>
</th>
<th class="cellrowborder" valign="top" width="13.672734546909378%" id="mcps1.2.8.1.4"><p id="p5503544193713"><a name="p5503544193713"></a>Model</p>
</th>
<th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.5"><p id="p19672186404"><a name="p19672186404"></a>YAML File Name</p>
</th>
<th class="cellrowborder" valign="top" width="15.433086617323463%" id="mcps1.2.8.1.6"><p id="p1096741894013"><a name="p1096741894013"></a>Download Link</p>
</th>
<th class="cellrowborder" valign="top" width="15.413082616523303%" id="mcps1.2.8.1.7"><p id="p2967518174012"><a name="p2967518174012"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row4503174412371"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p09365292408"><a name="p09365292408"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul129364297402"></a><ul id="ul129364297402"><li><span id="ph157633217501"><a name="ph157633217501"></a>Atlas 800T A2</span></li><li>Atlas 900 A2 PoD</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p319343422611"><a name="p319343422611"></a><span id="ph310231710274"><a name="ph310231710274"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p493616294406"><a name="p493616294406"></a><span id="ph22631282914"><a name="ph22631282914"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p893610293406"><a name="p893610293406"></a>pytorch_multinodes_acjob_910b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1987716427402"><a name="p1987716427402"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/release_v26.2.0/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml" target="_blank" rel="noopener noreferrer">pytorch_multinodes_acjob_910b.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p8936152964011"><a name="p8936152964011"></a>A two-server eight-chip job is presented in the example file by default.</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>Currently, some training frameworks do not provide resumable training example YAMLs for Atlas 900 A3 SuperPoD. You can add an annotations field under labels in the example YAML. An example is shown below:
>
>```yaml
>...
>  labels:
>...
>  annotations:
>    sp-block: "32"   # Number of chips in a logical super node. For details about the sp-block field, see the YAML parameter description.
>    huawei.com/schedule_policy: "chip2-node16-sp"    # Set the scheduling policy based on the hardware form
>...
>```

## Delivering a Job<a name="ZH-CN_TOPIC_0000002479226548"></a>

In the example YAML, the job is deployed in the default namespace. This chapter uses the PyTorch framework as an example to deliver a training job.

1. Log in to the management node and go to the path where the YAML file is stored.
2. Run the following command on the management node to deliver the training job using the YAML.

    ```shell
    kubectl apply -f XXX.yaml
    ```

    For example:

    ```shell
    kubectl apply -f pytorch_multinodes_acjob_910b.yaml
    ```

    Command output:

    ```ColdFusion
    configmap/reset-config-default-test-pytorch created
    ascendjob.mindxdl.gitee.com/default-test-pytorch created
    ```

## Viewing Job Processes<a name="ZH-CN_TOPIC_0000002511426461"></a>

After the training job is delivered successfully, it runs normally. You can check the running status of the training job as follows.

**Viewing All Training Jobs<a name="section16792164211375"></a>**

To view all training jobs running on the current node, perform the following steps.

1. Log in to the management node and go to the path where the YAML file is stored.
2. Run the following command to view the running status of the training job.

    ```shell
    kubectl get pods -A -o wide
    ```

    Command output:

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node1          <none>           <none>
    default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
    ……
    ```

**Viewing the Training Job of a Single Pod<a name="zh-cn_topic_0000001621551937_section1141119143319"></a>**

To view the training job running on one Pod, perform the following steps.

Run the following command to view the running status of the training job.

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

The following is an output example. The presence of `loss` indicates that the job is running properly.

![](../../../../../figures/scheduling/unnaming-71.png)

**Checking Whether CKPT Files Exist<a name="section979416428371"></a>**

The fault recovery function is implemented by referencing CKPT files, so you need to check whether CKPT files exist on the storage node.

After the training job has been running for longer than the CKPT saving interval you configured, check whether periodic CKPT files exist in the configured CKPT saving path. The steps are as follows.

1. Log in to the storage node and run the following command to go to the CKPT file path.

    ```shell
    cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/output/ckpt
    ```

2. Run the following command to check whether a periodic CKPT file exists in the current directory.

    ```shell
    ll ./
    ```

    The following is an output example, which indicates that a periodic CKPT file exists.

    ```ColdFusion
    total 8
    drwxr-xr-x  18 root root   8192 Jun 22 18:39 iter_0000100
    -rw-r--r--  1 root root    2    Jun 22 18:39 latest_checkpointed_iteration.txt
    ```

3. (Optional) If the dying gasp feature is used, run the following command in the CKPT saving path to check whether a dying gasp CKPT file exists in the current directory.

    ```shell
    ll ./
    ```

    The following is an output example, which indicates that a dying gasp CKPT file exists.

    ```ColdFusion
    total 8
    drwxr-xr-x  18 root root   8192 Jun 22 15:39 iter_0000009
    -rw-r--r--  1 root root    2    Jun 22 15:39 latest_checkpointed_iteration.txt
    ```

## Viewing Training Results<a name="ZH-CN_TOPIC_0000002479386554"></a>

### (Optional) Constructing Faults<a name="ZH-CN_TOPIC_0000002511426449"></a>

This chapter guides users through constructing simple faults, including node faults, parameter plane network faults, and service plane faults.

>[!NOTE]
>Constructing a chip fault involves security risks. If you need to construct one, contact Huawei technical support engineers.

**Constructing a Node Fault<a name="section173881558133914"></a>**

Simulate node status loss caused by node power-off by restarting the training node. This fault can be automatically recovered after the node restart is complete.

1. After the training job has trained for some iterations normally, log in to the node where training is running.
2. Run the following command to restart the training node to simulate a node status loss fault.

    ```shell
    reboot
    ```

3. Run the following command on the master node multiple times to view the Pod status.

    ```shell
    kubectl get pod -A
    ```

    You can see that the Pod status changes from `Terminating` to `Pending` and finally to `Running`, indicating that the training job has been restarted.

4. Run the following command on the master node to view the training logs and record the time when resumable training succeeds.

    ```shell
    kubectl logs -n <namespace_name> <pod_name>
    ```

    The following is an output example, which indicates that when the fault occurs, training is recovered from the most recently saved checkpoint at step 9, so that the training job continues from the 10th iteration.

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

**Constructing a Parameter Plane Network Fault<a name="section22113033919"></a>**

Simulate a parameter plane network fault by disconnecting the NPU network link. An NPU network fault does not affect single-node training jobs. You must manually recover the link after disconnecting it; otherwise, the fault persists.

1. After the training job has trained for some iterations normally, log in to the node where training is running.
2. Run the following command to construct an NPU network link fault.

    ```shell
    hccn_tool -i {device_id} -link -s down
    ```

    >[!NOTE]
    >`device_id` is the NPU ID. You can view the NPU ID by running the <b>npu-smi info</b> command.

3. Run the following command to view the NPU link status.

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    The following is an output example, which indicates that the NPU network link fault is constructed successfully.

    ```ColdFusion
    net health status: Fault
    ```

4. Run the following command on the master node multiple times to view the Pod status.

    ```shell
    kubectl get pod -A
    ```

    You can see that the Pod status changes from `Terminating` to `Pending` and finally to `Running`, indicating that the training job has been restarted.

5. Run the following command on the master node to view the training logs and record the time when resumable training succeeds.

    ```shell
    kubectl logs -n <namespace_name> <pod_name>
    ```

    The following is an output example, which indicates that when the fault occurs, training is recovered from the most recently saved checkpoint at step 9, so that the training job continues from the 10th iteration.

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

6. Run the following command to recover the NPU network link fault.

    ```shell
    hccn_tool -i {device_id} -cfg recovery
    ```

7. Run the following command to view the NPU link status.

    ```shell
    hccn_tool -i {device_id} -net_health -g
    ```

    The following is an output example, which indicates that the NPU network link fault has been recovered.

    ```ColdFusion
    net health status: Success
    ```

**Constructing a Service Plane Fault<a name="section9891038124213"></a>**

Simulate a service plane fault by deleting the training process.

1. After the training job has trained for some iterations normally, log in to the node where training is running.
2. Run the following command to query the training process information using the training startup script.

    ```shell
    ps -ef | grep python| grep <training_startup_script>.py
    ```

3. Run the following command to manually delete the training process with the smallest PID.

    ```shell
    kill -9 pid
    ```

4. Run the following command on the master node multiple times to view the Pod status.

    ```shell
    kubectl get pod -A
    ```

    You can see that the Pod status changes from `Terminating` to `Pending` and finally to `Running`, indicating that the training job has been restarted.

5. Run the following command on the master node to view the training logs and record the time when resumable training succeeds.

    ```shell
    kubectl logs -n <namespace_name> <pod_name>
    ```

    The following is an output example, which indicates that when the fault occurs, training is recovered from the most recently saved checkpoint at step 9, so that the training job continues from the 10th iteration.

    ```ColdFusion
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

### Rescheduling Mode<a name="ZH-CN_TOPIC_0000002479386534"></a>

**Rescheduling Status<a name="section87441013105513"></a>**

>[!NOTE]
>When a node becomes faulty, Volcano schedules the training job to another node that meets the conditions to continue running.

Log in to the management node and run the following command to view the running status of the training job.

```shell
kubectl get pods -A -o wide
```

Before the fault, if the training job is scheduled to node1 and node2, and a fault occurs on node1, Volcano reschedules the training job on node1 and node2 to node2 and node3. The output after rescheduling is as follows.

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node3          <none>           <none>
……
```

**Checking the Job Running Status on a Single Pod<a name="section28985295314"></a>**

Run the following command to check the running status of the training job on a single pod:

```shell
kubectl logs default-test-pytorch-worker-0 -n default -f
```

The sample output below shows that the latest checkpoint file from the ninth iteration is used to resume training at the tenth iteration after a fault occurs.

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

**Viewing Job Rescheduling Records<a name="section97707231547"></a>**

Run the following command to view the job rescheduling records.

```shell
kubectl describe cm -n mindx-dl job-reschedule-reason
```

    Command output:

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

### Graceful Fault Tolerance Mode (Sunset)<a name="ZH-CN_TOPIC_0000002511346479"></a>

This chapter guides users through viewing training information when the graceful fault tolerance mode of fault handling is used. When a chip becomes faulty, the process exits and graceful fault tolerance is performed, and the process is restarted after recovery.

**Log Description<a name="section83075820188"></a>**

The training logs of the restarted training process are stored in "_training_script_path_/newlog". The details are as follows.

- QWEN3 (PyTorch) training logs: `/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/alllogs`

**Procedure<a name="section25042117188"></a>**

1. Log in to the management node and run the following command to view the chip status.

    ```shell
    npu-smi info
    ```

    The following is an output example, which indicates that the training process occupies on-chip memory and training is running normally.

    ![](../../../../../figures/scheduling/1-13.png)

2. After the fault occurs, run the following command to view the chip information.

    ```shell
    npu-smi info
    ```

    The following is an output example, which indicates that the training process has exited and the on-chip memory is released.

    ![](../../../../../figures/scheduling/2.png)

3. After the fault is recovered, run the following command to view the chip information.

    ```shell
    npu-smi info
    ```

    The following is an output example, which indicates that the training process has been restarted and occupies the on-chip memory, and training is running normally.

    ![](../../../../../figures/scheduling/3.png)

## Constructing Faults and Verifying Fault Handling

This document provides verification methods for various fault handling policies. You can quickly jump to the corresponding verification section based on the fault handling policy actually configured.

**Quick Navigation**

- [Verifying Job-Level Rescheduling](#verifying-job-level-rescheduling)
- [Verifying Pod-Level Rescheduling](#verifying-pod-level-rescheduling)
- [Verifying Process-Level Rescheduling](#verifying-process-level-rescheduling)
- [Verifying Process-Level Online Recovery](#verifying-process-level-online-recovery)

### Verifying Job-Level Rescheduling

**Test Preparation**

In the job YAML for basic scheduling, add the configuration for job-level rescheduling. For the configuration description, see [Configuring Job-Level Rescheduling](../02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002479226580); for the principle, see [Job-Level Rescheduling](../01_solutions_principles/01_fault_handling.md#ZH-CN_TOPIC_0000002479226586).

**Test Operations**

1. Deliver the job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The Pod name and namespace vary with the configuration in the job YAML. The `taskmgr-npu-020-default-test-` and `trjob` shown below are example values, and the actual values vary with the configuration in the job YAML.

2. Check the job status and UID.

   1. Check the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The following is an output example. `Running` in the `STATUS` field indicates that the job is running properly.

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>    0          2s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>    0          3s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. Check the UIDs of the two Pods.

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      Command output:

      ```bash
      7286faf8-f029-450a-b302-5e6e94d4346c
      997add9e-6115-456c-9e8e-e05e4b70bb12
      ```

3. Construct a fault.

   1. Query the job process.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      Command output:

      ```ColdFusion
      2398104
      2398105
      2398107
      ```

   2. Terminate the process to simulate a fault.

      ```bash
      kill -9 2398104
      ```

4. Observe the rescheduling process.

   Monitor the status changes of the two Pods of the job.

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   The historical statuses of the two Pods of the job are as follows. By observing the changes of the bold fields, you can see that the two Pods go through `Terminating`, `Pending`, `ContainerCreating`, and `Running` and then run normally, indicating that job rescheduling is successful:

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running             0          3s      xx.xx.xx.xx       localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== Inject a fault ======================
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

**Expected Result**

1. Check the job status and UID.

   1. Check the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The following is an output example. `Running` in the `STATUS` field indicates that the job is running properly.

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          2s      xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          33s     xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. Check the UIDs of the two Pods:

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The following is an output example. The UIDs of both Pods of the job have changed, indicating that both Pods have been rescheduled, that is, job-level rescheduling is triggered:

      ```ColdFusion
      2a24eee8-88f1-4107-bc9d-dabcfb09dea9
      074f9f9c-35f1-4b9e-9298-5b2bcf3759e7
      ```

### Verifying Pod-Level Rescheduling

**Test Preparation**

In the job YAML for basic scheduling, add the configuration for Pod-level rescheduling. For the configuration description, see [Configuring Pod-Level Rescheduling](../02_configuration/01_configuring_fault_handling_policies.md#configuring-pod-level-rescheduling); for the principle, see [Pod-Level Rescheduling](../01_solutions_principles/01_fault_handling.md#ZH-CN_TOPIC_0000002511346429).

**Test Operations**

1. Deliver the job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The Pod name and namespace vary with the configuration in the job YAML. The `taskmgr-npu-020-default-test-` and `trjob` shown below are example values, and the actual values vary with the configuration in the job YAML.

2. Check the job status and UID.

   1. Check the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The following is an output example. The presence of `Running` indicates that the job is running properly.

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. Check the UIDs of the two Pods.

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      Command output:

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      47291595-85b0-47ff-8393-c922d0e2dfb2
      ```

3. Construct a fault.

   1. Query the job process.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      Command output:

      ```ColdFusion
      2398132
      2398144
      2398158
      ```

   2. Terminate the process to simulate a fault.

      ```bash
      kill -9 2398144
      ```

4. Observe the rescheduling process.

   Monitor the status changes of the two Pods of the job.

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   The historical statuses of the two Pods of the job are as follows. By observing the changes of the bold fields, you can see that the faulty Pod (`taskmgr-npu-020-default-test-1`) goes through `Error`, `Terminating`, `Pending`, `ContainerCreating`, and `Running` and then runs normally, indicating that Pod rescheduling is successful:

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running              0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running              0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== Inject a fault ======================
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

**Expected Result**

1. Check the job status and UID.

   1. Check the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The following is an output example. The presence of `Running` indicates that the job is running properly.

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          66s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          31s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. Check the UIDs of the two Pods again.

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The following is an output example. The UID of the `taskmgr-npu-020-default-test-0` Pod has not changed, while the UID of the `taskmgr-npu-020-default-test-1` Pod has changed, indicating that only the faulty Pod (`taskmgr-npu-020-default-test-1`) has been rescheduled, that is, Pod-level rescheduling is triggered:

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      6eb3c217-3b63-457a-9010-9d236d281634
      ```

### Verifying Process-Level Rescheduling

**Test Preparation**

In the job YAML for basic scheduling, add the configuration for process-level rescheduling. For the configuration description, see [Configuring Process-Level Rescheduling](../02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426407); for the principle, see [Process-Level Rescheduling](../01_solutions_principles/01_fault_handling.md#ZH-CN_TOPIC_0000002511346457).

**Test Operations**

1. Deliver the job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The Pod name and namespace vary with the configuration in the job YAML. The `process-reschedule-function-` and `trjob` shown below are example values, and the actual values vary with the configuration in the job YAML.

2. Check the job status.

   ```bash
   kubectl get pod -A -o wide
   ```

   The following is an output example. The presence of `Running` indicates that the job is running properly:
   <pre codetype="ColdFusion">
   trjob            process-reschedule-function-master-0   1/1     Running   0               14s   xx.xx.xx.xx     master-69-117   &lt;none&gt;           &lt;none&gt;
   trjob            process-reschedule-function-worker-0   1/1     Running   0               14s   xx.xx.xx.xx     work-69-115     &lt;none&gt;           &lt;none&gt;
   </pre>

3. Check the number of training log iterations to confirm that training iterates normally.

   ```bash
   kubectl logs -n trjob process-reschedule-function-worker-0|grep -Po '] iteration [[:space:]]*'|wc -l
   ```

      Command output:

   ```bash
   50
   ```

4. Check the process ID and construct a fault.

   1. Check the process ID.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      Command output:

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

   2. Terminate one of the processes to simulate a fault.

      ```bash
      kill -9 635777
      ```

**Expected Result**

1. Observe the training logs.

   ```bash
   kubectl logs -n trjob process-reschedule-function-master-0
   ```

      Command output:

   ```ColdFusion
   # The following information indicates that the ARF process starts
   Mindx calling notify do ARF repair

   # The following information indicates that ARF succeeds
   ... Mindio do repair operation ok ...
   ```

2. Check whether job information exists in the ConfigMap `job-reschedule-reason`.

   ```bash
   kubectl describe cm -n mindx-dl job-reschedule-reason |grep process-reschedule-function
   ```

   The following is an output example, which contains information such as the rescheduling time, the pod, node, and rank that trigger rescheduling, and the current number of rescheduling attempts of this job:

   ```ColdFusion
   {"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07":{"JobID":"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0603 05:16:52","RescheduleTimeStamp":1780435012,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"process-reschedule-function-worker-0","NodeName":"work-69-115","NodeRankIndex":"1"}]}]}}
   ```

### Verifying Process-Level Online Recovery

This chapter guides users through the adaptation steps for verifying process-level online recovery by instrumenting the training code to construct an on-chip memory UCE fault.

>[!NOTE]
>
>- The modifications in this chapter are only used to guide users in verifying the process-level online recovery function in a test environment. Do not release this instrumented version to the production environment.
>- Before performing the steps in this chapter, ensure that training can be started properly and that process-level online recovery has been configured.
>- To ensure that process-level online recovery works properly, keep the clocks of the master and worker nodes of the K8s cluster consistent.
>- The code below may differ from the actual version. The actual version code prevails.

#### MindCluster Adaptation<a name="ZH-CN_TOPIC_0000002479386410"></a>

1. <a name="li977718409381"></a>Clone the MindCluster code.

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/mind-cluster.git
    cd ./mind-cluster/component/clusterd
    git checkout release/v26.2.0   # release/v26.2.0 is the version branch of the code repository; switch to the target branch yourself
    ```

2. Modify the ClusterD code.
   1. Open the "pkg/application/faultmanager/jobprocess/faultrank/job\_fault\_rank\_processor.go" file.

      ```shell
      vi pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go
      ```

   2. Press `i` to enter the edit mode and add the following bold code.

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
               <strong>collector.ReportInfoCollector.ReportRetryInfo(podInfo.jobId, deviceInfo.RankID, constant.JobNotRecover, constant.UceFaultType)   // Set the service plane fault time to an invalid time to prevent a single fault from repeatedly triggering process-level online recovery</strong>
            }
        …
      </pre>

   3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit the editor.

3. <a name="li114977117517"></a>Build ClusterD.

   ```shell
   cd ./build/
   chmod +x build.sh && dos2unix build.sh
   sed -i 's|build_version="v[^"]\+"|build_version="xxx"|g' build.sh  # Replace xxx with the version number, for example, v26.2.0
   sed -i 's|export CGO_ENABLED=0|export CGO_ENABLED=1|g' build.sh  # Enable the CGO function
   ./build.sh # Build ClusterD. The Go SDK must be installed in advance; for the specific version, see the go.mod file of the ClusterD component code
   ```

   After the build succeeds, related files are generated in the "../output/" directory. Run the following command to view them:

   ```shell
   ll ../output/
   ```

    Command output:

   ```bash
   -r-x------. 1 root root 45891128 Aug 13 10:52 clusterd
   -r--------. 1 root root     4021 Aug 13 10:52 clusterd-v26.2.0.yaml
   -r--------. 1 root root      946 Aug 13 10:52 Dockerfile
   -r--------. 1 root root      209 Aug 13 10:52 faultDuration.json
   -r--------. 1 root root      207 Aug 13 10:52 fdConfig.yaml
   -r--------. 1 root root      467 Aug 13 10:52 publicFaultConfiguration.json
   -r--------. 1 root root      756 Aug 13 10:52 relationFaultCustomization.json
   ```

4. <a name="li89701053589"></a>Go to the output directory and build the ClusterD image.

   ```shell
   cd ../output/
   docker build --no-cache -t clusterd:{tag} ./  # {tag} must be the same as the value of build_version="xxx" in step 3
   ```

5. (Optional) Save the image and upload the saved image file and the `clusterd-{tag}.yaml` file to the master node. If [step 1](#li977718409381) to [step 4](#li89701053589) are performed on the master node, skip this step.

   ```shell
   docker save -o clusterd.tar clusterd:{tag}  # Save the image
   docker load -i clusterd.tar  # Import the image on the master node
   ```

6. Restart ClusterD on the master node.

   ```shell
   kubectl delete -f  clusterd-{tag}.yaml  # Delete the old ClusterD container
   kubectl apply -f  clusterd-{tag}.yaml  # Start the new container
   ```

#### Script Adaptation<a name="ZH-CN_TOPIC_0000002479226412"></a>

##### PyTorch Scenario Adaptation Example (Based on MindSpeed-LLM)<a name="ZH-CN_TOPIC_0000002511426361"></a>

1. Set up the training environment and start training. For details, see [PyTorch Scenario Adaptation Example (Based on MindSpeed-LLM)](#ZH-CN_TOPIC_0000002511346445).
2. Enable process-level online recovery. For details, see [Configuring Process-Level Online Recovery](../02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002479386492).
3. Add the following bold content to the "QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py" code to instrument and inject a fault. The new code obtains the fault injection iteration position and faulty rank information based on the environment variable "RAISE\_UCE\_ERROR\_STEP\_AND\_RANK".

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

4. Modify the startup script "QWEN3\_for\_PyTorch\_2.7\_code/scripts/train\_start.sh".

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # Configure the iterations and card numbers for fault injection; inject a UCE fault on rank 8 at the 3rd iteration and on rank 9 at the 10th iteration
   sed -i 's/check_memory_result = torch_npu.npu.check_uce_in_memory(device)/check_memory_result = ha_constant.UCE_HIGH_LEVEL/g' /job/code/mindspeed_llm/core/high_availability/tft_stop_clean.py # Modify the return value of the TorchNPU API to identify the exception thrown by the training code as a UCE fault
   …
   ```

#### Verification Process

The following example is based on a **dual-node 16-rank** environment (8 ranks per node, ranks 0 to 7 on master and ranks 8 to 15 on worker), consistent with the configuration in [Script Adaptation](#ZH-CN_TOPIC_0000002479226412) `RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"`. If you use a single node or another topology, adjust the environment variables and the rank and Pod name in `grep` accordingly.

**Test Preparation**

- In the job YAML for basic scheduling, add the configuration for process-level online recovery. For the configuration description, see [Configuring Process-Level Online Recovery](../02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002479386492); for the principle, see [Process-Level Online Recovery](../01_solutions_principles/01_fault_handling.md#ZH-CN_TOPIC_0000002479386460).
- MindCluster adaptation and script adaptation have been completed; the `RAISE_UCE_ERROR_STEP_AND_RANK` in the startup script is consistent with the rank and iteration steps in the verification commands below.

**Test Operations**

1. Deliver the job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file. If you adapted the QWEN3 script as described above, use the corresponding job YAML and Pod name.
   > - The Pod name and namespace vary with the configuration in the job YAML. The `process-online-recovery-` and `trjob` shown below are example values.

2. Check the job status.

   ```bash
   kubectl get pod -A -o wide
   ```

      The following is an output example. The presence of Running indicates that the job is running properly.

   ```ColdFusion
   trjob            process-online-recovery-master-0                   1/1     Running   0                 14s     xx.xx.xx.xx     master-x   <none>           <none>
   trjob            process-online-recovery-worker-0                   1/1     Running   0                 14s     xx.xx.xx.xx     worker-x   <none>           <none>
   ```

3. Monitor training logs
   1. Monitor the training logs to check whether a UCE fault is triggered.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "status error, rank:8"
      ```

      >[!NOTE]
      > In this example, the fault is injected into rank 8 at step 3. The rank in the `grep` keyword must be consistent with the global rank configured in the environment variable.

      The following is an output example, which indicates that a UCE fault is triggered.

      ```ColdFusion
      2026-06-04 09:24:31.767278 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 0
      2026-06-04 09:24:33.767422 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 1417
      ```

      >[!NOTE]
      >`step: 3` in the log indicates that the fault is triggered at the 3rd training step. `npu_status: 2` indicates that MindIO/TTP has entered the UCE handling state. In this instrumentation scenario, it is triggered by a software simulation path and does not represent a real hardware on-chip memory fault.

**Expected Result**

   1. Check the recovery result of the fault at step 3. If the output on any Pod of master or worker is greater than or equal to 1, the repair is successful.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "(0, 'Mindio do repair operation ok', {}, 'retry')"|wc -l
      ```

   2. Check whether the iterations are normal.

      1. Check the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      Command output:

      ```ColdFusion
      trjob            process-online-recovery-master-0                   1/1     Running   0                 110s    xx.xx.xx.xx     master-x   <none>           <none>
      trjob            process-online-recovery-worker-0                   1/1     Running   0                 110s    xx.xx.xx.xx     worker-x   <none>           <none>
      ```

      >[!NOTE]
      > Check the `RESTARTS` column here. The value must remain 0, which proves that the Pod container was never restarted during the entire UCE fault and repair process.

   3. Check the number of training iterations.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # Return: 0
      kubectl logs -n trjob process-online-recovery-worker-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # Return: 11
      ```

      >[!NOTE]
      > - The iteration step in the `grep` command above (for example, `iteration 4`) must be adjusted based on the step at which the fault is actually injected. If the fault is injected at step `N`, training should continue from step `N+1` after recovery, so the command should be `grep iteration [[:space:]]*{N+1}`. In this example, the fault at step 3 corresponds to `iteration 4`, and the fault at step 10 corresponds to `iteration 11`.
      > - In distributed multi-node training, due to the log redirection mechanism of the training framework, the iteration logs of each rank may be output only to the stdout of some nodes, or redirected to local physical log files.
      > - In this example, the master node returns 0 and the Worker node returns 11. As long as the count on any node is greater than 0, it proves that training has passed the corresponding fault step and continued after the hot repair.

## Deleting a Job<a name="ZH-CN_TOPIC_0000002479386566"></a>

**Procedure<a name="section324819211118"></a>**

Run the following command in the YAML directory where the job is delivered to delete the corresponding training job.

```shell
kubectl delete -f XXX.yaml
```

An example is shown below:

```shell
kubectl delete -f pytorch_multinodes_acjob_910b.yaml
```

Command output:

```ColdFusion
configmap "reset-config-default-test-pytorch" deleted
ascendjob.mindxdl.gitee.com "default-test-pytorch" deleted
```

## O&M Process<a name="ZH-CN_TOPIC_0000002479386520"></a>

**Prerequisites<a name="section18751194535314"></a>**

This function applies only to specific scenarios where you need to use the rescheduling function and the `autoStowing` parameter (sunset) in the Ascend Device Plugin startup YAML is set to `false`.

**Procedure<a name="section8557331115714"></a>**

- You can use the following command to put a chip whose health status has been restored from unhealthy to healthy back into the resource pool.

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-Recover-
    ```

    After this command is executed, the "**huawei.com/Ascend910-Recover**" label is deleted, and the chips in this label are put back into the resource pool for scheduling.

    >[!NOTE]
    >This command is only used to clear the Recover label information. Do not use it to add labels.

- You can use the following command to put a chip whose parameter plane network health status has been restored from unhealthy to healthy back into the resource pool.

    ```shell
    kubectl label nodes node_name huawei.com/Ascend910-NetworkRecover-
    ```

    After this command is executed, the "**huawei.com/Ascend910-NetworkRecover**" label is deleted, and the corresponding chip is also removed from "**huawei.com/Ascend910-NetworkUnhealthy**".

    >[!NOTE]
    >This command is only used to clear the `NetworkRecover` label information. Do not use it to add labels.
