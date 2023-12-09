package grlog

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
)

type RotatingFile struct {
	file        *os.File
	fileName    string
	maxFileSize int64
	backupCount int
	mutex       sync.Mutex
	async       bool //async write
	writeChan   chan []byte
	errorChan   chan error
}

const (
	defaultFileSize = 1 << 24 //16384 kb
)

func NewRotatingFile(fileName string, backupCount int, fileSize int64, async bool) (*RotatingFile, error) {
	if fileSize < 0 {
		fileSize = defaultFileSize
	} else if fileSize < 1024 {
		return nil, errors.New("file size must be than greater 1024")
	}
	filePath := path.Dir(fileName)
	if err := os.MkdirAll(filePath, 0664); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return nil, err
	}
	rf := &RotatingFile{
		file:        file,
		fileName:    fileName,
		maxFileSize: fileSize,
		backupCount: backupCount,
		async:       async,
	}
	if async {
		rf.writeChan = make(chan []byte, 10)
		rf.errorChan = make(chan error, 1)
		go rf.awaitWrite()
	}
	return rf, nil
}

func (self *RotatingFile) Write(p []byte) (n int, err error) {
	if err = self.rotating(int64(len(p))); err != nil {
		return
	}
	if self.async {
		select {
		case self.writeChan <- p:
			return len(p), nil
		case err = <-self.errorChan:
			return 0, err
		}
	}
	return self.file.Write(p)
}

func (self *RotatingFile) Close() error {
	if self.async {
		self.async = false
		close(self.writeChan)
		close(self.errorChan)
	}
	return self.file.Close()
}

func (self *RotatingFile) IsAsync() bool {
	return self.async
}

func (self *RotatingFile) rotating(wn int64) (err error) {
	if self.backupCount < 1 {
		return
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	fileInfo, err := self.file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size()+wn < self.maxFileSize {
		return
	}

	var oldPath, newPath string
	for i := self.backupCount - 1; i > 0; i-- {
		oldPath = fmt.Sprintf("%s-%d", self.fileName, i)
		newPath = fmt.Sprintf("%s-%d", self.fileName, i+1)
		_ = os.Rename(oldPath, newPath)
	}
	_ = self.file.Sync()
	_ = self.file.Close()
	newPath = self.fileName + "-1"
	if err = os.Rename(self.fileName, newPath); err != nil {
		return err
	}
	self.file, err = os.OpenFile(self.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	return err
}

func (self *RotatingFile) awaitWrite() {
	for data := range self.writeChan {
		if _, err := self.file.Write(data); err != nil {
			self.errorChan <- err
		}
	}
}
