/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.


   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taskd/common/constant"
)

const (
	fileMode = 0644
	dirMode  = 0755
)

func TestSaveProfilingDataIntoFileEmptyRecords(t *testing.T) {
	t.Run("empty records not save", func(t *testing.T) {

		ProfileRecordsKernel = make([]MsptiActivityKernel, 0)
		ProfileRecordsApi = make([]MsptiActivityApi, 0)
		ProfileRecordsMark = make([]MsptiActivityMark, 0)

		err := SaveProfilingDataIntoFile(0)

		assert.NoError(t, err)
	})
}

func TestSaveProfilingDataIntoFileWriteOK(t *testing.T) {
	t.Run("records ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		filePath := path.Join(t.TempDir(), "test_task_uid", strconv.Itoa(GlobalRankId))
		err := os.WriteFile(filePath, make([]byte, 0), fileMode)
		if err != nil {
			return
		}
		patches.ApplyFunc(getCurrentSavePath, func(rank int) (string, error) {
			return filePath, nil
		})

		patches.ApplyFunc(getNewestFileName, func(filePath string) (string, error) {
			return fmt.Sprintf("%d", time.Now().Unix()), nil
		})

		patches.ApplyFunc(saveProfileFile, func(file *os.File) error {
			return nil
		})

		patches.ApplyGlobalVar(&ProfileRecordsKernel, []MsptiActivityKernel{{}})

		err = SaveProfilingDataIntoFile(0)
		assert.NoError(t, err)
	})
}

func TestGetNewestFileNameEmptyDir(t *testing.T) {

	t.Run("empty directory get filename", func(t *testing.T) {
		filename, err := getNewestFileName(t.TempDir())
		assert.NoError(t, err)
		assert.Regexp(t, `^\d{10}$`, filename)
	})

}

func TestGetNewestFileNameWithFiles(t *testing.T) {

	t.Run("with files get filename", func(t *testing.T) {
		// Create test files
		testDir := t.TempDir()
		createTestFile(t, testDir, "1620000000") // 2021-05-03
		createTestFile(t, testDir, "1625000000") // 2021-06-30
		createTestFile(t, testDir, "invalid_name")

		filename, err := getNewestFileName(testDir)
		assert.NoError(t, err)
		assert.Equal(t, "1625000000", filename)
	})

}

func TestGetNewestFileNameExceedsLimit(t *testing.T) {

	t.Run("file size exceeds limit get filename", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(isFileOver10MB, func(filePath string) (bool, error) {
			return true, nil
		})

		testDir := t.TempDir()
		createTestFile(t, testDir, "1626000000")
		filename, err := getNewestFileName(testDir)
		assert.NoError(t, err)
		assert.NotEqual(t, "1626000000", filename)
	})
}

func TestIsFileOver10MB(t *testing.T) {
	t.Run("file size exceeds limit", func(t *testing.T) {
		testDir := t.TempDir()
		largeFilePath := filepath.Join(testDir, "1628000000")
		err := os.WriteFile(largeFilePath, make([]byte, constant.SizeLimitPerProfilingFile+1), fileMode)
		require.NoError(t, err)
		isOver, err := isFileOver10MB(largeFilePath)
		assert.NoError(t, err)
		assert.True(t, isOver)
	})
}

func TestGetCurrentSavePathWithUID(t *testing.T) {
	// Test case 1: With environment variable
	t.Run("with task UID", func(t *testing.T) {
		err := os.Setenv(constant.TaskUidKey, "test_task_123")
		assert.NoError(t, err)
		defer func(key string) {
			err := os.Unsetenv(key)
			if err != nil {
				assert.NoError(t, err)
			}
		}(constant.TaskUidKey)

		savePath, err := getCurrentSavePath(0)
		assert.NoError(t, err)
		assert.Contains(t, savePath, "test_task_123/0")
	})
}

func TestGetCurrentSavePathWithoutUID(t *testing.T) {
	// Test case 2: Without environment variable
	t.Run("without task UID", func(t *testing.T) {
		savePath, err := getCurrentSavePath(1)
		assert.NoError(t, err)
		assert.Contains(t, savePath, "default_task_id_")
		assert.Contains(t, savePath, "/1")
	})

}

func TestGetCurrentSavePathTooLongPath(t *testing.T) {
	t.Run("path too long", func(t *testing.T) {

		err := os.Setenv(constant.TaskUidKey, strings.Repeat("1", constant.PathLengthLimit+1))
		if err != nil {
			assert.NoError(t, err)
		}
		defer func(key string) {
			err := os.Unsetenv(key)
			if err != nil {
				assert.NoError(t, err)
			}
		}(constant.TaskUidKey)

		_, err = getCurrentSavePath(0)
		assert.Error(t, err)
	})
}

func TestGetDirSizeInMB(t *testing.T) {
	t.Run("get dir size with file", func(t *testing.T) {
		testDir := t.TempDir()
		createTestFiles(t, testDir, constant.BytesPerMB, 1)
		size, err := getDirSizeInMB(testDir)
		assert.NoError(t, err)
		assert.Greater(t, int(size), 0)
	})
}

func TestGetProfileFiles(t *testing.T) {
	t.Run("get profile files", func(t *testing.T) {
		tmpDir := t.TempDir()
		proDir := filepath.Join(tmpDir, strconv.Itoa(GlobalRankId))
		err := os.Mkdir(proDir, dirMode)
		if err != nil {
			t.Errorf("create dir failed: %v", err)
			return
		}

		fileCount := 3
		for i := 1; i <= fileCount; i++ {
			createTestFile(t, proDir, "profile"+strconv.Itoa(i))
		}
		profileFiles, err := getProfileFiles(tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, len(profileFiles), fileCount)
		assert.Equal(t, profileFiles[0].Name(), "profile1")
	})
}

func TestDeleteOldestFileForEachRank(t *testing.T) {
	t.Run("delete profile files", func(t *testing.T) {
		tmpDir := t.TempDir()
		profileDir := filepath.Join(tmpDir, strconv.Itoa(GlobalRankId))
		err := os.Mkdir(profileDir, dirMode)
		if err != nil {
			t.Errorf("create dir failed: %v", err)
			return
		}
		fileCount := 10
		for i := 0; i < fileCount; i++ {

		}
		createTestFiles(t, profileDir, fileCount, fileCount)
		err = deleteOldestFileForEachRank(tmpDir)
		assert.NoError(t, err)
		files, err := os.ReadDir(profileDir)
		assert.NoError(t, err)
		assert.Less(t, len(files), fileCount)
	})
}

func TestManageSaveProfiling(t *testing.T) {
	t.Run("save profile files ok", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		tmpDir := t.TempDir()
		patches.ApplyFunc(getCurrentSavePath, func(rank int) (string, error) {
			return tmpDir, nil
		})

		fileName := fmt.Sprintf("%d", time.Now().Unix())
		patches.ApplyFunc(getNewestFileName, func(filePath string) (string, error) {
			return fileName, nil
		})

		ProfileRecordsKernel = []MsptiActivityKernel{{}}
		go ManageSaveProfiling(ctx)
		time.Sleep(constant.CheckProfilingCacheInterval + time.Second)
		file, err := os.Stat(filepath.Join(tmpDir, fileName))
		assert.NoError(t, err)
		assert.Greater(t, file.Size(), int64(0))
	})

}

func TestManageSaveProfilingPanic(t *testing.T) {
	t.Run("manage profile files panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(SaveProfilingDataIntoFile, func(rank int) error {
			mu.Lock()
			called = true
			mu.Unlock()
			panic("panic error")
			return nil
		})

		go ManageSaveProfiling(ctx)
		time.Sleep(time.Second)
		cancel()
		mu.Lock()
		assert.True(t, called)
		mu.Unlock()
	})

}

func TestManageProfilingDiskUsagePanic(t *testing.T) {
	t.Run("manage disk files panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(getDirSizeInMB, func(path string) (float64, error) {
			mu.Lock()
			called = true
			mu.Unlock()
			panic("panic error")
			return 0.0, nil
		})
		testBaseDir := t.TempDir()
		go ManageProfilingDiskUsage(testBaseDir, ctx)
		time.Sleep(time.Second)
		cancel()
		mu.Lock()
		assert.True(t, called)
		mu.Unlock()
	})

}

func TestManageProfilingDiskUsage(t *testing.T) {
	t.Run("get profile files", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(dealWithDiskUsage, func(baseDir string, usedSize float64) {
			filePath := filepath.Join(baseDir, strconv.Itoa(0))
			_, err := os.Stat(filePath)
			if err != nil {
				return
			}
			err = os.Remove(filePath)
			assert.NoError(t, err)
		})
		// Setup
		testBaseDir := t.TempDir()
		SetDiskUsageUpperLimitMB(1)

		// Create test files exceeding limit
		fileCount := 3
		createTestFiles(t, testBaseDir, 1024*1024, fileCount)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Execute
		go ManageProfilingDiskUsage(testBaseDir, ctx)
		time.Sleep(constant.DiskUsageCheckInterval * time.Second) // Allow time for cleanup

		// Verify
		usedSize, err := getDirSizeInMB(testBaseDir)
		assert.NoError(t, err)
		assert.Less(t, usedSize, float64(fileCount))
	})
}

func TestWriteToBytes(t *testing.T) {
	t.Run("test records is null after write bytes", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Setup test data
		patches.ApplyGlobalVar(&ProfileRecordsMark, []MsptiActivityMark{
			{},
		})
		patches.ApplyGlobalVar(&ProfileRecordsApi, []MsptiActivityApi{
			{},
		})
		patches.ApplyGlobalVar(&ProfileRecordsKernel, []MsptiActivityKernel{
			{},
		})

		// Execute
		writeToBytes()

		// Verify
		assert.Equal(t, 0, len(ProfileRecordsMark))
		assert.Equal(t, 0, len(ProfileRecordsApi))
		assert.Equal(t, 0, len(ProfileRecordsKernel))
	})
}

func TestDealWithDiskUsageReadDirFailed(t *testing.T) {
	t.Run("test read dir failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(ioutil.ReadDir, func(dirname string) ([]fs.FileInfo, error) {
			return nil, errors.New("read dir failed")
		})
		testDir := t.TempDir()
		// Create test files exceeding limit
		fileCount := 3
		createTestFiles(t, testDir, 1, fileCount)
		dealWithDiskUsage(testDir, 1)
		files, err := os.ReadDir(testDir)
		assert.NoError(t, err)
		assert.Equal(t, len(files), fileCount)
	})
}

func TestDealWithDiskUsageGetDirFailed(t *testing.T) {
	t.Run("get dir failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(deleteOldestFileForEachRank, func(dirname string) error {
			return errors.New("read dir failed")
		})
		testDir := t.TempDir()
		// Create test files exceeding limit
		fileCount := 3
		createTestFiles(t, testDir, 1, fileCount)
		dealWithDiskUsage(testDir, 1)
		files, err := os.ReadDir(testDir)
		assert.NoError(t, err)
		assert.Equal(t, len(files), fileCount)
	})
}

func TestDealWithDiskUsageDeleteFileOK(t *testing.T) {
	t.Run("delete files ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(deleteOldestFileForEachRank, func(dirname string) error {
			files, err := os.ReadDir(dirname)
			assert.NoError(t, err)
			for _, file := range files {
				err := os.Remove(filepath.Join(dirname, file.Name()))
				assert.NoError(t, err)
			}
			return nil
		})
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "testDir")
		err := os.Mkdir(testDir, dirMode)
		assert.NoError(t, err)
		// Create test files exceeding limit
		fileCount := 3
		createTestFiles(t, testDir, 1, fileCount)
		dealWithDiskUsage(tmpDir, 1)
		files, err := os.ReadDir(testDir)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(files))
	})
}

func createTestFile(t *testing.T, dir string, filename string) {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte("test content"), fileMode)
	assert.NoError(t, err)
}

func createTestFiles(t *testing.T, dir string, size int, count int) {
	for i := 0; i < count; i++ {
		filePath := filepath.Join(dir, strconv.Itoa(i))
		err := os.WriteFile(filePath, make([]byte, size), fileMode)
		assert.NoError(t, err)
	}
}
