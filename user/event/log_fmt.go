package event

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"

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

func LogString(exe []byte, timestamp uint64, data []byte) string {
	logFmt := new(LogFmt)
	logFmt.Executable = unix.ByteSliceToString(exe)
	logFmt.Timestamp = timestamp

	var req *http.Request = nil

	frame, err := http2util.BytesToFrame(data)

	if err == nil && http2util.GetFrameType(frame) == http2.FrameHeaders {
		req, _ = http2util.FrameToHTTPRequest(frame.(*http2.MetaHeadersFrame))

	} else {
		
		rd := bytes.NewReader(data)
		bufRd := bufio.NewReader(rd)
		req, _ = http.ReadRequest(bufRd)

	}

	if req != nil {

		logFmt.Method = req.Method
		logFmt.Host = req.Host
		logFmt.Path = req.RequestURI

		return logFmt.String()
	}

	return ""

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
