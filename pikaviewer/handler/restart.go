package handler

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"coin-server/pikaviewer/utils"
)

type Restart struct {
	battleLog       string
	logicLog        string
	dungeonLog      string
	dungeonMatchLog string

	battleFlock       string
	logicFlock        string
	dungeonFlock      string
	dungeonMatchFlock string
}

func NewRestart(battle, logic, dungeon, dungeonMatchLog, battleFlock, logicFlock, dungeonFlock, dungeonMatchFlock string) *Restart {
	return &Restart{
		battleLog:       battle,
		logicLog:        logic,
		dungeonLog:      dungeon,
		dungeonMatchLog: dungeonMatchLog,

		battleFlock:       battleFlock,
		logicFlock:        logicFlock,
		dungeonFlock:      dungeonFlock,
		dungeonMatchFlock: dungeonMatchFlock,
	}
}

func (r *Restart) LogicServerPid() string {
	command := "ps -ef | grep gameserver | grep -v grep | awk '{print $2}'"
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	pid := string(out)
	if pid == "" {
		return "N/A"
	}
	return pid
}

func (r *Restart) BattleServerPid() string {
	command := "ps -ef | grep 'battle_main -n' | grep -v grep | awk '{print $2}'"
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	pid := string(out)
	if pid == "" {
		return "N/A"
	}
	return pid
}

func (r *Restart) DungeonServerPid() string {
	command := "ps -ef | grep 'battle_copy -s' | grep -v grep | awk '{print $2}'"
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	pid := string(out)
	if pid == "" {
		return "N/A"
	}
	return pid
}

func (r *Restart) DungeonMatchServerPid() string {
	command := "ps -ef | grep 'roguelikematchserver' | grep -v grep | awk '{print $2}'"
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	pid := string(out)
	if pid == "" {
		return "N/A"
	}
	return pid
}

func (r *Restart) RestartBattleServer(ip, tag string) error {
	flock := New(r.battleFlock)
	if err := flock.Lock(); err != nil {
		return utils.NewDefaultErrorWithMsg("其他用户正在重启战斗服务器，请稍后再试")
	}

	f, err := os.Create(r.battleLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.battleLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/sh"

	command := "source /opt/rh/devtoolset-11/enable && cd /root/battle && ./complier_and_restart.sh"
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
		defer flock.Unlock()
		if err := f.Sync(); err != nil {
			fmt.Println("sync error:", err)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println("wait error:", err)
		}
	}()
	_ = r.send2IC(tag, map[string]string{
		"content": "战斗服务开始重启" + "（" + ip + "）",
		"at_user": "all",
	})
	return nil
}

func (r *Restart) RestartLogicServer(ip, tag string) error {
	flock := New(r.logicFlock)
	if err := flock.Lock(); err != nil {
		return utils.NewDefaultErrorWithMsg("其他用户正在重启逻辑服务器，请稍后再试")
	}

	f, err := os.Create(r.logicLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.logicLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/bash"

	command := "cd /game && ./restart-gs.sh"
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
		defer flock.Unlock()
		if err := f.Sync(); err != nil {
			fmt.Println("sync error:", err)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println("wait error:", err)
		}
	}()
	_ = r.send2IC(tag, map[string]string{
		"content": "逻辑服务开始重启" + "（" + ip + "）",
		"at_user": "all",
	})
	return nil
}

func (r *Restart) RestartDungeonServer(ip, tag string) error {
	flock := New(r.dungeonFlock)
	if err := flock.Lock(); err != nil {
		return utils.NewDefaultErrorWithMsg("其他用户正在重启副本服务器，请稍后再试")
	}

	f, err := os.Create(r.dungeonLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.dungeonLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/sh"

	command := "source /opt/rh/devtoolset-11/enable && cd /root/battle && ./complier_and_restart_copy.sh"
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
		defer flock.Unlock()
		if err := f.Sync(); err != nil {
			fmt.Println("sync error:", err)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println("wait error:", err)
		}
	}()
	_ = r.send2IC(tag, map[string]string{
		"content": "副本服务开始重启" + "（" + ip + "）",
		"at_user": "all",
	})
	return nil
}

func (r *Restart) RestartDungeonMatchServer(ip, tag string) error {
	flock := New(r.dungeonMatchFlock)
	if err := flock.Lock(); err != nil {
		return utils.NewDefaultErrorWithMsg("其他用户正在重启匹配服务器，请稍后再试")
	}

	f, err := os.Create(r.dungeonMatchLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.dungeonMatchLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/sh"

	command := "cd /game && ./restart-rld.sh"
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
		defer flock.Unlock()
		if err := f.Sync(); err != nil {
			fmt.Println("sync error:", err)
		}
		if err := cmd.Wait(); err != nil {
			// if exitErr, ok := err.(*exec.ExitError); ok {
			// 	status := exitErr.Sys().(syscall.WaitStatus)
			// 	switch {
			// 	case status.Exited():
			// 		fmt.Printf("Return exit error: exit code=%d\n", status.ExitStatus())
			// 	case status.Signaled():
			// 		fmt.Printf("Return exit error: signal code=%d\n", status.Signal())
			// 	}
			// }
			f.Write([]byte("restart error"))
			fmt.Println("wait error:", err)
		}
	}()
	_ = r.send2IC(tag, map[string]string{
		"content": "匹配服务开始重启" + "（" + ip + "）",
		"at_user": "all",
	})
	return nil
}

func (r *Restart) ReadLog(typ utils.ServerType, tag string) []string {
	var filename string
	if typ == utils.Battle {
		filename = r.battleLog
	} else if typ == utils.Logic {
		filename = r.logicLog
	} else if typ == utils.Dungeon {
		filename = r.dungeonLog
	} else {
		filename = r.dungeonMatchLog
	}
	file, err := os.Open(filename)
	if err != nil {
		return []string{"日志文件不存在"}
	}
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	var (
		doneText bool
		errText  bool
	)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if !doneText {
			doneText = strings.Index(scanner.Text(), "restart done") != -1
		}
		if !errText {
			errText = strings.Index(scanner.Text(), "restart error") != -1
		}
	}
	if doneText || errText {
		var pid, content string
		if typ == utils.Battle {
			pid = r.BattleServerPid()
			content = "战斗服"
		} else if typ == utils.Logic {
			pid = r.LogicServerPid()
			content = "逻辑服"
		} else if typ == utils.Dungeon {
			pid = r.DungeonServerPid()
			content = "副本服"
		} else {
			pid = r.DungeonMatchServerPid()
			content = "匹配服"
		}
		// 战斗服pid有延迟（只要脚本没报错就认为成功）
		if typ == utils.Battle {
			if errText {
				content += "重启失败"
			} else {
				content += "重启成功"
			}
		} else {
			if pid == "N/A" || errText {
				content += "重启失败"
			} else {
				content += "重启成功"
			}
		}
		_ = r.send2IC(tag, map[string]string{
			"content": content,
		})
	}
	return lines
}

func (r *Restart) SyncMap(tag, ip string) error {
	// dungeonMatchLog 没用了，临时用它
	f, err := os.Create(r.dungeonMatchLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.dungeonMatchLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/sh"

	command := "cd /root/GitRepoes/MapSync && ./sync.sh"
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
			f.Write([]byte("sync error"))
			fmt.Println("wait error:", err)
		}
	}()
	_ = r.send2IC(tag, map[string]string{
		"content": "手动同步地图数据" + "（" + ip + "）",
		// "at_user": "all",
	}, "手动同步地图数据 ["+tag+"]")
	return nil
}

func (r *Restart) OverwriteDev(tag, ip string) error {
	f, err := os.Create(r.dungeonMatchLog)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.dungeonMatchLog, nil, 0666); err != nil {
		return err
	}
	name := "/bin/sh"

	command := "cd /root/GitRepoes/Overwrite/ && ./overwrite.sh"
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
			f.Write([]byte("sync error"))
			fmt.Println("wait error:", err)
		}
	}()
	// _ = r.send2IC(tag, map[string]string{
	// 	"content": "IP为 " + ip + " 的用户执行了Patch覆盖Dev的操作",
	// 	"at_user": "all",
	// }, "Patch覆盖Dev")
	return nil
}

func (r *Restart) send2IC(tag string, params map[string]string, title ...string) error {
	// IC_SEND_API="http://im-api.skyunion.net/msg"
	// IC_SEND_TOKEN="df8a445dd467a62bf1d7bdc5066dd918"
	// IC_SEND_TARGET="group"
	// IC_SEND_ROOM="10073164"
	// IC_SEND_TITLE="打卡提醒"
	// IC_SEND_CONTENT_TYPE=1
	// IC_SEND_CONTENT="----------"
	// IC_SEND_USER="all"
	// curl -X POST -H "Content-Type:application/x-www-form-urlencoded" ${IC_SEND_API}
	// -d "token=${IC_SEND_TOKEN}"
	// -d "target=${IC_SEND_TARGET}"
	// -d "room=${IC_SEND_ROOM}"
	// -d "title=${IC_SEND_TITLE}"
	// -d "content_type=${IC_SEND_CONTENT_TYPE}"
	// -d "content=${IC_SEND_CONTENT}"
	// -d "at_user=${IC_SEND_USER}"

	data := map[string]string{
		"token":        "df8a445dd467a62bf1d7bdc5066dd918",
		"target":       "group",
		"room":         "10073164",
		"title":        "服务器重启 [" + tag + "]",
		"content_type": "1",
	}
	if len(title) > 0 {
		data["title"] = title[0]
	}
	for k, v := range params {
		data[k] = v
	}
	// content := "逻辑服务器开始重启"
	// if battle {
	//	content = "战斗服务器开始重启"
	// }
	// data["content"] = content + "（" + ip + "）"
	// if atAll {
	//	data["at_user"] = "all"
	// }
	_, err := utils.NewRequest("http://im-api.skyunion.net/msg").Post(data)
	return err
}
