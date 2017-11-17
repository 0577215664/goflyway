package proxy

import (
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/coyove/goflyway/pkg/logg"
)

type IOConfig struct {
	Bucket  *TokenBucket
	Chunked bool
	Partial bool
}

// func (gc *Cipher) blockedIO(target, source interface{}, key []byte, options *IOConfig) {
// 	var wg sync.WaitGroup

// 	wg.Add(2)
// 	go gc.ioCopyOrWarn(target.(io.Writer), source.(io.Reader), key, options, &wg)
// 	go gc.ioCopyOrWarn(source.(io.Writer), target.(io.Reader), key, options, &wg)
// 	wg.Wait()
// }

func (iot *io_t) Bridge(target, source net.Conn, key []byte, options IOConfig) {

	var targetTCP *net.TCPConn
	var targetOK bool

	switch target.(type) {
	case *net.TCPConn:
		targetTCP, targetOK = target.(*net.TCPConn)
	case *connWrapper:
		targetTCP, targetOK = target.(*connWrapper).Conn.(*net.TCPConn)
	}

	sourceTCP, sourceOK := source.(*net.TCPConn)

	if !targetOK || !sourceOK || sourceTCP == nil || targetTCP == nil {
		logg.E("casting failed: ", target, " ", source)
		return
	}

	// if targetOK && sourceOK {
	// copy from source, decrypt, to target
	go iot.copyClose(targetTCP, sourceTCP, key, options)

	// copy from target, encrypt, to source
	go iot.copyClose(sourceTCP, targetTCP, key, options)
	// } else {
	// 	logg.D(target, source)
	// 	go func() {
	// 		gc.blockedIO(target, source, key, options)
	// 		source.Close()
	// 		target.Close()
	// 	}()
	// }
}

// func (gc *Cipher) WrapIO(dst io.Writer, src io.Reader, key []byte, options *IOConfig) *IOCopyCipher {

// 	return &IOCopyCipher{
// 		Dst:     dst,
// 		Src:     src,
// 		Key:     key,
// 		Partial: gc.Partial,
// 		Cipher:  gc,
// 		Config:  options,
// 	}
// }

func (iot *io_t) copyClose(dst, src *net.TCPConn, key []byte, config IOConfig) {
	ts := time.Now()

	if _, err := iot.Copy(dst, src, key, config); err != nil {
		logg.E("io copy ", int(time.Now().Sub(ts).Seconds()), "s: ", err)
	}

	dst.CloseWrite()
	src.CloseRead()
}

// func (gc *Cipher) ioCopyOrWarn(dst io.Writer, src io.Reader, key []byte, options *IOConfig, wg *sync.WaitGroup) {
// 	ts := time.Now()

// 	if _, err := gc.WrapIO(dst, src, key, options).DoCopy(); err != nil {
// 		logg.E("io.copy ", int(time.Now().Sub(ts).Seconds()), "s: ", err)
// 	}

// 	wg.Done()
// }

// type IOCopyCipher struct {
// 	Dst    io.Writer
// 	Src    io.Reader
// 	Key    []byte
// 	Cipher *Cipher
// 	Config *IOConfig

// 	// can be altered outside
// 	Partial bool
// }

type iface struct {
	tab  unsafe.Pointer
	data unsafe.Pointer
}

type conn_state_t struct {
	conn net.Conn
	last int64
}

var readConn = struct {
	sync.Mutex
	m map[uintptr]*conn_state_t
}{
	m: make(map[uintptr]*conn_state_t),
}

var errPeacefullyEnd = errors.New("i'm cool")

func mark(src io.Reader) {
	srcConn, _ := src.(net.Conn)
	if srcConn == nil {
		return
	}

	readConn.Lock()
	id := uintptr((*iface)(unsafe.Pointer(&src)).data)
	if readConn.m[id] == nil {
		readConn.m[id] = &conn_state_t{srcConn, time.Now().UnixNano()}
	} else {
		readConn.m[id].last = time.Now().UnixNano()
	}
	readConn.Unlock()
}

func init() {
	go func() {
		for {
			readConn.Lock()
			ns := time.Now().UnixNano()
			for id, state := range readConn.m {

				if (ns - state.last) > int64(timeoutDial) {
					state.conn.Close()
					delete(readConn.m, id)
				}
			}

			readConn.Unlock()
			time.Sleep(time.Second)
		}
	}()
}

var iid uint64 = 0

func (iot *io_t) Copy(dst io.Writer, src io.Reader, key []byte, config IOConfig) (written int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			logg.E("wtf, ", r)
		}
	}()

	buf := make([]byte, 32*1024)
	ctr := (*Cipher)(unsafe.Pointer(iot)).getCipherStream(key)
	encrypted := 0

	u := atomic.AddUint64(&iid, 1)

	logg.D("open ", u)
	// retries := 0
	// srcConn, _ := src.(*net.TCPConn)
	// if srcConn != nil {

	// }

	for {
		logg.D("mark ", u)
		mark(src)
		nr, er := src.Read(buf)
		if nr > 0 {
			xbuf := buf[0:nr]

			if config.Partial && encrypted == sslRecordLen {
				// goto direct_transmission
			} else if ctr != nil {

				if encrypted+nr > sslRecordLen && config.Partial {
					ctr.XorBuffer(xbuf[:sslRecordLen-encrypted])
					encrypted = sslRecordLen
					// we are done, the traffic coming later will be transfered as is
				} else {
					ctr.XorBuffer(xbuf)
					encrypted += nr
				}

			}

			if config.Bucket != nil {
				config.Bucket.Consume(int64(len(xbuf)))
			}

			var nw int
			var ew error
			if config.Chunked {
				hlen := strconv.FormatInt(int64(nr), 16)
				if _, ew = dst.Write([]byte(hlen + "\r\n")); ew == nil {
					if nw, ew = dst.Write(xbuf); ew == nil {
						_, ew = dst.Write([]byte("\r\n"))
					}
				}

			} else {
				nw, ew = dst.Write(xbuf)
			}

			if nw > 0 {
				written += int64(nw)
			}

			if ew != nil {
				err = ew
				break
			}

			if nr != nw {
				err = io.ErrShortWrite
				break
			}

			// retries = 0
		}

		if er != nil {
			// if ne, _ := er.(net.Error); ne != nil && ne.Timeout() {
			// 	logg.D(u, " ", retries, " ", addr)
			// 	if retries++; retries < 10 {
			// 		continue
			// 	}
			// }

			if er != io.EOF && !strings.Contains(er.Error(), "use of closed network connection") {
				err = er
			}
			break
		}
	}

	if config.Chunked {
		dst.Write([]byte("0\r\n\r\n"))
	}

	logg.D("close ", u)
	return written, err
}

func (iot *io_t) NewReadCloser(src io.ReadCloser, key []byte) *IOReadCloserCipher {
	return &IOReadCloserCipher{
		src: src,
		key: key,
		ctr: (*Cipher)(unsafe.Pointer(iot)).getCipherStream(key),
	}
}

type IOReadCloserCipher struct {
	src io.ReadCloser
	key []byte
	ctr *inplace_ctr_t
}

func (rc *IOReadCloserCipher) Read(p []byte) (n int, err error) {
	n, err = rc.src.Read(p)
	if n > 0 && rc.ctr != nil {
		rc.ctr.XorBuffer(p[:n])
	}

	return
}

func (rc *IOReadCloserCipher) Close() error {
	return rc.src.Close()
}
