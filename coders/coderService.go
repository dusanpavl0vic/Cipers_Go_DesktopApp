package coders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CipherType int

const (
	RailFence CipherType = iota
	XXTEA
	XXTEA_CBC
)

func (c CipherType) String() string {
	switch c {
	case RailFence:
		return "RailFence"
	case XXTEA:
		return "XXTEA"
	case XXTEA_CBC:
		return "XXTEA_CBC"
	default:
		return "UnknownCipher"
	}
}

var (
	Depth  int        = 3
	Cipher CipherType = 0
	Key    string     = "cao ja sam kljuc"
)

func EncodeFile(fileData []byte, filename string) error {

	var encrypted []byte

	switch Cipher {
	case RailFence:
		fmt.Println("Cipher is Railfence")
		encrypted = EncryptRailFence(fileData, Depth)
		//fmt.Println("Kriptovan", encrypted)
	case XXTEA:
		fmt.Println("Cipher is XXTEA")
		keyb := ToUint32([]byte(Key), false)

		//fmt.Println("Procitan tekst", string(fileData))

		data := ToUint32(fileData, true)
		encodeText := Btea(data, len(data), keyb)
		encrypted = ToBytes(encodeText, false)

		//encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)
		//fmt.Println("BASE64:", encryptedBase64)

		//fmt.Println("Kriptovan kod", string(encrypted))
	case XXTEA_CBC:
		fmt.Println("Cipher is XXTEA with CBC")

		encrypted = EncryptCBC(fileData)
	default:
		return fmt.Errorf("unsupported cipher")
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	newFileName := fmt.Sprintf("%s_%s%s", baseName, Cipher.String(), ext)
	newFilePath := filepath.Join("./X", newFileName)

	err2 := os.WriteFile(newFilePath, []byte(encrypted), 0644)
	if err2 != nil {
		return fmt.Errorf("failed to write encrypted file: %v", err2)
	}

	//log.Println("cao ovde sam prosoo 1")
	//fmt.Println("File saved successfully:", newFilePath)
	return nil
}

func DecodeFile(fileData []byte, fileName string) ([]byte, string, error) {

	var decoded []byte

	// Dekodiranje na osnovu algoritma
	switch Cipher {
	case RailFence:
		decoded = DecryptRailFence(fileData, Depth)
		//fmt.Println("Cipher is RailFence", decoded)
	case XXTEA:

		// fmt.Println(string(fileData))

		keyb := ToUint32([]byte(Key), false)
		data := ToUint32(fileData, false)

		// fmt.Println("Ovde je :  ", data)
		decodedText := Btea(data, -len(data), keyb)

		// fmt.Println("Ovde je :  ", decodedText)

		decoded = ToBytes(decodedText, true)

		// fmt.Println("U bajtovima: ", decoded)
		// fmt.Println("String: ", string(decoded))

		// for len(decoded) > 0 && decoded[len(decoded)-1] == 0 {
		// 	decoded = decoded[:len(decoded)-1]
		// }
	case XXTEA_CBC:
		fmt.Println("Cipher is XXTEA with CBC")

		decoded = DecryptCBC(fileData)

	default:
		return nil, "", fmt.Errorf("unsupported cipher")
	}

	// Dodaj "decoded" u naziv fajla
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	decodedFileName := fmt.Sprintf("%s_decoded%s", baseName, ext)

	fmt.Println("Decoded file name:", decodedFileName)
	return decoded, decodedFileName, nil
}

func EncodeData(fileData []byte, filename string) (string, []byte, error) {
	var encrypted []byte

	switch Cipher {
	case RailFence:
		fmt.Println("Cipher is Railfence")
		encrypted = EncryptRailFence(fileData, Depth)
		//fmt.Println("Kriptovan", encrypted)
	case XXTEA:
		fmt.Println("Cipher is XXTEA")
		keyb := ToUint32([]byte(Key), false)

		//fmt.Println("Procitan tekst", string(fileData))

		data := ToUint32(fileData, true)
		encodeText := Btea(data, len(data), keyb)
		encrypted = ToBytes(encodeText, false)

		//encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)
		//fmt.Println("BASE64:", encryptedBase64)

		//fmt.Println("Kriptovan kod", string(encrypted))
	case XXTEA_CBC:
		fmt.Println("Cipher is XXTEA with CBC")

		encrypted = EncryptCBC(fileData)
	default:
		return "", nil, fmt.Errorf("unsupported cipher")
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	newFileName := fmt.Sprintf("%s_%s%s", baseName, Cipher.String(), ext)

	return newFileName, encrypted, nil
}
