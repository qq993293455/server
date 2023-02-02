package env

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func SetRuleEnv() {
	suffix := os.Getenv(RULE_TAGE)
	if suffix == "" {
		var cmd *exec.Cmd
		command := "git symbolic-ref --short -q HEAD"
		if runtime.GOOS == "windows" {
			cmd = exec.Command("sh.exe", "-c", command)
		} else {
			cmd = exec.Command("/bin/sh", "-c", command)
		}
		bytes, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		curBranch := strings.TrimSpace(string(bytes))
		if err := os.Setenv(RULE_TAGE, curBranch); err != nil {
			panic(err)
		}
	}
}
