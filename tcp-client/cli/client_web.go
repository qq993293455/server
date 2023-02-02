package cli

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"path"

	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/tcp-client/nutsdb"
	proto2 "coin-server/tcp-client/proto"
	"coin-server/tcp-client/values"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartClientServer() {
	r := gin.Default()
	r.Use(Cors())
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.LoadHTMLGlob("../../coin-server/tcp-client/dist/index.html")
	r.StaticFS("/static", http.Dir("../../coin-server/tcp-client/dist/static"))
	r.StaticFile("/favicon.ico", "../../coin-server/tcp-client/favicon.ico")

	r.GET("/workdir/get", getWorkDir)
	r.GET("/push/get", getPush)
	r.GET("/request/names", getRequestNames)
	r.GET("/request/data", getRequestData)
	r.GET("/server/data", getServers)
	r.GET("/user/data", getUsers)
	r.GET("/listen", startListen)
	r.GET("/gen/proto", genProto)
	r.GET("/gen/rule", genRule)

	r.POST("/workdir/save", setWorkDir)
	r.POST("/request/save", saveRequestData)
	r.POST("/server/save", saveServerData)
	r.POST("/server/del", deleteServer)
	r.POST("/user/save", saveUser)
	r.POST("/user/del", deleteUser)
	r.POST("/login", beginRequest)
	r.POST("/request", sendRequest)
	r.POST("/logout", endRequest)
	r.POST("/push/set", switchPush)
	go func() {
		err := r.Run()
		if err != nil {
			panic(err)
		}
	}()
}

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}

func getRequestData(c *gin.Context) {
	protoName := c.Query("name")
	js := proto2.GetRequestJson(protoName)
	if js != nil {
		c.String(http.StatusOK, string(js))
		return
	}
	raw := proto2.PbToJson(protoName)
	c.String(http.StatusOK, raw)
	return
}

func getWorkDir(c *gin.Context) {
	dir := nutsdb.Db.Get(values.MAC)
	if dir == nil {
		c.String(http.StatusOK, "")
		return
	}
	workdir := dir[:len(dir)-6]
	c.String(http.StatusOK, string(workdir))
}

func setWorkDir(c *gin.Context) {
	dir := c.PostForm("dir")
	dir = path.Join(dir, "proto")
	//dir += "\\proto"
	//TODO： 有时候传过来会多三个编码字节？？？
	//b := []byte(dir)
	//dir = string(b[3:])
	values.SetWorkingDir(dir)
	proto2.InitProto(values.ProtoPath)
	err := nutsdb.Db.Set(values.MAC, []byte(dir))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func getPush(c *gin.Context) {
	if values.Push == true {
		c.String(http.StatusOK, "true")
	} else {
		c.String(http.StatusOK, "false")
	}
}

func saveRequestData(c *gin.Context) {
	js := values.RequestJson{}
	err := c.BindJSON(&js)
	if err != nil {
		return
	}
	success := proto2.SaveRequestJson(js.Key, js.Value)
	if success {
		c.JSON(http.StatusOK, gin.H{
			"key":   js.Key,
			"value": js.Value,
		})
	}
	return
}

func saveServerData(c *gin.Context) {
	js := values.ServerConfig{}
	err := c.BindJSON(&js)
	if err != nil {
		return
	}
	if js.ServerName == "" {
		return
	}
	for k, v := range values.ServerCfg {
		if v.ServerName == js.ServerName {
			values.ServerCfg[k] = values.ServerConfig{
				ServerId:    js.ServerId,
				ServerName:  js.ServerName,
				GateWayAddr: js.GateWayAddr,
				StatelessId: js.StatelessId,
				BattleId:    js.BattleId,
			}
			b, err1 := json.Marshal(values.ServerCfg)
			if err1 != nil {
				panic(err1)
			}
			err1 = nutsdb.Db.Set("server", b)
			if err1 != nil {
				panic(err1)
			}
			return
		}
	}
	values.ServerCfg = append(values.ServerCfg, values.ServerConfig{
		ServerId:    js.ServerId,
		ServerName:  js.ServerName,
		GateWayAddr: js.GateWayAddr,
		StatelessId: js.StatelessId,
		BattleId:    js.BattleId,
	})
	b, err1 := json.Marshal(values.ServerCfg)
	if err1 != nil {
		panic(err1)
	}
	err1 = nutsdb.Db.Set("server", b)
	if err1 != nil {
		panic(err1)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
	return
}

func getRequestNames(c *gin.Context) {
	names := proto2.Proto().GetRequestNames()
	c.JSON(http.StatusOK, gin.H{
		"requests": names,
	})
}

func getServers(c *gin.Context) {
	c.JSON(http.StatusOK, values.ServerCfg)
}

func deleteServer(c *gin.Context) {
	name := c.PostForm("name")
	for k, v := range values.ServerCfg {
		if v.ServerName == name {
			copy(values.ServerCfg[k:], values.ServerCfg[k+1:])
			values.ServerCfg = values.ServerCfg[:len(values.ServerCfg)-1]
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, values.Users)
}

func saveUser(c *gin.Context) {
	user := c.PostForm("user")
	values.Users = append(values.Users, user)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func deleteUser(c *gin.Context) {
	user := c.PostForm("user")
	for k, v := range values.Users {
		if v == user {
			copy(values.Users[k:], values.Users[k+1:])
			values.Users = values.Users[:len(values.Users)-1]
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func beginRequest(c *gin.Context) {
	js := values.SendRequest{}
	err := c.BindJSON(&js)
	if err != nil {
		return
	}
	if js.Server == "" || js.User == "" {
		return
	}
	if _, ok := TcpCli.Load(js.User); ok {
		return
	}
	for _, server := range values.ServerCfg {
		if js.Server == server.ServerName {
			InitClient(js.User, server)
			break
		}
	}
}

func sendRequest(c *gin.Context) {
	js := values.SendRequest{}
	err := c.BindJSON(&js)
	if err != nil {
		return
	}
	clienti, ok := TcpCli.Load(js.User)
	if !ok {
		return
	}
	client := clienti.(TcpClient)
	client.SendRequest(js.Proto)
}

func startListen(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	user := c.Query("user")
	if user == "" {
		return
	}
	userMap.Store(user, ws)
}

func endRequest(c *gin.Context) {
	js := values.SendRequest{}
	err := c.BindJSON(&js)
	if err != nil {
		return
	}
	clienti, ok := TcpCli.Load(js.User)
	if !ok {
		return
	}
	client := clienti.(TcpClient)
	proto := &lessservicepb.User_RoleLogoutPush{}
	client.Send(proto)
	client.Close()
}

func switchPush(c *gin.Context) {
	push := c.PostForm("push")
	if push == "false" {
		values.Push = false
	} else {
		values.Push = true
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func genProto(c *gin.Context) {
	command := "gen_proto.cmd"
	cmd := exec.Command("cmd.exe", "/c", "start", command)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func genRule(c *gin.Context) {
	command := "cd ../excel && start gen-server.bat"
	cmd := exec.Command("cmd.exe", "/c", command)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
