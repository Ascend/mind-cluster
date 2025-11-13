/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package utils is using for generating ranktable.
*/
package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	"ascend-operator/pkg/ranktable/utils"
	_ "ascend-operator/pkg/testtool"
)

func TestWriteToFile(t *testing.T) {
	convey.Convey("TestWriteToFile", t, func() {
		job := &mindxdlv1.AscendJob{}
		convey.Convey("01-empty dir should return error", func() {
			gen := New(job)
			err := gen.WriteToFile()
			convey.So(err, convey.ShouldBeNil)
		})
		patch := gomonkey.ApplyFunc(utils.GenRankTableDir, func(ascendJob *mindxdlv1.AscendJob) string {
			return "./"
		})
		defer patch.Reset()
		defer func() {
			err := os.Remove("./hccl.json")
			convey.So(err, convey.ShouldBeNil)
			err = os.Remove("./version")
			convey.So(err, convey.ShouldBeNil)
		}()
		gen := New(job)
		convey.Convey("02-open file failed should return error", func() {
			patch1 := gomonkey.ApplyFunc(os.OpenFile, func(string, int, os.FileMode) (*os.File, error) {
				return nil, errors.New("open file failed")
			})
			defer patch1.Reset()
			err := gen.WriteToFile()
			convey.So(err, convey.ShouldResemble, errors.New("open file failed"))
		})
		convey.Convey("03-write to file success return nil", func() {
			err := gen.WriteToFile()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDeleteFile(t *testing.T) {
	convey.Convey("TestDeleteFile", t, func() {
		job := &mindxdlv1.AscendJob{}
		convey.Convey("01-empty dir should return error", func() {
			gen := New(job)
			err := gen.DeleteFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-exist file will be remove", func() {
			patch := gomonkey.ApplyFunc(utils.GenRankTableDir, func(ascendJob *mindxdlv1.AscendJob) string {
				return "./"
			})
			defer patch.Reset()
			const defaultPerm = 0644
			file, err := os.OpenFile("./hccl.json", os.O_RDWR|os.O_CREATE, defaultPerm)
			convey.So(err, convey.ShouldBeNil)
			_, err = file.WriteString("xxxx")
			convey.So(err, convey.ShouldBeNil)
			err = file.Close()
			convey.So(err, convey.ShouldBeNil)
			gen := New(job)
			err = gen.DeleteFile()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGatherServerList(t *testing.T) {
	convey.Convey("TestGatherServerList", t, func() {
		job := &mindxdlv1.AscendJob{}
		gen := New(job)
		pod1 := &v1.Pod{}
		pod1.UID = "111"
		pod1.Status.PodIP = "192.168.1.1"
		pod1.Annotations = make(map[string]string)
		inst1 := newInstanceString("pod1", "127.0.0.1")
		pod1.Annotations[api.Pod910DeviceAnno] = inst1
		pod1.Annotations[api.PodRankIndexAnno] = "0"
		pod2 := &v1.Pod{}
		pod2.Status.PodIP = "192.168.1.2"
		pod2.UID = "222"
		pod2.Annotations = make(map[string]string)
		inst2 := newInstanceString("pod2", "127.0.0.2")
		pod2.Annotations[api.Pod910DeviceAnno] = inst2
		pod2.Annotations[api.PodRankIndexAnno] = "1"
		err := gen.AddPod(pod1)
		convey.So(err, convey.ShouldBeNil)
		err = gen.AddPod(pod2)
		convey.So(err, convey.ShouldBeNil)
		gen.GatherServerList()
		expected := 2
		convey.So(len(gen.ServerList), convey.ShouldEqual, expected)
		convey.So(gen.ServerList[0].ServerID, convey.ShouldEqual, "127.0.0.1")

		expected = 0
		gen.ServerList = nil
		convey.So(len(gen.ServerList), convey.ShouldEqual, expected)
		gen.ServerList = make([]*common.Server, 0)
		convey.So(len(gen.ServerList), convey.ShouldEqual, expected)
	})
}

func TestAddPod(t *testing.T) {
	convey.Convey("TestAddPod", t, func() {
		job := &mindxdlv1.AscendJob{}
		job.Annotations = map[string]string{"sp-block": "2"}
		job.Labels = map[string]string{"app": "mindie-ms-server", "jobID": "mindie-test"}
		gen := New(job)
		pod := &v1.Pod{}
		convey.Convey("01-pod without device-key annotation should return nil", func() {
			err := gen.AddPod(pod)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-json unmarshal failed should return error", func() {
			pod.Annotations = map[string]string{api.Pod910DeviceAnno: ""}
			err := gen.AddPod(pod)
			convey.So(err, convey.ShouldNotBeNil)
		})
		pod.Annotations = map[string]string{
			api.Pod910DeviceAnno: newInstanceString("pod1", "127.0.0.1")}
		convey.Convey("03-pod without ip should return error", func() {
			err := gen.AddPod(pod)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("pod(%s/%s) ip is empty", pod.Namespace, pod.Name))
		})
		pod.Status.PodIP = "192.168.1.1"
		convey.Convey("04-pod without rankIndex should return error", func() {
			err := gen.AddPod(pod)
			convey.So(err, convey.ShouldNotBeNil)
		})
		pod.Annotations[api.PodRankIndexAnno] = "0"
		convey.Convey("05-add pod success should return nil", func() {
			err := gen.AddPod(pod)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func newInstanceString(podName string, serverID string) string {
	inst := common.Instance{
		PodName:    podName,
		ServerID:   serverID,
		SuperPodId: 0,
		Devices: []common.Dev{
			{
				DeviceID: "0",
				DeviceIP: "127.0.1.1",
			},
		},
	}
	bt, err := json.Marshal(inst)
	if err != nil {
		return ""
	}
	return string(bt)
}
