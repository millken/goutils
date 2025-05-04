package crypto

import (
	"encoding/binary"
	"errors"
	"reflect"
	"unsafe"
)

func oneWay(a, b, c, d uint64) (uint64, uint64, uint64, uint64) {
	t := a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	a = (a << 31) | (a >> 33)
	b = (b << 27) | (b >> 37)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	c = (c << 23) | (c >> 41)
	d = (d << 19) | (d >> 45)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	a = (a << 15) | (a >> 49)
	b = (b << 11) | (b >> 53)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	c = (c << 7) | (c >> 57)
	d = (d << 5) | (d >> 59)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	return a, b, c, d

}

func mix(a, b, c, d uint64) (uint64, uint64, uint64, uint64) {
	t := a + b + c + d
	return a ^ t, b ^ t, c ^ t, d ^ t
}

func arx16(a, b, c, d uint16) (uint16, uint16, uint16, uint16) {

	t := a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t
	a = (a << 7) | (a >> 9)
	b = (b << 3) | (b >> 13)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t
	c = (c << 7) | (c >> 9)
	d = (d << 3) | (d >> 13)
	t = a + b + c + d
	a ^= t
	b ^= t
	c ^= t
	d ^= t

	return a, b, c, d
}

func invalidOverlap(a, b []byte) bool {

	if len(a) == 0 || len(b) == 0 || &a[0] == &b[0] {
		return false
	}
	return reflect.ValueOf(&a[0]).Pointer() <= reflect.ValueOf(&b[len(b)-1]).Pointer() &&
		reflect.ValueOf(&b[0]).Pointer() <= reflect.ValueOf(&a[len(a)-1]).Pointer()
}

type blockKey struct {
	k1, k2, k3     uint64
	k4, k5, k6, k7 uint64
	k8, k9, ka, kb uint64
	kc, kd, ke, kf uint64

	counter uint64
}

const (
	KeySize   int = 32
	NonceSize int = 24
	BlockSize int = 128

	w0 uint64 = 0x7468654265737450
	w1 uint64 = 0x726f74656374696f
	w2 uint64 = 0x6e49734f6e6c7946
	w3 uint64 = 0x726f6d416c6c6168
)

func newBlockKey(key, nonce []byte) (*blockKey, error) {
	if key == nil || len(key) != KeySize {

		return nil, errors.New("nlarx1w: wrong key size")
	}
	if nonce == nil || len(nonce) != NonceSize {

		return nil, errors.New("nlarx1w: wrong nonce size")
	}
	k := &blockKey{}
	k.k5 = binary.LittleEndian.Uint64(key[:8])
	k.k6 = binary.LittleEndian.Uint64(key[8:16])
	k.k7 = binary.LittleEndian.Uint64(key[16:24])
	k.k8 = binary.LittleEndian.Uint64(key[24:32])

	k.kd = binary.LittleEndian.Uint64(nonce[:8])
	k.ke = binary.LittleEndian.Uint64(nonce[8:16])
	k.kf = binary.LittleEndian.Uint64(nonce[16:24])

	k.k1 = w0
	k.k2 = w1
	k.k3 = w2
	k.k4 = w3

	//precompute

	//expand key
	k.k9, k.ka, k.kb, k.kc = oneWay(k.k4, k.k5, k.k6, k.k7)

	buf := make([]byte, BlockSize)
	k.nextKeyStream(0, buf)
	k.counter = (binary.LittleEndian.Uint64(buf[:8]) >> 16)
	k.k1 = binary.LittleEndian.Uint64(buf[8:16])
	k.k2 = binary.LittleEndian.Uint64(buf[16:24])
	k.k3 = binary.LittleEndian.Uint64(buf[24:32])
	k.k4 = binary.LittleEndian.Uint64(buf[32:40])
	k.k5 = binary.LittleEndian.Uint64(buf[40:48])
	k.k6 = binary.LittleEndian.Uint64(buf[48:56])
	k.k7 = binary.LittleEndian.Uint64(buf[56:64])
	k.k8 = binary.LittleEndian.Uint64(buf[64:72])
	k.k9 = binary.LittleEndian.Uint64(buf[72:80])
	k.ka = binary.LittleEndian.Uint64(buf[80:88])
	k.kb = binary.LittleEndian.Uint64(buf[88:96])
	k.kc = binary.LittleEndian.Uint64(buf[96:104])
	k.kd = binary.LittleEndian.Uint64(buf[104:112])
	k.ke = binary.LittleEndian.Uint64(buf[112:120])
	k.kf = binary.LittleEndian.Uint64(buf[120:128])

	return k, nil
}

func (k *blockKey) nextKeyStream(counter uint64, dst []byte) error {
	if len(dst) < BlockSize {
		return errors.New("nlarx1w: wrong dst size")
	}
	ds := (*[16]uint64)(unsafe.Pointer(&dst[0]))[:]

	k1, k2, k3 := k.k1, k.k2, k.k3
	k4, k5, k6, k7 := k.k4, k.k5, k.k6, k.k7
	k8, k9, ka, kb := k.k8, k.k9, k.ka, k.kb
	kc, kd, ke, kf := k.kc, k.kd, k.ke, k.kf

	//first round
	counter, k3, k4, k7 = oneWay(counter, k3, k4, k7)

	//first column and last column round
	counter, k4, k8, kc = mix(counter, k4, k8, kc)
	k3, k7, kb, kf = mix(k3, k7, kb, kf)

	//diagonal round
	counter, k5, ka, kf = mix(counter, k5, ka, kf)
	k3, k6, k9, kc = mix(k3, k6, k9, kc)

	//ring round
	k1, k7, ke, k8 = mix(k1, k7, ke, k8)
	k2, kb, kd, k4 = mix(k2, kb, kd, k4)

	counter = (counter << 31) | (counter >> 33)
	k3 = (k3 << 27) | (k3 >> 37)
	kc = (kc << 23) | (kc >> 41)
	kf = (kf << 19) | (kf >> 45)
	k5 = (k5 << 15) | (k5 >> 49)
	k6 = (k6 << 11) | (k6 >> 53)
	k9 = (k9 << 7) | (k9 >> 57)
	ka = (ka << 5) | (ka >> 59)

	k1 = (k1 << 31) | (k1 >> 33)
	k2 = (k2 << 27) | (k2 >> 37)
	k7 = (k7 << 23) | (k7 >> 41)
	kb = (kb << 19) | (kb >> 45)
	ke = (ke << 15) | (ke >> 49)
	kd = (kd << 11) | (kd >> 53)
	k8 = (k8 << 7) | (k8 >> 57)
	k4 = (k4 << 5) | (k4 >> 59)

	//diagonal round 2
	k1, k6, kb, kc = mix(k1, k6, kb, kc)
	k2, k5, k8, kf = mix(k2, k5, k8, kf)
	kd, ka, k7, counter = mix(kd, ka, k7, counter)
	ke, k9, k4, k3 = mix(ke, k9, k4, k3)

	counter = (counter << 31) | (counter >> 33)
	k3 = (k3 << 27) | (k3 >> 37)
	kc = (kc << 23) | (kc >> 41)
	kf = (kf << 19) | (kf >> 45)
	k5 = (k5 << 15) | (k5 >> 49)
	k6 = (k6 << 11) | (k6 >> 53)
	k9 = (k9 << 7) | (k9 >> 57)
	ka = (ka << 5) | (ka >> 59)

	k1 = (k1 << 31) | (k1 >> 33)
	k2 = (k2 << 27) | (k2 >> 37)
	k7 = (k7 << 23) | (k7 >> 41)
	kb = (kb << 19) | (kb >> 45)
	ke = (ke << 15) | (ke >> 49)
	kd = (kd << 11) | (kd >> 53)
	k8 = (k8 << 7) | (k8 >> 57)
	k4 = (k4 << 5) | (k4 >> 59)

	//column round
	counter, k4, k8, kc = mix(counter, k4, k8, kc)
	k1, k5, k9, kd = mix(k1, k5, k9, kd)
	k2, k6, ka, ke = mix(k2, k6, ka, ke)
	k3, k7, kb, kf = mix(k3, k7, kb, kf)

	counter = (counter << 31) | (counter >> 33)
	k3 = (k3 << 27) | (k3 >> 37)
	kc = (kc << 23) | (kc >> 41)
	kf = (kf << 19) | (kf >> 45)
	k5 = (k5 << 15) | (k5 >> 49)
	k6 = (k6 << 11) | (k6 >> 53)
	k9 = (k9 << 7) | (k9 >> 57)
	ka = (ka << 5) | (ka >> 59)

	k1 = (k1 << 31) | (k1 >> 33)
	k2 = (k2 << 27) | (k2 >> 37)
	k7 = (k7 << 23) | (k7 >> 41)
	kb = (kb << 19) | (kb >> 45)
	ke = (ke << 15) | (ke >> 49)
	kd = (kd << 11) | (kd >> 53)
	k8 = (k8 << 7) | (k8 >> 57)
	k4 = (k4 << 5) | (k4 >> 59)

	//diagonal round
	counter, k5, ka, kf = mix(counter, k5, ka, kf)
	k3, k6, k9, kc = mix(k3, k6, k9, kc)

	//ring round
	k1, k7, ke, k8 = mix(k1, k7, ke, k8)
	k2, kb, kd, k4 = mix(k2, kb, kd, k4)

	counter = (counter << 31) | (counter >> 33)
	k3 = (k3 << 27) | (k3 >> 37)
	kc = (kc << 23) | (kc >> 41)
	kf = (kf << 19) | (kf >> 45)
	k5 = (k5 << 15) | (k5 >> 49)
	k6 = (k6 << 11) | (k6 >> 53)
	k9 = (k9 << 7) | (k9 >> 57)
	ka = (ka << 5) | (ka >> 59)

	k1 = (k1 << 31) | (k1 >> 33)
	k2 = (k2 << 27) | (k2 >> 37)
	k7 = (k7 << 23) | (k7 >> 41)
	kb = (kb << 19) | (kb >> 45)
	ke = (ke << 15) | (ke >> 49)
	kd = (kd << 11) | (kd >> 53)
	k8 = (k8 << 7) | (k8 >> 57)
	k4 = (k4 << 5) | (k4 >> 59)

	//diagonal round 2
	k1, k6, kb, kc = mix(k1, k6, kb, kc)
	k2, k5, k8, kf = mix(k2, k5, k8, kf)
	kd, ka, k7, counter = mix(kd, ka, k7, counter)
	ke, k9, k4, k3 = mix(ke, k9, k4, k3)

	//column round
	counter, k4, k8, kc = mix(counter, k4, k8, kc)
	k1, k5, k9, kd = mix(k1, k5, k9, kd)
	k2, k6, ka, ke = mix(k2, k6, ka, ke)
	k3, k7, kb, kf = mix(k3, k7, kb, kf)

	ds[0] = counter
	ds[1] = k1
	ds[2] = k2
	ds[3] = k3
	ds[4] = k4
	ds[5] = k5
	ds[6] = k6
	ds[7] = k7
	ds[8] = k8
	ds[9] = k9
	ds[10] = ka
	ds[11] = kb
	ds[12] = kc
	ds[13] = kd
	ds[14] = ke
	ds[15] = kf
	return nil

}

type Cipher struct {
	k           *blockKey
	counter     uint64
	lastKSIndex int
	ks          []byte
	nli         uint64
}

func NewCipher(key, nonce []byte) (*Cipher, error) {

	bk, err := newBlockKey(key, nonce)
	if err != nil {
		return nil, err
	}
	c := &Cipher{}
	c.k = bk
	c.counter = bk.counter
	bk.counter = 0
	c.nli = c.counter >> 4
	c.lastKSIndex = 0
	c.ks = make([]byte, BlockSize)

	return c, nil
}

func (c *Cipher) genKs() {
	c.nli += 1
	a := uint16(c.nli)
	b := uint16(c.nli >> 16)
	cc := uint16(c.nli >> 32)
	d := uint16(c.nli >> 48)
	a, _, _, _ = arx16(a, b, cc, d)
	c.counter += uint64(((a >> 1) | 1))
	c.k.nextKeyStream(c.counter, c.ks)

}

func (c *Cipher) xorKeyStreamBig(dst, src []byte) {
	ds := (*[128]uint64)(unsafe.Pointer(&dst[0]))[:]
	sc := (*[128]uint64)(unsafe.Pointer(&src[0]))[:]
	ks := (*[16]uint64)(unsafe.Pointer(&(c.ks[0])))[:]
	var sds []uint64
	var ssc []uint64
	for i := 0; i < 128; i += 16 {
		sds = ds[i:]
		ssc = sc[i:]
		c.genKs()

		sds[0] = ssc[0] ^ ks[0]
		sds[1] = ssc[1] ^ ks[1]
		sds[2] = ssc[2] ^ ks[2]
		sds[3] = ssc[3] ^ ks[3]

		sds[4] = ssc[4] ^ ks[4]
		sds[5] = ssc[5] ^ ks[5]
		sds[6] = ssc[6] ^ ks[6]
		sds[7] = ssc[7] ^ ks[7]

		sds[8] = ssc[8] ^ ks[8]
		sds[9] = ssc[9] ^ ks[9]
		sds[10] = ssc[10] ^ ks[10]
		sds[11] = ssc[11] ^ ks[11]

		sds[12] = ssc[12] ^ ks[12]
		sds[13] = ssc[13] ^ ks[13]
		sds[14] = ssc[14] ^ ks[14]
		sds[15] = ssc[15] ^ ks[15]

	}

}

func (c *Cipher) XORKeyStream(dst, src []byte) {

	if len(dst) < len(src) {
		panic("nlarx1w: output smaller than input")
	}
	if invalidOverlap(dst, src) {
		panic("nlarx1w: invalid buffer overlap")
	}

	if c.lastKSIndex == 0 {
		srclen := len(src)
		bsx8 := BlockSize * 8
		if srclen >= bsx8 {
			i := 0
			for ; i < srclen; i += bsx8 {
				c.xorKeyStreamBig(dst[i:], src[i:])
			}
			if i == srclen {
				return
			}
			i -= bsx8
			src = src[i:]
			dst = dst[i:]

		}

		c.genKs()
	}
	j := c.lastKSIndex
	i := 0
	var ks []uint64
	var sc []uint64
	var ds []uint64
	for i < len(src) {
		if j >= BlockSize {
			j = 0
			c.genKs()
		}
		l := 0
		if len(src[i:]) > len(c.ks[j:]) {
			l = len(c.ks[j:])
		} else {
			l = len(src[i:])
		}
		if l%8 > 0 && l > 8 {
			l = l - (l % 8)
		}

		if l < 8 {
			for n := 0; n < l; n++ {
				dst[i+n] = src[i+n] ^ c.ks[j+n]
			}
		} else {
			switch l / 8 {

			case 1:
				ks = (*[1]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[1]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[1]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				break
			case 2:
				ks = (*[2]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[2]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[2]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				break
			case 3:
				ks = (*[3]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[3]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[3]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				break
			case 4:
				ks = (*[4]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[4]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[4]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]
				break
			case 5:
				ks = (*[5]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[5]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[5]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]
				ds[4] = sc[4] ^ ks[4]
				break
			case 6:
				ks = (*[6]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[6]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[6]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]
				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				break
			case 7:
				ks = (*[7]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[7]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[7]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				break
			case 8:
				ks = (*[8]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[8]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[8]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]
				break
			case 9:
				ks = (*[9]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[9]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[9]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				break
			case 10:
				ks = (*[10]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[10]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[10]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				break
			case 11:
				ks = (*[11]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[11]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[11]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				break
			case 12:
				ks = (*[12]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[12]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[12]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				ds[11] = sc[11] ^ ks[11]
				break
			case 13:
				ks = (*[13]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[13]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[13]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				ds[11] = sc[11] ^ ks[11]

				ds[12] = sc[12] ^ ks[12]
				break
			case 14:
				ks = (*[14]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[14]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[14]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				ds[11] = sc[11] ^ ks[11]

				ds[12] = sc[12] ^ ks[12]
				ds[13] = sc[13] ^ ks[13]
				break
			case 15:
				ks = (*[15]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[15]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[15]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				ds[11] = sc[11] ^ ks[11]

				ds[12] = sc[12] ^ ks[12]
				ds[13] = sc[13] ^ ks[13]
				ds[14] = sc[14] ^ ks[14]
				break
			default:
				ks = (*[16]uint64)(unsafe.Pointer(&(c.ks[j])))[:]
				sc = (*[16]uint64)(unsafe.Pointer(&src[i]))[:]
				ds = (*[16]uint64)(unsafe.Pointer(&dst[i]))[:]
				ds[0] = sc[0] ^ ks[0]
				ds[1] = sc[1] ^ ks[1]
				ds[2] = sc[2] ^ ks[2]
				ds[3] = sc[3] ^ ks[3]

				ds[4] = sc[4] ^ ks[4]
				ds[5] = sc[5] ^ ks[5]
				ds[6] = sc[6] ^ ks[6]
				ds[7] = sc[7] ^ ks[7]

				ds[8] = sc[8] ^ ks[8]
				ds[9] = sc[9] ^ ks[9]
				ds[10] = sc[10] ^ ks[10]
				ds[11] = sc[11] ^ ks[11]

				ds[12] = sc[12] ^ ks[12]
				ds[13] = sc[13] ^ ks[13]
				ds[14] = sc[14] ^ ks[14]
				ds[15] = sc[15] ^ ks[15]

			}

		}
		i += l
		j += l

	}
	if j >= BlockSize {
		j = 0
	}
	c.lastKSIndex = j
}
