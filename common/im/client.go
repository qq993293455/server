package im

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"coin-server/common/consulkv"
	"coin-server/common/utils"
	"coin-server/common/values/env"

	json "github.com/json-iterator/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var DefaultClient *Client
var tracer = otel.Tracer("im/client")

type Config struct {
	ImSys  string `json:"im_sys"`
	ImHttp string `json:"im_http"`
	ImTcp  string `json:"im_tcp"`
}

func Init(cnf *consulkv.Config) {
	cfg := &Config{}
	err := cnf.Unmarshal("im", cfg)
	utils.Must(err)
	DefaultClient = NewClient(cfg)
}

type Client struct {
	cfg      *Config
	baseAddr string
	client   *http.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{
		cfg:      cfg,
		baseAddr: cfg.ImSys,
		client:   utils.NewHttpClient(),
	}
}

func (c *Client) Config() *Config {
	return c.cfg
}

const (
	PathToken     = "/sys/token"
	PathPrivate   = "/sys/send/private"
	PathRoom      = "/sys/send/room"
	PathBroadcast = "/sys/send/broadcast"
	PathJoinRoom  = "/sys/join/room"
	PathLeaveRoom = "/sys/leave/room"
	PathBlackList = "/sys/blacklist"
)

// GetToken 供客户端获取im通信token
func (c *Client) GetToken(ctx context.Context, roleID, roleName string, rooms []string, extra string) (token string, err error) {
	if env.OpenTracing() {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "GetToken ")
		defer span.End()
	}

	q := url.Values{}
	q.Add("role_id", roleID)
	q.Add("role_name", roleName)
	for _, room := range rooms {
		q.Add("rooms", room)
	}
	q.Add("extra", extra)

	uri := utils.JoinURL(c.baseAddr, PathToken) + "?" + q.Encode()
	resp, err := c.client.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	ret := new(GetTokenResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}

	return ret.Result.Token, nil
}

const ContentTypeJSON = "application/json"

func (c *Client) PostJSON(ctx context.Context, url string, data interface{}) (body []byte, err error) {
	if env.OpenTracing() {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "PostJSON")
		defer span.End()
		span.SetAttributes(
			attribute.String("url", url),
		)
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return
	}
	r, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", ContentTypeJSON)
	r.Header.Set("keep-alive", "true")
	resp, err := c.client.Do(r)
	//resp, err := c.client.Post(url, ContentTypeJSON, bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	//if resp.StatusCode != http.StatusOK {
	//	return nil, errors.New(resp.Status)
	//}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

// SendMessage 发消息
func (c *Client) SendMessage(ctx context.Context, msg *Message) (err error) {
	var uri string

	switch msg.Type {
	case MsgTypePrivate:
		uri = utils.JoinURL(c.baseAddr, PathPrivate)
	case MsgTypeRoom:
		uri = utils.JoinURL(c.baseAddr, PathRoom)
	case MsgTypeBroadcast:
		uri = utils.JoinURL(c.baseAddr, PathBroadcast)
	default:
		return ErrUnknownMsgType
	}

	body, err := c.PostJSON(ctx, uri, msg)
	if err != nil {
		return
	}

	ret := new(BaseResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}
	if ret.Code != 200 {
		return errors.New(ret.Error)
	}

	return nil
}

// JoinRoom 进入房间
func (c *Client) JoinRoom(ctx context.Context, rm *RoomRole) (err error) {
	uri := utils.JoinURL(c.baseAddr, PathJoinRoom)

	body, err := c.PostJSON(ctx, uri, rm)
	if err != nil {
		return
	}

	ret := new(BaseResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}
	if ret.Code != 200 {
		return errors.New(ret.Error)
	}

	return nil
}

// LeaveRoom 离开房间
func (c *Client) LeaveRoom(ctx context.Context, rm *RoomRole) (err error) {
	uri := utils.JoinURL(c.baseAddr, PathLeaveRoom)

	body, err := c.PostJSON(ctx, uri, rm)
	if err != nil {
		return
	}

	ret := new(BaseResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}
	if ret.Code != 200 {
		return errors.New(ret.Error)
	}

	return nil
}

// BanPost 禁止发言 seconds表示禁言多少秒 0表示永封
func (c *Client) BanPost(ctx context.Context, roleId string, seconds int) (err error) {
	uri := utils.JoinURL(c.baseAddr, PathBlackList)

	body, err := c.PostJSON(ctx, uri, &BlackListOp{
		RoleID:  roleId,
		Type:    1,
		Seconds: seconds,
	})
	if err != nil {
		return
	}

	ret := new(BaseResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}
	if ret.Code != 200 {
		return errors.New(ret.Error)
	}

	return nil
}

// UnBanPost 解除禁言
func (c *Client) UnBanPost(ctx context.Context, roleId string) (err error) {
	uri := utils.JoinURL(c.baseAddr, PathBlackList)

	body, err := c.PostJSON(ctx, uri, &BlackListOp{
		RoleID: roleId,
		Type:   0,
	})
	if err != nil {
		return
	}

	ret := new(BaseResp)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return
	}
	if ret.Code != 200 {
		return errors.New(ret.Error)
	}

	return nil
}
