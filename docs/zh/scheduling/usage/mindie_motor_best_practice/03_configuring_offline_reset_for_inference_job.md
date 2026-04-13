# 配置推理任务场景下的离线复位<a name="ZH-CN_TOPIC_0000002479226442"></a>

当前仅支持Atlas 800I A2 推理服务器、Atlas 800I A3 超节点服务器的离线复位，开启此功能，芯片发生故障后，会进行热复位操作，让芯片恢复健康。

开启MindIE Motor推理任务的离线复位功能只需要将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”。

**表 1**  参数说明

<a name="table173461839165111"></a>

|参数|类型|默认值|说明|
|--|--|--|--|
|-hotReset|int|-1|设备热复位功能参数。开启此功能，芯片发生故障后，Ascend Device Plugin会进行热复位操作，使芯片恢复健康。<ul><li>-1：关闭芯片复位功能</li><li>0：开启推理设备复位功能</li><li>1：开启训练设备在线复位功能</li><li>2：开启训练/推理设备离线复位功能</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>取值为1对应的功能已经日落，请配置其他取值。</p></div></div>该参数支持的训练设备：<ul><li>Atlas 800 训练服务器（型号 9000）（NPU满配）</li><li>Atlas 800 训练服务器（型号 9010）（NPU满配）</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD（型号 9000）</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A2 PoD 集群基础单元</li><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li></ul>该参数支持的推理设备：<ul><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li><li>Atlas 300I Duo 推理卡</li><li>Atlas 300I 推理卡（型号 3000）（整卡）</li><li>Atlas 300I 推理卡（型号 3010）</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li><li>Atlas 800I A3 超节点服务器</li></ul>|

>[!NOTE] 
>Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>
>- 方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>- 方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。
