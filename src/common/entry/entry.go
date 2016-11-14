package entry

import (
	"fmt"
	"mae_proj/MAE/common/logging"
)

type EntryName interface {
	GetId() uint64
	GetName() string
	GetEntryName() string
}

type Entry struct {
	Id           uint64
	Name         string
	GetEntryName func() string
}

func (e *Entry) GetId() uint64 {
	return e.Id
}

func (e *Entry) GetName() string {
	return e.Name
}

func (e *Entry) SetId(id uint64) {
	e.Id = id
}

func (e *Entry) SetName(name string) {
	e.Name = name
}

type EntryInterface interface {
	GetId() uint64
	GetName() string
	SetId(id uint64)
	SetName(name string)
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

func (e *Entry) formatHead(format string) string {
	if e.GetEntryName != nil {
		return fmt.Sprintf("%s%s[%d,%d,%s]%s", logging.GetLogBtInfo(2), e.GetEntryName(), e.Id>>32, uint32(e.Id), e.Name, format)
	}
	return format
}

func (e *Entry) Debug(format string, v ...interface{}) {
	logging.Debug(e.formatHead(format), v...)
}

func (e *Entry) Info(format string, v ...interface{}) {
	logging.Info(e.formatHead(format), v...)
}

func (e *Entry) Error(format string, v ...interface{}) {
	logging.Error(e.formatHead(format), v...)
}
