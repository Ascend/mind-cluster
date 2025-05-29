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

// Package netfaultapi for fault network feature
package netfaultapi

/*
#cgo linux LDFLAGS: -ldl

#include <stdint.h>
#include <stdlib.h>

// 声明 Go 导出的回调函数
extern void netfaultResultCallBack(char* msg);

*/
import "C"
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/fdol/api"
	"ascend-faultdiag-online/pkg/fdol/context/contextdata"
	"ascend-faultdiag-online/pkg/fdol/context/diagcontext"
	"ascend-faultdiag-online/pkg/fdol/model"
	"ascend-faultdiag-online/pkg/fdol/sohandle"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature"
	"ascend-faultdiag-online/pkg/model/feature/netfault"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

const (
	apiNetwork       = "netfault"
	apiController    = "controller"
	apiStart         = "start"
	apiStop          = "stop"
	apiReload        = "reload"
	registerCallBack = "registerCallBack"
	name             = "netfault"
	byteSize         = 1024
	cluster          = "cluster"
	inputCommand     = "command"
	inputTarget      = "target"
	inputFunc        = "func"

	faultAssertionOnce      = "once"
	faultTypeNetwork        = "Network"
	publicFaultVersion      = "1.0"
	faultResource           = "fd-online"
	faultCodeNode           = "200001010"
	faultCodeSuperPod       = "200001011"
	faultCodeLink           = "200001012"
	rootCauseTypeNpu        = 0
	rootCauseTypeNode       = 4
	rootCauseTypeNetPlaneL2 = 6
	serverIDLeftMove        = 22
	serverIDMask            = 0x3FF
	deviceIDMask            = 0xFFFF
	grpcRetCodeSuccess      = 0
	splitLen                = 2
)

var dContext *diagcontext.DiagContext
var contextData *contextdata.CtxData

// ControllerStart 开启Cluster
func ControllerStart() *api.Api {
	return api.BuildApi(apiStart, &feature.Status{}, ControllerStartFunc, nil)
}

// ControllerStop 关闭Cluster
func ControllerStop() *api.Api {
	return api.BuildApi(apiStop, &feature.Status{}, ControllerStopFunc, nil)
}

// ControllerReload reload Cluster
func ControllerReload() *api.Api {
	return api.BuildApi(apiReload, &feature.Status{}, ControllerReloadFunc, nil)
}

// ControllerStartFunc start controller of cluster
func ControllerStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, model *feature.Status) error {
	dContext = diagCtx
	contextData = ctxData

	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}

	if err := registerCallback(handler); err != nil {
		hwlog.RunLog.Errorf("netfault controller register callback before starting failed, err: %v", err)
		return err
	}

	input, err := createJSONInput(apiStart, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output2 := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output2)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller start successfully"
	return nil
}

// ControllerStopFunc stop controller of cluster
func ControllerStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, model *feature.Status) error {
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	input, err := createJSONInput(apiStop, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller stop successfully"
	return nil
}

// ControllerReloadFunc reload controller of cluster
func ControllerReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, model *feature.Status) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}

	if err := registerCallback(handler); err != nil {
		hwlog.RunLog.Errorf("netfault controller register callback before reloading failed, err: %v", err)
		return err
	}

	input, err := createJSONInput(apiReload, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller reload successfully"
	return nil
}

func registerCallback(handler *sohandle.SoHandler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	input, err := createJSONInput(registerCallBack, uintptr(unsafe.Pointer(C.netfaultResultCallBack)), cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	return nil
}

// 回调将数据放到ctx中
//
//export netfaultResultCallBack
func netfaultResultCallBack(cMessage *C.char) {
	message := C.GoString(cMessage)
	defer C.free(unsafe.Pointer(cMessage))

	if dContext == nil || contextData == nil {
		hwlog.RunLog.Errorf("netfault result callback err: context is nil")
		return
	}
	if enum.Cluster != contextData.Framework.Config.Mode {
		return
	}
	var clusterResult []netfault.ClusterResult
	if err := json.Unmarshal([]byte(message), &clusterResult); err != nil {
		hwlog.RunLog.Errorf("error parsing clusterResult JSON: %v", err)
		return
	}
	if len(clusterResult) > 0 {
		hwlog.RunLog.Infof("the result of callback is %+v", clusterResult)
	}
	parseAndAddMetric(clusterResult, dContext)
	sendToPubFaultCenter(clusterResult)
	return
}

// 解析结果入库到MetricPool
func parseAndAddMetric(clusterResult []netfault.ClusterResult, context *diagcontext.DiagContext) {
	models := make([]*metricmodel.MetricReqModel, 0)
	// 通用的创建 MetricReqModel 的方法
	createMetric := func(domainType string, domainValue string, name string,
		valueType enum.MetricValueType, value string) *metricmodel.MetricReqModel {
		return &metricmodel.MetricReqModel{
			Domain: []*metricmodel.DomainItem{
				{
					DomainType: enum.MetricDomainType(domainType),
					Value:      domainValue,
				},
			},
			Name:      name,
			ValueType: valueType,
			Value:     value,
		}
	}
	// 从检测结果添加指标模型
	for i, result := range clusterResult {
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.TaskId,
			enum.StringMetric, result.TaskID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MinLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.MinLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MaxLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.MaxLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.AvgLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.AvgLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MinDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.MinDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MaxDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.MaxDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.AvgDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.AvgDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.FaultType,
			enum.FloatMetric, strconv.Itoa(result.FaultType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.SrcID,
			enum.StringMetric, result.SrcID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.SrcType,
			enum.FloatMetric, strconv.Itoa(result.SrcType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.DstID,
			enum.StringMetric, result.DstID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.DstType,
			enum.FloatMetric, strconv.Itoa(result.DstType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.Level,
			enum.FloatMetric, strconv.Itoa(result.Level)))
	}
	addMetricFromClusterResult(models, context)
}

func addMetricFromClusterResult(models []*metricmodel.MetricReqModel, context *diagcontext.DiagContext) {
	// 统一处理 MetricPool 添加逻辑
	for _, metric := range models {
		if slicetool.ValueIn(metric.ValueType, []enum.MetricValueType{enum.FloatMetric, enum.StringMetric}) != nil {
			hwlog.RunLog.Infof("Unknown Metric Type: %s", metric.ValueType)
			continue
		}
		domain := context.DomainFactory.GetInstance(metric.Domain)
		context.MetricPool.AddMetric(&diagcontext.Metric{Domain: domain, Name: metric.Name},
			metric.Value, metric.ValueType)
	}
}

func sendToPubFaultCenter(clusterResult []netfault.ClusterResult) {
	if len(clusterResult) == 0 {
		hwlog.RunLog.Debugf("fault result is empty, no need send to public fault center")
		return
	}
	if global.GrpcClient == nil {
		client, err := grpc.GetClient(utils.GetClusterIP())
		if err != nil {
			hwlog.RunLog.Errorf("get grpc client failed, err: %v", err)
			return
		}

		global.GrpcClient = client
	}
	req, err := createPubFault(clusterResult)
	if err != nil {
		hwlog.RunLog.Errorf("create public fault failed, err: %v", err)
		return
	}
	res, err := global.GrpcClient.SendToPubFaultCenter(req)
	if err != nil {
		hwlog.RunLog.Errorf("send to public fault center failed, err: %v", err)
		return
	}
	if res.GetCode() != grpcRetCodeSuccess {
		hwlog.RunLog.Errorf("send to public fault center failed, resp code: %v, msg: %v", res.GetCode(), res.GetInfo())
		return
	}
	hwlog.RunLog.Info("send to public fault center success")
}

func createPubFault(clusterResult []netfault.ClusterResult) (*pubfault.PublicFaultRequest, error) {
	if len(clusterResult) == 0 {
		return nil, errors.New("input cluster results is empty")
	}
	now := time.Now().UnixMilli()
	newFaultReq := pubfault.PublicFaultRequest{
		Id:        string(uuid.NewUUID()),
		Timestamp: now,
		Version:   publicFaultVersion,
		Resource:  faultResource,
		Faults:    make([]*pubfault.Fault, len(clusterResult)),
	}

	for i, result := range clusterResult {
		descData, err := json.Marshal(result)
		if err != nil {
			hwlog.RunLog.Warnf("unmarshal fault info failed, err: %v, fault: %v", err, result)
			continue
		}
		newFaultReq.Faults[i] = &pubfault.Fault{
			Assertion: faultAssertionOnce,
			FaultId:   generateFaultID(&result),
			FaultType: faultTypeNetwork,
			FaultCode: getFaultCode(&result),
			FaultTime: int64(result.TimeStamp),
			FaultLocation: map[string]string{
				constants.SrcType: strconv.Itoa(result.SrcType), constants.SrcID: result.SrcID,
				constants.DstType: strconv.Itoa(result.DstType), constants.DstID: result.DstID,
			},
			Influence:   getInfluence(&result),
			Description: string(descData),
		}
	}
	return &newFaultReq, nil
}

func generateFaultID(result *netfault.ClusterResult) string {
	if result == nil {
		hwlog.RunLog.Warn("result is empty")
		return ""
	}
	faultEntity := fmt.Sprintf("%d-%s->%d-%s", result.SrcType, result.SrcID, result.DstType, result.DstID)
	h := sha256.New()
	_, err := h.Write([]byte(faultEntity))
	if err != nil {
		hwlog.RunLog.Warnf("generateFaultID failed, err: %v", err)
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func getFaultCode(result *netfault.ClusterResult) string {
	if result == nil {
		hwlog.RunLog.Warn("result is empty")
		return ""
	}
	if result.SrcType != rootCauseTypeNpu || result.DstType != rootCauseTypeNpu {
		return faultCodeLink
	}
	if checkIsTheSameNode(result) {
		return faultCodeNode
	}
	return faultCodeSuperPod
}

func checkIsTheSameNode(result *netfault.ClusterResult) bool {
	if result == nil {
		return false
	}

	srcSdid, err := strconv.Atoi(result.SrcID)
	if err != nil {
		return false
	}

	dstSdid, err := strconv.Atoi(result.DstID)
	if err != nil {
		return false
	}

	srcServerID := (srcSdid >> serverIDLeftMove) & serverIDMask
	dstServerID := (dstSdid >> serverIDLeftMove) & serverIDMask
	return srcServerID == dstServerID
}

func getInfluence(result *netfault.ClusterResult) []*pubfault.PubFaultInfo {
	infoList := make([]*pubfault.PubFaultInfo, 1)
	if result == nil {
		return infoList
	}
	info := &pubfault.PubFaultInfo{
		NodeName:  strings.ToLower(result.SrcID),
		DeviceIds: []int32{int32(0)},
	}
	infoList[0] = info
	arr := strings.Split(result.SrcID, "-")
	if len(arr) != splitLen {
		return infoList
	}
	sdIdStr := arr[1]
	sdId, err := strconv.Atoi(sdIdStr)
	if err != nil {
		hwlog.RunLog.Debugf("invalid id, err: %v", err)
		return infoList
	}
	devId := sdId & deviceIDMask
	info.DeviceIds[0] = int32(devId)
	return infoList
}

// 构造 JSON 输入
func createJSONInput(command string, funcPtr uintptr, target string) (string, error) {
	input := map[string]interface{}{
		inputCommand: command,
		inputTarget:  target,
	}

	if funcPtr != 0 {
		input[inputFunc] = funcPtr
	}
	jsonData, err := json.Marshal(input)
	return string(jsonData), err
}
