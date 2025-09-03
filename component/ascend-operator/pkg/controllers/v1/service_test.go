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

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

const (
	fakeSvcClusterIP = "127.0.0.1"
	fakePort         = 8080
)

// TestGetServiceIpAndPort test getServiceIpAndPort
func TestGetServiceIpAndPort(t *testing.T) {
	convey.Convey("getServiceIpAndPort", t, func() {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{Ports: make([]corev1.ServicePort, 1), ClusterIP: fakeSvcClusterIP},
		}
		defaultPort := corev1.ServicePort{
			Name: api.DefaultPortName,
			Port: mindxdlv1.DefaultPort,
		}
		convey.Convey("01-service with no port should return empty ip and port", func() {
			ip, port := getServiceIpAndPort(svc)
			convey.So(ip, convey.ShouldEqual, "")
			convey.So(port, convey.ShouldEqual, "")
		})
		convey.Convey("02-service with no default port should return empty ip and port", func() {
			svc.Spec.Ports[0] = corev1.ServicePort{
				Name: "fake",
				Port: 0,
			}
			ip, port := getServiceIpAndPort(svc)
			convey.So(ip, convey.ShouldEqual, "")
			convey.So(port, convey.ShouldEqual, "")
		})
		convey.Convey("03-service with default port should return right ip and port", func() {
			svc.Spec.Ports[0] = defaultPort
			ip, port := getServiceIpAndPort(svc)
			convey.So(ip, convey.ShouldEqual, fakeSvcClusterIP)
			convey.So(port, convey.ShouldEqual, strconv.Itoa(mindxdlv1.DefaultPort))
		})
		convey.Convey("04-nil service should return right empty ip and port", func() {
			ip, port := getServiceIpAndPort(nil)
			convey.So(ip, convey.ShouldEqual, "")
			convey.So(port, convey.ShouldEqual, "")
		})
	})
}

// TestGetMangerSvc test getMangerSvc
func TestGetMangerSvc(t *testing.T) {
	convey.Convey("getMangerSvc", t, func() {
		rc := &ASJobReconciler{}
		services := make([]*corev1.Service, 1)
		convey.Convey("01-empty services should return nil", func() {
			res := rc.getMangerSvc([]*corev1.Service{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("02-services with svc which has no label should return nil", func() {
			services[0] = &corev1.Service{}
			res := rc.getMangerSvc(services)
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("03-services with svc which has no require label should return nil", func() {
			services[0] = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"xxx": "yyy"},
				},
			}
			res := rc.getMangerSvc(services)
			convey.So(res, convey.ShouldBeNil)
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
			convey.So(res, convey.ShouldBeNil)
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
			convey.So(res, convey.ShouldNotBeNil)
		})
	})
}

// TestGetMngSvcIpAndPortWithError test getMngSvcIpAndPort with error
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
			convey.So(err, convey.ShouldResemble, errors.New("not found"))
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
			convey.So(err, convey.ShouldResemble, fmt.Errorf("job<%s/%s> chief service Ip<> or port<> is empty",
				job.Namespace, job.Name))
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
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("job<%s/%s> chief service Ip<%s> or port<%s> is empty",
					job.Namespace, job.Name, ip, port))
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

// TestGetMngSvcIpAndPortNormal get service ip and port
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
								Name: api.DefaultPortName,
							}},
							ClusterIP: fakeSvcClusterIP,
						},
					}, nil
				})
			defer patch.Reset()
			ip, port, err := rc.getMngSvcIpAndPort(job, mindxdlv1.PytorchFrameworkName, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(ip, convey.ShouldEqual, fakeSvcClusterIP)
			convey.So(port, convey.ShouldEqual, "2222")
		})
	})
}

// TestGetIpFromSvcName get service ip from service name
func TestGetIpFromSvcName(t *testing.T) {
	convey.Convey("getIpFromSvcName", t, func() {
		rc := &ASJobReconciler{}
		defaultDomain := "test.svc.cluster.local"
		convey.Convey("01-get service from api-server failed will return defaultDomain", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver",
				func(_ *ASJobReconciler, _, _ string) (*corev1.Service, error) {
					return nil, fmt.Errorf("get service failed")
				})
			defer patch.Reset()
			ip := rc.getIpFromSvcName("svc", "default", defaultDomain)
			convey.So(ip, convey.ShouldEqual, defaultDomain)
		})
		convey.Convey("02-get service from api-server success will return service ip", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver",
				func(_ *ASJobReconciler, _, _ string) (*corev1.Service, error) {
					return &corev1.Service{
						Spec: corev1.ServiceSpec{
							ClusterIP: fakeSvcClusterIP,
						},
					}, nil
				})
			defer patch.Reset()
			ip := rc.getIpFromSvcName("svc", "default", defaultDomain)
			convey.So(ip, convey.ShouldEqual, fakeSvcClusterIP)
		})
	})
}

// TestGetClusterDSvcIp test getClusterDSvcIp
func TestGetClusterDSvcIp(t *testing.T) {
	convey.Convey("getClusterDSvcIp", t, func() {
		rc := &ASJobReconciler{}
		convey.Convey("01-get service from api-server success will return service ip", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getIpFromSvcName",
				func(_ *ASJobReconciler, _, _, _ string) string {
					return fakeSvcClusterIP
				})
			defer patch.Reset()
			ip := rc.getClusterDSvcIp()
			convey.So(ip, convey.ShouldEqual, fakeSvcClusterIP)
		})
	})
}

// TestGenServiceLabels test genServiceLabels
func TestGenServiceLabels(t *testing.T) {
	convey.Convey("genServiceLabels", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		convey.Convey("01-genServiceLabels", func() {
			fakeType := "master"
			fakeIndex := "0"
			expectedLables := map[string]string{
				commonv1.OperatorNameLabel:           "ascendjob-controller",
				commonv1.GroupNameLabelDeprecated:    "mindxdl.gitee.com",
				commonv1.JobNameLabel:                "ascendjob-test",
				commonv1.JobNameLabelDeprecated:      "ascendjob-test",
				commonv1.ReplicaTypeLabel:            fakeType,
				commonv1.ReplicaTypeLabelDeprecated:  fakeType,
				commonv1.ReplicaIndexLabel:           fakeIndex,
				commonv1.ReplicaIndexLabelDeprecated: fakeIndex,
			}
			labelMap := rc.genServiceLabels(job, commonv1.ReplicaType(fakeType), fakeIndex)
			convey.So(labelMap, convey.ShouldResemble, expectedLables)
		})
	})
}

// TestGenServicePorts test genServicePorts
func TestGenServicePorts(t *testing.T) {
	convey.Convey("genServicePorts", t, func() {
		rc := newCommonReconciler()
		spec := newCommonSpec()
		container := newCommonContainer()
		convey.Convey("01-spec without default container will return err", func() {
			spec.Template.Spec.Containers[0] = container
			ports, err := rc.genServicePorts(spec)
			convey.So(ports, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("01-spec with default container will return correct ports", func() {
			container.Name = api.DefaultContainerName
			spec.Template.Spec.Containers[0] = container
			expectedPorts := []corev1.ServicePort{
				{
					Name: api.DefaultPortName,
					Port: fakePort,
				},
			}
			ports, err := rc.genServicePorts(spec)
			convey.So(ports, convey.ShouldResemble, expectedPorts)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGenService test genService
func TestGenService(t *testing.T) {
	convey.Convey("genService", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		fakeType := "Master"
		spec := newCommonSpec()
		container := newCommonContainer()
		container.Name = api.DefaultContainerName
		spec.Template.Spec.Containers[0] = container
		convey.Convey("01-genServicePorts failed will return error", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "genServicePorts",
				func(_ *commonv1.ReplicaSpec) ([]corev1.ServicePort, error) {
					return nil, errors.New("genServicePorts failed")
				})
			defer patch.Reset()
			service, err := rc.genService(job, commonv1.ReplicaType(fakeType), spec)
			convey.So(service, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-genServiceLabels failed will return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "genServicePorts",
				func(_ *commonv1.ReplicaSpec) ([]corev1.ServicePort, error) {
					return []corev1.ServicePort{
						{
							Name: api.DefaultPortName,
							Port: fakePort,
						},
					}, nil
				})
			defer patch1.Reset()
			service, _ := rc.genService(job, commonv1.ReplicaType(fakeType), spec)
			convey.So(service, convey.ShouldNotBeNil)
		})
		convey.Convey("03-genService will return service", func() {
			_, err := rc.genService(job, commonv1.ReplicaType(fakeType), spec)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func newCommonSpec() *commonv1.ReplicaSpec {
	return &commonv1.ReplicaSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: corev1.PodSpec{
				Containers: make([]corev1.Container, 1),
			},
		},
	}
}
