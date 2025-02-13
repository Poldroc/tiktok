// Code generated by hertz generator.

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzUtils "github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/hertz-contrib/gzip"
	hertzSentinel "github.com/hertz-contrib/opensergo/sentinel/adapter"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"github.com/ozline/tiktok/cmd/api/biz/middleware/es"
	"github.com/ozline/tiktok/cmd/api/biz/rpc"
	"github.com/ozline/tiktok/config"
	"github.com/ozline/tiktok/pkg/constants"
	"github.com/ozline/tiktok/pkg/errno"
	"github.com/ozline/tiktok/pkg/tracer"
	"github.com/ozline/tiktok/pkg/utils"
)

var (
	path       *string
	listenAddr string // listen port
)

func Init() {
	// config init
	path = flag.String("config", "./config", "config path")
	flag.Parse()
	config.Init(*path, constants.APIServiceName)

	rpc.Init()
	tracer.InitJaeger(constants.APIServiceName)

	es.Init()

	// set log
	klog.SetLevel(klog.LevelDebug)
	klog.SetLogger(kitexlogrus.NewLogger(kitexlogrus.WithHook(es.EsHookLog())))
}

func main() {
	Init()

	// get available port from config set
	for index, addr := range config.Service.AddrList {
		if ok := utils.AddrCheck(addr); ok {
			listenAddr = addr
			break
		}

		if index == len(config.Service.AddrList)-1 {
			klog.Fatal("not available port from config")
		}
	}

	r := server.New(
		server.WithHostPorts(listenAddr),
		server.WithHandleMethodNotAllowed(true),
		server.WithMaxRequestBodySize(1<<31),
	)

	// Recovery 错误恢复
	r.Use(recovery.Recovery(recovery.WithRecoveryHandler(recoveryHandler)))

	// Gzip
	r.Use(gzip.Gzip(gzip.BestSpeed))

	// Sentinel 流量治理
	r.Use(hertzSentinel.SentinelServerMiddleware(
		hertzSentinel.WithServerResourceExtractor(func(c context.Context, ctx *app.RequestContext) string {
			return "server_test"
		}),
		hertzSentinel.WithServerBlockFallback(func(ctx context.Context, c *app.RequestContext) {
			hlog.CtxInfof(ctx, "frequent requests have been rejected by the gateway. clientIP: %v\n", c.ClientIP())
			c.AbortWithStatusJSON(400, hertzUtils.H{
				"status_msg":  "too many request; the quota used up",
				"status_code": -1,
			})
		}),
	))

	register(r)

	r.Spin()
}

func recoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {

	hlog.CtxInfof(ctx, "[Recovery] InternalServiceError err=%v\n stack=%s\n", err, stack)
	c.JSON(consts.StatusInternalServerError, map[string]interface{}{
		"code":    errno.ServiceErrorCode,
		"message": fmt.Sprintf("[Recovery] err=%v\nstack=%s", err, stack),
	})
}
