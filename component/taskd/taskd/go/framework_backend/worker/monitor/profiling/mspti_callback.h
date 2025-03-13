/**
* @file mspti_callback.h
*
* Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
*/


#ifndef MSPTI_CALLBACK_H
#define MSPTI_CALLBACK_H

#include <stdint.h>
#include "mspti_result.h"
#include "mspti_cbid.h"

#if defined(__cplusplus)
extern "C" {
#endif

#if defined(__GNUC__) && defined(MSPTI_LIB)
#pragma GCC visibility push(default)
#endif

/**
* @brief Callback domains.
*
* Callback domains. Each domain represents callback points for a
* group of related API functions or CANN driver activity.
*/
typedef enum {
    /**
   * Invalid domain.
   */
    MSPTI_CB_DOMAIN_INVALID           = 0,
    /**
   * Domain containing callback points for all runtime API functions.
   */
    MSPTI_CB_DOMAIN_RUNTIME           = 1,
    MSPTI_CB_DOMAIN_HCCL              = 2,
    MSPTI_CB_DOMAIN_SIZE,
    MSPTI_CB_DOMAIN_FORCE_INT         = 0x7fffffff
} msptiCallbackDomain;

typedef uint32_t msptiCallbackId;

/**
 * @brief Specifies the point in an API call that a callback is issued.
 *
 * Specifies the point in an API call that a callback is issued. This
 * value is communicated to the callback function by @ref
 * msptiCallbackData::callbackSite.
 */
typedef enum {
    /**
    * The callback is at the entry of the API call.
    */
    MSPTI_API_ENTER                 = 0,
    /**
    * The callback is at the exit of the API call.
    */
    MSPTI_API_EXIT                  = 1,
    MSPTI_API_CBSITE_FORCE_INT     = 0x7fffffff
} msptiApiCallbackSite;

typedef struct {
    /**
    * Point in the runtime or driver function from where the callback
    * was issued.
    */
    msptiApiCallbackSite callbackSite;

    /**
    * Name of the runtime or driver API function which issued the
    * callback.
    */
    const char *functionName;

    /**
    * Params of the runtime or driver API function which issued the
    * callback.
    */
    const void *functionParams;

    /**
    * Pointer to the return value of the runtime or driver API
    * call.
    */
    const void *functionReturnValue;

    /**
   * Name of the symbol operated on by the runtime or driver API
   * function which issued the callback. This entry is valid only for
   * driver and runtime launch callbacks, where it returns the name of
   * the kernel.
   */
    const char *symbolName;

    /**
    * The activity record correlation ID for this callback. For a
    * driver domain callback (i.e. @p domain
    * MSPTI_CB_DOMAIN_DRIVER_API) this ID will equal the correlation ID
    * in the MSPTI_ActivityAPI record corresponding to the CANN driver
    * function call. For a runtime domain callback (i.e. @p domain
    * MSPTI_CB_DOMAIN_RUNTIME_API) this ID will equal the correlation
    * ID in the MSPTI_ActivityAPI record corresponding to the CANN
    * runtime function call. Within the callback, this ID can be
    * recorded to correlate user data with the activity record.
    */
    uint64_t correlationId;

    /**
    * Undefined. Reserved for internal use.
    */
    uint64_t reserved1;

    /**
   * Undefined. Reserved for internal use.
   */
    uint64_t reserved2;

    /**
    * Pointer to data shared between the entry and exit callbacks of
    * a given runtime or drive API function invocation. This field
    * can be used to pass 64-bit values from the entry callback to
    * the corresponding exit callback.
    */
    uint64_t *correlationData;
} msptiCallbackData;

/**
 * @brief Function type for a callback.
 *
 * Function type for a callback. The type of the data passed to the
 * callback in @p cbdata depends on the @p domain. If @p domain is
 * MSPTI_CB_DOMAIN_RUNTIME the type
 * of @p cbdata will be msptiCallbackData.
 *
 * @param userdata User data supplied at subscription of the callback
 * @param domain The domain of the callback
 * @param cbid The ID of the callback
 * @param cbdata Data passed to the callback.
 */
typedef void (*msptiCallbackFunc)(
    void *userdata,
    msptiCallbackDomain domain,
    msptiCallbackId cbid,
    const msptiCallbackData *cbdata);

struct msptiSubscriber_st;

/**
 * @brief A callback subscriber.
 */
typedef struct msptiSubscriber_st *msptiSubscriberHandle;

/**
 * @brief Initialize a callback subscriber with a callback function
 * and user data.
 *
 * Initializes a callback subscriber with a callback function and
 * (optionally) a pointer to user data. The returned subscriber handle
 * can be used to enable and disable the callback for specific domains
 * and callback IDs.
 * @note Only a single subscriber can be registered at a time. To ensure
 * that no other MSPTI client interrupts the profiling session, it's the
 * responsibility of all the MSPTI clients to call this function before
 * starting the profling session.
 * @note This function does not enable any callbacks.
 * @note @b Thread-safety: this function is thread safe.
 *
 * @param subscriber handle to initialize subscriber
 * @param callback The callback function
 * @param userdata A pointer to user data. This data will be passed to
 * the callback function via the @p userdata paramater.
 *
 * @retval MSPTI_SUCCESS on success
 * @retval MSPTI_ERROR_INNER if unable to initialize MSPTI
 * @retval MSPTI_ERROR_MULTIPLE_SUBSCRIBERS_NOT_SUPPORTED if there is already a MSPTI subscriber
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p subscriber is NULL
 */
msptiResult msptiSubscribe(msptiSubscriberHandle *subscriber, msptiCallbackFunc callback, void* userdata);

/**
 * @brief Unregister a callback subscriber.
 *
 * Removes a callback subscriber so that no future callbacks will be
 * issued to that subscriber.
 *
 * @param subscriber Handle to the initialize subscriber
 *
 * @retval MSPTI_SUCCESS on success
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p subscriber is NULL or not initialized
 */
msptiResult msptiUnsubscribe(msptiSubscriberHandle subscriber);

/**
 * @brief Enable or disabled callbacks for a specific domain and
 * callback ID.
 *
 * Enable or disabled callbacks for a subscriber for a specific domain
 * and callback ID.
 *
 * @note @b Thread-safety: a subscriber must serialize access to
 * msptiEnableCallback, msptiEnableDomain.
 *
 * @param enable New enable state for the callback. Zero disables the
 * callback, non-zero enables the callback.
 * @param subscriber Handle to callback subscription
 * @param domain The domain of the callback
 * @param cbid The ID of the callback
 *
 * @retval MSPTI_SUCCESS on success
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p subscriber, @p domain or @p
 * cbid is invalid.
 */
msptiResult msptiEnableCallback(
    uint32_t enable, msptiSubscriberHandle subscriber, msptiCallbackDomain domain, msptiCallbackId cbid);

/**
 * @brief Enable or disabled callbacks for a specific domain
 *
 * Enable or disabled callbacks for a subscriber for a specific domain
 *
 * @note @b Thread-safety: a subscriber must serialize access to
 * msptiEnableCallback, msptiEnableDomain.
 *
 * @param enable New enable state for the callback. Zero disables the
 * callback, non-zero enables the callback.
 * @param subscriber Handle to callback subscription
 * @param domain The domain of the callback
 *
 * @retval MSPTI_SUCCESS on success
 * @retval MSPTI_ERROR_INVALID_PARAMETER if @p subscriber, @p domain is invalid.
 */
msptiResult msptiEnableDomain(
    uint32_t enable, msptiSubscriberHandle subscriber, msptiCallbackDomain domain);

#if defined(__GNUC__) && defined(MSPTI_LIB)
#pragma GCC visibility pop
#endif

#if defined(__cplusplus)
}
#endif

#endif
