package coders

import (
	"strings"
)

// TODO: ISPRAVI DEKODIRANJE SLIKE JER NESTO NE VALJA ISPISUJE SVE 10-TKE
func DecryptRailFence(cipher []byte, depth int) []byte {
	if depth <= 1 {
		return cipher
	}

	railPattern := make([]int, len(cipher))
	downward := false
	currentRow := 0

	for i := 0; i < len(cipher); i++ {
		railPattern[i] = currentRow
		if currentRow == 0 {
			downward = true
		}
		if currentRow == depth-1 {
			downward = false
		}
		if downward {
			currentRow++
		} else {
			currentRow--
		}
	}

	railPositions := make([]int, depth)
	for _, row := range railPattern {
		railPositions[row]++
	}

	rails := make([][]byte, depth)
	currentIndex := 0
	for i, count := range railPositions {
		rails[i] = cipher[currentIndex : currentIndex+count]
		currentIndex += count
	}

	plainText := make([]byte, 0, len(cipher))
	currentRow = 0
	railIndices := make([]int, depth)

	for i := 0; i < len(cipher); i++ {
		plainText = append(plainText, rails[currentRow][railIndices[currentRow]])
		railIndices[currentRow]++
		if currentRow == 0 {
			downward = true
		}
		if currentRow == depth-1 {
			downward = false
		}
		if downward {
			currentRow++
		} else {
			currentRow--
		}
	}

	return plainText
}

func EncryptRailFence(message []byte, depth int) []byte {
	if depth <= 1 {
		return message
	}

	rails := make([][]byte, depth)
	for i := range rails {
		rails[i] = make([]byte, 0, len(message))
	}

	downward := false
	currentRow := 0

	for _, char := range message {
		rails[currentRow] = append(rails[currentRow], char)
		if currentRow == 0 || currentRow == depth-1 {
			downward = !downward
		}
		if downward {
			currentRow++
		} else {
			currentRow--
		}
	}

	cipherText := make([]byte, 0, len(message))
	for i := 0; i < depth; i++ {
		cipherText = append(cipherText, rails[i]...)
	}

	return cipherText
}

func PrintRails(rails [][]byte) string {
	var sb strings.Builder

	for _, row := range rails {
		for _, char := range row {
			if char != 0xFF {
				sb.WriteByte(char)
			} else {
				sb.WriteByte('_')
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// func main() {

// 	//original := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
// 	original := "HELLO WORLD"
// 	cipher := EncryptRailFence([]byte(original), 3)
// 	decoded := DecryptRailFence(cipher, 3)
// 	fmt.Println(reflect.DeepEqual([]byte(original), decoded))

// }
