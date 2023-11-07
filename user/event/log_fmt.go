package event

import (
	"encoding/json"
)

type LogFmt struct {
	Timestamp  string `json:"timestamp"`
	Executable string `json:"executable"`
	Data       string `json:"data"`
}

func (l *LogFmt) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}
