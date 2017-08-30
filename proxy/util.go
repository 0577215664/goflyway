package proxy

import (
	. "../config"
	"../logg"
	"../shoco"

	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	OK200   = []byte("HTTP/1.0 200 OK\r\n\r\n")
	tlsSkip = &tls.Config{InsecureSkipVerify: true}

	rkeyHeader  = "X-Request-ID"
	rkeyHeader2 = "X-Request-HTTP-ID"
	dnsHeaderID = "X-Host-Lookup-ID"
	dnsHeader   = "X-Host-Lookup"

	hostHeadHolder  = "%s.%s.com"
	dummyHosts      = append([]string{"baidu", "qq", "taobao", "sina", "163", "youku"}, strings.Split(*G_Dummies, "|")...)
	hostHeadExtract = regexp.MustCompile(`(\S+)\.(?:` + strings.Join(dummyHosts, "|") + `)\.com`)
	urlExtract      = regexp.MustCompile(`\?q=(\S+)$`)
	hasPort         = regexp.MustCompile(`:\d+$`)

	primes = []byte{
		11, 13, 17, 19, 23, 29, 31, 37, 41, 43,
		47, 53, 59, 61, 67, 71, 73, 79, 83, 89,
	}
)

func NewRand() *rand.Rand {
	k := int64(binary.BigEndian.Uint64(G_KeyBytes[:8]))
	return rand.New(rand.NewSource(time.Now().UnixNano() ^ k))
}

func RandomKey() string {
	_rand := NewRand()
	retB := make([]byte, 16)

	for i := 0; i < 16; i++ {
		retB[i] = byte(_rand.Intn(255) + 1)
	}

	return base64.StdEncoding.EncodeToString(Skip32Encode(G_KeyBytes, retB, false))
}

func ReverseRandomKey(key string) []byte {
	if key == "" {
		return nil
	}

	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil
	}

	return Skip32Decode(G_KeyBytes, k, false)
}

func processBody(req *http.Request, enc bool) {

	for _, c := range req.Cookies() {
		if enc {
			c.Value = Skip32EncodeString(G_KeyBytes, c.Value)
		} else {
			c.Value = Skip32DecodeString(G_KeyBytes, c.Value)
		}
	}

	var rkey string
	if enc {
		rkey = RandomKey()
		SafeAddHeader(req, rkeyHeader2, rkey)
	} else {
		rkey = SafeGetHeader(req, rkeyHeader2)
	}

	req.Body = ioutil.NopCloser((&IOReaderCipher{Src: req.Body, Key: ReverseRandomKey(rkey)}).Init())
}

func SafeAddHeader(req *http.Request, k, v string) {
	if orig := req.Header.Get(k); orig != "" {
		req.Header.Set(k, v+" "+orig)
	} else {
		req.Header.Add(k, v)
	}
}

func SafeGetHeader(req *http.Request, k string) string {
	v := req.Header.Get(k)
	if s := strings.Index(v, " "); s > 0 {
		req.Header.Set(k, v[s+1:])
		v = v[:s]
	}

	return v
}

func EncryptRequest(req *http.Request) string {
	req.Host = EncryptHost(req.Host)
	req.URL, _ = url.Parse("http://" + req.Host + "/?q=" + Skip32EncodeString(G_KeyBytes, req.URL.String()))

	rkey := RandomKey()
	SafeAddHeader(req, rkeyHeader, rkey)

	if !*G_UnsafeHttp {
		processBody(req, true)
	}

	return rkey
}

func DecryptRequest(req *http.Request) string {
	req.Host = DecryptHost(req.Host)
	if p := urlExtract.FindStringSubmatch(req.URL.String()); len(p) > 1 {
		req.URL, _ = url.Parse(Skip32DecodeString(G_KeyBytes, p[1]))
	}

	rkey := SafeGetHeader(req, rkeyHeader)

	if !*G_UnsafeHttp {
		processBody(req, false)
	}

	return rkey
}

func SplitHostPort(host string) (string, int) {

	if idx := strings.Index(host, ":"); idx > 0 {
		n, err := strconv.Atoi(host[idx+1:])
		if err != nil {
			logg.E("cannot split: ", host)
			return host, 0
		}

		return host[:idx], n
	} else {
		return host, 0
	}
}

func EncryptHost(host string) string {
	h, p := SplitHostPort(host)
	dummy := dummyHosts[NewRand().Intn(len(dummyHosts))]

	if *G_NoShoco {
		t := fmt.Sprintf(hostHeadHolder, Skip32EncodeString(G_KeyBytes, h), dummy)
		if p > 0 {
			t += ":" + strconv.Itoa(p)
		}
		return t
	}

	x := Base36Encode(Skip32Encode(G_KeyBytes, shoco.CompressHost(h), false))
	t := fmt.Sprintf(hostHeadHolder, x, dummy)
	if p > 0 {
		t += ":" + strconv.Itoa(p)
	}
	return t
}

func DecryptHost(host string) string {
	defer func() {
		if r := recover(); r != nil {
			logg.E("[SHOCO] - ", r)
		}
	}()

	h, p := SplitHostPort(host)

	if s := hostHeadExtract.FindStringSubmatch(h); len(s) > 1 {
		if *G_NoShoco {
			if p > 0 {
				return Skip32DecodeString(G_KeyBytes, s[1]) + ":" + strconv.Itoa(p)
			} else {
				return Skip32DecodeString(G_KeyBytes, s[1])
			}
		}

		if p > 0 {
			return shoco.DecompressHost(Skip32Decode(G_KeyBytes, Base36Decode(s[1]), false)) + ":" + strconv.Itoa(p)
		} else {
			return shoco.DecompressHost(Skip32Decode(G_KeyBytes, Base36Decode(s[1]), false))
		}
	}

	return ""
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func getAuth(r *http.Request) string {
	pa := r.Header.Get("Proxy-Authorization")
	if pa == "" {
		pa = r.Header.Get("X-Authorization")
	}

	return pa
}

func basicAuth(token string) bool {
	parts := strings.Split(token, " ")
	if len(parts) != 2 {
		return false
	}

	pa, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
	if err != nil {
		return false
	}

	return string(pa) == *G_Username+":"+*G_Password
}

func tryClose(b io.ReadCloser) {
	if err := b.Close(); err != nil {
		logg.W("can't close response body - ", err)
	}
}

func bytesStartWith(buf []byte, prefix []byte) bool {
	if len(prefix) > len(buf) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if prefix[i] != buf[i] {
			return false
		}
	}

	return true
}
