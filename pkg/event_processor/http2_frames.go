// Copyright 2022 CFC4N <cfc4n.cs@gmail.com>. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event_processor

import (
	"bufio"
	"bytes"
	"fmt"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

type HTTP2Frame struct {
	frame      http2.Frame
	packerType PacketType
	isDone     bool
	isInit     bool
	reader     *bytes.Buffer
	bufReader  *bufio.Reader
}

func (hr *HTTP2Frame) Init() {
	hr.reader = bytes.NewBuffer(nil)
	hr.bufReader = bufio.NewReader(hr.reader)
	hr.packerType = PacketTypeHTTP2Frame
}

func (hr *HTTP2Frame) Name() string {
	return "HTTP2Frame"
}

func (hr *HTTP2Frame) PacketType() PacketType {
	return hr.packerType
}

func (hr *HTTP2Frame) ParserType() ParserType {
	return ParserTypeHttp2Frame
}

func (hr *HTTP2Frame) Write(b []byte) (int, error) {
	// 如果未初始化
	if !hr.isInit {
		n, e := hr.reader.Write(b)
		if e != nil {
			return n, e
		}

		fr := http2.NewFramer(nil, hr.bufReader)
		fr.ReadMetaHeaders = hpack.NewDecoder(0, nil)
		// fr.ReadMetaHeaders.SetEmitFunc(emitFunc)
		// fr.ReadMetaHeaders.SetEmitEnabled(true)

		frame, err := fr.ReadFrame()
		if err != nil {
			return 0, err
		}
		hr.frame = frame
		hr.isInit = true
		return n, nil
	}

	// 如果已初始化
	l, e := hr.reader.Write(b)
	if e != nil {
		return 0, e
	}
	// TODO 检测是否接收完整个包
	if false {
		hr.isDone = true
	}

	return l, nil
}

func (hr *HTTP2Frame) detect(payload []byte) error {
	//hr.Init()
	rd := bytes.NewReader(payload)
	buf := bufio.NewReader(rd)

	fr := http2.NewFramer(nil, buf)
	fr.ReadMetaHeaders = hpack.NewDecoder(0, nil)
	// fr.ReadMetaHeaders.SetEmitFunc(emitFunc)
	// fr.ReadMetaHeaders.SetEmitEnabled(true)
	f, err := fr.ReadFrame()
	if err != nil {
		return err
	}

	hr.frame = f
	return nil
}

func (hr *HTTP2Frame) IsDone() bool {
	return hr.isDone
}

func (hr *HTTP2Frame) Reset() {
	hr.isDone = false
	hr.isInit = false
	hr.reader.Reset()
	hr.bufReader.Reset(hr.reader)
}

func (hr *HTTP2Frame) Display() []byte {
	// b, e := httputil.DumpRequest(hr.request, true)
	// if e != nil {
	// 	log.Println("DumpRequest error:", e)
	// 	return hr.reader.Bytes()
	// }

	display := fmt.Sprintf("HTTP2Frame: %#v", hr.frame)

	return []byte(display)
}

func init() {
	hr := &HTTP2Frame{}
	hr.Init()
	Register(hr)
}

func BytesToHTTP2Frame(b []byte) (http2.Frame, error) {

	rd := bytes.NewReader(b)
	buf := bufio.NewReader(rd)

	fr := http2.NewFramer(nil, buf)
	fr.ReadMetaHeaders = hpack.NewDecoder(0, nil)
	// fr.ReadMetaHeaders.SetEmitFunc(emitFunc)
	// fr.ReadMetaHeaders.SetEmitEnabled(true)
	f, err := fr.ReadFrame()
	if err != nil {
		return nil, err
	}
	return f, nil
}
