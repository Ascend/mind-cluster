/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package nodelevel is used for file reading and writing, as well as data processing.
package nodelevel

/* 节点侧任务级每张npu卡单独一份的steptime时延数据 */
const stepTimeFileName string = "steptime.csv"

/* 节点侧任务级npu卡并行域信息 */
const rankTopofileName string = "parallel_group.json"

/* topo中tp并行域字段 group_name == "tp" */
const tpParallelDomainName string = "tp"

/* 同质化组中至少有 2 张卡才可以进行均质化对比 */
const minRanksInGroup int = 2

/* steptime csv文件中数据必须为两列 */
const stepTimeFileMinColumns int = 2

/* 根据当前30s检测一次，15分钟连续不更新则认为卡死，打印相关日志 900/30 */
const maxContinuousNotUpdate int = 30

/* 数据字节 */
const byteLength int = 64

/* maxDetectionLoop */
const maxDetectionLoop int = 10000000

/* 节点侧任务级数据文件中PP_device列数据 */
const ppDataColumn string = "PP_device"

/* 节点侧任务级数据文件中ZP_device列数据 */
const zpDataColumn string = "ZP_device"

/* 节点侧任务级数据文件中ZP_host列数据 */
const zpHostDataColumn string = "ZP_host"

/* 节点侧任务级数据文件中dataloader_host列数据 */
const dataLoaderDataColumn string = "dataloader_host"

/* step time 数据 */
const stepTimeData string = "stepTime"

/* 节点侧任务级数据文件名 */
const nodeLevelNpuLatencyDataFile string = "comm.csv"

/* 节点侧任务级topo数据文件中group_name字段 */
const dataFIleFieldGroupName string = "group_name"

/* 节点侧任务级topo数据文件中global_ranks字段 */
const dataFIleFieldGlobalRanks string = "global_ranks"
