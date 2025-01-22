package coders

import (
	"crypto/rand"
	"log"
)

var (
	KeyCBC = []byte("cbckljuc")
	IV     []byte
)

func generateIV(blockSize int) {
	IV = make([]byte, blockSize)
	_, err := rand.Read(IV)
	if err != nil {
		log.Fatal("Failed to generate IV:", err)
	}
}

func xorBlocks(block1, block2 []byte) []byte {
	xored := make([]byte, len(block1))
	for i := 0; i < len(block1); i++ {
		xored[i] = block1[i] ^ block2[i]
	}
	return xored
}

func EncryptCBC(plainText []byte) []byte {
	blockSize := 8 // Block size for TEA algorithm (64 bits or 8 bytes)
	generateIV(blockSize)
	plainBlocks := splitIntoBlocks(plainText, blockSize)
	cipherText := make([]byte, 0, len(plainText)+blockSize)

	cipherText = append(cipherText, IV...)
	previousBlock := IV

	for _, block := range plainBlocks {
		xoredBlock := xorBlocks(block, previousBlock)
		encryptedBlock := Btea(ToUint32(xoredBlock, false), len(xoredBlock)/4, ToUint32(KeyCBC, false))
		encryptedBytes := ToBytes(encryptedBlock, false)
		cipherText = append(cipherText, encryptedBytes...)
		previousBlock = encryptedBytes
	}
	return cipherText
}

func DecryptCBC(cipherText []byte) []byte {
	blockSize := 8
	if len(cipherText) < blockSize {
		log.Fatal("CipherText too short")
	}
	IV = cipherText[:blockSize] // Postavlja IV iz šifrovanog teksta
	cipherBlocks := splitIntoBlocks(cipherText[blockSize:], blockSize)
	plainText := make([]byte, 0, len(cipherText)-blockSize)
	previousBlock := IV

	for _, block := range cipherBlocks {
		decryptedBlock := Btea(ToUint32(block, false), -len(block)/4, ToUint32(KeyCBC, false))
		decryptedBytes := ToBytes(decryptedBlock, false)
		xoredBlock := xorBlocks(decryptedBytes, previousBlock)
		plainText = append(plainText, xoredBlock...)
		previousBlock = block
	}
	return plainText
}

func splitIntoBlocks(data []byte, blockSize int) [][]byte {
	var blocks [][]byte
	for len(data) > 0 {
		if len(data) < blockSize {
			padding := make([]byte, blockSize-len(data))
			data = append(data, padding...)
		}
		blocks = append(blocks, data[:blockSize])
		data = data[blockSize:]
	}
	return blocks
}

// func main() {

// 	plainText := []byte("Cao ja sam dusan")

// 	cipherText := EncryptCBC(plainText)
// 	decryptedText := DecryptCBC(cipherText)

// 	fmt.Println("Original text:", string(plainText))

// 	fmt.Println("Encrypted text:", cipherText)

// 	fmt.Println("Decrypted text:", string(decryptedText))

// }
