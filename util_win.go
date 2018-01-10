// +build windows

package kubegen

import (
    "fmt"
    "os"
    "os/exec"
)

const SHELL_EXE = "cmd"
const SHELL_ARG = "/c"

func setFileModeAndOwnership(f *os.File, fi os.FileInfo) error {
    return nil
}

func moveFile(src string, dest string) error {
    moveCmd := "move " + src + " " + dest
    cmd := exec.Command(SHELL_EXE, SHELL_ARG, moveCmd)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error creating output file: %v", err)
    }
    return nil
}
