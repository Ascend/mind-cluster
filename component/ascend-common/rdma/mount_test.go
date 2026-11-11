/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package rdma

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ascend-common/cdi/mount"
)

func TestReadRdmaMountEntryOptionalProviderFailures(t *testing.T) {
	optional := "/usr/lib64/libibverbs/libhrn5-rdmav34.so"
	for _, test := range []struct {
		entry string
		cause error
		fail  bool
	}{
		{optional, os.ErrNotExist, false},
		{optional, os.ErrPermission, true},
		{optional + ".1", os.ErrNotExist, true},
		{"/usr/lib64/libhrn5-rdmav34.so", os.ErrNotExist, true},
		{optional + ":/container/provider.so", os.ErrNotExist, true},
	} {
		lstat := func(path string) (os.FileInfo, error) {
			return nil, &os.PathError{Op: "lstat", Path: path, Err: test.cause}
		}
		_, _, err := readRdmaMountEntry(test.entry, lstat)
		if (err != nil) != test.fail {
			t.Errorf("entry %q with %v returned %v", test.entry, test.cause, err)
		}
	}
}

func TestReadRdmaListWithOnlyAbsentOptionalProvider(t *testing.T) {
	var sources []string
	lstat := func(sourcePath string) (os.FileInfo, error) {
		sources = append(sources, sourcePath)
		return nil, os.ErrNotExist
	}
	for _, entry := range []string{optionalRdmaProviderPath, "  " + optionalRdmaProviderPath + "  "} {
		files, dirs, err := readRdmaMountEntries(strings.NewReader(entry+"\n"), "optional.list", lstat)
		if err != nil || len(files) != 0 || len(dirs) != 0 {
			t.Fatalf("absent optional entry must be skipped: files=%v dirs=%v err=%v", files, dirs, err)
		}
	}
	if !reflect.DeepEqual(sources, []string{optionalRdmaProviderPath, optionalRdmaProviderPath}) {
		t.Fatalf("unexpected optional sources: %v", sources)
	}
}

func TestReadMountEntriesEmptyAndMissingList(t *testing.T) {
	dir := secureRdmaTestDir(t)
	path := filepath.Join(dir, "driver.list")
	for _, contents := range []string{"", "\n# intentionally empty\n\n"} {
		mustWriteFile(t, path, contents)
		if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
			t.Fatal("empty lists must fail")
		}
	}
	if err := readTestRdmaCollections(dir, []string{"../escape"}); err == nil {
		t.Fatal("collection traversal must fail before list path construction")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("missing list must fail")
	}
}

func TestReadMountEntriesChecksEveryMatchedCollection(t *testing.T) {
	dir := secureRdmaTestDir(t)
	mustWriteFile(t, filepath.Join(dir, rdmaDeviceTableName),
		"cc08 8200 0102 driver\ncc08 8200 0102 extra\n")
	sysfs := t.TempDir()
	replaceRdmaSysfs(t, sysfs)
	fakeRdmaHCA(t, sysfs, "hca0", "cc08", "8200", "0102")
	collections, err := DiscoverRdmaCollections(dir, []string{"hca0"})
	if err != nil || len(collections) != 2 {
		t.Fatalf("expected both collections, got %v: %v", collections, err)
	}
	source := filepath.Join(dir, "driver.so")
	mustWriteFile(t, source, "driver")
	mustWriteFile(t, filepath.Join(dir, "driver.list"), source+"\n")
	if err := readTestRdmaCollections(dir, collections); err == nil || !strings.Contains(err.Error(), "extra") {
		t.Fatalf("missing second list must fail: %v", err)
	}
	mustWriteFile(t, filepath.Join(dir, "extra.list"), filepath.Join(dir, "missing.so")+"\n")
	if err := readTestRdmaCollections(dir, collections); err == nil || !strings.Contains(err.Error(), "missing.so") {
		t.Fatalf("missing dependency in second list must fail: %v", err)
	}
	mustWriteFile(t, filepath.Join(dir, "extra.list"), source+"\n")
	if err := readTestRdmaCollections(dir, collections); err != nil {
		t.Fatal(err)
	}
}

func TestReadMountEntriesLiteralMappingAndBounds(t *testing.T) {
	dir := secureRdmaTestDir(t)
	source := filepath.Join(dir, "libdriver.so")
	mustWriteFile(t, source, "library")
	path := filepath.Join(dir, "driver.list")
	contents := "# ignored\n\n" + strings.Repeat(source+":/usr/lib64/libdriver.so\n", mount.MaxEntriesPerFile)
	mustWriteFile(t, path, contents)
	if err := readTestRdmaCollections(dir, []string{"driver"}); err != nil {
		t.Fatalf("128 effective entries and a mapped source must pass: %v", err)
	}
	mustWriteFile(t, path, contents+source+"\n")
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil ||
		!strings.Contains(err.Error(), "mount list too long") {
		t.Fatalf("129 effective entries must fail: %v", err)
	}
	for _, entry := range []string{
		filepath.Join(dir, "missing.so"), "relative.so", dir + "/../outside.so",
		source + ":relative", strings.Repeat("x", 70*1024),
	} {
		mustWriteFile(t, path, entry+"\n")
		if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
			t.Errorf("invalid/missing source %q must fail", entry)
		}
	}
}

func TestReadMountEntriesDriverSymlinkChains(t *testing.T) {
	dir := setupDriverSymlinkChain(t)
	if err := readTestRdmaCollections(dir, []string{"driver"}); err != nil {
		t.Fatalf("relative driver link chains must pass: %v", err)
	}
}

func TestReadMountEntriesLiteralSymlinkParent(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("symlink ownership cases require root")
	}
	dir := secureRdmaTestDir(t)
	realDir, linkedDir := filepath.Join(dir, "real"), filepath.Join(dir, "linked")
	requireRdmaTestSuccess(t, os.Mkdir(realDir, 0755))
	requireRdmaTestSuccess(t, os.Symlink(realDir, linkedDir))
	mustWriteFile(t, filepath.Join(realDir, "libdriver.so"), "library")
	mustWriteFile(t, filepath.Join(dir, "driver.list"), filepath.Join(linkedDir, "libdriver.so")+"\n")
	if err := readTestRdmaCollections(dir, []string{"driver"}); err != nil {
		t.Fatalf("literal sources within a root-owned symlink parent must pass: %v", err)
	}
}

func TestReadMountEntriesUnsafeDriverSymlinkChains(t *testing.T) {
	dir := setupDriverSymlinkChain(t)
	if err := os.Lchown(filepath.Join(dir, "libhrn5.so.1"), 1001, 1001); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("a listed non-root intermediate link must fail")
	}
	if err := os.Lchown(filepath.Join(dir, "libhrn5.so.1"), 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "libibv_extend.so.4.2.0")); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("a missing driver target must break the required link chain")
	}
}

func setupDriverSymlinkChain(t *testing.T) string {
	t.Helper()
	if os.Getuid() != 0 {
		t.Skip("driver symlink ownership cases require root")
	}
	dir := secureRdmaTestDir(t)
	for _, name := range []string{"hrn5.driver", "libhrn5-rdmav34.so", "libibv_extend.so.4.2.0"} {
		mustWriteFile(t, filepath.Join(dir, name), "driver")
	}
	for link, target := range map[string]string{
		"libhrn5.so":         "libhrn5.so.1",
		"libhrn5.so.1":       "libhrn5-rdmav34.so",
		"libibv_extend.so":   "libibv_extend.so.4",
		"libibv_extend.so.4": "libibv_extend.so.4.2.0",
	} {
		if err := os.Symlink(target, filepath.Join(dir, link)); err != nil {
			t.Fatal(err)
		}
	}
	var entries []string
	for _, name := range []string{"hrn5.driver", "libhrn5-rdmav34.so", "libhrn5.so", "libhrn5.so.1",
		"libibv_extend.so", "libibv_extend.so.4", "libibv_extend.so.4.2.0"} {
		entries = append(entries, filepath.Join(dir, name))
	}
	mustWriteFile(t, filepath.Join(dir, "driver.list"), strings.Join(entries, "\n")+"\n")
	return dir
}

func TestReadMountEntriesPermissions(t *testing.T) {
	dir := secureRdmaTestDir(t)
	source := filepath.Join(dir, "library.so")
	mustWriteFile(t, source, "library")
	path := filepath.Join(dir, "driver.list")
	mustWriteFile(t, path, source+"\n")
	if err := os.Chmod(path, 0666); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("writable-by-others list must fail")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(source, 0666); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("writable-by-others source must fail")
	}
}

func TestReadMountEntriesRootSymlinkFailures(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("root ownership cases require root")
	}
	dir := secureRdmaTestDir(t)
	source := filepath.Join(dir, "libdriver-v1.so")
	mustWriteFile(t, source, "library")
	link := filepath.Join(dir, "libdriver.so")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "driver.list")
	mustWriteFile(t, path, link+"\n")
	if err := readTestRdmaCollections(dir, []string{"driver"}); err != nil {
		t.Fatalf("existing root-owned literal symlink must pass: %v", err)
	}
	if err := os.Lchown(link, 1001, 1001); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("non-root-owned link must fail")
	}
	if err := os.Lchown(link, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := readTestRdmaCollections(dir, []string{"driver"}); err == nil {
		t.Fatal("broken literal symlink must fail")
	}

}

func readTestRdmaCollections(dir string, collectionNames []string) error {
	for _, collectionName := range collectionNames {
		if _, _, err := ReadMountEntries(dir, collectionName); err != nil {
			return err
		}
	}
	return nil
}

func TestReadMountEntriesReturnsValidatedFilesAndDirectories(t *testing.T) {
	dir := secureRdmaTestDir(t)
	sourcePath := filepath.Join(dir, "library.so")
	mustWriteFile(t, sourcePath, "library")
	subdir := filepath.Join(dir, "config")
	requireRdmaTestSuccess(t, os.Mkdir(subdir, 0755))
	mustWriteFile(t, filepath.Join(dir, "driver.list"), sourcePath+":/usr/lib64/library.so\n"+subdir+"\n")
	files, dirs, err := ReadMountEntries(dir, "driver")
	if err != nil || !reflect.DeepEqual(files, []string{sourcePath + ":/usr/lib64/library.so"}) ||
		!reflect.DeepEqual(dirs, []string{subdir}) {
		t.Fatalf("files=%v dirs=%v err=%v", files, dirs, err)
	}
}

func TestReadRdmaMountEntryRejectsWildcards(t *testing.T) {
	for _, entry := range []string{
		"/usr/lib64/*.so", "/usr/lib64/lib?.so", "/usr/lib64/lib[12].so",
		"/usr/lib64/*/library.so", "/usr/lib64/*.so:/container/library.so",
	} {
		files, dirs, err := readRdmaMountEntry(entry, os.Lstat)
		if err == nil || !strings.Contains(err.Error(), "wildcards") || files != nil || dirs != nil {
			t.Errorf("wildcard entry %q must be rejected: files=%v dirs=%v err=%v", entry, files, dirs, err)
		}
	}
}

func TestReadMountEntriesRejectsPartialResults(t *testing.T) {
	dir := secureRdmaTestDir(t)
	sourcePath := filepath.Join(dir, "library.so")
	mustWriteFile(t, sourcePath, "library")
	entry := sourcePath + "\n"
	for _, reader := range []io.Reader{
		strings.NewReader(entry + filepath.Join(dir, "missing.so") + "\n"),
		io.MultiReader(strings.NewReader(entry), rdmaFailReader{}),
	} {
		files, dirs, err := readRdmaMountEntries(reader, "driver.list", os.Lstat)
		if err == nil || files != nil || dirs != nil {
			t.Fatalf("partial results: %v %v %v", files, dirs, err)
		}
	}
}
