package tcp

import (
	"ZI_Desktop_App/coders"
	"ZI_Desktop_App/hash"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

var (
	ServerPort int = 3000
	ClientPort int = 3000
)

type FileServer struct{}

func (fs *FileServer) Start() {

	serverPortString := fmt.Sprintf(":%d", ServerPort)
	ln, err := net.Listen("tcp", serverPortString)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server started on ", serverPortString)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go fs.ReadLoop(conn)
	}
}

func (fs *FileServer) ReadLoop(conn net.Conn) {
	defer conn.Close()
	buf := new(bytes.Buffer)

	var fileNameLen int32
	if err := binary.Read(conn, binary.LittleEndian, &fileNameLen); err != nil {
		log.Println("Error reading file name length:", err)
		return
	}

	fileName := make([]byte, fileNameLen)
	if _, err := io.ReadFull(conn, fileName); err != nil {
		log.Println("Error reading file name:", err)
		return
	}
	// fmt.Printf("Received file name: %s\n", string(fileName))

	var fileSize int64
	if err := binary.Read(conn, binary.LittleEndian, &fileSize); err != nil {
		log.Println("Error reading file size:", err)
		return
	}
	// fmt.Printf("Received file size: %d bytes\n", fileSize)

	var hashLen int32
	if err := binary.Read(conn, binary.LittleEndian, &hashLen); err != nil {
		log.Println("Error reading hash length:", err)
		return
	}
	// fmt.Printf("Received hash length: %d bytes\n", hashLen)

	hashh := make([]byte, hashLen)
	if _, err := io.ReadFull(conn, hashh); err != nil {
		log.Println("Error reading hash:", err)
		return
	}
	fmt.Printf("Received hash: %x\n", hashh)

	if _, err := io.CopyN(buf, conn, fileSize); err != nil && err != io.EOF {
		log.Println("Error reading file data:", err)
		return
	}

	fileContent := buf.Bytes()

	hashTest := hash.TigerHash(fileContent)

	// fmt.Println("Hashes 1: ", hashh)
	// fmt.Println("Hashes 2: ", hashTest[:])

	if bytes.Equal(hashh, hashTest[:]) {
		fmt.Println("Hashes match!")
	} else {
		fmt.Println("Hashes don't match")
		return
	}

	// FIXME: Dekodiranje fajla koji je primljen

	dataNew, fileNameNew, errDecodeData := coders.DecodeFile(fileContent, string(fileName))
	if errDecodeData != nil {
		log.Print(errDecodeData)
	}

	newFilePath := filepath.Join("./Decoded_TCP", fileNameNew)

	err2 := os.WriteFile(newFilePath, []byte(dataNew), 0644)
	if err2 != nil {
		log.Println(fmt.Errorf("failed to write encrypted file: %v", err2))
	}

	// fmt.Printf("Received %d bytes of file content\n", len(fileContent))
	// fmt.Println("File content:")
	// fmt.Println(string(fileContent))
}

func SendFile(fileName string, data []byte) error {

	// FIXME: Kodiranje fajla pre slanja
	// FIXME: Smisliti po kom algoritmu cu da kodiram na koji nacin da prosledim

	fileNameNew, dataNew, errEncodeData := coders.EncodeData(data, fileName)
	if errEncodeData != nil {
		return errEncodeData
	}
	hash := hash.TigerHash(dataNew)
	hashLen := int32(len(hash))

	clientPortString := fmt.Sprintf(":%d", ClientPort)
	conn, err := net.Dial("tcp", clientPortString)
	if err != nil {
		return err
	}
	fmt.Println("Message sent on PORT: ", clientPortString)
	defer conn.Close()

	// Duzina imena fajla
	fileNameBytes := []byte(fileNameNew)
	if err := binary.Write(conn, binary.LittleEndian, int32(len(fileNameBytes))); err != nil {
		return err
	}

	//Slanje imena fajla
	if _, err := conn.Write(fileNameBytes); err != nil {
		return err
	}
	// fmt.Printf("Written %d bytes of file name\n", len(fileNameBytes))

	// Velicna fajla
	fileSize := int64(len(dataNew))
	if err := binary.Write(conn, binary.LittleEndian, fileSize); err != nil {
		return err
	}
	// fmt.Printf("Written %d bytes of file size\n", 8)

	// Duzina hesha
	if err := binary.Write(conn, binary.LittleEndian, hashLen); err != nil {
		return err
	}
	// fmt.Printf("Written %d bytes of hash length\n", 4)

	//Hesh ceo
	if _, err := conn.Write(hash[:]); err != nil {
		return err
	}
	// fmt.Printf("Written %d bytes of hash data\n", hashLen)

	n, err := conn.Write(dataNew)
	if err != nil {
		return err
	}
	fmt.Printf("Written %d bytes of file data\n", n)

	return nil
}

// func main() {
// 	file := []byte("Cao moje ime je dusan")

// 	go func() {
// 		time.Sleep(4 * time.Second)
// 		if err := SendFile("name.txt", file); err != nil {
// 			log.Fatal("Error sending file:", err)
// 		}
// 	}()

// 	go func() {
// 		time.Sleep(15 * time.Second)
// 		if err := SendFile("numbers.txt", file); err != nil {
// 			log.Fatal("Error sending file:", err)
// 		}
// 	}()

// 	server := &FileServer{}
// 	server.Start()
// }
