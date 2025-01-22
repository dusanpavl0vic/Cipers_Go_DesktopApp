package coders

const DELTA = 0x9E3779B9

func mx(sum, y, z, p, e uint32, key []uint32) uint32 {
	return ((z>>5 ^ y<<2) + (y>>3 ^ z<<4)) ^ ((sum ^ y) + (key[p&3^e] ^ z))
}

func adjustKey(key []uint32) []uint32 {
	if len(key) < 4 {
		newKey := make([]uint32, 4)
		copy(newKey, key)
		return newKey
	}
	return key
}

func Btea(v []uint32, n int, key []uint32) []uint32 {
	key = adjustKey(key)
	var y, z, sum uint32
	var p, rounds, e uint32

	if n > 1 { // Encryption
		rounds = 6 + 52/uint32(n)
		sum = 0
		z = v[uint32(n)-1]
		for rounds > 0 {
			sum += DELTA
			e = (sum >> 2) & 3
			for p = 0; p < uint32(n)-1; p++ {
				y = v[p+1]
				v[p] += mx(sum, y, z, p, e, key)
				z = v[p]
			}
			y = v[0]
			v[uint32(n)-1] += mx(sum, y, z, p, e, key)
			z = v[uint32(n)-1]
			rounds--
		}
	} else if n < -1 { // Decryption
		n = -n
		rounds = 6 + 52/uint32(n)
		sum = rounds * DELTA
		y = v[0]
		for rounds > 0 {
			e = (sum >> 2) & 3
			for p = uint32(n) - 1; p > 0; p-- {
				z = v[p-1]
				v[p] -= mx(sum, y, z, p, e, key)
				y = v[p]
			}
			z = v[uint32(n)-1]
			v[0] -= mx(sum, y, z, p, e, key)
			y = v[0]
			sum -= DELTA
			rounds--
		}
	}
	return v
}

func ToBytes(data []uint32, includeLength bool) []byte {
	length := uint32(len(data))
	n := length << 2
	if includeLength {
		last := data[length-1]
		n -= 4
		if last < n-3 || last > n {
			return nil
		}
		n = last
	}
	result := make([]byte, n)
	for i := uint32(0); i < n; i++ {
		result[i] = byte(data[i>>2] >> ((i & 3) << 3))
	}
	return result
}

func ToUint32(data []byte, includeLength bool) []uint32 {
	length := uint32(len(data))
	n := length >> 2
	if length&3 != 0 {
		n++
	}
	var result []uint32
	if includeLength {
		result = make([]uint32, n+1)
		result[n] = length
	} else {
		result = make([]uint32, n)
	}
	for i := uint32(0); i < length; i++ {
		result[i>>2] |= uint32(data[i]) << ((i & 3) << 3)
	}
	return result
}

// func main() {

// 	key2 := "cao ja sam kljuc"

// 	// Tekst za šifrovanje
// 	plainText2 := "cao moje ime je dusn kako ste danas. .dsds sd d sds"

// 	//
// 	//
// 	//
// 	//
// 	//
// 	//
// 	//
// 	//
// 	//
// 	//

// 	keyUint32 := ToUint32([]byte(key2), false)
// 	plainUint32 := ToUint32([]byte(plainText2), true)
// 	encryptedUint32 := Btea(plainUint32, len(plainUint32), keyUint32)
// 	encryptedText := ToBytes(encryptedUint32, false)
// 	fmt.Println("Kriptovan tekst:", string(encryptedText))

// 	// Konvertovanje u Base64
// 	encryptedBase64 := base64.StdEncoding.EncodeToString(encryptedText)
// 	fmt.Println("BASE64:", encryptedBase64)

// 	// Dešifrovanje
// 	encryptedUint32 = ToUint32(encryptedText, false)
// 	decryptedUint32 := Btea(encryptedUint32, -len(encryptedUint32), keyUint32)
// 	decryptedText := ToBytes(decryptedUint32, false)
// 	fmt.Println("Dekodirani tekst:", decryptedText)

// 	fmt.Println("Dekodirani tekst:", string(decryptedText))

// }
