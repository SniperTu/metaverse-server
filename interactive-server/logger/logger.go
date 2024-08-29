package logger

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var lgr = logrus.New()

func Debugf(f string, args ...interface{}) {
	lgr.Debugf(f, args...)
}

func Infof(f string, args ...interface{}) {
	lgr.Infof(f, args...)
}

func Warnf(f string, args ...interface{}) {
	lgr.Warnf(f, args...)
}

func Errorf(f string, args ...interface{}) {
	lgr.Errorf(f, args...)
}

func Fatalf(f string, args ...interface{}) {
	lgr.Fatalf(f, args...)
}

func callerFileLineStack(start, depth int) []string {
	stacks := []string{}
	for {
		_, file, line, ok := runtime.Caller(start)
		if !ok || len(stacks) >= depth {
			break
		}
		if strings.HasSuffix(file, ".go") {
			stacks = append(stacks, fmt.Sprintf("%s:%d", suffixPath(file), line))
		}
		start++
	}
	return stacks
}

func suffixPath(abs string) string {
	idents := strings.Split(abs, "/")
	var path string
	if l := len(idents); l > 3 {
		path = filepath.Join(idents[l-3:]...)
	} else {
		path = abs
	}
	return path
}

func init() {
	logFileName := "logs/server.log"
	rl, err := rotatelogs.New(logFileName+".%Y%m%d",
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithLinkName(logFileName),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithRotationSize(100*1024*1024),
	)
	if err != nil {
		log.Fatal(err)
	}
	lgr.SetOutput(rl)
	lgr.SetReportCaller(true)
	lgr.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		PrettyPrint:     true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			skip := 9
			p, _, _, _ := runtime.Caller(skip)
			function = runtime.FuncForPC(p).Name()
			file = strings.Join(callerFileLineStack(skip+1, 10), "|")
			return
		},
	})
}
