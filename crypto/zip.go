package crypto

import (
	"hash/crc32"
	"io"
)

type ZipCrypto struct {
	password []byte
	Keys     [3]uint32
}

func NewZipCrypto(passphrase []byte) *ZipCrypto {
	z := &ZipCrypto{}
	z.password = passphrase
	z.Reset()
	return z
}

func (z *ZipCrypto) Reset() {
	z.Keys[0] = 0x12345678
	z.Keys[1] = 0x23456789
	z.Keys[2] = 0x34567890

	for i := 0; i < len(z.password); i++ {
		z.updateKeys(z.password[i])
	}
}

func (z *ZipCrypto) updateKeys(byteValue byte) {
	z.Keys[0] = crc32update(z.Keys[0], byteValue)
	z.Keys[1] += z.Keys[0] & 0xff
	z.Keys[1] = z.Keys[1]*134775813 + 1
	z.Keys[2] = crc32update(z.Keys[2], (byte)(z.Keys[1]>>24))
}

func (z *ZipCrypto) magicByte() byte {
	var t uint32 = z.Keys[2] | 2
	return byte((t * (t ^ 1)) >> 8)
}

func (z *ZipCrypto) Encrypt(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	chiper := make([]byte, length)
	for i := 0; i < length; i++ {
		v := data[i]
		chiper[i] = v ^ z.magicByte()
		z.updateKeys(v)
	}
	return chiper
}

func (z *ZipCrypto) Decrypt(chiper []byte) []byte {
	length := len(chiper)
	if length == 0 {
		return nil
	}
	plain := make([]byte, length)
	for i, c := range chiper {
		v := c ^ z.magicByte()
		z.updateKeys(v)
		plain[i] = v
	}
	return plain
}

type ZipCryptoReader struct {
	r  io.Reader
	zc *ZipCrypto
}

func NewZipCryptoReader(r io.Reader, zc *ZipCrypto) *ZipCryptoReader {
	return &ZipCryptoReader{r: r, zc: zc}
}

func (zr *ZipCryptoReader) Read(p []byte) (int, error) {
	n, err := zr.r.Read(p)
	if n > 0 {
		for i := 0; i < n; i++ {
			c := p[i]
			plain := c ^ zr.zc.magicByte()
			zr.zc.updateKeys(plain)
			p[i] = plain
		}
	}
	return n, err
}

func (zr *ZipCryptoReader) WriteTo(w io.Writer) (int64, error) {
	var total int64
	buf := make([]byte, 32*1024)
	for {
		n, err := zr.Read(buf)
		if n > 0 {
			nw, ew := w.Write(buf[:n])
			total += int64(nw)
			if ew != nil {
				return total, ew
			}
			if nw != n {
				return total, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return total, err
		}
	}
	return total, nil
}

type ZipCryptoWriter struct {
	w  io.Writer
	zc *ZipCrypto
}

func NewZipCryptoWriter(w io.Writer, zc *ZipCrypto) *ZipCryptoWriter {
	return &ZipCryptoWriter{w: w, zc: zc}
}

func (zw *ZipCryptoWriter) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		b := p[i]
		buf[i] = b ^ zw.zc.magicByte()
		zw.zc.updateKeys(b) // 用明文推进 key
	}
	return zw.w.Write(buf)
}

func (zw *ZipCryptoWriter) ReadFrom(r io.Reader) (int64, error) {
	var total int64
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			nw, ew := zw.Write(buf[:n])
			total += int64(nw)
			if ew != nil {
				return total, ew
			}
			if nw != n {
				return total, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return total, err
		}
	}
	return total, nil
}

func crc32update(pCrc32 uint32, bval byte) uint32 {
	return crc32.IEEETable[(pCrc32^uint32(bval))&0xff] ^ (pCrc32 >> 8)
}
