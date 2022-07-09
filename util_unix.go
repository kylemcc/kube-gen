//go:build !windows

package kubegen

import (
	"fmt"
	"os"
	"syscall"
)

const SHELL_EXE = "/bin/sh"
const SHELL_ARG = "-c"

func setFileModeAndOwnership(f *os.File, fi os.FileInfo) error {
	if err := f.Chmod(fi.Mode()); err != nil {
		return fmt.Errorf("error setting file permissions: %v", err)
	}
	if err := f.Chown(int(fi.Sys().(*syscall.Stat_t).Uid), int(fi.Sys().(*syscall.Stat_t).Gid)); err != nil {
		return fmt.Errorf("error changing file owner: %v", err)
	}
	return nil
}

func moveFile(src *os.File, dest string) error {
	if err := os.Rename(src.Name(), dest); err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	return nil
}
