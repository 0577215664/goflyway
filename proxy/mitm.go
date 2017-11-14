package proxy

import (
	"github.com/coyove/goflyway/pkg/logg"

	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type bufioConn struct {
	m io.Reader
	net.Conn
}

func (c *bufioConn) Read(buf []byte) (int, error) {
	return c.m.Read(buf)
}

func (proxy *ProxyClient) manInTheMiddle(client net.Conn, host string) {
	_host, _ := splitHostPort(host)
	// try self signing a cert of this host
	cert := sign(_host)
	if cert == nil {
		return
	}

	client.Write(okHTTP)

	go func() {

		tlsClient := tls.Server(client, &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{*cert},
		})

		if err := tlsClient.Handshake(); err != nil {
			logg.E("handshake failed: ", host, ", ", err)
			return
		}

		bufTLSClient := bufio.NewReader(tlsClient)

		for {
			var err error
			var rURL string
			var buf []byte
			if buf, err = bufTLSClient.Peek(3); err == io.EOF || len(buf) != 3 {
				break
			}

			// switch string(buf) {
			// case "GET", "POS", "HEA", "PUT", "OPT", "DEL", "PAT", "TRA":
			// 	// good
			// default:
			// 	proxy.dialUpstreamAndBridge(&bufioConn{Conn: tlsClient, m: bufTLSClient}, host, auth, []byte{})
			// 	return
			// }

			req, err := http.ReadRequest(bufTLSClient)
			if err != nil {
				logg.E("cannot read request: ", err)
				break
			}

			rURL = req.URL.String()
			req.Header.Del("Proxy-Authorization")
			req.Header.Del("Proxy-Connection")

			if !isHTTPSSchema.MatchString(req.URL.String()) {
				// we can ignore 443 since it's by default
				h := req.Host
				if strings.HasSuffix(h, ":443") {
					h = h[:len(h)-4]
				}

				req.URL, err = url.Parse("https://" + h + req.URL.String())
				rURL = req.URL.String()
			}

			logg.D(req.Method, "^ ", rURL)

			resp, rkeybuf, err := proxy.encryptAndTransport(req)
			if err != nil {
				logg.E("proxy pass: ", rURL, ", ", err)
				tlsClient.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n" + err.Error()))
				break
			}

			resp.Header.Del("Content-Length")
			resp.Header.Set("Transfer-Encoding", "chunked")

			if strings.ToLower(resp.Header.Get("Connection")) != "upgrade" {
				resp.Header.Set("Connection", "close")
				tlsClient.Write([]byte("HTTP/1.1 " + resp.Status + "\r\n"))
			} else {
				// we don't support upgrade in mitm
				tlsClient.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
				break
			}

			// buf, _ := httputil.DumpResponse(resp, true)
			_ = httputil.DumpResponse

			hdr := http.Header{}
			copyHeaders(hdr, resp.Header, proxy.Cipher, false, rkeybuf)
			if err := hdr.Write(tlsClient); err != nil {
				logg.W("write header: ", err)
				break
			}
			if _, err = io.WriteString(tlsClient, "\r\n"); err != nil {
				logg.W("write header: ", err)
				break
			}

			nr, err := proxy.Cipher.IO.Copy(tlsClient, resp.Body, rkeybuf, IOConfig{Partial: false, Chunked: true})
			if err != nil {
				logg.E("copy ", nr, "bytes: ", err)
			}
			tryClose(resp.Body)
		}

		tlsClient.Close()
	}()
}
