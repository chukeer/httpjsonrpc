package httpjsonrpc

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type RWC struct {
	io.Reader
	io.Writer
	io.Closer
}

func newRWC(r io.Reader, w io.Writer, c io.Closer) *RWC {
	return &RWC{
		Reader: r,
		Writer: w,
		Closer: c,
	}
}

type serverCodec struct {
	r         *bufio.Reader
	conn      net.Conn
	jsonCodec rpc.ServerCodec
	httpReq   *http.Request
	reply     *bytes.Buffer
}

func NewServerCodec(conn net.Conn) rpc.ServerCodec {
	return &serverCodec{
		r:     bufio.NewReader(conn),
		conn:  conn,
		reply: &bytes.Buffer{},
	}
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	var err error
	if c.httpReq, err = http.ReadRequest(c.r); err != nil {
		return err
	}

	c.jsonCodec = jsonrpc.NewServerCodec(newRWC(c.httpReq.Body, c.reply, c.conn))
	return c.jsonCodec.ReadRequestHeader(r)
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	defer c.httpReq.Body.Close()
	return c.jsonCodec.ReadRequestBody(x)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if err := c.jsonCodec.WriteResponse(r, x); err != nil {
		return err
	}

	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBuffer(c.reply.Bytes())),
		ContentLength: int64(c.reply.Len()),
		Request:       c.httpReq,
		Header:        make(http.Header, 0),
	}

	return resp.Write(c.conn)
}

func (c *serverCodec) Close() error {
	return c.conn.Close()
}
