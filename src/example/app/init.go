package app 

import (
	
)

func Init(file string) {
	InitConfigure(file)
	DoInitMysql(Cfg.Server.Mysql)
}