/**
* @file mspti_activity.h
*
* Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
*/

#ifndef MSPTI_ACTIVITY_H
#define MSPTI_ACTIVITY_H

#define ACTIVITY_STRUCT_ALIGNMENT 8
#if defined(_WIN32)
#define START_PACKED_ALIGNMENT __pragma(pack(push, 1))
#define PACKED_ALIGNMENT __declspec(align(ACTIVITY_STRUCT_ALIGNMENT))
#define END_PACKED_ALIGNMENT __pragma(pack(pop))
#elif defined(__GNUC__)
#define START_PACKED_ALIGNMENT
#define PACKED_ALIGNMENT __attribute__((__packed__)) __attribute__((aligned(ACTIVITY_STRUCT_ALIGNMENT)))
#define END_PACKED_ALIGNMENT
#else
#define START_PACKED_ALIGNMENT
#define PACKED_ALIGNMENT
#define END_PACKED_ALIGNMENT
#endif

#include <stddef.h>
#include <stdint.h>
#include "mspti_result.h"

#define MSPTI_INVALID_DEVICE_ID ((uint32_t) 0xFFFFFFFFU)
#define MSPTI_INVALID_STREAM_ID ((uint32_t) 0xFFFFFFFFU)
#define MSPTI_INVALID_CORRELATION_ID ((uint64_t) 0)

#if defined(__cplusplus)
extern "C" {
#endif

#if defined(__GNUC__) && defined(MSPTI_LIB)
#pragma GCC visibility push(default)
#endif

/**
 * @brief The kinds of activity records.
 *
 * Each kind is associated with a
 * activity record structure that holds the information associated
 * with the kind.
 */
typedef enum {
    /**
    * The activity record is invalid.
    */
    MSPTI_ACTIVITY_KIND_INVALID = 0,
    MSPTI_ACTIVITY_KIND_MARKER = 1,
    MSPTI_ACTIVITY_KIND_KERNEL = 2,
    MSPTI_ACTIVITY_KIND_API = 3,
    MSPTI_ACTIVITY_KIND_HCCL = 4,
    MSPTI_ACTIVITY_KIND_MEMORY = 5,
    MSPTI_ACTIVITY_KIND_MEMSET = 6,
    MSPTI_ACTIVITY_KIND_MEMCPY = 7,
    MSPTI_ACTIVITY_KIND_EXTERNAL_CORRELATION = 8,
    MSPTI_ACTIVITY_KIND_COUNT,
    MSPTI_ACTIVITY_KIND_FORCE_INT = 0x7fffffff
} msptiActivityKind;

/**
 * @brief The source kinds of mark data.
 *
 * Each mark activity record kind represents information about host or device
 */
typedef enum {
    MSPTI_ACTIVITY_SOURCE_KIND_HOST = 0,
    MSPTI_ACTIVITY_SOURCE_KIND_DEVICE = 1
} msptiActivitySourceKind;

/**
 * @brief The kind of external APIs supported for correlation.
 *
 * Custom correlation kinds are reserved for usage in external tools.
 */
typedef enum {
    MSPTI_EXTERNAL_CORRELATION_KIND_INVALID = 0,
    MSPTI_EXTERNAL_CORRELATION_KIND_UNKNOWN = 1,
    MSPTI_EXTERNAL_CORRELATION_KIND_CUSTOM0 = 2,
    MSPTI_EXTERNAL_CORRELATION_KIND_CUSTOM1 = 3,
    MSPTI_EXTERNAL_CORRELATION_KIND_CUSTOM2 = 4,
    MSPTI_EXTERNAL_CORRELATION_KIND_SIZE,
    MSPTI_EXTERNAL_CORRELATION_KIND_FORCE_INT = 0x7fffffff,
} msptiExternalCorrelationKind;

/**
 * @brief Flags linked to activity records.
 *
 * These are the Flags that pertain to activity records.
 * Flags can be combined by bitwise OR to
 * associated multiple flags with an activity record.
 */
typedef enum {
    /**
    * Signifies that the activity record lacks any flags.
    */
    MSPTI_ACTIVITY_FLAG_NONE = 0,
    /**
    * Represents the activity as a pure host instantaneous marker. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_INSTANTANEOUS = 1 << 0,
    /**
    * Represents the activity as a pure host region start marker. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_START = 1 << 1,
    /**
    * Represents the activity as a pure host region end marker. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_END = 1 << 2,
    /**
    * Represents the activity as an instantaneous marker with device. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_INSTANTANEOUS_WITH_DEVICE = 1 << 3,
    /**
    * Represents the activity as a pure start marker with device. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_START_WITH_DEVICE = 1 << 4,
    /**
    * Represents the activity as a pure end marker with device. Works with
    * MSPTI_ACTIVITY_KIND_MARKER.
    */
    MSPTI_ACTIVITY_FLAG_MARKER_END_WITH_DEVICE = 1 << 5
} msptiActivityFlag;

typedef enum {
    /**
    * Memory is allocated.
    */
    MSPTI_ACTIVITY_MEMORY_OPERATION_TYPE_ALLOCATION = 0,
    /**
    * Memory is released.
    */
    MSPTI_ACTIVITY_MEMORY_OPERATION_TYPE_RELEASE = 1
} msptiActivityMemoryOperationType;

typedef enum {
    /**
    * The memory kind is unknown.
    */
    MSPTI_ACTIVITY_MEMORY_UNKNOWN = 0,
    /**
    * The memory is on the device.
    */
    MSPTI_ACTIVITY_MEMORY_DEVICE = 1
} msptiActivityMemoryKind;

typedef enum {
    /**
    * The memory copy kind is not known.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_UNKNOWN = 0,
    /**
    * A host to host memory copy.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_HTOH = 1,
    /**
    * A host to device memory copy.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_HTOD = 2,
    /**
    * A device to host memory copy.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_DTOH = 3,
    /**
    * A device to device memory copy.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_DTOD = 4,
    /**
    * A system-determined memory copy by address.
    */
    MSPTI_ACTIVITY_MEMCPY_KIND_DEFAULT = 5
} msptiActivityMemcpyKind;

START_PACKED_ALIGNMENT

typedef struct PACKED_ALIGNMENT {
    msptiActivityKind kind;
} msptiActivity;

typedef union PACKED_ALIGNMENT {
    /**
    * A thread object requires that we identify both the process and
    * thread ID.
    */
    struct {
        uint32_t processId;
        uint32_t threadId;
    } pt;
    /**
    * A stream object requires that we identify device and stream ID.
    */
    struct {
        uint32_t deviceId;
        uint32_t streamId;
    } ds;
} msptiObjectId;

/**
 * @brief This activity record serves as a marker, representing a specific moment in time.
 *
 * The marker is characterized by a distinctive name and a unique identifier
 */
typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, always be MSPTI_ACTIVITY_KIND_MARKER.
    */
    msptiActivityKind kind;

    /**
    * The flags associated with the marker.
    * @see msptiActivityFlag
    */
    msptiActivityFlag flag;

    /**
    * The source kinds of mark data.
    * @see msptiActivitySourceKind
    */
    msptiActivitySourceKind sourceKind;

    /**
    * The timestamp for the marker, in ns. A value of 0 indicates that
    * timestamp information could not be collected for the marker.
    */
    uint64_t timestamp;

    /**
    * The marker ID.
    */
    uint64_t id;

    /**
    * The identifier for the activity object associated with this
    * marker. 'objectKind' indicates which ID is valid for this record.
    */
    msptiObjectId objectId;

    /**
    * The marker name for an instantaneous or start marker.
    * This will be NULL for an end marker.
    */
    const char *name;

    /**
    * The name of the domain to which this marker belongs to.
    * This will be NULL for default domain.
    */
    const char *domain;
} msptiActivityMarker;

typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_API.
    */
    msptiActivityKind kind;

    /**
    * The start timestamp for the api, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the api, in ns.
    */
    uint64_t end;

    /**
    * A thread object requires that we identify both the process and
    * thread ID.
    */
    struct {
        uint32_t processId;
        uint32_t threadId;
    } pt;

    /**
    * The correlation ID of the kernel.
    */
    uint64_t correlationId;

    /**
    * The api name.
    */
    const char *name;
} msptiActivityApi;

typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_KERNEL.
    */
    msptiActivityKind kind;

    /**
    * The start timestamp for the kernel, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the kernel, in ns.
    */
    uint64_t end;

    /**
    * A stream object requires that we identify device and stream ID.
    */
    struct {
        uint32_t deviceId;
        uint32_t streamId;
    } ds;

    /**
    * The correlation ID of the kernel.
    */
    uint64_t correlationId;

    /**
    * The kernel type.
    */
    const char *type;

    /**
    * The kernel name.
    */
    const char *name;
} msptiActivityKernel;

typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_HCCL.
    */
    msptiActivityKind kind;

    /**
    * The start timestamp for the hccl, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the hccl, in ns.
    */
    uint64_t end;

    /**
    * A stream object requires that we identify device and stream ID.
    */
    struct {
        uint32_t deviceId;
        uint32_t streamId;
    } ds;

    /**
    * The bandWidth the hccl op, in GB/S
    */
    double bandWidth;

    /**
    * The hccl op name.
    */
    const char *name;

    /**
    * The comm name.
    */
    const char *commName;
} msptiActivityHccl;

/**
 * @brief The activity record for memory.
 *
 * This activity record represents a memory allocation and release operation
 * (MSPTI_ACTIVITY_KIND_MEMORY).
 * This activity record provides separate records for memory allocation and
 * memory release operations.
 */
typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_MEMORY.
    */
    msptiActivityKind kind;

    /**
    * The memory operation requested by the user.
    * @see msptiActivityMemoryOperationType
    */
    msptiActivityMemoryOperationType memoryOperationType;

    /**
    * The memory kind requested by the user.
    * @see msptiActivityMemoryKind
    */
    msptiActivityMemoryKind memoryKind;

    /**
    * The correlation ID of the memory operation.
    */
    uint64_t correlationId;

    /**
    * The start timestamp for the memory operation, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the memory operation, in ns.
    */
    uint64_t end;

    /**
    * The virtual device address of the memory operation.
    */
    uint64_t address;

    /**
    * The count of bytes of the memory operation.
    */
    uint64_t bytes;

    /**
    * The process ID to which this operation belongs to.
    */
    uint32_t processId;

    /**
    * The device ID where this operation is taking place.
    */
    uint32_t deviceId;

    /**
    * The ID of the stream. If memory operation is not async,
    * streamId is set to MSPTI_INVALID_STREAM_ID.
    */
    uint32_t streamId;
} msptiActivityMemory;

/**
 * @brief The activity record for memset.
 *
 * This activity record represents a memory set operation
 * (MSPTI_ACTIVITY_KIND_MEMSET).
 */
typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_MEMSET.
    */
    msptiActivityKind kind;

    /**
    * The value being set to memory.
    */
    uint32_t value;

    /**
    * The count of bytes being set.
    */
    uint64_t bytes;

    /**
    * The start timestamp for the memory set operation, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the memory set operation, in ns.
    */
    uint64_t end;

    /**
    * The device ID where this operation is occurring.
    */
    uint32_t deviceId;

    /**
    * The stream ID where this operation is occurring.
    */
    uint32_t streamId;

    /**
    * The correlation ID of the memory set operation.
    */
    uint64_t correlationId;
    /**
    * Whether the memory set operation happens through async memory APIs.
    */
    uint8_t isAsync;
} msptiActivityMemset;

/**
 * @brief The activity record for memory copies.
 *
 * This activity record represents a memory copy
 * (MSPTI_ACTIVITY_KIND_MEMCPY).
 */
typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_MEMCPY.
    */
    msptiActivityKind kind;

    /**
    * The kind of the memory copy operation.
    * @see msptiActivityMemcpyKind
    */
    msptiActivityMemcpyKind copyKind;

    /**
    * The count of bytes transferred by the memory copy operation.
    */
    uint64_t bytes;

    /**
    * The start timestamp for the memory copy operation, in ns.
    */
    uint64_t start;

    /**
    * The end timestamp for the memory copy operation, in ns.
    */
    uint64_t end;

    /**
    * The device ID where this operation is occurring.
    */
    uint32_t deviceId;

    /**
    * The stream ID where this operation is occurring.
    */
    uint32_t streamId;

    /**
    * The correlation ID of the memory copy operation.
    */
    uint64_t correlationId;

    /**
    * Whether memory operation happens through async memory APIs.
    */
    uint8_t isAsync;
} msptiActivityMemcpy;

typedef struct PACKED_ALIGNMENT {
    /**
    * The activity record kind, must be MSPTI_ACTIVITY_KIND_EXTERNAL_CORRELATION.
    */
    msptiActivityKind kind;
    /**
    * The kind of external API this record correlated to.
    */
    msptiExternalCorrelationKind externalKind;
    /**
    * The exact field in the associated external record depends on that record’s activity kind
    */
    uint64_t externalId;
    /**
    * The correlation ID of the associated MSPTI runtime API record.
    */
    uint64_t correlationId;
} msptiActivityExternalCorrelation;

END_PACKED_ALIGNMENT

/**
 * @brief Function type for callback used by MSPTI to request an empty
 * buffer for storing activity records.
 *
 * This callback function signals the MSPTI client that an activity
 * buffer is needed by MSPTI. The activity buffer is used by MSPTI to
 * store activity records. The callback function can decline the
 * request by setting **buffer to NULL. In this case MSPTI may drop
 * activity records.
 *
 * @param buffer Returns the new buffer. If set to NULL then no buffer
 * is returned.
 * @param size Returns the size of the returned buffer.
 * @param maxNumRecords Returns the maximum number of records that
 * should be placed in the buffer. If 0 then the buffer is filled with
 * as many records as possible. If > 0 the buffer is filled with at
 * most that many records before it is returned.
 */
typedef void(*msptiBuffersCallbackRequestFunc)(uint8_t **buffer, size_t *size, size_t *maxNumRecords);

/**
 * @brief Function type for callback used by MSPTI to return a buffer
 * of activity records.
 *
 * This callback function returns to the MSPTI client a buffer
 * containing activity records.  The buffer contains @p validSize
 * bytes of activity records which should be read using
 * msptiActivityGetNextRecord. After this call MSPTI
 * relinquished ownership of the buffer and will not use it
 * anymore. The client may return the buffer to MSPTI using the
 * msptiBuffersCallbackRequestFunc callback.
 *
 * @param buffer The activity record buffer.
 * @param size The total size of the buffer in bytes as set in
 * MSPTI_BuffersCallbackRequestFunc.
 * @param validSize The number of valid bytes in the buffer.
 */
typedef void(*msptiBuffersCallbackCompleteFunc)(uint8_t *buffer, size_t size, size_t validSize);

/**
 * @brief Registers callback functions with MSPTI for activity buffer
 * handling.
 *
 * This function registers two callback functions to be used in asynchronous
 * buffer handling. If registered, activity record buffers are handled using
 * asynchronous requested/completed callbacks from MSPTI.
 *
 * @param funcBufferRequested callback which is invoked when an empty
 * buffer is requested by MSPTI
 * @param funcBufferCompleted callback which is invoked when a buffer
 * containing activity records is available from MSPTI
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_INVALID_PARAMETER if either
 * funcBufferRequested or funcBufferCompleted is NULL
 */
msptiResult msptiActivityRegisterCallbacks(
    msptiBuffersCallbackRequestFunc funcBufferRequested, msptiBuffersCallbackCompleteFunc funcBufferCompleted);

/**
 * @brief Enable collection of a specific kind of activity record.
 *
 * Enable collection of a specific kind of activity record. Multiple
 * kinds can be enabled by calling this function multiple times.
 * By default, the collection of all activity types is inactive.
 *
 * @param kind The kind of activity record to collect
 *
 * @retval MSPTI_SUCCESS
 */
msptiResult msptiActivityEnable(msptiActivityKind kind);

/**
 * @brief Disable collection of a specific kind of activity record.
 *
 * Disable collection of a specific kind of activity record. Multiple
 * kinds can be disabled by calling this function multiple times.
 * By default, the collection of all activity types is inactive.
 *
 * @param kind The kind of activity record to stop collecting
 *
 * @retval MSPTI_SUCCESS
 */
msptiResult msptiActivityDisable(msptiActivityKind kind);

/**
 * @brief Enable collection of Domain marker.
 *
 * Enable collection of Domain mark. Multiple
 * Domain can be enabled by calling this function multiple times.
 * By default, the Domain of all activity types is inactive.
 *
 * @param name The name of domain to collect
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p name is NULL
 */
msptiResult msptiActivityEnableMarkerDomain(const char* name);

/**
 * @brief Disable collection of Domain marker.
 *
 * Disable collection of Domain mark. Multiple
 * Domain can be disable by calling this function multiple times.
 * By default, the Domain of all activity types is inactive.
 *
 * @param name The name of domain to collect
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p name is NULL
 */
msptiResult msptiActivityDisableMarkerDomain(const char* name);

/**
 * @brief Iterate over the activity records in a buffer.
 *
 * This is a function to iterate over the activity records in buffer.
 *
 * @param buffer The buffer containing activity records
 * @param validBufferSizeBytes The number of valid bytes in the buffer.
 * @param record Inputs the previous record returned by
 * msptiActivityGetNextRecord and returns the next activity record
 * from the buffer. If input value is NULL, returns the first activity
 * record in the buffer.
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_MAX_LIMIT_REACHED if no more records in the buffer
 * @retval MSPTI_ERROR_INVALID_PARAMETER if buffer is NULL.
 */
msptiResult msptiActivityGetNextRecord(uint8_t *buffer, size_t validBufferSizeBytes, msptiActivity **record);

/**
 * @brief Request to deliver activity records via the buffer completion callback.
 *
 * This function returns the activity records associated with all contexts/streams
 * (and the global buffers not associated with any stream) to the MSPTI client
 * using the callback registered in msptiActivityRegisterCallbacks. It return all
 * activity buffers that contain completed activity records, even if these buffers are not completely filled.
 *
 * Before calling this function, the buffer handling callback api must be activated
 * by calling msptiActivityRegisterCallbacks.
 *
 * @param flag Reserved for internal use.
 *
 * @retval MSPTI_SUCCESS
 */
msptiResult msptiActivityFlushAll(uint32_t flag);

/**
 * @brief Periodically Request to deliver activity records via the buffer completion callback.
 *
 * This function periodically returns the activity records associated with
 * all contexts/streams (and the global buffers not associated with any stream)
 * to the MSPTI client using the callback registered in msptiActivityRegisterCallbacks.
 * Periodic flush can return only those activity buffers which are full.
 *
 * Before calling this function, the buffer handling callback api must be activated
 * by calling msptiActivityRegisterCallbacks.
 *
 * @param time Period flush time.Disable period flush when time is set to 0.
 *
 * @retval MSPTI_SUCCESS
 */
msptiResult msptiActivityFlushPeriod(uint32_t time);

/**
 * @brief Push an external correlation id for the calling thread.
 *
 * This function notifies MSPTI that the calling thread is entering an external API region.
 * When a MSPTI activity API record is created while within an external API region and
 * MSPTI_ACTIVITY_KIND_EXTERNAL_CORRELATION is enabled, the activity API record will
 * be preceded by a msptiActivityExternalCorrelation record for each msptiExternalCorrelationKind.
 *
 * @param kind The kind of external API activities should be correlated with.
 * @param id External correlation id.
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_INVALID_PARAMETER
 */
msptiResult msptiActivityPushExternalCorrelationId(msptiExternalCorrelationKind kind, uint64_t id);

/**
 * @brief Pop an external correlation id for the calling thread.
 *
 * This function notifies MSPTI that the calling thread is leaving an external API region.
 *
 * @param kind The kind of external API activities should be correlated with.
 * @param lastId If the function returns successful, contains the last external correlation id for this kind,
 * can be NULL.
 *
 * @retval MSPTI_SUCCESS
 * @retval MSPTI_ERROR_INVALID_PARAMETER
 * @retval MSPTI_ERROR_QUEUE_EMPTY
 */
msptiResult msptiActivityPopExternalCorrelationId(msptiExternalCorrelationKind kind, uint64_t *lastId);

#if defined(__GNUC__) && defined(MSPTI_LIB)
#pragma GCC visibility pop
#endif

#if defined(__cplusplus)
}
#endif

#endif
