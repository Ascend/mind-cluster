/**
* @file mspti_result.h
*
* Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
*/

#ifndef MSPTI_BASE_H
#define MSPTI_BASE_H

#include <stddef.h>
#include <stdint.h>

/**
 * @brief MSPTI result codes.
 *
 * Error and result codes returned by MSPTI functions.
 */
typedef enum {
    MSPTI_SUCCESS                                       = 0,
    MSPTI_ERROR_INVALID_PARAMETER                       = 1,
    MSPTI_ERROR_MULTIPLE_SUBSCRIBERS_NOT_SUPPORTED      = 2,
    MSPTI_ERROR_MAX_LIMIT_REACHED                       = 3,
    MSPTI_ERROR_DEVICE_OFFLINE                          = 4,
    MSPTI_ERROR_INNER                                   = 999,
    MSPTI_ERROR_FOECE_INT                               = 0x7fffffff
} msptiResult;

#endif
