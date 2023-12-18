package grlog

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

type TimedRotateFile struct {
	file        *os.File
	fileName    string
	maxFileSize int64
	backupCount int
	mutex       sync.Mutex
	async       bool //async write
	writeChan   chan []byte
	errorChan   chan error
	filePattern *regexp.Regexp
	rotateTime  time.Time
}

// backup yesterday's files at 00:00 every day
// fileName: log file path: a/b/c.log
// backupCount: backup files, if backupCount=3: a.log  a.log-2023-12-01  a.log-2023-12-02  a.log-2023-12-03
// fileSize: log file max size, default size 16m
// async: asynchronous write
func NewTimedRotateFile(fileName string, backupCount int, fileSize int64, async bool) (*TimedRotateFile, error) {
	if fileSize <= 0 {
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

	rf := &TimedRotateFile{
		file:        file,
		fileName:    fileName,
		maxFileSize: fileSize,
		backupCount: backupCount,
		async:       async,
		filePattern: regexp.MustCompile(fmt.Sprintf("^%s-\\d{4}-\\d{2}-\\d{2}$", fileName)),
	}
	stat, _ := os.Stat(fileName)
	rf.setRotateTime(stat.ModTime())
	if async {
		rf.writeChan = make(chan []byte, 10)
		rf.errorChan = make(chan error, 1)
		go rf.awaitWrite()
	}
	return rf, nil
}

func (self *TimedRotateFile) Write(p []byte) (n int, err error) {
	if err = self.rotate(int64(len(p))); err != nil {
		panic(err)
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

func (self *TimedRotateFile) Close() error {
	if self.async {
		self.async = false
		close(self.writeChan)
		close(self.errorChan)
	}
	return self.file.Close()
}

func (self *TimedRotateFile) IsAsync() bool {
	return self.async
}

func (self *TimedRotateFile) awaitWrite() {
	for data := range self.writeChan {
		if _, err := self.file.Write(data); err != nil {
			self.errorChan <- err
		}
	}
}

func (self *TimedRotateFile) setRotateTime(t time.Time) {
	y, m, d := t.Date()
	self.rotateTime = time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
}

func (self *TimedRotateFile) rotate(wn int64) (err error) {
	if self.backupCount < 1 {
		return
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	fileInfo, err := self.file.Stat()
	if err != nil {
		return err
	}
	now := time.Now()
	if now.Before(self.rotateTime) {
		if fileInfo.Size()+wn < self.maxFileSize {
			return nil
		}
	}

	date := fileInfo.ModTime().Format("2006-01-02")
	newPath := fmt.Sprintf("%s-%s", self.fileName, date)
	for i := 1; true; i++ {
		if _, err = os.Stat(newPath); err != nil {
			if os.IsNotExist(err) {
				err = os.Rename(self.fileName, newPath)
				break
			} else {
				return err
			}
		}
		newPath = fmt.Sprintf("%s-%s-%d", self.fileName, date, i)
	}
	self.file.Sync()
	self.file.Close()

	self.file, err = os.OpenFile(self.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	self.setRotateTime(now)
	self.prune()
	return
}

// delete expired files
func (self *TimedRotateFile) prune() {
	dir := path.Dir(self.fileName)
	// sorted fs
	fs, _ := os.ReadDir(dir)
	files := make([]string, 0, len(fs))
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		if self.filePattern.MatchString(f.Name()) {
			files = append(files, f.Name())
		}
	}

	if len(files) > self.backupCount {
		files = files[:len(files)-self.backupCount]
		for _, f := range files {
			err := os.Remove(f)
			fmt.Println("delete", err)
		}
	}
}
