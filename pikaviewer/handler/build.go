package handler

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"coin-server/pikaviewer/env"
	"coin-server/pikaviewer/utils"
)

type Build struct {
	Branch string `json:"branch" binding:"required"`
}

const logFile = "./pikaviewer/log/build.log"

var scripts = map[string]string{
	"develop": fmt.Sprintf("cd %s && ./cmake_android_to_client.sh", os.Getenv(env.SO_BUILD_DIR_DEV)),
	"patch":   fmt.Sprintf("cd %s && ./cmake_android_to_client_patch.sh", os.Getenv(env.SO_BUILD_DIR_PATCH)),
}

func (h *Build) Build(branch string) error {
	script, ok := scripts[branch]
	if !ok {
		return utils.NewDefaultErrorWithMsg("无效的分支")
	}
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(logFile, nil, 0666); err != nil {
		return err
	}
	name := "/bin/bash"

	command := script
	cmd := exec.Command(name, "-c", command)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		defer f.Close()
		reader := bufio.NewReader(stdout)
		for {
			line, err2 := reader.ReadBytes('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			_, err = f.Write(line)
			if err != nil {
				break
			}
		}
		if err := f.Sync(); err != nil {
			fmt.Println("sync error:", err)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println("wait error:", err)
		}
	}()

	return nil
}

func (h *Build) ReadLog() []string {
	file, err := os.Open(logFile)
	if err != nil {
		return []string{"日志文件不存在"}
	}
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
