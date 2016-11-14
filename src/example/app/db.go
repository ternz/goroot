package app

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"regexp"
	"strings"

	"common/goredis"
	"common/logging"
)

func InitRedis(redisurl string) *goredis.Redis {
	redis_handle, err := goredis.DialURL(redisurl)
	if err != nil {
		logging.Error("redis init fail:%s,%s", redisurl, err.Error())
		panic("InitRedis failed")
		return nil
	} else {
		logging.Info("redis conn ok:%s", redisurl)
	}
	return redis_handle
}

func InitMysql(mysqlurl string) *sql.DB {

	if ok, err := regexp.MatchString("^mysql://.*:.*@.*/.*$", mysqlurl); ok == false || err != nil {
		logging.Error("mysql config syntax err:mysql_zone,%s,shutdown", mysqlurl)
		panic("InitMysql conf error")
		return nil
	}
	mysqlurl = strings.Replace(mysqlurl, "mysql://", "", 1)
	db, err := sql.Open("mysql", mysqlurl)
	if err != nil {
		logging.Error("InitMysql failed mysqlurl=" + mysqlurl + ",err=" + err.Error())
		panic("InitMysql failed mysqlurl=" + mysqlurl)
		return nil
	} else {
		logging.Info("mysql conn ok:%s", mysqlurl)
	}
	return db
}
