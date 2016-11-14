package app

import (
	"common/libutil"
	"common/logging"
	"flag"
	"time"
)

type Configure struct {
	Log struct {
		File   string
		Level  string
		Name   string
		Suffix string
	}

	Prog struct {
		CPU        int
		Daemon     bool
		HealthPort string
	}

	Server struct {
		Redis    			string
		Mysql    			string
		PortInfo 			string
		AuthUrl				string
		AuthCheckTimeout	time.Duration
	}
}

func NewConfigure(path string) *Configure {
	config := flag.String("a", path, "config file")
	flag.Parse()
	Cfg := &Configure{}
	err := libutil.ParseJSON(*config, &Cfg)
	if err != nil {
		logging.Error("parse config %s error: %s\n", *config, err.Error())
		return nil
	}
	return Cfg
}

var Cfg *Configure

func InitConfigure(file string) {
	Cfg = NewConfigure(file)
	if Cfg == nil {
		panic("init configure failed")
	}
}
