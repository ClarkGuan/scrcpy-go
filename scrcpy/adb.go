package scrcpy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func adbPush(serial, local, remote string) error {
	return adbExec(serial, "push", local, remote)
}

func adbInstall(serial, local string) error {
	return adbExec(serial, "install", "-r", local)
}

func adbRemovePath(serial, path string) error {
	return adbExec(serial, "shell", "rm", "-rf", path)
}

func adbReverse(serial, sockName string, localPort int) error {
	return adbExec(serial, "reverse",
		fmt.Sprintf("localabstract:%s", sockName),
		fmt.Sprintf("tcp:%d", localPort))
}

func adbReverseRemove(serial, sockName string) error {
	return adbExec(serial, "reverse", "--remove",
		fmt.Sprintf("localabstract:%s", sockName))
}

func adbForward(serial string, localPort int, sockName string) error {
	return adbExec(serial, "forward",
		fmt.Sprintf("tcp:%d", localPort),
		fmt.Sprintf("localabstract:%s", sockName))
}

func adbForwardRemove(serial string, localPort int) error {
	return adbExec(serial, "forward", "--remove",
		fmt.Sprintf("tcp:%d", localPort))
}

func adbExec(serial string, params ...string) error {
	if cmd, err := adbExecAsync(serial, params...); err != nil {
		return err
	} else {
		return cmd.Wait()
	}
}

func adbExecAsync(serial string, params ...string) (*exec.Cmd, error) {
	args := make([]string, 0, 8)
	if len(serial) > 0 {
		args = append(args, "-s", serial)
	}
	args = append(args, params...)

	adbCmdOnce.Do(getAdbCommand)
	if debugOpt.Debug() {
		log.Printf("执行 %s %s\n", adbCmd, strings.Join(args, " "))
	}
	cmd := exec.Command(adbCmd, args...)
	if debugOpt.Debug() {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

var adbCmd = "adb"
var adbCmdOnce sync.Once

func getAdbCommand() {
	adbEnv := os.Getenv("ADB")
	if len(adbEnv) > 0 {
		adbCmd = adbEnv
	}
}
