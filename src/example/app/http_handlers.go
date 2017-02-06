package app

import (
	"common/logging"
	"net/http"
	"strconv"
	"time"
)

func HandleHistoryOptQuery(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//w.Header().Add("Access-Control-Allow-Origin", "*")

	logging.Debug("request: %+v", r)

	session := r.Header.Get("X-Auth-Token")
	if session == "" {
		logging.Warning("parameter session is null")
		http.Error(w, ERR_STR_NULL_SESSION, http.StatusUnauthorized)
		return
	}

	user_id, err := CheckSession(session)
	if err != nil {
		logging.Error("check session error: %s", err.Error())
		if err == ErrInvalidSession {
			http.Error(w, ERR_STR_INVALID_SESSION, http.StatusUnauthorized)
		} else {
			http.Error(w, ERR_STR_INTERNAL_SERVER, http.StatusInternalServerError)
		}
		return
	}

	var mintime, maxtime uint64
	var page, size int
	params := r.URL.Query()
	mintime_str := params.Get("mintime")
	maxtime_str := params.Get("maxtime")
	page_str := params.Get("page")
	size_str := params.Get("size")
	if mintime_str != "" {
		mintime, err = strconv.ParseUint(mintime_str, 10, 64)
		if err != nil {
			logging.Error("%s ParseUint error: %s", mintime_str, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if maxtime_str != "" {
		maxtime, err = strconv.ParseUint(maxtime_str, 10, 64)
		if err != nil {
			logging.Error("%s ParseUint error: %s", maxtime_str, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	page, err = strconv.Atoi(page_str)
	if err != nil {
		logging.Error("%s Atoi error: %s", page_str, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	size, err = strconv.Atoi(size_str)
	if err != nil {
		logging.Error("%s Atoi error: %s", size_str, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, total, err := DbQueryHistoryOpt(user_id, mintime, maxtime, page, size)
	if err != nil {
		logging.Error("db error: %s", err.Error())
		http.Error(w, ERR_STR_INTERNAL_SERVER, http.StatusInternalServerError)
		return
	}

	for i, _ := range list {
		list[i].CreateTimeStr = formatDate(int64(list[i].CreateTime))
	}

	resp := make(map[string]interface{})
	resp["total"] = total
	resp["history_opt_list"] = list
	logAndResponse(w, resp)
}

func HandleHistoryOptCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Add("Access-Control-Allow-Origin", "*")

	logging.Debug("request: %+v", r)

	urlpara := r.URL.Query()
	user_id := urlpara.Get(":user_id")
	if user_id == "" {
		logging.Warning("parameter user_id is null")
		http.Error(w, "parameter user_id is null", http.StatusBadRequest)
		return
	}

	bodypara := make(map[string]string)
	ParsePostJsonBody(r.Body, &bodypara)
	content := bodypara["content"]
	if content == "" {
		logging.Warning("parameter content is null")
		http.Error(w, "parameter content is null", http.StatusBadRequest)
		return
	}

	h := &HistoryOpt{
		UserId:     user_id,
		Content:    content,
		CreateTime: uint64(time.Now().Unix()),
	}

	err := DbCreateHistoryOpt(h)
	if err != nil {
		logging.Error("db error: %s", err.Error())
		http.Error(w, ERR_STR_INTERNAL_SERVER, http.StatusInternalServerError)
		return
	}

	logging.Debug("response OK")
	w.Write(nil)
}

func HandleTest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logging.Info("handle test")
	w.Write([]byte("test ok"))
}
