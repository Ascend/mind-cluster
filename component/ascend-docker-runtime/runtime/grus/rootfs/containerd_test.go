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

// Package rootfs, ut for containerd
package rootfs

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
)

func TestGetSnapshotClient(t *testing.T) {
	tests := []struct {
		name          string
		ns            string
		mockNewClient bool
		newClientErr  error
		expectedErr   bool
	}{
		{
			name:          "successful connection",
			ns:            "test-ns",
			mockNewClient: true,
			newClientErr:  nil,
			expectedErr:   false,
		},
		{
			name:          "failed connection",
			ns:            "test-ns",
			mockNewClient: true,
			newClientErr:  os.ErrNotExist,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock patches
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Mock newContainerdClient function
			if tt.mockNewClient {
				patches.ApplyFunc(newContainerdClient, func(ctx context.Context, namespace, address string,
					opts ...containerd.ClientOpt) (*containerd.Client, context.Context, context.CancelFunc, error) {
					if tt.newClientErr != nil {
						return nil, nil, nil, tt.newClientErr
					}
					return &containerd.Client{}, context.Background(), func() {}, nil
				})
			}
			// Create ContainerdRootfs instance
			c := &ContainerdRootfs{socket: "/test/socket"}
			// Call function
			ctx, cancel, client, err := c.GetSnapshotClient(tt.ns)
			// Verify error
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
			// Verify return values
			if !tt.expectedErr {
				if ctx == nil {
					t.Fatalf("Expected ctx to be non-nil")
				}
				if cancel == nil {
					t.Fatalf("Expected cancel to be non-nil")
				}
				if client == nil {
					t.Fatalf("Expected client to be non-nil")
				}
			}
		})
	}
}

func TestCreateDiff(t *testing.T) {
	tests := []struct {
		name              string
		lower             string
		upper             string
		mockView          bool
		mockMounts        bool
		mockRemove        bool
		mockWithTempMount bool
		viewErr           error
		mountsErr         error
		removeErr         error
		withTempMountErr  error
		expectedErr       bool
	}{
		{
			name:              "successful diff creation",
			lower:             "lower-snap",
			upper:             "upper-snap",
			mockView:          true,
			mockMounts:        true,
			mockRemove:        true,
			mockWithTempMount: true,
			expectedErr:       false,
		},
		{
			name:        "view failed",
			lower:       "lower-snap",
			upper:       "upper-snap",
			mockView:    true,
			viewErr:     os.ErrNotExist,
			expectedErr: true,
		},
		{
			name:        "mounts failed",
			lower:       "lower-snap",
			upper:       "upper-snap",
			mockView:    true,
			mockMounts:  true,
			mountsErr:   os.ErrNotExist,
			expectedErr: true,
		},
		{
			name:              "with temp mount failed",
			lower:             "lower-snap",
			upper:             "upper-snap",
			mockView:          true,
			mockMounts:        true,
			mockRemove:        true,
			mockWithTempMount: true,
			withTempMountErr:  os.ErrNotExist,
			expectedErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock patches
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Mock sn.View method
			if tt.mockView {
				patches.ApplyMethodFunc(&mockSnapshotter{}, "View", func(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
					if tt.viewErr != nil {
						return nil, tt.viewErr
					}
					return []mount.Mount{{Source: "/test/lower"}}, nil
				})
			}
			// Mock sn.Mounts method
			if tt.mockMounts {
				patches.ApplyMethodFunc(&mockSnapshotter{}, "Mounts", func(ctx context.Context, key string) ([]mount.Mount, error) {
					if tt.mountsErr != nil {
						return nil, tt.mountsErr
					}
					return []mount.Mount{{Source: "/test/upper"}}, nil
				})
			}
			// Mock sn.Remove method
			if tt.mockRemove {
				patches.ApplyMethodFunc(&mockSnapshotter{}, "Remove", func(ctx context.Context, key string) error {
					return tt.removeErr
				})
			}
			// Mock mount.WithTempMount function
			if tt.mockWithTempMount {
				patches.ApplyFunc(mount.WithTempMount, func(ctx context.Context, mounts []mount.Mount, fn func(root string) error) error {
					return tt.withTempMountErr
				})
			}
			// Mock getRootfsIncre function
			patches.ApplyFunc(getRootfsIncre, func(ctx context.Context, diffInfo DiffInfo) (string, error) {
				if tt.withTempMountErr != nil {
					return "", tt.withTempMountErr
				}
				return "test-digest", nil
			})
			// Create mock snapshotter
			sn := &mockSnapshotter{}
			// Create writer
			writer := &bytes.Buffer{}
			// Call function
			digest, err := CreateDiff(context.Background(), sn, writer, tt.lower, tt.upper)
			// Verify error
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
			// Verify return values
			if !tt.expectedErr {
				if digest == "" {
					t.Fatalf("Expected digest to be non-empty")
				}
			}
		})
	}
}

func TestGetRootfsIncre(t *testing.T) {
	tests := []struct {
		name                      string
		diffInfo                  DiffInfo
		mockWithTempMount         bool
		mockWithReadonlyTempMount bool
		mockCompressStream        bool
		mockWriteDiff             bool
		withTempMountErr          error
		withReadonlyTempMountErr  error
		compressStreamErr         error
		writeDiffErr              error
		expectedErr               bool
	}{
		{
			name: "successful rootfs increment",
			diffInfo: DiffInfo{
				Lower:  []mount.Mount{{Source: "/test/lower"}},
				Upper:  []mount.Mount{{Source: "/test/upper"}},
				Writer: &bytes.Buffer{},
			},
			mockWithTempMount:         true,
			mockWithReadonlyTempMount: true,
			mockCompressStream:        true,
			mockWriteDiff:             true,
			expectedErr:               false,
		},
		{
			name: "with temp mount failed",
			diffInfo: DiffInfo{
				Lower:  []mount.Mount{{Source: "/test/lower"}},
				Upper:  []mount.Mount{{Source: "/test/upper"}},
				Writer: &bytes.Buffer{},
			},
			mockWithTempMount: true,
			withTempMountErr:  os.ErrNotExist,
			expectedErr:       true,
		},
		{
			name: "with readonly temp mount failed",
			diffInfo: DiffInfo{
				Lower:  []mount.Mount{{Source: "/test/lower"}},
				Upper:  []mount.Mount{{Source: "/test/upper"}},
				Writer: &bytes.Buffer{},
			},
			mockWithTempMount:         true,
			mockWithReadonlyTempMount: true,
			withReadonlyTempMountErr:  os.ErrNotExist,
			expectedErr:               true,
		},
		{
			name: "compress stream failed",
			diffInfo: DiffInfo{
				Lower:  []mount.Mount{{Source: "/test/lower"}},
				Upper:  []mount.Mount{{Source: "/test/upper"}},
				Writer: &bytes.Buffer{},
			},
			mockWithTempMount:         true,
			mockWithReadonlyTempMount: true,
			mockCompressStream:        true,
			compressStreamErr:         os.ErrNotExist,
			expectedErr:               true,
		},
		{
			name: "write diff failed",
			diffInfo: DiffInfo{
				Lower:  []mount.Mount{{Source: "/test/lower"}},
				Upper:  []mount.Mount{{Source: "/test/upper"}},
				Writer: &bytes.Buffer{},
			},
			mockWithTempMount:         true,
			mockWithReadonlyTempMount: true,
			mockCompressStream:        true,
			mockWriteDiff:             true,
			writeDiffErr:              os.ErrNotExist,
			expectedErr:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock patches
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Mock mount.WithTempMount function
			if tt.mockWithTempMount {
				patches.ApplyFunc(mount.WithTempMount, func(ctx context.Context, mounts []mount.Mount, fn func(root string) error) error {
					if tt.withTempMountErr != nil {
						return tt.withTempMountErr
					}
					return fn("/test/lower-root")
				})
			}
			// Mock mount.WithReadonlyTempMount function
			if tt.mockWithReadonlyTempMount {
				patches.ApplyFunc(mount.WithReadonlyTempMount, func(ctx context.Context, mounts []mount.Mount, fn func(root string) error) error {
					if tt.withReadonlyTempMountErr != nil {
						return tt.withReadonlyTempMountErr
					}
					return fn("/test/upper-root")
				})
			}
			// Mock compression.CompressStream function
			if tt.mockCompressStream {
				patches.ApplyFunc(compression.CompressStream, func(w io.Writer, algorithm compression.Compression) (io.WriteCloser, error) {
					if tt.compressStreamErr != nil {
						return nil, tt.compressStreamErr
					}
					return &mockWriteCloser{}, nil
				})
			}
			// Mock archive.WriteDiff function
			if tt.mockWriteDiff {
				patches.ApplyFunc(archive.WriteDiff, func(ctx context.Context, w io.Writer, lower, upper string, opts ...archive.WriteDiffOpt) error {
					return tt.writeDiffErr
				})
			}
			// Call function
			digest, err := getRootfsIncre(context.Background(), tt.diffInfo)
			// Verify error
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
			// Verify return values
			if !tt.expectedErr {
				if digest == "" {
					t.Fatalf("Expected digest to be non-empty")
				}
			}
		})
	}
}

var ApplyTarDiffTests = []struct {
	name           string
	diffTarPath    string
	lower          string
	decompressType string
	mockOpen       bool
	mockApply      bool
	mockCopy       bool
	openErr        error
	applyErr       error
	copyErr        error
	expectedErr    bool
}{
	{
		name:           "successful apply",
		diffTarPath:    "/test/diff.tar",
		lower:          "/test/lower",
		decompressType: "",
		mockOpen:       true,
		mockApply:      true,
		mockCopy:       true,
		expectedErr:    false,
	},
	{
		name:           "open failed",
		diffTarPath:    "/test/diff.tar",
		lower:          "/test/lower",
		decompressType: "",
		mockOpen:       true,
		openErr:        os.ErrNotExist,
		expectedErr:    true,
	},
	{
		name:           "apply failed",
		diffTarPath:    "/test/diff.tar",
		lower:          "/test/lower",
		decompressType: "",
		mockOpen:       true,
		mockApply:      true,
		applyErr:       os.ErrNotExist,
		expectedErr:    true,
	},
	{
		name:           "copy failed",
		diffTarPath:    "/test/diff.tar",
		lower:          "/test/lower",
		decompressType: "",
		mockOpen:       true,
		mockApply:      true,
		mockCopy:       true,
		copyErr:        os.ErrNotExist,
		expectedErr:    true,
	},
}

func TestApplyTarDiff(t *testing.T) {
	for _, tt := range ApplyTarDiffTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock patches
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Mock os.Open function
			if tt.mockOpen {
				patches.ApplyFunc(os.Open, func(name string) (*os.File, error) {
					if tt.openErr != nil {
						return nil, tt.openErr
					}
					return &os.File{}, nil
				})
			}
			// Mock archive.Apply function
			if tt.mockApply {
				patches.ApplyFunc(archive.Apply, func(ctx context.Context, dest string, r io.Reader, opts ...archive.ApplyOpt) (int64, error) {
					return 0, tt.applyErr
				})
			}
			// Mock io.Copy function
			if tt.mockCopy {
				patches.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (int64, error) {
					return 0, tt.copyErr
				})
			}
			// Call function
			err := ApplyTarDiff(context.Background(), tt.diffTarPath, tt.lower, tt.decompressType)
			// Verify error
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestNewContainerdClient(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		address         string
		mockCheckSocket bool
		mockNew         bool
		checkSocketErr  error
		newErr          error
		expectedErr     bool
	}{
		{
			name:            "successful client creation",
			namespace:       "test-ns",
			address:         "unix:///test/socket",
			mockCheckSocket: true,
			mockNew:         true,
			expectedErr:     false,
		},
		{
			name:            "check socket failed",
			namespace:       "test-ns",
			address:         "unix:///test/socket",
			mockCheckSocket: true,
			checkSocketErr:  os.ErrNotExist,
			expectedErr:     true,
		},
		{
			name:            "containerd new failed",
			namespace:       "test-ns",
			address:         "unix:///test/socket",
			mockCheckSocket: true,
			mockNew:         true,
			newErr:          os.ErrNotExist,
			expectedErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock patches
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Mock checkSocket function
			if tt.mockCheckSocket {
				patches.ApplyFunc(checkSocket, func(s string) error {
					return tt.checkSocketErr
				})
			}
			// Mock containerd.New function
			if tt.mockNew {
				patches.ApplyFunc(containerd.New, func(address string, opts ...containerd.ClientOpt) (*containerd.Client, error) {
					if tt.newErr != nil {
						return nil, tt.newErr
					}
					return &containerd.Client{}, nil
				})
			}
			// Call function
			client, ctx, cancel, err := newContainerdClient(context.Background(), tt.namespace, tt.address)
			// Verify error
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
			// Verify return values
			if !tt.expectedErr {
				if client == nil {
					t.Fatalf("Expected client to be non-nil")
				}
				if ctx == nil {
					t.Fatalf("Expected ctx to be non-nil")
				}
				if cancel == nil {
					t.Fatalf("Expected cancel to be non-nil")
				}
			}
		})
	}
}

// mockSnapshotter mocks snapshots.Snapshotter interface
type mockSnapshotter struct{}

func (m *mockSnapshotter) Stat(ctx context.Context, key string) (snapshots.Info, error) {
	return snapshots.Info{}, nil
}
func (m *mockSnapshotter) Usage(ctx context.Context, key string) (snapshots.Usage, error) {
	return snapshots.Usage{}, nil
}
func (m *mockSnapshotter) Mounts(ctx context.Context, key string) ([]mount.Mount, error) {
	return []mount.Mount{}, nil
}
func (m *mockSnapshotter) Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	return []mount.Mount{}, nil
}
func (m *mockSnapshotter) Commit(ctx context.Context, name, key string, opts ...snapshots.Opt) error {
	return nil
}
func (m *mockSnapshotter) View(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	return []mount.Mount{}, nil
}
func (m *mockSnapshotter) Remove(ctx context.Context, key string) error {
	return nil
}
func (m *mockSnapshotter) Walk(ctx context.Context, fn snapshots.WalkFunc, filters ...string) error {
	return nil
}
func (m *mockSnapshotter) Update(ctx context.Context, info snapshots.Info, fieldpaths ...string) (snapshots.Info, error) {
	return snapshots.Info{}, nil
}
func (m *mockSnapshotter) Close() error {
	return nil
}

// mockWriteCloser mocks io.WriteCloser interface
type mockWriteCloser struct{}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (m *mockWriteCloser) Close() error {
	return nil
}
