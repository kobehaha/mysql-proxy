package mysql

import (
    "crypto/sha1"
    "time"
    "math/rand"
)

func CheckPassword(scramble, password []byte) []byte {

    if len(password) == 0 {
        return nil
    }

    crypt := sha1.New()
    crypt.Write(password)
    stage1 := crypt.Sum(nil)

    crypt.Reset()
    crypt.Write(stage1)
    hash := crypt.Sum(nil)

    // outer Hash
    crypt.Reset()
    crypt.Write(scramble)
    crypt.Write(hash)
    scramble = crypt.Sum(nil)

    for i := range scramble {
        scramble[i] ^= stage1[i]
    }
    return scramble
}


func PutLengthEncodedInt(n uint64) []byte {
    switch {
    case n <= 250:
        return []byte{byte(n)}

    case n <= 0xffff:
        return []byte{0xfc, byte(n), byte(n >> 8)}

    case n <= 0xffffff:
        return []byte{0xfd, byte(n), byte(n >> 8), byte(n >> 16)}

    case n <= 0xffffffffffffffff:
        return []byte{0xfe, byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
            byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56)}
    }
    return nil
}


func RandomBuf(size int) []byte {
    buf := make([]byte, size)
    rand.Seed(time.Now().UTC().UnixNano())
    for i := 0; i < size; i++ {
        buf[i] = byte(rand.Intn(127))
        if buf[i] == 0 || buf[i] == byte('$') {
            buf[i]++
        }
    }
    return buf
}


