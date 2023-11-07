package event

import (
	"encoding/json"

	"github.com/h0x0er/http2util"
	"golang.org/x/net/http2"
	"golang.org/x/sys/unix"
)

type LogFmt struct {
	Timestamp  uint64 `json:"timestamp"`
	Executable string `json:"executable"`
	Data       string `json:"data"`
}

func (l *LogFmt) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

func LogString(exe []byte, timestamp uint64, data []byte) string {
	logFmt := new(LogFmt)
	logFmt.Executable = unix.ByteSliceToString(exe)
	logFmt.Timestamp = timestamp
	logFmt.Data = ""

	frame, err := http2util.BytesToHTTP2Frame(data)

	if err == nil && http2util.GetFrameType(frame) == http2.FrameHeaders {

		s, err := http2util.Dump(frame)
		if err == nil {
			logFmt.Data = s
		}

	} else {
		logFmt.Data = unix.ByteSliceToString(data)
	}

	if len(logFmt.Data) > 0 {
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
