package main

import (
	"fmt"
	"net/http"
	"os"
	"parent-child-api/api"
	"parent-child-api/resx"

	"github.com/cmstar/go-errx"
	"github.com/rs/cors"
)

func main() {

	defer func() {
		err := errx.PreserveRecover("", recover())
		if err == nil {
			return
		}

		msg := errx.Describe(err)
		if resx.Log == nil {
			os.Stderr.Write([]byte(msg))
		} else {
			resx.Log.Root.Error(msg)
		}
	}()

	resx.Log.Root.Info("start.....")
	resx.InitDb(resx.Conf.DB)

	runWebApi()
}

func runWebApi() {
	mux := http.NewServeMux()

	reg := register{}
	api.RegisterRoutes(mux, resx.Conf.JwtSecret)

	err := http.ListenAndServe(fmt.Sprintf(":%d", resx.Conf.ApiPort), reg.corsEngine(mux))
	if err != nil {
		panic(err)
	}
}

type register struct{}

// 在给定的 Handler 上添加 CORS 支持。
func (x register) corsEngine(h http.Handler) http.Handler {
	op := cors.Options{
		AllowedOrigins: []string{"*"}, // MVP 阶段先放开，正式环境再收敛域名。
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:         86400,
	}

	return cors.New(op).Handler(h)
}
