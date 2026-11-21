//go:build !windows

package edit

import "golang.org/x/sys/unix"

func editorAvailableSpace(path string) (uint64, error) {
	var stats unix.Statfs_t
	if err := unix.Statfs(path, &stats); err != nil {
		return 0, err
	}
	return uint64(stats.Bavail) * uint64(stats.Bsize), nil
}
