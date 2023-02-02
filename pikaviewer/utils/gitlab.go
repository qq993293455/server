package utils

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
)

const (
	gitlabToken = "hqsLLRHTXmsnVXysGKmy"
	gitlabAddr  = "http://gitlab.cdl5.org/"
)

var Gitlab *gitlab.Client

func InitGitlab() {
	client, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabAddr))
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}
	Gitlab = client
}
