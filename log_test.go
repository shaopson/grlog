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
	log := New(os.Stderr, LevelDebug, DefaultFlag)

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warning message")
	log.Error("error message")

}

func TestRotatingFile(t *testing.T) {
	writer, err := NewRotatingFile("test.log", 5, 1024, false)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	log := New(writer, LevelDebug, DefaultFlag)

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
	writer, err := NewRotatingFile("async.log", 5, 1024, true)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	log := New(writer, LevelDebug, DefaultFlag)

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
