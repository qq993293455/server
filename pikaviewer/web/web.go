package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	selfEnv "coin-server/pikaviewer/env"
	"github.com/gin-gonic/gin"
)

//go:embed dist
var web embed.FS

func Init(engine *gin.Engine) {
	fsys, err := fs.Sub(web, "dist")
	if err != nil {
		panic(err)
	}

	cfs := onlyFilesFS{
		fs: http.FS(fsys),
	}
	h := http.FileServer(cfs)
	engine.Any("/", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})
	engine.StaticFS("/78fc91638d13863e", http.FS(fsys))

	engine.Any("/static/*any", gin.WrapH(h))
	engine.StaticFS("/client", http.Dir(os.Getenv(selfEnv.CLIENT_STATIC_FILE)))
}

type onlyFilesFS struct {
	fs http.FileSystem
}

type neuteredReaddirFile struct {
	http.File
}

func (fs onlyFilesFS) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}
