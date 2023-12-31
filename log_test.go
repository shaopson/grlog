package grlog

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	log := New(os.Stderr, "", FlagStd|FlagSFile, LevelInfo)
	log.SetLevel(LevelInfo)
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warning message")
	log.Error("error message")
	SetLevel(LevelWarn)
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("Error")
}

func TestRotatingFile(t *testing.T) {
	writer, err := NewRotateFile("test.log", 5, 1024, false)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	log := New(os.Stderr, "", FlagStd|FlagLevel, LevelInfo)

	var wait sync.WaitGroup
	buf := bytes.NewBuffer(nil)
	for i := 0; i < 100; i++ {
		buf.WriteString(strconv.Itoa(i))
	}
	s := buf.String()
	now := time.Now()
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			log.Info(s)
			wait.Done()
		}()
	}
	wait.Wait()
	fmt.Println("sync write time:", time.Since(now))
}

func TestAsyncWrite(t *testing.T) {
	writer, err := NewRotateFile("async.log", 5, 1024, true)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	log := New(os.Stderr, "", FlagStd|FlagLevel, LevelInfo)

	var wait sync.WaitGroup
	buf := bytes.NewBuffer(nil)
	for i := 0; i < 100; i++ {
		buf.WriteString(strconv.Itoa(i))
	}
	s := buf.String()
	now := time.Now()
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			log.Info(s)
			wait.Done()
		}()
	}
	wait.Wait()
	fmt.Println("async write time:", time.Since(now))
}

func TestTimedRotateFile_Write(t *testing.T) {

	writer, err := NewTimedRotateFile("a.log", 3, 1024, false)
	if err != nil {
		t.Fatal(err)
	}
	log := Default()
	log.SetOutput(writer)
	for i := 0; i < 10; i++ {
		log.Info("TestTimedRotateFile_Write")
		//time.Sleep(time.Second)
	}
}
