package controller

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"coin-server/common/proto/models"
	selfEnv "coin-server/pikaviewer/env"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var versionHandler = new(handler.Version)

type Version struct {
}

func (c *Version) Get(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	v := ctx.Query("v")
	device := ctx.Query("device")
	typ := ctx.Query("type")
	var (
		version *models.Version
		err     error
	)
	if typ == "audit" {
		version, err = versionHandler.GetByGM(typ)
	} else {
		version, err = versionHandler.Get(ctx, v, device)
	}
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(version)
}

func (c *Version) Save(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Version{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	if err := versionHandler.Save(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Version) UploadVersionFile(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	targetOS := ctx.PostForm("os")
	file, err := ctx.FormFile("file")
	if err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg(err.Error()))
		return
	}
	// 1M
	if file.Size > 1024*1024 {
		resp.Send(utils.NewDefaultErrorWithMsg("文件太大了"))
		return
	}
	index := strings.LastIndex(file.Filename, ".")
	if index == -1 {
		resp.Send(utils.NewDefaultErrorWithMsg("文件格式有误"))
		return
	}
	ext := strings.ToLower(file.Filename[index:len(file.Filename)])
	if ext != ".txt" {
		resp.Send(utils.NewDefaultErrorWithMsg("文件格式有误"))
		return
	}
	path := fmt.Sprintf("%s/l5assets/%s/Release/", os.Getenv(selfEnv.CLIENT_STATIC_FILE), targetOS)
	dst := path + "/" + file.Filename
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg(err.Error()))
		return
	}
	resp.Send()
}

func (c *Version) UploadHotUpdateFile(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	targetOS := ctx.PostForm("os")
	file, err := ctx.FormFile("file")
	if err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg(err.Error()))
		return
	}
	// 1M
	// if file.Size > 1024*1024 {
	//	resp.Send(utils.NewDefaultErrorWithMsg("文件太大了"))
	//	return
	// }

	if file.Filename != "Release.zip" {
		resp.Send(utils.NewDefaultErrorWithMsg("文件格式有误"))
		return
	}
	path := fmt.Sprintf("%s/l5assets/%s/", os.Getenv(selfEnv.CLIENT_STATIC_FILE), targetOS)
	dst := path + "/" + file.Filename
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg(err.Error()))
		return
	}
	command := fmt.Sprintf("cd %s && unzip -o %s", path, file.Filename)
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(string(out))
}
