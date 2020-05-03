// Copyright 2013 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package listener

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/event"
	pkgLine "github.com/prometheus/statsd_exporter/pkg/line"
	protobufmessage "github.com/prometheus/statsd_exporter/pkg/protobufmessage"
)

type StatsDUDPListener struct {
	Conn            *net.UDPConn
	EventHandler    event.EventHandler
	Logger          log.Logger
	UDPPackets      prometheus.Counter
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
}

func (l *StatsDUDPListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDUDPListener) Listen() {
	buf := make([]byte, 65535)
	for {
		n, _, err := l.Conn.ReadFromUDP(buf)
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log("error", err)
			return
		}
		l.HandlePacket(buf[0:n])
	}
}

func (l *StatsDUDPListener) HandlePacket(packet []byte) {
	l.UDPPackets.Inc()
	lines := strings.Split(string(packet), "\n")
	for _, line := range lines {
		l.LinesReceived.Inc()
		l.EventHandler.Queue(pkgLine.LineToEvents(line, l.SampleErrors, l.SamplesReceived, l.TagErrors, l.TagsReceived, l.Logger))
	}
}

type StatsDTCPListener struct {
	Conn            *net.TCPListener
	EventHandler    event.EventHandler
	Logger          log.Logger
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
	TCPConnections  prometheus.Counter
	TCPErrors       prometheus.Counter
	TCPLineTooLong  prometheus.Counter
}

func (l *StatsDTCPListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDTCPListener) Listen() {
	for {
		c, err := l.Conn.AcceptTCP()
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log("msg", "AcceptTCP failed", "error", err)
			os.Exit(1)
		}
		go l.HandleConn(c)
	}
}

func (l *StatsDTCPListener) HandleConn(c *net.TCPConn) {
	defer c.Close()

	l.TCPConnections.Inc()

	var buf bytes.Buffer
	io.Copy(&buf, c)
	level.Error(l.Logger).Log("msg", "Read", "addr", c.RemoteAddr(), "buf.Len()", buf.Len())
	message := &protobufmessage.TraceMetric{}
	err := proto.Unmarshal(buf.Bytes(), message)
	if err != nil {
		return
	}

	l.EventHandler.Queue(protobufmessage.MessageToEvent(*message))
	l.LinesReceived.Inc()
}

type StatsDUnixgramListener struct {
	Conn            *net.UnixConn
	EventHandler    event.EventHandler
	Logger          log.Logger
	UnixgramPackets prometheus.Counter
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
}

func (l *StatsDUnixgramListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDUnixgramListener) Listen() {
	buf := make([]byte, 65535)
	for {
		n, _, err := l.Conn.ReadFromUnix(buf)
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log(err)
			os.Exit(1)
		}
		l.HandlePacket(buf[:n])
	}
}

func (l *StatsDUnixgramListener) HandlePacket(packet []byte) {
	l.UnixgramPackets.Inc()
	lines := strings.Split(string(packet), "\n")
	for _, line := range lines {
		l.LinesReceived.Inc()
		l.EventHandler.Queue(pkgLine.LineToEvents(line, l.SampleErrors, l.SamplesReceived, l.TagErrors, l.TagsReceived, l.Logger))
	}
}
