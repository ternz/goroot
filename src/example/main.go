// It includes skill, equipment, card and so on
package main

import (
	"example/app"
	"fmt"
	"time"

	"net/http"
	"runtime"

	"common/libutil"
	"common/logging"

	_ "github.com/go-sql-driver/mysql"
)

//注册http回调
func registerHttpHandle() {
	http.HandleFunc("/test", app.HandleTest)
	//mux := routes.New()
	//mux.Get("/publish/history/opt/", app.HandleHistoryOptQuery)
	//mux.Post("/publish/history/opt/:user_id", app.HandleHistoryOptCreate)
	//http.Handle("/", mux)
}

func main() {
	//配置解析
	app.Init("conf/config.json")

	//日志
	if err := libutil.TRLogger(app.Cfg.Log.File, app.Cfg.Log.Level, app.Cfg.Log.Name, app.Cfg.Log.Suffix, app.Cfg.Prog.Daemon); err != nil {
		fmt.Printf("init time rotate logger error: %s\n", err.Error())
		return
	}
	if app.Cfg.Prog.CPU == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU()) //配0就用所有核
	} else {
		runtime.GOMAXPROCS(app.Cfg.Prog.CPU)
	}

	logging.Debug("server start")

	libutil.InitSignal()

	logging.Debug("server init finish")
	/*go func() {
		err := http.ListenAndServe(app.Cfg.Prog.HealthPort, nil)

		fmt.Printf("err:%+v", err)
		if err != nil {
			logging.Error("ListenAndServe: %s", err.Error())
		}
	}()*/

	registerHttpHandle()

	go func() {
		err := http.ListenAndServe(app.Cfg.Server.PortInfo, nil)
		//err := http.ListenAndServeTLS(cfg.Server.PortInfo, "cert_server/server.crt",
		//"cert_server/server.key", nil)
		if err != nil {
			logging.Error("ListenAndServe port:%s failed", app.Cfg.Server.PortInfo)
		}
	}()

	file, err := libutil.DumpPanic("gsrv")
	if err != nil {
		logging.Error("init dump panic error: %s", err.Error())
	}

	defer func() {
		logging.Info("server stop...:%d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		logging.Info("server stop...,ok")
		if err := libutil.ReviewDumpPanic(file); err != nil {
			logging.Error("review dump panic error: %s", err.Error())
		}

	}()
	<-libutil.ChanRunning

}
