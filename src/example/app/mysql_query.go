package app 

import (
	"fmt"
	"common/logging"
	"database/sql"
)

var mysql_db *sql.DB

func DoInitMysql(url string) {
	mysql_db = InitMysql(url)
}

type HistoryOpt struct {
	Id				uint64		`json:"id"`
	UserId			string		`json:"user_id,omitempty"`
	Content			string		`json:"content"`
	CreateTime		uint64		`json:"-"`
	CreateTimeStr	string		`json:"create_time"`
}

func DbCreateHistoryOpt(h *HistoryOpt) error {
	sqlstr := `insert into history_opt(user_id,content,create_time) values(?,?,?);`
	logging.Debug("DbCreateHistoryOpt %s (%+v)", sqlstr, h)
	_, err := mysql_db.Exec(sqlstr, h.UserId, h.Content, h.CreateTime)
	return err
}

func DbQueryHistoryOpt(userid string, mintime, maxtime uint64, page, size int) ([]*HistoryOpt, int, error) {
	var queryparams []interface{}
	wheresegment := "user_id=?"
	queryparams = append(queryparams, userid)
	countsql := `select count(*) from history_opt where %s;`
	querysql := `select id,content,create_time from history_opt where %s order by id desc limit ?,?;`
	if mintime != 0 {
		wheresegment += " and create_time>=?"
		queryparams = append(queryparams, mintime)
	}
	if maxtime != 0 {
		wheresegment += " and create_time<=?"
		queryparams = append(queryparams, maxtime)
	}
	countsql = fmt.Sprintf(countsql, wheresegment)
	querysql = fmt.Sprintf(querysql, wheresegment)
	
	logging.Debug("DbQueryHistoryOpt query sql:%s %s parameters user_is=%s mintime=%d maxtime=%d page=%d size=%d", 
		countsql, querysql, userid, mintime, maxtime, page, size)
		
	countRow := mysql_db.QueryRow(countsql, queryparams...)
	var total_count int
	err := countRow.Scan(&total_count)
	if err != nil {
		return nil, 0, err
	}
	
	queryparams = append(queryparams, size*(page-1))
	queryparams = append(queryparams, size)
	
	rows, err := mysql_db.Query(querysql, queryparams...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*HistoryOpt
	for rows.Next() {
		ho := &HistoryOpt{}
		err := rows.Scan(&ho.Id, &ho.Content, &ho.CreateTime)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, ho)
	}
	return list, total_count, nil
}