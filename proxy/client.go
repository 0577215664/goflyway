package proxy

import (
	"github.com/coyove/goflyway/pkg/logg"
	"github.com/coyove/goflyway/pkg/lookup"
	"github.com/coyove/goflyway/pkg/lru"

	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type ClientConfig struct {
	Upstream        string
	DNSCache        *lru.Cache
	Dummies         *lru.Cache
	ProxyAllTraffic bool
	UseChinaList    bool
	DisableConsole  bool
	UserAuth        string

	UDPRelayPort   int
	UDPRelayNoHdr  bool
	UDPRelayCoconn int

	*GCipher
}

type ProxyClient struct {
	Tr       *http.Transport
	TrDirect *http.Transport

	*ClientConfig

	udp struct {
		relay    *net.UDPConn
		response []byte

		upstream struct {
			sync.Mutex
			conns map[string]net.Conn
		}
	}
}

var GClientProxy *ProxyClient

func (proxy *ProxyClient) DialUpstreamAndBridge(downstreamConn net.Conn, host, auth string, options int) {
	upstreamConn, err := net.Dial("tcp", proxy.Upstream)
	if err != nil {
		logg.E("[UPSTREAM] - ", err)
		return
	}

	rkey, rkeybuf := proxy.GCipher.RandomIV()

	if (options & DO_SOCKS5) != 0 {
		host = EncryptHost(proxy.GCipher, host, HOST_SOCKS_CONNECT)
	} else {
		host = EncryptHost(proxy.GCipher, host, HOST_HTTP_CONNECT)
	}

	payload := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\n", host)

	proxy.Dummies.Info(func(k lru.Key, v interface{}, h int64) {
		if v.(string) != "" && proxy.GCipher.Rand.Intn(5) > 1 {
			payload += k.(string) + ": " + v.(string) + "\r\n"
		}
	})

	payload += fmt.Sprintf("%s: %s\r\n", RKEY_HEADER, rkey)

	if auth != "" {
		payload += fmt.Sprintf("%s: %s\r\n", AUTH_HEADER, proxy.GCipher.EncryptString(auth))
	}

	upstreamConn.Write([]byte(payload + "\r\n"))
	proxy.GCipher.Bridge(downstreamConn, upstreamConn, rkeybuf, nil)
}

func (proxy *ProxyClient) DialHostAndBridge(downstreamConn net.Conn, host string, options int) {
	targetSiteConn, err := net.Dial("tcp", host)
	if err != nil {
		logg.E("[HOST] - ", err)
		return
	}

	if (options & DO_SOCKS5) != 0 {
		downstreamConn.Write(OK_SOCKS)
	} else {
		// response HTTP 200 OK to downstream, and it will not be xored in IOCopyCipher
		downstreamConn.Write(OK_HTTP)
	}

	proxy.GCipher.Bridge(downstreamConn, targetSiteConn, nil, nil)
}

func (proxy *ProxyClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/?goflyway-console" && !proxy.DisableConsole {
		handleWebConsole(w, r)
		return
	}

	var auth string
	if proxy.UserAuth != "" {
		if auth = proxy.basicAuth(getAuth(r)); auth == "" {
			w.Header().Set("Proxy-Authenticate", "Basic realm=goflyway")
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}
	}

	if r.Method == "CONNECT" {
		// dig tunnel
		hij, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		proxyClient, _, err := hij.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// we are inside GFW and should pass data to upstream
		host := r.URL.Host
		if !hasPort.MatchString(host) {
			host += ":80"
		}

		if proxy.CanDirectConnect(auth, host) {
			proxy.DialHostAndBridge(proxyClient, host, DO_NOTHING)
		} else {
			proxy.DialUpstreamAndBridge(proxyClient, host, auth, DO_NOTHING)
		}
	} else {
		// normal http requests
		var err error
		var rkeybuf []byte

		if !r.URL.IsAbs() {
			http.Error(w, "abspath only", http.StatusInternalServerError)
			return
		}

		direct := false
		rUrl := r.URL.String()
		// encrypt req to pass GFW
		if proxy.CanDirectConnect(auth, r.Host) {
			direct = true
		} else {
			rkeybuf = proxy.EncryptRequest(r)
		}

		r.Header.Del("Proxy-Authorization")
		r.Header.Del("Proxy-Connection")
		if auth != "" {
			SafeAddHeader(r, AUTH_HEADER, auth)
		}

		var resp *http.Response

		if direct {
			resp, err = proxy.TrDirect.RoundTrip(r)
		} else {
			resp, err = proxy.Tr.RoundTrip(r)
		}

		if err != nil {
			logg.E("[HTTP] - ", rUrl, " - ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode >= 400 {
			logg.D("[", resp.Status, "] - ", rUrl)
		}

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)

		iocc := proxy.GCipher.WrapIO(w, resp.Body, rkeybuf, nil)
		iocc.Partial = false

		nr, err := iocc.DoCopy()
		tryClose(resp.Body)

		if err != nil {
			logg.E("[COPYC] ", err, " - bytes: ", nr)
		}
	}
}

func (proxy *ProxyClient) CanDirectConnect(auth, host string) bool {
	host, _ = splitHostPort(host)

	isChineseIP := func(ip string) bool {
		if proxy.ProxyAllTraffic {
			return false
		}

		return lookup.IsChineseIP(ip)
	}

	if lookup.IsChineseWebsite(host) {
		return !proxy.ProxyAllTraffic
	}

	if ip, ok := proxy.DNSCache.Get(host); ok && ip.(string) != "" { // we have cached the host
		return lookup.IsPrivateIP(ip.(string)) || isChineseIP(ip.(string))
	}

	// lookup at local in case host points to a private ip
	ip, err := lookup.LookupIP(host)
	if err != nil {
		logg.E("[DNS] ", err)
	}

	if lookup.IsPrivateIP(ip) {
		proxy.DNSCache.Add(host, ip)
		return true
	}

	// if it is a foreign ip, just return false
	// but if it is a chinese ip, we withhold and query the upstream to double check
	maybeChinese := isChineseIP(ip)
	if !maybeChinese {
		proxy.DNSCache.Add(host, ip)
		return false
	}

	// lookup at upstream
	client := http.Client{Timeout: time.Second}
	req, _ := http.NewRequest("GET", "http://"+proxy.Upstream, nil)
	req.Header.Add(DNS_HEADER, EncryptHost(proxy.GCipher, host, HOST_DOMAIN_LOOKUP))
	if auth != "" {
		req.Header.Add(AUTH_HEADER, auth)
	}
	resp, err := client.Do(req)

	if err != nil {
		if !err.(net.Error).Timeout() {
			logg.W("[REMOTE LOOKUP] ", err)
		}
		return maybeChinese
	}

	ipbuf, err := ioutil.ReadAll(resp.Body)
	tryClose(resp.Body)

	if err != nil {
		logg.W("[REMOTE LOOKUP] ", err)
		return maybeChinese
	}

	proxy.DNSCache.Add(host, string(ipbuf))
	return isChineseIP(string(ipbuf))
}

func (proxy *ProxyClient) authConnection(conn net.Conn) (string, bool) {
	buf := make([]byte, 1+1+255+1+255)
	var n int
	var err error

	if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
		logg.E(CANNOT_READ_BUF, err)
		return "", false
	}

	if buf[0] != 0x01 {
		return "", false
	}

	username_len := int(buf[1])
	if 2+username_len+1 > n {
		return "", false
	}

	username := string(buf[2 : 2+username_len])
	password_len := int(buf[2+username_len])

	if 2+username_len+1+password_len > n {
		return "", false
	}

	password := string(buf[2+username_len+1 : 2+username_len+1+password_len])
	pu := username + ":" + password
	return pu, proxy.UserAuth == pu
}

func (proxy *ProxyClient) HandleSocks(conn net.Conn) {
	var err error
	log_close := func(args ...interface{}) {
		logg.E(args...)
		conn.Close()
	}

	buf := make([]byte, 2)
	if _, err = io.ReadAtLeast(conn, buf, 2); err != nil {
		log_close(CANNOT_READ_BUF, err)
		return
	}

	if buf[0] != socks5Version {
		log_close(NOT_SOCKS5)
		return
	}

	numMethods := int(buf[1])
	methods := make([]byte, numMethods)
	if _, err = io.ReadAtLeast(conn, methods, numMethods); err != nil {
		log_close(CANNOT_READ_BUF, err)
		return
	}

	var (
		auth string
		ok   bool
	)

	if proxy.UserAuth != "" {
		conn.Write([]byte{socks5Version, 0x02}) // username & password auth

		if auth, ok = proxy.authConnection(conn); !ok {
			conn.Write([]byte{0x01, 0x01})
			conn.Close()
			return
		} else {
			// auth success
			conn.Write([]byte{0x1, 0})
		}
	} else {
		conn.Write([]byte{socks5Version, 0})
	}
	// handshake over
	// tunneling start
	method, addr, ok := ParseDstFrom(conn, nil, false)
	if !ok {
		conn.Close()
		return
	}

	host := addr.String()

	if method == 0x01 {
		if proxy.CanDirectConnect(auth, host) {
			proxy.DialHostAndBridge(conn, host, DO_SOCKS5)
		} else {
			proxy.DialUpstreamAndBridge(conn, host, auth, DO_SOCKS5)
		}
	} else if method == 0x03 && proxy.udp.relay != nil {
		// UDP relay test
		conn.Write(proxy.udp.response)

		for {
			b := make([]byte, 2048)
			n, src, _ := proxy.udp.relay.ReadFrom(b)
			go proxy.HandleUDPtoTCP(b[:n], auth, src)
		}
	}
}

func StartClient(localaddr string, config *ClientConfig) {
	var mux net.Listener
	var err error

	upstreamUrl, err := url.Parse("http://" + config.Upstream)
	if err != nil {
		logg.F(err)
	}

	proxy := &ProxyClient{
		Tr: &http.Transport{
			TLSClientConfig: tlsSkip,
			Proxy:           http.ProxyURL(upstreamUrl),
		},
		TrDirect:     &http.Transport{TLSClientConfig: tlsSkip},
		ClientConfig: config,
	}

	if config.UDPRelayPort != 0 {
		r := []byte{socks5Version, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		binary.BigEndian.PutUint16(r[8:], uint16(config.UDPRelayPort))
		relay, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv6zero, Port: config.UDPRelayPort})
		if err != nil {
			logg.F("[UDP] - ", err)
		}

		if proxy.UDPRelayCoconn <= 0 {
			proxy.UDPRelayCoconn = 1
		}

		proxy.udp.response = r
		proxy.udp.relay = relay
	}

	GClientProxy = proxy

	if port, lerr := strconv.Atoi(localaddr); lerr == nil {
		mux, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv6zero, Port: port})
		localaddr = "localhost:" + localaddr
	} else {
		mux, err = net.Listen("tcp", localaddr)
	}

	if err != nil {
		logg.F(err)
	}

	logg.L("proxy is listening at ", localaddr, ", upstream is ", config.Upstream)
	logg.F(http.Serve(&listenerWrapper{Listener: mux, proxy: proxy, obpool: NewOneBytePool(1024)}, proxy))
}
