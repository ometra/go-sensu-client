package main

// contains the code for logging to the android syslog
// borrowed from go.mobile/app

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

type infoWriter struct{}

var (
	ctagLog = C.CString("SensuClient")
)

func (infoWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	cstr := C.CString(string(p))
	C.__android_log_write(C.ANDROID_LOG_INFO, ctagLog, cstr)
	C.free(unsafe.Pointer(cstr))
	return len(p), nil
}

func init() {
	log.SetOutput(infoWriter{})
	// android logcat includes all of log.LstdFlags
	log.SetFlags(log.Flags() &^ log.LstdFlags)
	logOutput = infoWriter{}
}
