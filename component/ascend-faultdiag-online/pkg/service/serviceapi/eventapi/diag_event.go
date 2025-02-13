package eventapi

import (
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/serviceapi"
)

const apiDiagEvent = "diag"

// GetDiagEventApi 获取添加指标的api
func GetDiagEventApi() *serviceapi.Api {
	return serviceapi.BuildApi(apiDiagEvent, &diagmodel.DiagModel{}, apiDiagEventFunc, nil)
}

// apiDiagEventFunc 诊断事件
func apiDiagEventFunc(fdCtx *context.FaultDiagContext, reqCtx *request.Context, model *diagmodel.DiagModel) error {
	fdCtx.DiagCtx.StartDiag(fdCtx)
	return nil
}
