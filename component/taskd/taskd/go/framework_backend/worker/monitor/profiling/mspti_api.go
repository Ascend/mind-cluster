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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

/*
	#cgo CFLAGS: -I.
	//#cgo LDFLAGS: -L. -lmspti
	#cgo LDFLAGS: -ldl
	#include <stddef.h>
    #include <dlfcn.h>
    #include <stdlib.h>
    #include <stdio.h>
	#include <pthread.h>
	#include "mspti_activity.h"
	static void *dcmiHandle;
    #define SO_NOT_FOUND  -99999
    #define FUNCTION_NOT_FOUND  -99998
    #define SUCCESS  0
    #define ERROR_UNKNOWN  -99997
	// Go functions declared in C to act as callback functions
	void goBufferRequested(uint8_t **buffer, size_t *size, size_t *maxNumRecords);
	void goBufferCompleted(uint8_t *buffer, size_t size, size_t validSize);
	// Wrapper functions that call the Go callbacks
	static void bufferRequestedCallback(uint8_t **buffer, size_t *size, size_t *maxNumRecords) {
		goBufferRequested(buffer, size, maxNumRecords);
	}

	static void bufferCompletedCallback(uint8_t *buffer, size_t size, size_t validSize) {
		goBufferCompleted(buffer, size, validSize);
	}

	static msptiResult (*cgo_mspti_activity_register_callbacks)(msptiBuffersCallbackRequestFunc funcBufferRequested,
		msptiBuffersCallbackCompleteFunc funcBufferCompleted);
	static msptiResult msptiActivityRegisterCallbacksWrapper() {
		return cgo_mspti_activity_register_callbacks(bufferRequestedCallback, bufferCompletedCallback);
	}

	static int (*cgo_mspti_activity_flush_all)(uint32_t flag);
	static msptiResult mspti_activity_flush_all(uint32_t flag){
		return cgo_mspti_activity_flush_all(flag);
	}

    // dcmi
    static int (*cgo_mspti_activity_enable)(msptiActivityKind kind);
    static msptiResult mspti_activity_enable(msptiActivityKind kind){
		return cgo_mspti_activity_enable(kind);
	}

    static int (*cgo_mspti_activity_dis_enable)(msptiActivityKind kind);
    static msptiResult mspti_activity_dis_enable(msptiActivityKind kind){
		return cgo_mspti_activity_dis_enable(kind);
	}

	static int (*cgo_mspti_activity_get_next_record)(uint8_t *buffer, size_t validBufferSizeBytes,
		msptiActivity **record);
    static msptiResult mspti_activity_get_next_record(uint8_t *buffer, size_t validBufferSizeBytes,
		msptiActivity **record){
		return cgo_mspti_activity_get_next_record(buffer,validBufferSizeBytes,record);
	}

	static int (*cgo_mspti_mstx_domain_enable)(const char* domainName);
    static msptiResult mspti_mstx_domain_enable(const char* domainName){
		return cgo_mspti_mstx_domain_enable(domainName);
	}

	static int (*cgo_mspti_mstx_domain_disable)(const char* domainName);
    static msptiResult mspti_mstx_domain_disable(const char* domainName){
		return cgo_mspti_mstx_domain_disable(domainName);
	}

	 // load .so files and functions
	static int CgoInitMspti(const char* dcmiLibPath){
		if (dcmiLibPath == NULL) {
			fprintf (stderr,"lib path is null\n");
			return 1;
		}
		dcmiHandle = dlopen(dcmiLibPath, RTLD_NOW | RTLD_GLOBAL | RTLD_DEEPBIND );
		if (dcmiHandle == NULL){
			fprintf (stderr,"%s\n",dlerror());
			return 2;
		}

		//MSPTI_API msptiResult msptiActivityEnable(msptiActivityKind kind);
		cgo_mspti_activity_enable = dlsym(dcmiHandle,"msptiActivityEnable");
		cgo_mspti_activity_dis_enable = dlsym(dcmiHandle,"msptiActivityDisable");
		cgo_mspti_activity_get_next_record = dlsym(dcmiHandle,"msptiActivityGetNextRecord");
		cgo_mspti_activity_register_callbacks = dlsym(dcmiHandle,"msptiActivityRegisterCallbacks");
		cgo_mspti_activity_flush_all = dlsym(dcmiHandle,"msptiActivityFlushAll");
		cgo_mspti_mstx_domain_enable = dlsym(dcmiHandle,"msptiActivityEnableMarkerDomain");
		cgo_mspti_mstx_domain_disable = dlsym(dcmiHandle,"msptiActivityDisableMarkerDomain");
		return SUCCESS;
	}

    static char* serialize_msptiActivityMark(msptiActivity **pRecord) {
		msptiActivityMarker* activity = (msptiActivityMarker*)(*pRecord);

		if (pRecord == NULL || *pRecord == NULL) {
			printf("pRecord or *pRecord is NULL\n");
			return NULL;
		}

		char* result = (char*)malloc(1000);
		if (result == NULL) {
			return NULL;
		}

		snprintf(result, 1000, "{\"Kind\":%d,\"Flag\":%d,\"SourceKind\":%d,\"Timestamp\":%llu,\"Id\":%llu,"
				"\"MsptiObjectId\":{\"Pt\":{\"ProcessId\":%u,\"ThreadId\":%u},\"Ds\":{\"DeviceId\":%u,\"StreamId\":%u}}"
				",\"Name\":\"%s\",\"Domain\":\"%s\"}",activity->kind, activity->flag,
				activity->sourceKind,(unsigned long long)activity->timestamp,(unsigned long long)activity->id,
				activity->objectId.pt.processId,activity->objectId.pt.threadId,activity->objectId.ds.deviceId,
				activity->objectId.ds.streamId,activity->name,activity->domain);

		return result;

	}

    static char* serialize_msptiActivityApi(msptiActivity **pRecord) {
		msptiActivityApi* activity = (msptiActivityApi*)(*pRecord);
		char* result = (char*)malloc(300);
		if (result == NULL) {
			return NULL;
		}
		snprintf(result, 300, "{\"Kind\":%d, \"Start\":%llu,\"End\":%llu,\"Pt\":{\"ProcessId\":%u,\"ThreadId\":%u},"
				"\"CorrelationId\":%llu,\"Name\":\"%s\" }",activity->kind, (unsigned long long)activity->start,
				(unsigned long long)activity->end,activity->pt.processId,activity->pt.threadId,
				(unsigned long long)activity->correlationId,activity->name);
		return result;
	}

    static char* serialize_msptiActivityKernel(msptiActivity **pRecord) {
		msptiActivityKernel* activity = (msptiActivityKernel*)(*pRecord);
		char* result = (char*)malloc(300);
		if (result == NULL) {
			return NULL;
		}

		snprintf(result, 300, "{\"Kind\":%d, \"Start\":%llu,\"End\":%llu,\"Ds\":{\"DeviceId\":%u,\"StreamId\":%u},"
				"\"CorrelationId\":%llu,\"Type\":\"%s\",\"Name\":\"%s\" }",activity->kind,
				(unsigned long long)activity->start,(unsigned long long)activity->end,activity->ds.deviceId,
				activity->ds.streamId,(unsigned long long)activity->correlationId,activity->type,activity->name);
		return result;
	}

	static void free_serialized_data(const char* data) {
		free((void*)data);
	}

	static uint64_t getThreadID() {
		return (uint64_t)pthread_self();
	}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"taskd/common/constant"
)

var requestSem = make(chan struct{}, constant.MaxRequestBufferNum)

const int2 int32 = 2

// InitMspti found mspti so and init it
func InitMspti() error {
	libMsptiName := "libmspti.so"
	libPath, err := utils.GetDriverLibPath(libMsptiName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get mspti lib path, error: %v, will try default path", err)
		libPath = constant.MsptiLibPath + libMsptiName
	}
	libPathCString := C.CString(libPath)
	defer C.free(unsafe.Pointer(libPathCString))
	if retCode := C.CgoInitMspti(libPathCString); retCode != C.SUCCESS {
		hwlog.RunLog.Errorf("failed to init mspti so %s with recode:%v", libPath, retCode)
		return fmt.Errorf("failed to init mspti so with recode:%v", retCode)
	}
	hwlog.RunLog.Infof("successfully init mspti lib, libPath:%s", libPath)
	return nil
}

// EnableMarkerDomain to enable or disable specific domain
func EnableMarkerDomain(domainName string, enable bool) error {
	cDomainName := C.CString(domainName)
	defer C.free(unsafe.Pointer(cDomainName))
	if !enable {
		if result := C.mspti_mstx_domain_disable(cDomainName); result != C.MSPTI_SUCCESS {
			hwlog.RunLog.Errorf("failed to disable domain %s with retCode:%v", domainName, result)
			return fmt.Errorf("failed to disable domain %s with retCode:%v", domainName, result)
		}
	} else {
		if result := C.mspti_mstx_domain_enable(cDomainName); result != C.MSPTI_SUCCESS {
			hwlog.RunLog.Errorf("failed to enable domain %s with retCode:%v", domainName, result)
			return fmt.Errorf("failed to enable domain %s with retCode:%v", domainName, result)
		}
	}
	hwlog.RunLog.Infof("successfully changed domain %s status to %v, rank:%d", domainName, enable, GlobalRankId)
	return nil
}

// ProfileRecordsMark cache all Marker kind records
var ProfileRecordsMark = make([]MsptiActivityMark, 0)

// ProfileRecordsApi cache all Api kind records
var ProfileRecordsApi = make([]MsptiActivityApi, 0)

// ProfileRecordsKernel cache all Kernel kind records
var ProfileRecordsKernel = make([]MsptiActivityKernel, 0)

// MsptiActivityRegisterCallbacksWrapper to register callbacks to mspti
func MsptiActivityRegisterCallbacksWrapper() error {
	if result := C.msptiActivityRegisterCallbacksWrapper(); result != C.MSPTI_SUCCESS {
		hwlog.RunLog.Errorf("failed to register callba cks with retCode:%v", result)
		return fmt.Errorf("failed to register callbacks with retCode:%v", result)
	}
	hwlog.RunLog.Infof("successfully registered profiling callbacks")
	return nil
}

// EnableMsptiMarkerActivity enable marker profile activity
func EnableMsptiMarkerActivity() error {
	if retCode := C.mspti_activity_enable(C.MSPTI_ACTIVITY_KIND_MARKER); retCode != C.SUCCESS {
		return fmt.Errorf("failed to enable profiling marker data, error code: %d", int32(retCode))
	}
	hwlog.RunLog.Infof("successfully enabled profiling")
	return nil
}

// DisableMsptiActivity disable all mspti kinds
func DisableMsptiActivity() error {
	if retCode := C.mspti_activity_dis_enable(C.MSPTI_ACTIVITY_KIND_MARKER); retCode != C.SUCCESS {
		return fmt.Errorf("failed to disable profiling maker data, err code: %d", int32(retCode))
	}
	if retCode := C.mspti_activity_dis_enable(C.MSPTI_ACTIVITY_KIND_KERNEL); retCode != C.SUCCESS {
		return fmt.Errorf("failed to enable profiling kernel data, error code: %d", int32(retCode))
	}
	if retCode := C.mspti_activity_dis_enable(C.MSPTI_ACTIVITY_KIND_API); retCode != C.SUCCESS {
		return fmt.Errorf("failed to enable profiling api data, error code: %d", int32(retCode))
	}
	hwlog.RunLog.Infof("rank:%v successfully disabled profiling", GlobalRankId)
	return nil
}

// FlushAllActivity flush will be called for each step, while each step finished
func FlushAllActivity() error {
	if retCode := C.mspti_activity_flush_all(C.uint32_t(1)); retCode != C.SUCCESS {
		hwlog.RunLog.Errorf("failed to flush all activities, errCode:%v", retCode)
		return fmt.Errorf("failed to flush all activties, errCode:%v", retCode)
	}
	hwlog.RunLog.Debugf("rank:%v successfully flush all activities", GlobalRankId)
	return nil
}

// goBufferRequested mspti will request for memory, after fulfilled it will call goBufferCompleted
//
//export goBufferRequested
func goBufferRequested(buffer **C.uint8_t, size *C.size_t, maxNumRecords *C.size_t) {
	if len(requestSem) > constant.MaxRequestBufferNum/constant.HalfSize {
		hwlog.RunLog.Warnf("requeste for buffer, current requested buffer num:%v", len(requestSem))
	}
	requestSem <- struct{}{}
	maxRecords := 0
	*buffer = (*C.uint8_t)(C.malloc(C.size_t(constant.NormalBufferSizeInBytes)))
	*size = C.size_t(constant.NormalBufferSizeInBytes)
	*maxNumRecords = C.size_t(maxRecords)
}

// goBufferCompleted  fulfilled it will call goBufferCompleted
//
//export goBufferCompleted
func goBufferCompleted(buffer *C.uint8_t, size C.size_t, validSize C.size_t) {
	ProfileTaskQueue.AddTask(dealBufferCompleted, buffer, size, validSize)
}

func dealBufferCompleted(buffer *C.uint8_t, size C.size_t, validSize C.size_t) {
	defer func() {
		<-requestSem
		hwlog.RunLog.Debugf("the buffer free status is: %v", buffer == nil)
		if buffer != nil {
			hwlog.RunLog.Debugf("will free current buffer, the buffer address is %v", buffer)
			// free address
			C.free(unsafe.Pointer(buffer))
			buffer = nil
		}
	}()
	if validSize > 0 {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var status C.msptiResult
		count := 0
		for {
			var pRecord *C.msptiActivity
			status = C.mspti_activity_get_next_record(buffer, validSize, &pRecord)
			if status == C.MSPTI_SUCCESS {
				count++
				handleActivityRecord(pRecord)
			} else if status == C.MSPTI_ERROR_MAX_LIMIT_REACHED {
				hwlog.RunLog.Infof("there is no more records in the buffer,the current mark size is %v, count is: %v",
					len(ProfileRecordsMark), count)
				break
			} else if status == C.MSPTI_ERROR_INVALID_PARAMETER {
				hwlog.RunLog.Warnf("given buffer is nil, code: %v", status)
				break
			} else {
				hwlog.RunLog.Warnf("received code is not SUCCESS, code: %v", status)
				break
			}
		}
	}
	hwlog.RunLog.Debugf("the buffer free status is: %v", buffer == nil)
}

func handleActivityRecord(pRecord *C.msptiActivity) {
	if pRecord.kind == C.MSPTI_ACTIVITY_KIND_MARKER {
		handleMarkerRecord(pRecord)
	} else if pRecord.kind == C.MSPTI_ACTIVITY_KIND_API {
		handleApiRecord(pRecord)
	} else if pRecord.kind == C.MSPTI_ACTIVITY_KIND_KERNEL {
		handleKernelRecord(pRecord)
	} else if pRecord.kind == C.MSPTI_ACTIVITY_KIND_HCCL {
		hwlog.RunLog.Debugf("current will not deal with hccl record")
	} else {
		hwlog.RunLog.Errorf("receive unsupported activity type, activity: %#v", pRecord)
	}
}

// handleMarkerRecord if the record is marker, deal with it, push it into a cache array with mutex lock
func handleMarkerRecord(pRecord *C.msptiActivity) {
	cString := C.serialize_msptiActivityMark(&pRecord)
	jsonStr := C.GoString(cString)
	C.free_serialized_data(cString)
	hwlog.RunLog.Debugf("got a marker kind string: %s", jsonStr)
	var mark MsptiActivityMark
	err := json.Unmarshal([]byte(jsonStr), &mark)
	if err != nil {
		hwlog.RunLog.Errorf("failed to decode record %v err:%v", jsonStr, err)
		return
	}
	hwlog.RunLog.Debugf("got a marker kind record: %v", mark.Timestamp)
	appendMark(mark)
}

func handleApiRecord(pRecord *C.msptiActivity) {
	cString := C.serialize_msptiActivityApi(&pRecord)
	jsonStr := C.GoString(cString)
	C.free_serialized_data(cString)
	var api MsptiActivityApi
	err := json.Unmarshal([]byte(jsonStr), &api)
	if err != nil {
		hwlog.RunLog.Errorf("failed to decode record %v err:%v", jsonStr, err)
		return
	}
	hwlog.RunLog.Debugf("got a api kind record: %#v", api.Marshal())
	appendApi(api)
}

func handleKernelRecord(pRecord *C.msptiActivity) {
	cString := C.serialize_msptiActivityKernel(&pRecord)
	jsonStr := C.GoString(cString)
	C.free_serialized_data(cString)
	var kernel MsptiActivityKernel
	err := json.Unmarshal([]byte(jsonStr), &kernel)
	if err != nil {
		hwlog.RunLog.Errorf("failed to decode record %v err:%v", jsonStr, err)
		return
	}
	hwlog.RunLog.Debugf("got a kernel kind record: %#v", kernel.Marshal())
	appendKernel(kernel)
}

func appendKernel(kernel MsptiActivityKernel) {
	constant.MuKernal.Lock()
	defer constant.MuKernal.Unlock()
	ProfileRecordsKernel = append(ProfileRecordsKernel, kernel)
}

func appendApi(api MsptiActivityApi) {
	constant.MuApi.Lock()
	defer constant.MuApi.Unlock()
	ProfileRecordsApi = append(ProfileRecordsApi, api)
}

func appendMark(mark MsptiActivityMark) {
	constant.MuMark.Lock()
	defer constant.MuMark.Unlock()
	ProfileRecordsMark = append(ProfileRecordsMark, mark)
}
