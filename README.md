# grlog - A golang rotate file logging package

grlog is a simple log package that is extended from the log standard library.

## Features
*  rotate file and timed rotate file
*  support asynchronous writing
*  log level

## Use
download
```shell
go get -u github.com/shaopson/grlog
```
import
```go
import "github.com/shaopson/grlog"
```

### Example
```go
package main

import "github.com/shaopson/grlog"

func main() {
    //shortcuts
    //default stderr writer
    grlog.Debug("debug")
    grlog.Info("info")
    grlog.Warn("warn")
    grlog.Error("error")
    grlog.SetLevel(grlog.LevelDebug)
    grlog.Log(grlog.LevelInfo, "hello")
}

```

### Rotate file
```go
//5 backup files,  default file size,  sync write mode
writer, err := grlog.NewRotateFile("test.log", 5, 0, false)
defer writer.Close()
if err != nil {
    panic(err)
}
log := grlog.Default()
log.SetOutput(writer)
log.Info("debug")

// timed rotate file,
writer, err := grlog.NewTimedRotateFile("test.log", 5, 0, false)
```

### Async Write
```go
writer, err := grlog.NewRotatingFile("test.log", 5, -1, true)
// don`t forget close!!!
defer writer.Close()
```
