package proxy

import (
	"io"
	"net"
)

// prefetchReader can prefetch one byte from the reader before Read()
type prefetchReader struct {
	src io.Reader

	first   byte
	reseted bool
	err     error

	obpool *OneBytePool
}

func (ur *prefetchReader) prefetch() (byte, error) {
	buf := ur.obpool.Get()
	defer ur.obpool.Free(buf)

	n, err := ur.src.Read(buf)
	ur.err = err

	if n == 1 {
		ur.first = buf[0]
		return buf[0], nil
	}

	return buf[0], err
}

func (ur *prefetchReader) Read(p []byte) (int, error) {
	if ur.err != nil {
		return 0, ur.err
	}

	if !ur.reseted {
		p[0] = ur.first
		xp := p[1:]

		n, err := ur.src.Read(xp)
		ur.reseted = true

		return n + 1, err
	}

	return ur.src.Read(p)
}

type OneBytePool chan []byte

func NewOneBytePool(s int) *OneBytePool {
	p := OneBytePool(make(chan []byte, s))
	return &p
}

func (p *OneBytePool) Get() []byte {
	select {
	case buf := <-*p:
		return buf
	default:
		return make([]byte, 1)
	}
}

func (p *OneBytePool) Free(buf []byte) {
	select {
	case *p <- buf:
	default:
	}
}

type connWrapper struct {
	net.Conn
	sbuffer *prefetchReader
}

func (cw *connWrapper) Read(p []byte) (int, error) {
	return cw.sbuffer.Read(p)
}

type listenerWrapper struct {
	net.Listener

	proxy     *ProxyClient
	httpConn  connWrapper
	socksConn connWrapper

	obpool *OneBytePool
}

func (l *listenerWrapper) Accept() (net.Conn, error) {
CONTINUE:
	c, err := l.Listener.Accept()
	wrapper := &connWrapper{Conn: c, sbuffer: &prefetchReader{src: c, obpool: l.obpool}}

	switch b, _ := wrapper.sbuffer.prefetch(); b {
	case 0x04, 0x05:
		// we are accepting SOCKS4 in case it goes to the HTTP handler
		go l.proxy.handleSocks(wrapper)
		goto CONTINUE
	default:
		return wrapper, err
	}
}
