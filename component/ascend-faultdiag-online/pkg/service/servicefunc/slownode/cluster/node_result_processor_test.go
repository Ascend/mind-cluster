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

// Package cluster a DT collection for func in node_result_processor
package cluster

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/dataparse"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

const (
	testDir      = "testdata"
	testFileName = "testfile.json"
)

func TestDataProfilingResultProcessor(t *testing.T) {
	var profile = &slownode.NodeDataProfilingResult{
		Namespace:                "testNamespace",
		FinishedInitialProfiling: true,
	}
	profile.JobId = "test-job-id"
	profile.JobName = "test-job-name"

	var ctx = &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "test-job-id"
	var channelCapacity = 10
	ctx.NodeReportSignal = make(chan struct{}, channelCapacity)
	slownodejob.GetJobCtxMap().Clear()
	convey.Convey("test DataProfilingResultProcessor", t, func() {
		testDataProfilingResultProcessorWithError(profile, ctx)
		testDataProfilingResultProcessorWithSuccess(profile, ctx)
	})
}

func testDataProfilingResultProcessorWithError(
	profile *slownode.NodeDataProfilingResult,
	ctx *slownodejob.JobContext,
) {
	defer slownodejob.GetJobCtxMap().Clear()
	convey.Convey("test DataProfilingResultProcessor with error", func() {
		// operator is delete
		DataProfilingResultProcessor(nil, nil, watch.Deleted)
		// profiling is not finished
		DataProfilingResultProcessor(nil, &slownode.NodeDataProfilingResult{}, watch.Added)
		// no ctx found
		DataProfilingResultProcessor(nil, profile, watch.Added)
		// ctx found, but not runing
		slownodejob.GetJobCtxMap().Insert(profile.KeyGenerator(), ctx)
		DataProfilingResultProcessor(nil, profile, watch.Added)
		// job is running, but wrong steps
		setUnexportedFiled(ctx, "isRunning", true)
		DataProfilingResultProcessor(nil, profile, watch.Added)
		// write file failed
		setUnexportedFiled(ctx, "step", slownodejob.ClusterStep1)
		patches := gomonkey.ApplyMethod(
			reflect.TypeOf(&dataparse.Controller{}),
			"MergeParallelGroupInfoWatcher",
			func(*dataparse.Controller) {
				fmt.Println("mock MergeParallelGroupInfoWatcher")
			})
		patches.ApplyFunc(writeFile, func(string, string, map[string]any) error {
			fmt.Println("mock write file failed")
			return errors.New("mock write file failed")
		})
		defer patches.Reset()
		output := captureOutput(func() {
			DataProfilingResultProcessor(nil, profile, watch.Added)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock MergeParallelGroupInfoWatcher")
		convey.So(output, convey.ShouldContainSubstring, "mock write file failed")
		jobOnceMap.Delete(profile.JobId)
	})

}

func testDataProfilingResultProcessorWithSuccess(
	profile *slownode.NodeDataProfilingResult,
	ctx *slownodejob.JobContext,
) {
	slownodejob.GetJobCtxMap().Insert(profile.KeyGenerator(), ctx)
	setUnexportedFiled(ctx, "isRunning", true)
	setUnexportedFiled(ctx, "step", slownodejob.ClusterStep1)
	patches := gomonkey.ApplyMethod(
		reflect.TypeOf(&dataparse.Controller{}),
		"MergeParallelGroupInfoWatcher",
		func(*dataparse.Controller) {
			fmt.Println("mock MergeParallelGroupInfoWatcher")
		})
	patches.ApplyFunc(writeFile, func(string, string, map[string]any) error {
		fmt.Println("mock write file")
		return nil
	})
	defer patches.Reset()
	convey.Convey("test testDataProfilingResultProcessor with success", func() {
		// call 2 times, only the first time call merge parallel group info watcher
		output := captureOutput(func() {
			DataProfilingResultProcessor(nil, profile, watch.Added)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock MergeParallelGroupInfoWatcher")
		convey.So(output, convey.ShouldContainSubstring, "mock write file")
		// second time
		output = captureOutput(func() {
			DataProfilingResultProcessor(nil, profile, watch.Added)
		})
		convey.So(output, convey.ShouldNotContainSubstring, "mock MergeParallelGroupInfoWatcher")
		convey.So(output, convey.ShouldContainSubstring, "mock write file")
	})
}

func TestAlgoResultProcessor(t *testing.T) {
	patch := gomonkey.ApplyFunc(writeFile, func(string, string, map[string]any) error {
		fmt.Println("mock write file")
		return nil
	})
	defer patch.Reset()
	var algoResult = &slownode.NodeAlgoResult{
		Namespace: "testNamespace",
		NodeRank:  "127.0.1",
	}
	algoResult.JobId = "test-job-id"
	algoResult.JobName = "test-job-name"
	var ctx = &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "test-job-id"
	slownodejob.GetJobCtxMap().Clear()
	convey.Convey("test AlgoResultProcessor", t, func() {
		// operator is delete
		AlgoResultProcessor(nil, nil, watch.Deleted)
		// no ctx found
		AlgoResultProcessor(nil, algoResult, watch.Added)
		// ctx is not running
		slownodejob.GetJobCtxMap().Insert(algoResult.KeyGenerator(), ctx)
		AlgoResultProcessor(nil, algoResult, watch.Added)
		// no nodeRank found -> deprecated
		setUnexportedFiled(ctx, "isRunning", true)
		AlgoResultProcessor(nil, algoResult, watch.Added)
		// add the ip
		ctx.AddReportedNodeIp(algoResult.NodeRank)
		output := captureOutput(func() {
			AlgoResultProcessor(nil, algoResult, watch.Added)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock write file")
	})
}

func TestWriteFile(t *testing.T) {
	convey.Convey("test clusterWriteFile func", t, func() {
		// make sure the test directory is clean before running tests
		convey.So(os.RemoveAll(testDir), convey.ShouldBeNil)
		defer os.RemoveAll(testDir)

		testData := map[string]any{
			"key": "value",
			"num": 123,
		}

		testCreateDirectoryAndWriteFile(testData)
		testDirectoryAlreadyExists(testData)
		testSymlinkPath(testData)
		testJsonMarshalFailure()
	})
}

func testCreateDirectoryAndWriteFile(testData map[string]any) {
	convey.Convey("dir is not existed", func() {
		targetDir := filepath.Join(testDir, "newdir")
		targetFile := filepath.Join(targetDir, testFileName)

		convey.Convey("create file and write successfully", func() {
			err := writeFile(targetDir, testFileName, testData)
			convey.So(err, convey.ShouldBeNil)
			_, err = os.Stat(targetDir)
			convey.So(err, convey.ShouldBeNil)

			_, err = os.Stat(targetFile)
			convey.So(err, convey.ShouldBeNil)

			content, err := os.ReadFile(targetFile)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(content), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func testDirectoryAlreadyExists(testData map[string]any) {
	convey.Convey("dir exists", func() {
		targetDir := filepath.Join(testDir, "existingdir")
		targetFile := filepath.Join(targetDir, testFileName)

		// pre-create the directory
		convey.So(os.MkdirAll(targetDir, os.ModePerm), convey.ShouldBeNil)

		convey.Convey("write data", func() {
			writeFile(targetDir, testFileName, testData)

			// verify the file exists and is written correctly
			_, err := os.Stat(targetFile)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func testSymlinkPath(testData map[string]any) {
	convey.Convey("file is symlink", func() {
		targetDir := filepath.Join(testDir, "realdir")
		symlinkDir := filepath.Join(testDir, "symlinkdir")

		// create the target directory and symlink
		convey.So(os.MkdirAll(targetDir, os.ModePerm), convey.ShouldBeNil)
		convey.So(os.Symlink(targetDir, symlinkDir), convey.ShouldBeNil)

		convey.Convey("write failed", func() {
			writeFile(symlinkDir, testFileName, testData)
			_, err := os.Stat(filepath.Join(symlinkDir, testFileName))
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})
}

func testJsonMarshalFailure() {
	convey.Convey("testfile.json marshal failed", func() {
		invalidData := map[string]any{
			"channel": make(chan int), // channels cannot be marshaled to JSON
		}

		convey.Convey("write failed", func() {

			targetDir := filepath.Join(testDir, "marshalerrordir")
			writeFile(targetDir, testFileName, invalidData)
			_, err := os.Stat(targetDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})
}
