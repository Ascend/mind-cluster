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

// Package rootfs, containerd implementation of rootfs
package rootfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/go-digest"
	"golang.org/x/sys/unix"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
)

// ContainerdRootfs implements RootfsSnapshot interface using containerd
type ContainerdRootfs struct {
	socket string // Containerd socket address
}

// NewContainerdRootfs creates a new ContainerdRootfs instance
func NewContainerdRootfs(socket string) RootfsSnapshot {
	return &ContainerdRootfs{socket: socket}
}

// GetSnapshotClient creates a new containerd client with the specified namespace
func (c *ContainerdRootfs) GetSnapshotClient(ns string) (context.Context, context.CancelFunc, *containerd.Client, error) {
	cClient, ctx, cancel, err := newContainerdClient(context.Background(), ns, c.socket)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to connect to containerd, error: %v", err)
		return nil, nil, nil, err
	}
	return ctx, cancel, cClient, nil
}

// Checkpoint creates a snapshot of the container's rootfs
func (c *ContainerdRootfs) Checkpoint(ckptPath, containerID, ns string) (string, error) {
	tarFile := filepath.Join(ckptPath, common.ROOTFS_DIFF)
	hwlog.RunLog.Infof("exporting rootfs diff to tar, dst: %s", tarFile)
	file, err := os.OpenFile(tarFile, os.O_CREATE|os.O_RDWR, os.FileMode(common.SAFE_CONFIG_MODE))
	if err != nil {
		hwlog.RunLog.Errorf("Cannot create file dst: %s, error: %v", tarFile, err)
		return "", err
	}
	defer file.Close()
	var tarBuf bytes.Buffer
	w := io.MultiWriter(file, &tarBuf)

	ctx, cancel, dclient, err := c.GetSnapshotClient(ns)
	if err != nil {
		return "", err
	}
	defer func() {
		cancel()
		dclient.Close()
	}()
	//snapshot service build file system for container running
	sn := dclient.SnapshotService(containerd.DefaultSnapshotter)
	info, err := sn.Stat(ctx, containerID)
	if err != nil {
		hwlog.RunLog.Errorf("Cannot get snapshot info, error: %v", err)
		return "", err
	}

	return CreateDiff(ctx, sn, w, info.Parent, containerID)
}

// Restore applies a snapshot to the container's rootfs
func (c *ContainerdRootfs) Restore(ckptPath, rootfsPath, decompressType string) error {
	tarFile := filepath.Join(ckptPath, common.ROOTFS_DIFF)
	return ApplyTarDiff(context.Background(), tarFile, rootfsPath, decompressType)
}

// DiffInfo holds mount information for diff computation
type DiffInfo struct {
	Lower  []mount.Mount // Lower snapshot mounts (base layer)
	Upper  []mount.Mount // Upper snapshot mounts (top layer)
	Writer io.Writer     // Writer for output
}

// CreateDiff creates a diff between two snapshots and writes it to the writer
func CreateDiff(ctx context.Context, sn snapshots.Snapshotter, writer io.Writer, lower, upper string) (string, error) {
	var err error

	var lowerMounts, upperMounts []mount.Mount
	// use upper as key, to support concurrent checkpoint on same image containers
	key := fmt.Sprintf("grus-agent-%s-lower-key", upper)

	lowerMounts, err = sn.View(ctx, key, lower)
	if err != nil {
		return "", err
	}
	defer sn.Remove(ctx, key)

	hwlog.RunLog.Infof("Got lower snapshot mount, Mount: %v", lowerMounts)

	upperMounts, err = sn.Mounts(ctx, upper)
	if err != nil {
		return "", err
	}
	hwlog.RunLog.Infof("Got upper snapshot mount, Mount: %v", upperMounts)

	rootfsDiffInfo := DiffInfo{
		Lower:  lowerMounts,
		Upper:  upperMounts,
		Writer: writer,
	}
	return getRootfsIncre(ctx, rootfsDiffInfo)
}

// getRootfsIncre gets the incremental rootfs changes between two snapshots
// and writes the incremental changes to a tar file.
func getRootfsIncre(ctx context.Context, diffInfo DiffInfo) (string, error) {
	lowerMounts := diffInfo.Lower
	upperMounts := diffInfo.Upper

	var tarDigest string
	err := mount.WithTempMount(ctx, lowerMounts, func(lowerRoot string) error {
		hwlog.RunLog.Infof("Mounted lower snapshot to temp dir, path:%s", lowerRoot)

		return mount.WithReadonlyTempMount(ctx, upperMounts, func(upperRoot string) error {
			hwlog.RunLog.Infof("Mounted upper snapshot to temp dir, path:%s", upperRoot)

			dgstr := digest.SHA256.Digester()
			compressed, err := compression.CompressStream(diffInfo.Writer, compression.Uncompressed)
			if err != nil {
				hwlog.RunLog.Errorf("Failed to get compressed stream, error: %v", err)
				return err
			}
			defer compressed.Close()

			hwlog.RunLog.Infof("start writing rootfs-diff,  lowerRoot: %s, upperRoot: %s", lowerRoot, upperRoot)
			err = archive.WriteDiff(ctx, io.MultiWriter(compressed, dgstr.Hash()), lowerRoot, upperRoot)
			if err != nil {
				hwlog.RunLog.Errorf("Failed to write compressed diff, error: %v", err)
				return err
			}
			tarDigest = dgstr.Digest().String()
			hwlog.RunLog.Infof("Diff written successfully, digest: %s", tarDigest)
			return nil
		})
	})
	if err != nil {
		return "", err
	}

	return tarDigest, nil
}

// ApplyTarDiff applies a tar diff to the specified path
func ApplyTarDiff(ctx context.Context, diffTarPath, lower, decompressType string, opts ...archive.ApplyOpt) error {
	var r io.Reader

	tarFile, err := os.Open(diffTarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()
	r = tarFile

	hwlog.RunLog.Infof("Applying changes to path, path: %s, diffTarPath: %s", lower, diffTarPath)
	_, err = archive.Apply(ctx, lower, r, opts...)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to apply changes, error: %v", err)
		return err
	}

	if _, err = io.Copy(io.Discard, r); err != nil {
		return err
	}
	return nil
}

func checkSocket(s string) error {
	// set AT_EACCESS
	return unix.Faccessat(-1, s, unix.R_OK|unix.W_OK, unix.AT_EACCESS)
}

func newContainerdClient(ctx context.Context, namespace, address string, opts ...containerd.ClientOpt) (*containerd.Client, context.Context, context.CancelFunc, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	address = strings.TrimPrefix(address, "unix://")
	if err := checkSocket(address); err != nil {
		err = fmt.Errorf("access containerd socket %q, err: %v", address, err)
		return nil, nil, nil, err
	}
	client, err := containerd.New(address, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	return client, ctx, cancel, nil
}
