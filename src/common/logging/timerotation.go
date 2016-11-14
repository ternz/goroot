package logging

import (
	//"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type TimeRotationHandler struct {
	*Handler
	localData        map[string]string
	fileChangeHandle HeaderFunc
}

type HeaderFunc func(w io.Writer)

func NewTimeRotationHandler(shortfile string, suffix string, handle HeaderFunc) (*TimeRotationHandler, error) {
	h := &TimeRotationHandler{}
	h.fileChangeHandle = handle
	fullfile := strings.Join([]string{shortfile, time.Now().Format(suffix)}, ".")
	file, err := h.openFile(fullfile, shortfile)
	if err != nil {
		return nil, err
	}
	h.Handler = NewHandler(file)
	h.before = h.rotate
	h.localData = make(map[string]string)
	h.localData["oldfilepath"] = fullfile
	h.localData["linkpath"] = shortfile
	h.localData["suffix"] = suffix
	return h, nil
}

func (h *TimeRotationHandler) openFile(filepath, linkpath string) (*os.File, error) {
	bNew := false
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Create(filepath); err != nil {
				return nil, err
			} else {
				bNew = true
			}

		} else {
			return nil, err
		}
	}
	_ = os.Remove(linkpath)
	var fn string
	if err := os.Symlink(filepath, linkpath); err != nil {
		fn = filepath
	} else {
		fn = linkpath
	}
	file, err := os.OpenFile(fn, FileCreateFlag, FileCreatePerm)
	if err != nil {
		return nil, err
	}

	if bNew && h.fileChangeHandle != nil {
		h.fileChangeHandle(file)
		bNew = false
	}

	return file, nil
}

func (h *TimeRotationHandler) rotate(*Record, io.ReadWriter) {
	filepath := h.localData["linkpath"] + "." + time.Now().Format(h.localData["suffix"])
	if filepath != h.localData["oldfilepath"] {
		_ = h.writer.(io.Closer).Close()
		file, err := h.openFile(filepath, h.localData["linkpath"])
		if err != nil {
			return
		}
		h.writer = file
		h.localData["oldfilepath"] = filepath
	}

}
