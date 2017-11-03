package proxy

import (
	"github.com/coyove/goflyway/pkg/bitsop"
	"github.com/coyove/goflyway/pkg/counter"
	"github.com/coyove/goflyway/pkg/logg"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"math/rand"
)

const (
	IV_LENGTH          = 16
	SSL_RECORD_MAX     = 18 * 1024 // 18kb
	STREAM_BUFFER_SIZE = 512
)

var primes = []int16{
	11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
	73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151,
	157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229, 233, 239,
	241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317, 331, 337,
	347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409, 419, 421, 431, 433,
	439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499, 503, 509, 521, 523, 541,
	547, 557, 563, 569, 571, 577, 587, 593, 599, 601, 607, 613, 617, 619, 631, 641,
	643, 647, 653, 659, 661, 673, 677, 683, 691, 701, 709, 719, 727, 733, 739, 743,
	751, 757, 761, 769, 773, 787, 797, 809, 811, 821, 823, 827, 829, 839, 853, 857,
	859, 863, 877, 881, 883, 887, 907, 911, 919, 929, 937, 941, 947, 953, 967, 971,
	977, 983, 991, 997, 1009, 1013, 1019, 1021, 1031, 1033, 1039, 1049, 1051, 1061, 1063, 1069,
	1087, 1091, 1093, 1097, 1103, 1109, 1117, 1123, 1129, 1151, 1153, 1163, 1171, 1181, 1187, 1193,
	1201, 1213, 1217, 1223, 1229, 1231, 1237, 1249, 1259, 1277, 1279, 1283, 1289, 1291, 1297, 1301,
	1303, 1307, 1319, 1321, 1327, 1361, 1367, 1373, 1381, 1399, 1409, 1423, 1427, 1429, 1433, 1439,
	1447, 1451, 1453, 1459, 1471, 1481, 1483, 1487, 1489, 1493, 1499, 1511, 1523, 1531, 1543, 1549,
	1553, 1559, 1567, 1571, 1579, 1583, 1597, 1601, 1607, 1609, 1613, 1619, 1621, 1627, 1637, 1657,
}

type Cipher struct {
	Key       []byte
	KeyString string
	Block     cipher.Block
	Partial   bool
	Rand      *rand.Rand
}

type inplace_ctr_t struct {
	b       cipher.Block
	ctr     []byte
	out     []byte
	outUsed int
	ptr     int
}

// From src/crypto/cipher/ctr.go
func (x *inplace_ctr_t) XorBuffer(buf []byte) {
	for i := 0; i < len(buf); i++ {
		if x.outUsed >= len(x.out)-x.b.BlockSize() {
			// refill
			remain := len(x.out) - x.outUsed
			copy(x.out, x.out[x.outUsed:])
			x.out = x.out[:cap(x.out)]
			bs := x.b.BlockSize()
			for remain <= len(x.out)-bs {
				x.b.Encrypt(x.out[remain:], x.ctr)
				remain += bs

				// Increment counter
				for i := len(x.ctr) - 1; i >= 0; i-- {
					x.ctr[i]++
					if x.ctr[i] != 0 {
						break
					}
				}
			}
			x.out = x.out[:remain]
			x.outUsed = 0
		}

		if x.ptr == 0 && len(buf) > 10 {
			if buf[0] == 0x1f && buf[1] == 0x8b {
				// ignore the first 10 bytes of gzip
				x.ptr = 10
				x.XorBuffer(buf[10:])
				return
			}
		}

		buf[i] ^= x.out[x.outUsed]
		x.outUsed++
		x.ptr++
	}
}

func dup(in []byte) (out []byte) {
	out = make([]byte, len(in))
	copy(out, in)
	return
}

func xor(blk cipher.Block, iv, buf []byte) []byte {
	iv = dup(iv)
	bsize := blk.BlockSize()
	x := make([]byte, len(buf)/bsize*bsize+bsize)

	for i := 0; i < len(x); i += bsize {
		blk.Encrypt(x[i:], iv)

		for i := len(iv) - 1; i >= 0; i-- {
			if iv[i]++; iv[i] != 0 {
				break
			}
		}
	}

	for i := 0; i < len(buf); i++ {
		buf[i] ^= x[i]
	}

	return buf
}

func (gc *Cipher) getCipherStream(key []byte) *inplace_ctr_t {
	if key == nil {
		return nil
	}

	if len(key) != IV_LENGTH {
		logg.E("iv is not 128bit long: ", key)
		return nil
	}

	return &inplace_ctr_t{
		b: gc.Block,
		// key must be duplicated because it gets modified during XorBuffer
		ctr:     dup(key),
		out:     make([]byte, 0, STREAM_BUFFER_SIZE),
		outUsed: 0,
	}
}

func (gc *Cipher) New() (err error) {
	gc.Key = []byte(gc.KeyString)

	for len(gc.Key) < 32 {
		gc.Key = append(gc.Key, gc.Key...)
	}

	gc.Block, err = aes.NewCipher(gc.Key[:32])
	gc.Rand = gc.NewRand()

	return
}

func (gc *Cipher) genIV(ss ...byte) []byte {
	if len(ss) == IV_LENGTH {
		return ss
	}

	ret := make([]byte, IV_LENGTH)

	var mul uint32 = 1
	for _, s := range ss {
		mul *= uint32(primes[s])
	}

	var seed uint32 = binary.LittleEndian.Uint32(gc.Key[:4])

	for i := 0; i < IV_LENGTH/4; i++ {
		seed = (mul * seed) % 0x7fffffff
		binary.LittleEndian.PutUint32(ret[i*4:], seed)
	}

	return ret
}

func (gc *Cipher) Encrypt(buf []byte, ss ...byte) []byte {
	r := gc.NewRand()

	if ss == nil || len(ss) < 2 {
		b, b2 := byte(r.Intn(256)), byte(r.Intn(256))
		return append(xor(gc.Block, gc.genIV(b, b2), buf), b, b2)
	}

	return xor(gc.Block, gc.genIV(ss...), buf)
}

func (gc *Cipher) Decrypt(buf []byte, ss ...byte) []byte {
	if buf == nil || len(buf) < 2 {
		return []byte{}
	}

	if ss == nil || len(ss) < 2 {
		b, b2 := byte(buf[len(buf)-2]), byte(buf[len(buf)-1])
		return xor(gc.Block, gc.genIV(b, b2), buf[:len(buf)-2])
	}

	return xor(gc.Block, gc.genIV(ss...), buf)
}

func (gc *Cipher) EncryptString(text string, rkey ...byte) string {
	return Base32Encode(gc.Encrypt([]byte(text), rkey...), true)
}

func (gc *Cipher) DecryptString(text string, rkey ...byte) string {
	buf, _ := Base32Decode(text, true)
	return string(gc.Decrypt(buf, rkey...))
}

func (gc *Cipher) EncryptCompress(str string, rkey ...byte) string {
	return Base32Encode(gc.Encrypt(bitsop.Compress(str), rkey...), false)
}

func (gc *Cipher) DecryptDecompress(str string, rkey ...byte) string {
	buf, _ := Base32Decode(str, false)
	return bitsop.Decompress(gc.Decrypt(buf, rkey...))
}

func (gc *Cipher) NewRand() *rand.Rand {
	var k int64 = int64(binary.BigEndian.Uint64(gc.Key[:8]))
	var k2 int64 = counter.GetCounter()

	return rand.New(rand.NewSource(k2 ^ k))
}

func checksum1b(buf []byte) byte {
	s := int16(1)
	for _, b := range buf {
		s *= primes[b]
	}
	return byte(s>>12) + byte(s&0x00f0)
}

func (gc *Cipher) NewIV(options byte, payload []byte, auth string) (string, []byte) {
	_rand := gc.NewRand()

	// +------------+-------------+-----------+-- -  -   -
	// | options 1b | checksum 1b | iv 128bit | auth data ...
	// +------------+-------------+-----------+-- -  -   -

	var retB, ret []byte
	retB = make([]byte, 1+1+IV_LENGTH+len(auth))

	if payload == nil {
		ret = make([]byte, IV_LENGTH)
		for i := 2; i < IV_LENGTH+2; i++ {
			retB[i] = byte(_rand.Intn(255) + 1)
			ret[i-2] = retB[i]
		}
	} else {
		ret = payload
		copy(retB[2:], payload)
	}

	if auth != "" {
		copy(retB[2+IV_LENGTH:], []byte(auth))
	}

	retB[0], retB[1] = options, checksum1b(retB[2:])
	s1, s2, s3, s4 := byte(_rand.Intn(256)), byte(_rand.Intn(256)), byte(_rand.Intn(256)), byte(_rand.Intn(256))

	return base64.StdEncoding.EncodeToString(
		append(
			xor(
				gc.Block, gc.genIV(s1, s2, s3, s4), retB,
			), s1, s2, s3, s4,
		),
	), ret
}

func (gc *Cipher) ReverseIV(key string) (options byte, iv []byte, auth []byte) {
	options = 0xff
	if key == "" {
		return
	}

	buf, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(buf) < 5 {
		return
	}

	b, b2, b3, b4 := buf[len(buf)-4], buf[len(buf)-3], buf[len(buf)-2], buf[len(buf)-1]
	buf = xor(gc.Block, gc.genIV(b, b2, b3, b4), buf[:len(buf)-4])

	if len(buf) < IV_LENGTH+2 {
		return
	}

	if buf[1] != checksum1b(buf[2:]) {
		return
	}

	return buf[0], buf[2 : 2+IV_LENGTH], buf[2+IV_LENGTH:]
}
