package proxy

import (
	"fmt"
	"github.com/coyove/goflyway/pkg/rand"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
)

const (
	ivLen            = 16
	sslRecordLen     = 18 * 1024 // 18kb
	streamBufferSize = 512
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
	IO io_t

	Key     string
	Block   cipher.Block
	Rand    rand.Rand
	Partial bool
	Alias   string

	keyBuf []byte
}

type inplace_ctr_t struct {
	b       cipher.Block
	ctr     [ivLen]byte
	out     []byte
	outUsed int
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
				x.b.Encrypt(x.out[remain:], x.ctr[:])
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

		buf[i] ^= x.out[x.outUsed]
		x.outUsed++
	}
}

func xor(blk cipher.Block, iv [ivLen]byte, buf []byte) []byte {
	bsize := blk.BlockSize()
	x := make([]byte, len(buf)/bsize*bsize+bsize)

	for i := 0; i < len(x); i += bsize {
		blk.Encrypt(x[i:], iv[:])

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

func (gc *Cipher) getCipherStream(key *[16]byte) *inplace_ctr_t {
	if key == nil {
		return nil
	}

	return &inplace_ctr_t{
		b:       gc.Block,
		ctr:     *key,
		out:     make([]byte, 0, streamBufferSize),
		outUsed: 0,
	}
}

// Init inits the Cipher struct with key
func (gc *Cipher) Init(key string) (err error) {
	gc.Key = key
	gc.keyBuf = []byte(key)

	for len(gc.keyBuf) < 32 {
		gc.keyBuf = append(gc.keyBuf, gc.keyBuf...)
	}

	gc.Block, err = aes.NewCipher(gc.keyBuf[:32])

	alias := make([]byte, 16)
	gc.Block.Encrypt(alias, gc.keyBuf)
	gc.Alias = fmt.Sprintf("%X", alias[:3])

	return
}

func (gc *Cipher) Jibber() string {
	const (
		vowels = "aeioutr" // t, r are specials
		cons   = "bcdfghlmnprst"
	)

	s := gc.Rand.Intn(20)
	b := (vowels + cons)[s]

	var ret buffer
	ret.WriteByte(b)
	ln := gc.Rand.Intn(10) + 5

	if s < 7 {
		ret.WriteByte(cons[gc.Rand.Intn(13)])
	}

	for i := 0; i < ln; i += 2 {
		ret.WriteByte(vowels[gc.Rand.Intn(7)])
		ret.WriteByte(cons[gc.Rand.Intn(13)])
	}

	ret.Truncate(ln)
	return ret.String()
}

type clientRequest struct {
	Query  string  `json:"q,omitempty"`
	Auth   string  `json:"a,omitempty"`
	Opt    Options `json:"o"`
	Filler uint64  `json:"f"`

	iv [ivLen]byte
}

func (gc *Cipher) newRequest() *clientRequest {
	r := &clientRequest{}
	gc.Rand.Read(r.iv[:])
	// generate a random uint64 number, this will make encryptHost() outputs different lengths of data
	r.Filler = 2 << uint64(gc.Rand.Intn(21)*3+1)
	return r
}

func (gc *Cipher) genIV(init *[4]byte, out *[ivLen]byte) {
	mul := uint32(primes[init[0]]) * uint32(primes[init[1]]) * uint32(primes[init[2]]) * uint32(primes[init[3]])
	seed := binary.LittleEndian.Uint32(gc.keyBuf[:4])

	for i := 0; i < ivLen/4; i++ {
		seed = (mul * seed) % 0x7fffffff
		binary.LittleEndian.PutUint32(out[i*4:], seed)
	}
}

// Xor is an inplace xor method, and it just returns buf then
func (gc *Cipher) Xor(buf []byte, full *[ivLen]byte, quad *[4]byte) []byte {
	if full != nil {
		return xor(gc.Block, *full, buf)
	}

	var iv [ivLen]byte
	gc.genIV(quad, &iv)
	return xor(gc.Block, iv, buf)
}

// Encrypt encrypts a string
func (gc *Cipher) Encrypt(text string, iv *[ivLen]byte) string {
	return base64.URLEncoding.EncodeToString(gc.Xor([]byte(text), iv, nil))
}

// Decrypt decrypts a string
func (gc *Cipher) Decrypt(text string, iv *[ivLen]byte) string {
	buf, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return ""
	}
	return string(gc.Xor(buf, iv, nil))
}

func checksum1b(buf []byte) byte {
	s := int16(1)
	for _, b := range buf {
		s *= primes[b]
	}
	return byte(s>>12) + byte(s&0x00f0)
}
