package easylog

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

const (
	Rdate         = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime         = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	Rmicroseconds = `\.[0-9][0-9][0-9][0-9][0-9][0-9]`
)

var output = map[int]func(v ...interface{}){
	Ldebug: Debug,
	Linfo:  Info,
	Lwarn:  Warn,
	Lerror: Error,
}

var outputF = map[int]func(format string, v ...interface{}){
	Ldebug: Debugf,
	Linfo:  Infof,
	Lwarn:  Warnf,
	Lerror: Errorf,
}

func testPrint(t *testing.T, flag int, prefix string, pattern string, level int, useFormat bool) {
	buf := new(bytes.Buffer)
	SetOutput(buf)
	SetFlags(flag)
	SetPrefix(prefix)
	SetLevel(Ldebug)
	if useFormat {
		outputF[level]("hello %s", "world")
	} else {
		output[level]("hello world")
	}
	line := buf.String()
	line = line[0 : len(line)-1]
	pattern = "^" + pattern + "\\" + levels[level] + "\\" + " " + "hello world$"
	matched, err := regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatal("pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern, line)
	}
	SetOutput(os.Stderr)
}

func TestAll(t *testing.T) {
	tests := []struct {
		flag    int
		prefix  string
		pattern string
		level   int
	}{
		{0, "", "", Linfo},
		{0, "XXX", "XXX", Linfo},
		{Ldate, "", Rdate + " ", Linfo},
		{Ltime, "", Rtime + " ", Linfo},
		{Ltime | Lmicroseconds, "", Rtime + Rmicroseconds + " ", Ldebug},
		{Lmicroseconds, "", Rtime + Rmicroseconds + " ", Lwarn},
		{Ldate | Ltime | Lmicroseconds, "XXX", "XXX" + Rdate + " " + Rtime + Rmicroseconds + " ", Lerror},
	}

	for _, test := range tests {
		testPrint(t, test.flag, test.prefix, test.pattern, test.level, false)
		testPrint(t, test.flag, test.prefix, test.pattern, test.level, true)
	}
}

func ExampleAll() {
	logger := New(os.Stdout, "", 0)
	logger.SetLevel(Ldebug)
	logger.Debug("hello world")
	logger.Debugf("hello %s", "world")
	logger.Info("hello world")
	logger.Infof("hello %s", "world")
	logger.Warn("hello world")
	logger.Warnf("hello %s", "world")
	logger.Error("hello world")
	logger.Errorf("hello %s", "world")
	// Output:
	// [DEBUG] hello world
	// [DEBUG] hello world
	// [INFO] hello world
	// [INFO] hello world
	// [WARN] hello world
	// [WARN] hello world
	// [ERROR] hello world
	// [ERROR] hello world
}
