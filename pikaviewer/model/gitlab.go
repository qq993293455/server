package model

import (
	selfEnv "coin-server/pikaviewer/env"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"os"
	"strconv"
)

type GitlabUser struct {
	UserId      int
	Username    string
	ProjectId   int
	AccessLevel int
}

func NewGitlabUser(userId int, username string, projectId, accessLevel int) *GitlabUser {
	return &GitlabUser{
		UserId:      userId,
		Username:    username,
		ProjectId:   projectId,
		AccessLevel: accessLevel,
	}
}

func (g *GitlabUser) genFileName() string {
	return fmt.Sprintf("%d_%d", g.UserId, g.ProjectId)
}

func (g *GitlabUser) Save() error {
	name := g.genFileName()
	filename := fmt.Sprintf("%s/gitlab/%s", os.Getenv(selfEnv.CLIENT_STATIC_FILE), name)
	return ioutil.WriteFile(filename, []byte(strconv.Itoa(g.AccessLevel)), 0666)
}

func (g *GitlabUser) Get() (gitlab.AccessLevelValue, error) {
	name := g.genFileName()
	filename := fmt.Sprintf("%s/gitlab/%s", os.Getenv(selfEnv.CLIENT_STATIC_FILE), name)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return gitlab.NoPermissions, err
	}
	al, err := strconv.Atoi(string(b))
	if err != nil {
		return gitlab.NoPermissions, err
	}
	return gitlab.AccessLevelValue(al), nil
}
