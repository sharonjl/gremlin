package gremlin

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
	"encoding/base64"
	"github.com/davecgh/go-spew/spew"
)

var ErrConnClosed = errors.New("gremlinx: connection closed")
var ErrNoConnectionsAvailable = errors.New("gremlinx: all connections closed")

type Gremlin struct {
	conns                    chan *websocket.Conn
	mu                       sync.Mutex
	host, username, password string
	connections              uint
	maxConnections           uint
	minConnections           uint
	io, writer               chan []byte
	done                     chan bool
}

func (g *Gremlin) newConnection() (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(g.host, http.Header{})
	return conn, err
}

func (g *Gremlin) acquireConn() (*websocket.Conn, error) {
	select {
	case c := <-g.conns:
		// Return one of the buffered connections.
		if c != nil {
			return c, nil
		}
	default:
		// We haven't reached the max connection count,
		g.mu.Lock()
		if g.connections < g.maxConnections {
			defer g.mu.Unlock()

			c, err := g.newConnection()
			if err != nil {
				return nil, err
			}
			g.connections++
			return c, err
		}
		g.mu.Unlock()

		// Already at maximum number of connections, wait until
		// a connection is available.
		c := <-g.conns
		if c != nil {
			return c, nil
		}
	}
	return nil, ErrNoConnectionsAvailable
}

func (g *Gremlin) releaseConn(conn *websocket.Conn) error {
	if conn == nil {
		return nil
	}

	select {
	case g.conns <- conn:
		// Push connection back into the pool.
		return nil
	default:
		g.mu.Lock()
		defer g.mu.Unlock()

		// Keep these connections, if we still need to meet the
		// minimum connection requirement.
		if g.connections <= g.minConnections {
			return nil
		}

		// Close connections we don't need to keep around.
		err := conn.Close()
		if err != nil {
			return err
		}
		g.connections--

	}
	return nil
}

func (g *Gremlin) exec(b []byte) (data []byte, err error) {
	conn, err := g.acquireConn()
	defer g.releaseConn(conn)
	if err != nil {
		return
	}
	return exec(conn, b, g.username, g.password)
}

func (g *Gremlin) Eval(in *EvalInput) (RawOutput, error) {
	in.Language = "gremlin-groovy"
	req := &Request{
		RequestID: uuid.New(),
		Args:      in,
		Processor: "",
		Op:        OpEval,
	}
	b, err := build(req)
	if err != nil {
		return nil, err
	}
	return g.exec(b)
}

func exec(conn *websocket.Conn, b []byte, username, password string) (data []byte, err error) {
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		if err == websocket.ErrCloseSent {
			return nil, ErrConnClosed
		}
		return nil, err
	}
	data, err = readResponse(conn, username, password)
	return
}

func readResponse(src *websocket.Conn, username, password string) (data []byte, err error) {
	var batched bool
	var batchedItems []json.RawMessage
	for {
		resp := &Response{}
		err = src.ReadJSON(resp)
		if err != nil {
			return
		}
		var d []json.RawMessage
		switch resp.Status.Code {
		case StatusNoContent:
			return // No more messages to read
		case StatusPartialContent:
			batched = true
			if err = json.Unmarshal(resp.Result.Data, &d); err != nil {
				return
			}
			batchedItems = append(batchedItems, d...)
		case StatusSuccess:
			if batched {
				if err = json.Unmarshal(resp.Result.Data, &d); err != nil {
					return
				}
				batchedItems = append(batchedItems, d...)
				data, err = json.Marshal(batchedItems)
			} else {
				data = resp.Result.Data
			}
			return
		case StatusAuthenticate:
			var sasl []byte
			sasl = append(sasl, 0)
			sasl = append(sasl, []byte(username)...)
			sasl = append(sasl, 0)
			sasl = append(sasl, []byte(password)...)
			saslEnc := base64.StdEncoding.EncodeToString(sasl)

			var b []byte
			b, err = build(&Request{
				RequestID: resp.RequestID,
				Op:        OpAuthentication,
				Processor: "",
				Args: &AuthenticationInput{
					SASL: saslEnc,
				},
			})
			if err != nil {
				return
			}

			err = src.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				if err == websocket.ErrCloseSent {
					return nil, ErrConnClosed
				}
				return nil, err
			}
		default:
			if m, ok := StatusMessages[resp.Status.Code]; ok {
				err = errors.New(m)
			} else {
				err = errors.New("an unknown error occured")
			}
			return
		}
	}
}

func build(req *Request) ([]byte, error) {
	spew.Dump(req)
	message, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var reqMsg []byte
	mimeType := []byte("application/json")
	mimeTypeLen := byte(len(mimeType))
	reqMsg = append(reqMsg, mimeTypeLen)
	reqMsg = append(reqMsg, mimeType...)
	reqMsg = append(reqMsg, message...)
	return reqMsg, nil
}

const (
	MaxConnections = 5
	MinConnections = 3
)

func Open(urlStr string, options ...func(client *Gremlin)) (*Gremlin, error) {
	g := &Gremlin{
		host:           urlStr,
		writer:         make(chan []byte),
		io:             make(chan []byte),
		done:           make(chan bool),
		conns:          make(chan *websocket.Conn, MaxConnections),
		maxConnections: MaxConnections,
		minConnections: MinConnections,
	}

	for _, opt := range options {
		opt(g)
	}

	for i := uint(0); i < g.minConnections; i++ {
		go func() {
			c, err := g.newConnection()
			if err == nil {
				g.conns <- c
			}
		}()
	}
	return g, nil
}

func WithAuthentication(username, password string) func(client *Gremlin) {
	return func(c *Gremlin) {
		c.username = username
		c.password = password
	}
}
