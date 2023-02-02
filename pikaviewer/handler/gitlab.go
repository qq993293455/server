package handler

import (
	"fmt"
	"log"
	"strings"

	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
	"github.com/xanzy/go-gitlab"
)

var projectNameMap = map[int]string{
	3:  "coin-server",
	40: "l5client",
	7:  "share",
}

type Gitlab struct {
	AccessLevel string `json:"access_level" binding:"required"`
	Projects    []int  `json:"projects" binding:"required"`
	Users       []int  `json:"users" binding:"required"`
}

type GitlabMember struct {
	Id       int    `json:"id"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (h *Gitlab) Members() ([]*GitlabMember, error) {
	list, _, err := utils.Gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			Page:    0,
			PerPage: 1000,
		},
	})
	if err != nil {
		return nil, err
	}
	members := make([]*GitlabMember, 0, len(list))
	for _, member := range list {
		members = append(members, &GitlabMember{
			Id:       member.ID,
			Username: member.Username,
			Name:     member.Name,
		})
	}
	return members, nil
}

func (h *Gitlab) ModifyAccessLevel(req *Gitlab) ([]string, error) {
	members := make([]*model.GitlabUser, 0)
	for _, project := range req.Projects {
		for _, user := range req.Users {
			member, _, err := utils.Gitlab.ProjectMembers.GetProjectMember(project, user)
			if err != nil {
				return nil, err
			}
			members = append(members, model.NewGitlabUser(member.ID, member.Username, project, int(member.AccessLevel)))
		}
	}
	success := make([]string, 0)
	// 将用户当前的选项写入文件，用于推送完之后恢复用户的权限
	for _, member := range members {
		if err := member.Save(); err != nil {
			return nil, err
		}
	}

	accessLevel := h.GetAccessLevel(req.AccessLevel)
	var status string
	if req.AccessLevel == "Developer" {
		status = "关闭"
	} else if req.AccessLevel == "Maintainer" {
		status = "开启"
	}
	// 修改权限
	for _, member := range members {
		if _, _, err := utils.Gitlab.ProjectMembers.EditProjectMember(member.ProjectId, member.UserId, &gitlab.EditProjectMemberOptions{
			AccessLevel: &accessLevel,
		}); err != nil {
			return success, err
		}
		success = append(success, member.Username)
		projectName := projectNameMap[member.ProjectId]
		h.Send2IC(projectName+"权限变动", fmt.Sprintf("\n%s用户%s的patch分支提交权限已%s", member.Username, projectName, status))
	}

	return success, nil
}

type RestoreForm struct {
	UserId      int    `json:"user_id"`
	UserName    string `json:"user_name"`
	UsrUsername string `json:"usr_username"`
	Project     struct {
		Id   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"project,omitempty"`
	Commits []struct {
		Id      string `json:"id,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"commits,omitempty"`
	Ref string `json:"ref,omitempty"`
}

func (h *Gitlab) RestoreAccessLevel(req *RestoreForm) error {
	if !strings.Contains(req.Ref, "patch") {
		return nil
	}
	user := model.NewGitlabUser(req.UserId, req.UsrUsername, req.Project.Id, 0)
	// 只有Developer和Maintainer，只会修改Developer的权限
	// Maintainer的权限不会修改，如果是Maintainer提交这里会返回error（文件不存在）
	level, _ := user.Get()
	var str string
	if level > 0 {
		if _, _, err := utils.Gitlab.ProjectMembers.EditProjectMember(req.Project.Id, req.UserId, &gitlab.EditProjectMemberOptions{
			AccessLevel: &level,
		}); err != nil {
			return err
		}
	}
	if level == gitlab.DeveloperPermissions {
		str = "提交权限已自动关闭"
	}
	var commit string
	for _, c := range req.Commits {
		commit += c.Id + "\n"
		commit += c.Message
	}
	name := req.UsrUsername
	if name == "" {
		name = req.UserName
	}
	content := fmt.Sprintf("\n%s向%s的patch分支提交了新内容：\n %s \n %s", name, req.Project.Name, commit, str)
	h.Send2IC(req.Project.Name+"提交信息", content)
	return nil
}

func (h *Gitlab) GetAccessLevel(aLevel string) gitlab.AccessLevelValue {
	if aLevel == "Developer" {
		return gitlab.DeveloperPermissions
	}
	if aLevel == "Maintainer" {
		return gitlab.MaintainerPermissions
	}

	return gitlab.NoPermissions
}

func (h *Gitlab) Send2IC(title, content string) {
	data := map[string]string{
		"token":        "df8a445dd467a62bf1d7bdc5066dd918",
		"target":       "group",
		"room":         "10073164",
		"title":        title,
		"content_type": "1",
		"content":      content,
	}
	if _, err := utils.NewRequest("http://im-api.skyunion.net/msg").Post(data); err != nil {
		log.Println("gitlab send to ic err:", data)
	}
}
