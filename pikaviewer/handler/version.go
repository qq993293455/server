package handler

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	client_version "coin-server/common/client-version"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/pikaviewer/env"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/selector"
	"coin-server/pikaviewer/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Version struct {
	Version      string `json:"version"`
	MinVersion   string `json:"min_version"`
	Gateway      string `json:"gateway"`
	CDN          string `json:"cdn"`
	Announcement string `json:"announcement"`
	VersionFile  string `json:"version_file"`
	Activate     bool   `json:"activate"`
	Audit        bool   `json:"audit"`
}

type Resp struct {
	Code int
	Data *models.Version
}

func (h *Version) Get(ctx *gin.Context, version, device string) (*models.Version, error) {
	bwl := &BetaWhiteList{}
	ok, err := bwl.IsInBetaWhiteList(device)
	if err != nil {
		return nil, err
	}
	verModel := model.NewVersion()
	// 正常版本信息
	normalData, err1 := verModel.GetPB(client_version.ClientVersionKey)
	if err1 != nil {
		return nil, utils.NewDefaultErrorWithMsg(err1.Error())
	}
	// 送审版本信息
	auditData, err1 := verModel.GetPB(client_version.AuditVersionKey)
	if err1 != nil {
		return nil, utils.NewDefaultErrorWithMsg(err1.Error())
	}
	data := normalData
	var audit bool
	// 送审版本已开启且版本号匹配，返回送审版本
	if auditData.Activate && h.isEqual(version, auditData.Version) {
		data = auditData
		audit = true
	}

	cdn := normalData.Cdn
	ret := &models.Version{
		Version:      data.Version,
		MinVersion:   data.MinVersion,
		Gateway:      data.Gateway,
		Cdn:          data.Cdn,
		Announcement: data.Announcement,
		VersionFile:  data.VersionFile,
	}
	// 在灰度服白名单里，需要返回灰度服的版本信息（热更文件地址还是返回正式服的）
	if ok {
		// caCert, err := ioutil.ReadFile("./pikaviewer/cer.crt")
		// if err != nil {
		// 	return nil, err
		// }
		// caCertPool := x509.NewCertPool()
		// caCertPool.AppendCertsFromPEM(caCert)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// RootCAs:            caCertPool,
					InsecureSkipVerify: true,
				},
			},
		}

		b, err := client.Get(os.Getenv(env.BETA_ADDR))
		if err != nil {
			return nil, err
		}
		defer b.Body.Close()
		body, err := ioutil.ReadAll(b.Body)
		if err != nil {
			return nil, err
		}
		resp := &Resp{}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}
		ret = resp.Data
		ret.Cdn = cdn
	} else if !audit {
		ret.Gateway = selector.Gateway.Get(ctx)
	}
	logger.DefaultLogger.Debug("get version info",
		zap.String("version", version),
		zap.String("device", device),
		zap.Bool("white_list", ok),
		zap.Bool("audit", audit),
		zap.Any("data", ret))

	return ret, nil
}

func (h *Version) GetByGM(typ string) (*models.Version, error) {
	key := client_version.ClientVersionKey
	if typ == "audit" {
		key = client_version.AuditVersionKey
	}
	verModel := model.NewVersion()
	data, err1 := verModel.GetPB(key)
	if err1 != nil {
		return nil, utils.NewDefaultErrorWithMsg(err1.Error())
	}
	return &models.Version{
		Version:      data.Version,
		Gateway:      data.Gateway,
		Cdn:          data.Cdn,
		Announcement: data.Announcement,
		VersionFile:  data.VersionFile,
		Activate:     data.Activate,
	}, nil
}

func (h *Version) Save(v *Version) error {
	key := client_version.ClientVersionKey
	if v.Audit {
		key = client_version.AuditVersionKey
	}
	data, err := model.NewVersion().GetPB(key)
	if err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	if data == nil {
		data = &dao.ClientVersion{
			Key:          key,
			Version:      "",
			MinVersion:   "",
			Gateway:      "",
			Cdn:          "",
			Announcement: "",
			VersionFile:  "",
			Activate:     false,
		}
	}
	if v.Version != "" && v.Version != data.Version {
		data.Version = v.Version
	}
	if v.MinVersion != "" && v.MinVersion != data.MinVersion {
		data.MinVersion = v.MinVersion
	}
	if v.Gateway != "" && v.Gateway != data.Gateway {
		data.Gateway = v.Gateway
	}
	if v.CDN != "" && v.CDN != data.Cdn {
		data.Cdn = v.CDN
	}
	if v.Announcement != "" && v.Announcement != data.Announcement {
		data.Announcement = v.Announcement
	}
	if v.VersionFile != "" && v.VersionFile != data.VersionFile {
		data.VersionFile = v.VersionFile
	}
	data.Activate = v.Activate
	if err := model.NewVersion().SavePB(data); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *Version) isEqual(client, server string) bool {
	index := strings.LastIndex(server, ".")
	if index == -1 {
		return false
	}
	server = server[0:index]
	return client == server
}
