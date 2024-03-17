//go:build !windows

package kubegen

import (
	"fmt"
	"os"
	"syscall"
)

const (
	shellExe = "/bin/sh"
	shellArg = "-c"
)

func setFileModeAndOwnership(f *os.File, fi os.FileInfo) error {
	if err := f.Chmod(fi.Mode()); err != nil {
		return fmt.Errorf("error setting file permissions: %w", err)
	}

	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("could not convert FileInfo.Sys() to *syscall.Stat_t")
	}

	if err := f.Chown(int(s.Uid), int(s.Gid)); err != nil {
		return fmt.Errorf("error changing file owner: %w", err)
	}
	return nil
}

func moveFile(src *os.File, dest string) error {
	if err := os.Rename(src.Name(), dest); err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	return nil
}
