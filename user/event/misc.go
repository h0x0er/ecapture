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

package event

import (
	"bufio"
	"bytes"
	"fmt"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// 格式化输出相关

const ChunkSize = 16
const ChunkSizeHalf = ChunkSize / 2

const (
	COLORRESET  = "\033[0m"
	COLORRED    = "\033[31m"
	COLORGREEN  = "\033[32m"
	COLORYELLOW = "\033[33m"
	COLORBLUE   = "\033[34m"
	COLORPURPLE = "\033[35m"
	COLORCYAN   = "\033[36m"
	COLORWHITE  = "\033[37m"
)

func dumpByteSlice(b []byte, perfix string) *bytes.Buffer {
	var a [ChunkSize]byte
	bb := new(bytes.Buffer)
	n := (len(b) + (ChunkSize - 1)) &^ (ChunkSize - 1)

	for i := 0; i < n; i++ {

		// 序号列
		if i%ChunkSize == 0 {
			bb.WriteString(perfix)
			bb.WriteString(fmt.Sprintf("%04d", i))
		}

		// 长度的一半，则输出4个空格
		if i%ChunkSizeHalf == 0 {
			bb.WriteString("    ")
		} else if i%(ChunkSizeHalf/2) == 0 {
			bb.WriteString("  ")
		}

		if i < len(b) {
			bb.WriteString(fmt.Sprintf(" %02X", b[i]))
		} else {
			bb.WriteString("  ")
		}

		// 非ASCII 改为 .
		if i >= len(b) {
			a[i%ChunkSize] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%ChunkSize] = '.'
		} else {
			a[i%ChunkSize] = b[i]
		}

		// 如果到达size长度，则换行
		if i%ChunkSize == (ChunkSize - 1) {
			bb.WriteString(fmt.Sprintf("    %s\n", string(a[:])))
		}
	}
	return bb
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
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
