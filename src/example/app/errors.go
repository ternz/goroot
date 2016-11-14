package app

import (
	//"fmt"
)

const (
	ERR_SUCCESS = 0
	ERR_SERVER = iota + 10000
	ERR_REQUEST_PARA
	ERR_DATABASE
	ERR_HTTP_REQUEST
	
	ERR_ALREADY_EXIST
)

func initErrorCode() (m map[int32]string) {
	m = make(map[int32]string)
	
	m[ERR_SUCCESS] = "success"
	m[ERR_SERVER] = "server internal error"
	m[ERR_REQUEST_PARA] = "request parameters error"
	m[ERR_DATABASE] = "database error"
	m[ERR_HTTP_REQUEST] = "http request error"

	m[ERR_ALREADY_EXIST] = "%s already exist"
	
	//for k, v := range m {
	//	fmt.Println(k, v)
	//}

	return
}

var ErrMap = initErrorCode()

const (
	ERR_STR_NULL_SESSION = "session is null"
	ERR_STR_INVALID_SESSION = "invalid session"
	ERR_STR_INTERNAL_SERVER = "internal server error"
)

type MyError struct {
	msg		string
}
func (e MyError) Error() string {
	return e.msg
}

var ErrInvalidSession = MyError{msg:"invalid session"}
