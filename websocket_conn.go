// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
)

// ErrNotBinary may be returned by WebSocket connection when a message is
// received that is not binary.
var ErrNotBinary = errors.New("received web socket message is not binary")

var closeMessage = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")

type webSocketStream struct {
	conn   *websocket.Conn
	reader io.Reader
}

func (s *webSocketStream) Read(p []byte) (int, error) {
	total := 0
	buf := p

	for {
		// get next reader
		if s.reader == nil {
			messageType, reader, err := s.conn.NextReader()
			if _, ok := err.(*websocket.CloseError); ok {
				return 0, io.EOF
			} else if err != nil {
				return 0, err
			} else if messageType != websocket.BinaryMessage {
				return 0, ErrNotBinary
			}

			// set current reader
			s.reader = reader
		}

		// read data
		n, err := s.reader.Read(buf)

		// increment counter
		total += n
		buf = buf[n:]

		// handle EOF
		if err == io.EOF {
			// clear reader
			s.reader = nil

			continue
		}

		// handle other errors
		if err != nil {
			return total, err
		}

		return total, err
	}
}

func (s *webSocketStream) Write(p []byte) (n int, err error) {
	// create writer if missing
	writer, err := s.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	// write packet to writer
	n, err = writer.Write(p)
	if err != nil {
		return n, err
	}

	// close temporary writer
	err = writer.Close()
	if err != nil {
		return n, err
	}

	return n, nil
}

func (s *webSocketStream) Close() error {
	// write close message
	err := s.conn.WriteMessage(websocket.CloseMessage, closeMessage)
	if err != nil {
		return err
	}

	return s.conn.Close()
}

func (s *webSocketStream) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

// The WebSocketConn wraps a websocket.Conn. The implementation supports packets
// that are chunked over several WebSocket messages and packets that are coalesced
// to one WebSocket message.
type WebSocketConn struct {
	BaseConn

	conn *websocket.Conn
}

// NewWebSocketConn returns a new WebSocketConn.
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	s := &webSocketStream{
		conn: conn,
	}

	return &WebSocketConn{
		BaseConn: BaseConn{
			carrier: s,
			stream:  packet.NewStream(s, s),
		},
		conn: conn,
	}
}

// TODO: Move LocalAddr and RemoteAddr to Stream?

// LocalAddr returns the local network address.
func (c *WebSocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *WebSocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// UnderlyingConn returns the underlying websocket.Conn.
func (c *WebSocketConn) UnderlyingConn() *websocket.Conn {
	return c.conn
}
