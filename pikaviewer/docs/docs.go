package docs

import "github.com/gin-gonic/gin"

func Serve(engine *gin.Engine) {
	engine.Static("/docs", "./pikaviewer/static/_site")
}
