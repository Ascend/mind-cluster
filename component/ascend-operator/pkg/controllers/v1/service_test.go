/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func TestGetServiceIpAndPort(t *testing.T) {
	convey.Convey("getServiceIpAndPort", t, func() {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{Ports: make([]corev1.ServicePort, 1), ClusterIP: "127.0.0.1"},
		}
		defaultPort := corev1.ServicePort{
			Name: mindxdlv1.DefaultPortName,
			Port: mindxdlv1.DefaultPort,
		}
		convey.Convey("01-service with no port should return empty ip and port", func() {
			ip, port := getServiceIpAndPort(svc)
			convey.ShouldEqual(ip, "")
			convey.ShouldEqual(port, "")
		})
		convey.Convey("02-service with no default port should return empty ip and port", func() {
			svc.Spec.Ports[0] = corev1.ServicePort{
				Name: "fake",
				Port: 0,
			}
			ip, port := getServiceIpAndPort(svc)
			convey.ShouldEqual(ip, "")
			convey.ShouldEqual(port, "")
		})
		convey.Convey("03-service with default port should return right ip and port", func() {
			svc.Spec.Ports[0] = defaultPort
			ip, port := getServiceIpAndPort(svc)
			convey.ShouldEqual(ip, "127.0.0.1")
			convey.ShouldEqual(port, strconv.Itoa(mindxdlv1.DefaultPort))
		})
	})
}

func TestGetMangerSvc(t *testing.T) {
	convey.Convey("getMangerSvc", t, func() {
		rc := &ASJobReconciler{}
		services := make([]*corev1.Service, 1)
		convey.Convey("01-empty services should return nil", func() {
			res := rc.getMangerSvc([]*corev1.Service{})
			convey.ShouldBeNil(res)
		})
		convey.Convey("02-services with svc which has no label should return nil", func() {
			services[0] = &corev1.Service{}
			res := rc.getMangerSvc(services)
			convey.ShouldBeNil(res)
		})
		convey.Convey("03-services with svc which has no require label should return nil", func() {
			services[0] = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"xxx": "yyy"},
				},
			}
			res := rc.getMangerSvc(services)
			convey.ShouldBeNil(res)
		})
		convey.Convey("04-services with svc which has worker label should return nil", func() {
			services[0] = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						commonv1.ReplicaTypeLabel: strings.ToLower(string(mindxdlv1.ReplicaTypeWorker)),
					},
				},
			}
			res := rc.getMangerSvc(services)
			convey.ShouldBeNil(res)
		})
		convey.Convey("04-services with svc which has master label should return not be nil", func() {
			services[0] = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						commonv1.ReplicaTypeLabel: strings.ToLower(string(mindxdlv1.PytorchReplicaTypeMaster)),
					},
				},
			}
			res := rc.getMangerSvc(services)
			convey.ShouldNotBeNil(res)
		})
	})
}

func TestGetMngSvcIpAndPortWithError(t *testing.T) {
	convey.Convey("getMngSvcIpAndPort", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		convey.Convey("01-get job ref services failed, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getOrCreateSvc",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) (*corev1.Service, error) {
					return nil, errors.New("not found")
				})
			defer patch.Reset()
			_, _, err := rc.getMngSvcIpAndPort(job, mindxdlv1.PytorchFrameworkName, "")
			convey.ShouldEqual(err, errors.New("not found"))
		})
		convey.Convey("02-job has no manager svc, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getOrCreateSvc",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) (*corev1.Service, error) {
					return &corev1.Service{ObjectMeta: metav1.ObjectMeta{
						Labels: make(map[string]string),
					}}, nil
				})
			defer patch.Reset()
			_, _, err := rc.getMngSvcIpAndPort(job, mindxdlv1.PytorchFrameworkName, "")
			convey.ShouldEqual(err, fmt.Errorf("get job<%s/%s> chief service failed", job.Namespace, job.Name))
		})
		convey.Convey("03-service with manager label has no ip should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getOrCreateSvc",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) (*corev1.Service, error) {
					return &corev1.Service{ObjectMeta: metav1.ObjectMeta{
						Labels: make(map[string]string),
					}}, nil
				})
			defer patch.Reset()
			ip, port, err := rc.getMngSvcIpAndPort(job, mindxdlv1.PytorchFrameworkName, "")
			convey.ShouldEqual(err, fmt.Errorf("job<%s/%s> chief service Ip<%s> or port<%s> is empty", job.Namespace, job.Name,
				ip, port))
		})
		convey.Convey("04-mindspore single npu task should return empty", func() {
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{mindxdlv1.ReplicaTypeWorker: {}}
			ip, port, err := rc.getMngSvcIpAndPort(job, mindxdlv1.MindSporeFrameworkName, mindxdlv1.ReplicaTypeWorker)
			convey.So(ip, convey.ShouldEqual, "")
			convey.So(port, convey.ShouldEqual, "")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetMngSvcIpAndPortNormal(t *testing.T) {
	convey.Convey("getMngSvcIpAndPort", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		convey.Convey("01-service with manager label has ip and port should not return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getOrCreateSvc",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) (*corev1.Service, error) {
					return &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{commonv1.ReplicaTypeLabel: "master"},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{{
								Port: 2222,
							}},
							ClusterIP: "127.0.0.1",
						},
					}, nil
				})
			defer patch.Reset()
			ip, port, err := rc.getMngSvcIpAndPort(job, mindxdlv1.PytorchFrameworkName, "")
			convey.ShouldBeNil(err)
			convey.ShouldEqual(ip, "127.0.0.1")
			convey.ShouldEqual(port, "2222")
		})
	})
}
