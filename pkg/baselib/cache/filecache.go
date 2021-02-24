package cache

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/log"
)

type fileCacheInfo struct {
	file     *os.File
	lastTime time.Time
}

var ticker *time.Ticker
var files sync.Map //map[string]*fileCacheInfo

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getFile(filename string) (file *os.File, err error) {
	var exist bool
	// var cacheInfo *fileCacheInfo
	value, exist := files.Load(filename)
	if exist {
		cacheInfo := value.(*fileCacheInfo)
		file = cacheInfo.file
		return
	}

	dirname := path.Dir(filename)
	if dirname != "" && !IsExist(dirname) {
		err = os.MkdirAll(dirname, 0666)
		if err != nil {
			log.Errorf("MkdirAll(%s) failed! err: %v", dirname, err)
			return
		}
	}
	// file, err = os.Open(filename)
	file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("os.Open(%s) failed! err: %v", filename, err)
		return
	}

	cacheInfo := &fileCacheInfo{file: file, lastTime: time.Now()}
	files.Store(filename, cacheInfo)

	return
}

func putFile(filename string, file *os.File) {
	// 记录最后修改时间，用于对长时间没访问的文件进行清理。
	cacheInfo := &fileCacheInfo{file: file, lastTime: time.Now()}
	files.Store(filename, cacheInfo)
}

func CreateKeyValuePairsFromMap(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"  ", key, value)
	}
	return b.String()
}

func BizLogWrite(filename string, logstr string) (err error) {
	file, err := getFile(filename)
	if err != nil {
		return
	}
	loginfo := logstr + "\n"
	_, err = file.WriteString(loginfo)

	if err != nil {
		log.Errorf("failed to write file [%s], err: %v", filename, err)
	}

	putFile(filename, file)

	return
}

func cleanTimeout(timeout time.Duration) (count int) {
	now := time.Now().Add(timeout * -1)
	files.Range(func(key, value interface{}) bool {
		filename := key.(string)
		fileinfo := value.(*fileCacheInfo)
		lastTime := fileinfo.lastTime
		if lastTime.Before(now) {
			file := fileinfo.file
			err := file.Close()
			if err != nil {
				log.Errorf("failed to close file '%s', err: %v", filename, err)
			} else {
				log.Infof("success to close file '%s'", filename)
			}
			files.Delete(filename)
			count = count + 1
		}
		return true
	})
	return
}

func StartFileCleanupTimer(fdExpireTime time.Duration) {
	if ticker != nil {
		ticker.Stop()
	}

	ticker = time.NewTicker(fdExpireTime / 2)
	go func() {
		for range ticker.C {
			cleanTimeout(fdExpireTime)
		}
	}()
}
