package app 

import (
	"time"
	"net/http"
	"io/ioutil"
	"crypto/sha1"
	"errors"
	"encoding/json"
	"common/logging"
	"hash/crc32"
	"fmt"
	"io"
	"crypto/rand"
	"crypto/md5"
	"strconv"
)

type ResponseJson struct {
	Code int32			`json:"result"`
	Msg string 			`json:"msg,omitempty"`
	Data interface{}	`json:"data,omitempty"`
}

func GetResponseJson(code int32, msg string, data interface{}) (b []byte, e error) {
	r := ResponseJson{
		Code: code,
		Msg: msg,
		Data: data,
	}
	b, e = json.Marshal(&r)
	if e != nil {
		return
	}
	return
}

func JsonSuccess(data interface{}) ([]byte) {
	bytes, err := GetResponseJson(ERR_SUCCESS, ErrMap[ERR_SUCCESS], data)
	if err != nil {
		logging.Error("json error: %s", err.Error())
		return nil
	}
	return bytes
}

func JsonError(code int32) ([]byte) {
	bytes, err := GetResponseJson(code, ErrMap[code], nil)
	if err != nil {
		logging.Error("json error: %s", err.Error())
		return nil
	}
	return bytes
}

func JsonErrorf(code int32, a...interface{}) []byte {
	msg := fmt.Sprintf(ErrMap[code], a...)
	bytes, err := GetResponseJson(code, msg, nil)
	if err != nil {
		logging.Error("json error: %s", err.Error())
		return nil
	}
	return bytes
}

func CheckPowerOfTwo(num uint32) bool {
	if num == 1 {
		return true
	}
	for num > 0 {
		if num % 2 == 1 {
			return false
		} else if num == 2 {
			return true
		}
		num /= 2
	}
	return false
}

func MyHash32(src string) uint32 {
	return crc32.ChecksumIEEE([]byte(src))
}

func randomId(length int) (string, error) {
	retry := 3
	b := make([]byte, length)
	var err error
	for retry > 0 {
		_, err = io.ReadFull(rand.Reader, b)
		if err == nil {
			break
		}
		logging.Error("generate random id error %s, retry", err.Error())
		retry --
	}
	if err != nil {
		return "", errors.New("generate random id error")
	}
	return fmt.Sprintf("%x", b), nil
}

func generateSignature(src string) string {
	hs := sha1.New()
	io.WriteString(hs, src)
	return string(hs.Sum(nil))
}

func md5sum(src string) string {
	hs := md5.New()
	io.WriteString(hs, src)
	return fmt.Sprintf("%x", hs.Sum(nil))
}

func ParsePostJsonBody(body io.Reader, v interface{}) error {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, v)
	if err != nil {
		return errors.New(fmt.Sprintf("error: %s, input: %s", err.Error(), string(bytes)))
	}
	return nil
}

func logAndResponse(w http.ResponseWriter, v interface{}) {
	content, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
	
	logging.Debug("Response %s", content)
}

func CheckSession(session string) (string, error) {
	url := Cfg.Server.AuthUrl
	timeout := Cfg.Server.AuthCheckTimeout * time.Second
	if url == "" {
		return session, nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Auth-Token", session)
	client := &http.Client{
		Timeout: timeout,
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logging.Error("http request %+v reponse %+v", req, res)
		if res.StatusCode == 401 {
			return "", ErrInvalidSession
		} else {
			return "", errors.New(fmt.Sprintf("http response statues: %s", res.Status))
		}
	}
	
	respJson := &ResponseJson{}
	
	err = ParsePostJsonBody(res.Body, respJson)
	if err != nil {
		return "", err
	}
	if respJson.Code != 0 {
		logging.Error("http request %+v reponse %+v", req, respJson)
		return "", ErrInvalidSession
	}
	user_id := respJson.Data.(map[string]interface{})["user_id"].(string)
	if user_id == "" {
		logging.Error("http request %+v reponse %+v", req, respJson)
		return "", ErrInvalidSession
	}
	return user_id, nil
}

func formatDate(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
