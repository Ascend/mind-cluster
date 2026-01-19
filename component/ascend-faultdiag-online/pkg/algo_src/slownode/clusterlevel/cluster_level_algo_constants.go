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

// Package clusterlevel is used for file reading and writing, as well as data processing.
package clusterlevel

/* pp并行域 group_name == "pp" */
const ppParallelDomainName string = "pp"

/* tp并行域 group_name == "tp" */
const tpParallelDomainName string = "tp"

/* 数据字节 */
const decimalLen int = 64

/* max loop */
const maxLoopDetection int = 10000000

/* degradation level */
const degradationLevelZero string = "0.0"

/* 链路判断常量 */
const linkHalfStandard float64 = 0.5

/* 集群侧任务级topo数据文件中字段 */
const dataFIleFieldGroupName string = "group_name"

/* 集群侧任务级topo数据文件中字段 */
const dataFIleFieldGlobalRanks string = "global_ranks"
