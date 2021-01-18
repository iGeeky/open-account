package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/spaolacci/murmur3"
)

func MMHash(buf []byte) uint64 {
	return murmur3.Sum64WithSeed(buf, 0)
}

func Uint64ToStrID(id int64, m1, m2, m3 string) string {
	tmp := fmt.Sprintf("%s:%d:%s", m1, id, m2)
	// fmt.Printf("tmp: %s\n", tmp)
	tmpMMHash := MMHash([]byte(tmp))
	// fmt.Printf("tmpMMHash: %x\n", tmpMMHash)
	src := make([]byte, 8)
	binary.LittleEndian.PutUint64(src, uint64(tmpMMHash))
	// fmt.Printf("id: %d src bts: %x\n", id, src)
	// binary.BigEndian.PutUint64(src, uint64(tmpMMHash))
	sid := base64.RawURLEncoding.EncodeToString(src)
	h := sha1.New()
	sum := fmt.Sprintf("%x", h.Sum([]byte(sid+m3)))
	checkSum := sum[0:3]
	return fmt.Sprintf("%s%s", sid, checkSum)
}

func Uint64ToUID(id int64) (uid string) {
	m1 := "52f95f59c9fa"
	m2 := "aae4c7acf5"
	m3 := "022eeaf818"
	uid = Uint64ToStrID(id, m1, m2, m3)
	return
}
