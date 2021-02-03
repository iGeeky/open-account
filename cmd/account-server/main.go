package main

import (
	"context"
	_ "expvar"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"open-account/configs"
	"open-account/internal/api/routers"
	apiutils "open-account/internal/api/utils"
	"open-account/pkg/baselib/db"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/json-iterator/go/extra"
)

func rlimitInit() {
	var rlim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("get rlimit error: " + err.Error())
	}
	fmt.Printf("limit before cur:%v max:%v\n", rlim.Cur, rlim.Max)
	rlim.Cur = 10000
	rlim.Max = 50000
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("set rlimit error: " + err.Error())
		return
	}
	fmt.Printf("limit after cur:%v max:%v\n", rlim.Cur, rlim.Max)
}

func startServer(config *configs.ServerConfig) (err error) {
	router := routers.InitRouter(config)

	s := &http.Server{
		Addr:         config.Listen,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// make sure idle connections returned
	processed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); nil != err {
			log.Fatalf("server shutdown failed, err: %v\n", err)
		}
		log.Infof("server gracefully shutdown")

		close(processed)
	}()

	err = gracehttp.Serve(s)
	if err != nil && err != http.ErrServerClosed {
		//log.Errorf("gracehttp.Serve %v start error:%v ", core.BuildTime, err)
		log.Errorf("gracehttp.Serve start error:%v ", err)
	} else if err == http.ErrServerClosed {
		err = nil
	}
	<-processed

	return err
}

func main() {
	rlimitInit()
	extra.RegisterFuzzyDecoders()
	var configFilename string
	flag.StringVar(&configFilename, "config", "./configs/config.yml", "yaml config filename")

	flag.Parse()

	config, err := configs.LoadConfig(configFilename)
	if err != nil {
		return
	}

	log.InitLogger(config.LogLevel, config.LogFileName, config.Debug, config.DisableStacktrace)
	debugInfo := fmt.Sprintf("### Build Time: %s ###  GitBranch: %s, GitCommit: %s,  Start Time: %s",
		utils.BuildTime, utils.GitBranch, utils.GitCommit, utils.Datetime())
	log.Infof(debugInfo)

	db.InitDBConfig(config.AccountDB.MaxOpenConns, config.AccountDB.MaxIdleConns)
	apiutils.InitRedisStore(config)

	ginplus.InitSign(config.SignHeaders, config.CustomHeaderPrefix)
	signUrls := make(map[string]bool, 10)
	ginplus.SignConfig = ginplus.APISignConfig{config.Debug, config.CheckSign, config.SuperSignKey, config.AppKeys, signUrls}
	ginplus.InitGinPlus(config.CustomHeaderPrefix)

	runtime.GOMAXPROCS(runtime.NumCPU())

	err = startServer(config)
}
