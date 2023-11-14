package event

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/h0x0er/http2util"
	"golang.org/x/net/http2"
	"golang.org/x/sys/unix"
)

type LogFmt struct {
	Timestamp  uint64 `json:"timestamp"`
	Executable string `json:"exe"`
	Host       string `json:"host"`
	Path       string `json:"path"`
	Method     string `json:"method"`
}

func (l *LogFmt) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

func LogSSLEvent(probeType int64, exe []byte, timestamp uint64, data [4096]byte, dataLen int32) string {

	if AttachType(probeType) != ProbeRet {
		return ""
	}

	return LogString(exe, timestamp, data[:dataLen])

}

func LogGoTLSEvent(payloadType uint8, exe []byte, timestamp uint64, data []byte, dataLen int32) string {

	if payloadType != 0 {
		return ""
	}
	return LogString(exe, timestamp, data[:dataLen])

}

func LogString(exe []byte, timestamp uint64, data []byte) string {
	logFmt := new(LogFmt)
	logFmt.Executable = unix.ByteSliceToString(exe)

	t, err := DecodeKtime(int64(timestamp), false)
	if err == nil {
		logFmt.Timestamp = uint64(t.Unix())
	} else {
		logFmt.Timestamp = timestamp

	}

	var req *http.Request = nil

	frame, err := http2util.BytesToFrame(data)

	if err == nil && http2util.GetFrameType(frame) == http2.FrameHeaders {
		req, _ = http2util.FrameToHTTPRequest(frame)

	} else {

		rd := bytes.NewReader(data)
		bufRd := bufio.NewReader(rd)
		req, _ = http.ReadRequest(bufRd)

	}

	if req != nil {

		logFmt.Method = req.Method
		logFmt.Host = req.Host
		logFmt.Path = req.RequestURI

		out := logFmt.String()

		go writeLog(out)

		return out
	}

	return ""

}

var logMutex sync.Mutex

func writeLog(message string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	f, _ := os.OpenFile("/home/runner/work/_temp/network_events.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer f.Close()

	location, _ := time.LoadLocation("Etc/GMT")
	f.WriteString(fmt.Sprintf("%s:%s\n", time.Now().In(location).Format("Mon, 02 Jan 2006 15:04:05 MST"), message))

}

func DecodeKtime(ktime int64, monotonic bool) (time.Time, error) {
	var clk int32
	if monotonic {
		clk = int32(unix.CLOCK_MONOTONIC)
	} else {
		clk = int32(unix.CLOCK_BOOTTIME)
	}
	currentTime := unix.Timespec{}
	if err := unix.ClockGettime(clk, &currentTime); err != nil {
		return time.Time{}, err
	}
	diff := ktime - currentTime.Nano()
	t := time.Now().Add(time.Duration(diff))
	return t, nil
}
