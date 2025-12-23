package main

import (
	"admin/internal/middleware"
	"admin/internal/router"
	"admin/internal/task"
	"context"
	"log"
	"net/http"
	"time"
	"wallet/common-lib/app"
	"wallet/common-lib/config"
	"wallet/common-lib/dbs"
	"wallet/common-lib/natsx"
	"wallet/common-lib/rdb"
	"wallet/common-lib/rpcx/kms_rpcx"
	"wallet/common-lib/rpcx/member_rpcx"
	"wallet/common-lib/rpcx/trade_rpcx"
	"wallet/common-lib/rpcx/wallet_rpcx"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

const (
	svrName = "admin"
)

func main() {
	app.Run(svrName, entry)
}

func entry(conf *config.BaseConf, svrConf *config.ServiceConfig) func() {
	dbs.Member = dbs.Init(&conf.DBMember)
	dbs.Trade = dbs.Init(&conf.DBTrade)
	dbs.Wallet = dbs.Init(&conf.DBWallet)
	dbs.Admin = dbs.Init(&conf.DBAdmin)
	dbs.System = dbs.Init(&conf.DBSystem)
	dbs.Merchant = dbs.Init(&conf.DBMerchant)
	rdb.Client = rdb.Init(&conf.RedisInstance)

	etcdAddr := conf.Etcd.Endpoints
	trade_rpcx.InitConn(svrName, etcdAddr)
	member_rpcx.InitConn(svrName, etcdAddr)
	wallet_rpcx.InitConn(svrName, etcdAddr)
	kms_rpcx.InitConn(svrName, etcdAddr)

	natsx.Init(svrName, &conf.Nats)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.Cors(),
		otelgin.Middleware(svrName),
		middleware.LoggerWithZap(zap.L()),
		middleware.RequestHeaders(),
		middleware.Response(),
	)
	router.Init(r)
	s := &http.Server{
		Addr:           svrConf.Service.Http.Address,
		Handler:        r,
		ReadTimeout:    svrConf.HttpReadTimeout(),
		WriteTimeout:   svrConf.HttpWriteTimeout(),
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		log.Printf("http listens on %s.", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			zap.L().Error("http listen error", zap.Error(err))
		}
	}()

	task.Run()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			zap.L().Error("http shutdown error", zap.Error(err))
			_ = s.Close()
		}
		dbs.Close()
		rdb.Close()
		trade_rpcx.CloseConn()
		member_rpcx.CloseConn()
		wallet_rpcx.CloseConn()
		kms_rpcx.CloseConn()
		natsx.Close()
	}
}
