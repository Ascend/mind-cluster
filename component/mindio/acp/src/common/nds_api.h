/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Description: dpcpfs_stat
 * Author: k00617263
 * Create: 2024-08-21
 */

#ifndef OCK_DFS_NDS_API_H_
#define OCK_DFS_NDS_API_H_

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif /*  __cpluscplus */
#endif /*  __cpluscplus */

/**
 * error code for user face
 */
enum NdsFileOpError {
    NDS_FILE_SUCCESS = 0,
    NDS_FILE_DRIVER_ERROR = -1
};


typedef struct NdsFileError_s {
    enum NdsFileOpError err; // ndsfile error
} NdsFileError_t;

typedef struct NdsFileDescr_s {
    int fd; /* Linux   */
} NdsFileDescr_t;

typedef void *NdsFileHandle_t;


#ifdef __cplusplus
#if __cplusplus
}
#endif /*  __cpluscplus */
#endif /*  __cpluscplus */

#endif
