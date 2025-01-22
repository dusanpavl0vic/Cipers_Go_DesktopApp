package main

import (
	"ZI_Desktop_App/coders"
	"ZI_Desktop_App/tcp"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/fsnotify/fsnotify"
)

var (
	doneTarget = make(chan bool)
	doneX      = make(chan bool)

	startButton, stopButton *widget.Button
	statusLabel             *widget.Label
	statusLabelTCP          *widget.Label
	CBC_enable              bool = false

	leftListContainer  = container.NewGridWithRows(500)
	rightListContainer = container.NewGridWithRows(500)

	leftFilesSet = make(map[string]struct{})

	startedWileWatcher bool = false

	startOnce sync.Once

	checkedCBC = false
)

func TCPSettings(a fyne.App) fyne.Window {
	settingsWindow := a.NewWindow("TCP Settings")

	serverPortEntry := widget.NewEntry()
	serverPortEntry.SetText(fmt.Sprintf("%d", tcp.ServerPort))
	serverPortEntry.OnChanged = func(input string) {
		valid := ""
		for _, r := range input {
			if r >= '0' && r <= '9' {
				valid += string(r)
			}
		}
		if input != valid {
			serverPortEntry.SetText(valid)
		}
	}
	serverPortButton := widget.NewButton("Set Server Port", func() {
		if port, err := strconv.Atoi(serverPortEntry.Text); err == nil {
			tcp.ServerPort = port
			fmt.Println("Server Port Set:", tcp.ServerPort)
		} else {
			fmt.Println("Invalid Server Port")
		}
	})

	serverPortContainer := container.NewGridWithColumns(3,
		widget.NewLabel("Server Port:"),
		serverPortEntry,
		serverPortButton,
	)

	clientPortEntry := widget.NewEntry()
	clientPortEntry.SetText(fmt.Sprintf("%d", tcp.ClientPort))
	clientPortEntry.OnChanged = func(input string) {
		valid := ""
		for _, r := range input {
			if r >= '0' && r <= '9' {
				valid += string(r)
			}
		}
		if input != valid {
			clientPortEntry.SetText(valid)
		}
	}
	clientPortButton := widget.NewButton("Set Client Port", func() {
		if port, err := strconv.Atoi(clientPortEntry.Text); err == nil {
			tcp.ClientPort = port
			fmt.Println("Client Port Set:", tcp.ClientPort)
		} else {
			fmt.Println("Invalid Client Port")
		}
	})

	clientPortContainer := container.NewGridWithColumns(3,
		widget.NewLabel("Client Port:"),
		clientPortEntry,
		clientPortButton,
	)

	settingsLayout := container.NewVBox(
		serverPortContainer,
		clientPortContainer,
	)

	settingsWindow.SetContent(settingsLayout)
	settingsWindow.Resize(fyne.NewSize(400, 200))

	return settingsWindow
}
func TCPContainer(win fyne.Window, a fyne.App) *fyne.Container {

	tcpButton := widget.NewButton("Start TCP Server", func() {
		startOnce.Do(func() {
			server := &tcp.FileServer{}
			go server.Start()

			statusLabelTCP.SetText("TCP Status: Started")
			fmt.Println("TCP server started.")
		})
	})

	fileButton := widget.NewButton("Select File", func() {
		dialog.NewFileOpen(
			func(file fyne.URIReadCloser, err error) {
				if err != nil || file == nil {
					log.Println("No file selected.")
					return
				}

				// Read file form client
				data, err := io.ReadAll(file)
				if err != nil {
					log.Println("Error reading file:", err)
					return
				}

				// Send file to client
				go func() {
					if err := tcp.SendFile(file.URI().Name(), data); err != nil {
						log.Fatal("Error sending file:", err)
					}
					log.Println("File sent:", file.URI().Name())
				}()
			}, win,
		).Show()
	})

	tcpsettingButton := widget.NewButton("TCP Settings", func() {
		settingsWindow := TCPSettings(a)
		settingsWindow.Show()
	})

	return container.NewVBox(tcpButton, fileButton, tcpsettingButton)
}

func fileDecodeButton(win fyne.Window) *widget.Button {
	decodeButton := widget.NewButton("Select and Decode File", func() {
		dialog.NewFileOpen(
			func(file fyne.URIReadCloser, err error) {
				if err != nil || file == nil {
					log.Println("No file selected or error occurred:", err)
					return
				}

				go func() {
					defer file.Close()

					data, err := io.ReadAll(file)
					if err != nil {
						log.Println("Error reading file:", err)
						return
					}

					decodedData, newFileName, err1 := coders.DecodeFile(data, file.URI().Name())
					if err1 != nil {
						log.Println("Error decoding file:", err1)
						return
					}
					saveDialog := dialog.NewFileSave(
						func(saveFile fyne.URIWriteCloser, err error) {
							if err != nil || saveFile == nil {
								log.Println("No save location selected or error occurred:", err)
								return
							}
							defer saveFile.Close()

							_, err = saveFile.Write(decodedData)
							if err != nil {
								log.Println("Error saving file:", err)
								return
							}

							log.Println("File successfully saved as:", saveFile.URI().Path())
						}, win,
					)
					saveDialog.SetFileName(newFileName)
					saveDialog.Show()
				}()
			}, win,
		).Show()
	})

	return decodeButton
}

func uploadAndEncodeContainerCreate(w fyne.Window) *fyne.Container {
	uploadButton := widget.NewButton("Upload File and encode", func() {
		if !startedWileWatcher {
			dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if err != nil || uc == nil {
					return
				}

				// Gorutina za obradu fajla

				go func() {

					// Čitanje sadržaja fajla
					fileContent, readErr := io.ReadAll(uc)
					if readErr != nil {
						dialog.ShowError(readErr, w)
						return
					}

					// Izdvajanje imena fajla
					fileName := uc.URI().Name()

					// Pozivanje funkcije za kodiranje
					encodedErr := coders.EncodeFile(fileContent, fileName)
					if encodedErr != nil {
						dialog.ShowError(encodedErr, w)
						return
					}

					uc.Close()
					dialog.ShowInformation("Success", "File encoded successfully!", w)
				}()

			}, w).Show()
		}
	})

	return container.NewVBox(uploadButton)
}

func uploadContainerCreate(w fyne.Window) *fyne.Container {
	uploadButton := widget.NewButton("Upload File", func() {
		dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}

			go func() {
				defer uc.Close()

				fileContent, readErr := io.ReadAll(uc)
				if readErr != nil {
					dialog.ShowError(readErr, w)
					return
				}

				targetDir := "./Target"
				dirErr := os.MkdirAll(targetDir, os.ModePerm)
				if dirErr != nil {
					dialog.ShowError(dirErr, w)
					return
				}

				filePath := targetDir + "/" + uc.URI().Name()
				writeErr := os.WriteFile(filePath, fileContent, 0644)
				if writeErr != nil {
					dialog.ShowError(writeErr, w)
					return
				}

				dialog.ShowInformation("Success", "File uploaded and encode successfully!", w)
			}()
		}, w).Show()
	})

	return container.NewVBox(uploadButton)
}

func createCBCCheckbox() *widget.Check {
	cbcCheckbox := widget.NewCheck("CBC", func(checked bool) {
		CBC_enable = checked
		if CBC_enable {
			println("CBC is enabled")
		} else {
			println("CBC is disabled")
		}
	})
	return cbcCheckbox
}

// func listFiles(directory string) ([]string, error) {
// 	files, err := os.ReadDir(directory)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var fileNames []string
// 	for _, file := range files {
// 		fileNames = append(fileNames, file.Name())
// 	}
// 	return fileNames, nil
// }

// func contains(slice []string, item string) bool {
// 	for _, v := range slice {
// 		if v == item {
// 			return true
// 		}
// 	}
// 	return false
// }

func leftListContainerAdd(label string) {
	if _, exists := leftFilesSet[label]; exists {
		return
	}
	leftFilesSet[label] = struct{}{}
	leftListContainer.Add(widget.NewLabel(label))
	//fmt.Printf("Added to leftListContainer: %s\n", label)
}

func rightListContainerAdd(label string) {
	rightListContainer.Add(widget.NewLabel(label))

	//fmt.Printf("Added to rightListContainer: %s\n", label)
}

func loadExistingFiles(dir string, addToContainer func(string)) {
	leftFilesSet = make(map[string]struct{})
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dir, err)
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			addToContainer(file.Name())
		}
	}
}

func watchTargetDirectory() {

	dir := "./Target"

	leftListContainer.Objects = nil

	absPath, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for %s: %v", dir, err)
	}
	loadExistingFiles(absPath, leftListContainerAdd)

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(absPath); err != nil {
		log.Fatalf("Failed to watch directory %s: %v", absPath, err)
	}
	fmt.Printf("Watching directory: %s\n", absPath)

	fmt.Println("Starting file watcher for Target...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Printf("Created file in Target: %s\n", event.Name)
				fileName := filepath.Base(event.Name)
				leftListContainerAdd(fileName)

				log.Println("File created:", event.Name)

				go func(filePath string) {
					fmt.Println("File path: ", filePath)
					fileData, err1 := readFile(filePath)
					if err1 != nil {
						log.Printf("Error reading file %s: %v", filePath, err1)
						return
					}
					err2 := coders.EncodeFile(fileData, fileName)
					if err2 != nil {
						log.Printf("Error encoding file %s: %v", fileName, err2)
					}
				}(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Error: %v\n", err)
		case d := <-doneTarget:
			if d {
				return
			}
		}
	}
}

func watchXDirectory() {

	dir := "./X"

	rightListContainer.Objects = nil

	absPath, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for %s: %v", dir, err)
	}

	files, err := listFiles(absPath)
	if err != nil {
		log.Fatalf("Failed to list files in directory %s: %v", absPath, err)
	}
	for _, file := range files {
		rightListContainerAdd(file)
	}

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(absPath); err != nil {
		log.Fatalf("Failed to watch directory %s: %v", absPath, err)
	}
	fmt.Printf("Watching directory: %s\n", absPath)

	fmt.Println("Starting file watcher for X...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Printf("File event in X: %s\n", event.Name)
				files, err := listFiles(absPath)
				if err != nil {
					log.Printf("Failed to refresh file list in directory %s: %v", absPath, err)
					continue
				}
				rightListContainer.Objects = nil
				for _, file := range files {
					rightListContainerAdd(file)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Error: %v\n", err)
		case d := <-doneX:
			if d {
				return
			}
		}
	}
}

func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	return files, err
}

func stopFileWatcherTarget() {
	fmt.Println("Stopping file watcher...")

	doneTarget <- true
}

// func stopFileWatcherX() {
// 	fmt.Println("Stopping file watcher...")
// 	uploadAndEncodeContainer.Show()
// 	doneX <- true
// }

func createStartButton() *widget.Button {
	return widget.NewButton("Start", func() {
		go watchTargetDirectory()
		startedWileWatcher = true
		startButton.Disable()
		stopButton.Enable()
		statusLabel.SetText("Status: File Watcher Running")
	})
}

func createStopButton() *widget.Button {
	return widget.NewButton("Stop", func() {
		go stopFileWatcherTarget()
		startedWileWatcher = false
		startButton.Enable()
		stopButton.Disable()
		statusLabel.SetText("Status: File Watcher Stopped")
	})
}

func createSettingsWindow(a fyne.App) fyne.Window {
	name := coders.Cipher.String() + " Settings"
	settingsWindow := a.NewWindow(name)

	// Key Entry
	keyEntry := widget.NewEntry()
	keyEntry.SetText(coders.Key)
	keyButton := widget.NewButton("Set Key", func() {
		coders.Key = keyEntry.Text
	})

	keyContainer := container.NewGridWithColumns(3,
		widget.NewLabel("Key:"),
		keyEntry,
		keyButton,
	)

	cbcKeyEntry := widget.NewEntry()
	cbcKeyEntry.SetText(string(coders.KeyCBC))
	cbcKeyButton := widget.NewButton("Set CBC Key", func() {
		coders.KeyCBC = []byte(cbcKeyEntry.Text)
	})

	cbcKeyContainer := container.NewGridWithColumns(3,
		widget.NewLabel("CBC Key:"),
		cbcKeyEntry,
		cbcKeyButton,
	)

	cbcSettingsContainer := container.NewVBox(cbcKeyContainer)
	cbcSettingsContainer.Hide()

	cbcCheckbox := widget.NewCheck("Enable CBC Mode", func(checked bool) {
		if checked {
			cbcSettingsContainer.Show()
			coders.Cipher = coders.XXTEA_CBC
			fmt.Println("Selektovan je: ", coders.Cipher)
			checkedCBC = true

		} else {
			cbcSettingsContainer.Hide()
			coders.Cipher = coders.XXTEA
			fmt.Println("Selektovan je: ", coders.Cipher)
			checkedCBC = false
		}
	})
	if checkedCBC {
		cbcCheckbox.SetChecked(checkedCBC)
	}

	xxteaContainer := container.NewVBox(
		keyContainer,
		cbcCheckbox,
		cbcSettingsContainer,
	)

	// Depth Entry
	depthEntry := widget.NewEntry()
	depthEntry.SetText("3")
	depthEntry.OnChanged = func(input string) {
		valid := ""
		for _, r := range input {
			if r >= '0' && r <= '9' {
				valid += string(r)
			}
		}
		if input != valid {
			depthEntry.SetText(valid)
		}
	}
	depthButton := widget.NewButton("Set Depth", func() {
		if d, err := strconv.Atoi(depthEntry.Text); err == nil {
			coders.Depth = d
		}
	})

	depthContainer := container.NewGridWithColumns(3,
		widget.NewLabel("Depth:"),
		depthEntry,
		depthButton,
	)

	railFenceContainer := container.NewVBox(
		depthContainer,
	)

	if coders.Cipher == coders.RailFence {
		railFenceContainer.Show()
		xxteaContainer.Hide()
	} else if coders.Cipher == coders.XXTEA || coders.Cipher == coders.XXTEA_CBC {
		railFenceContainer.Hide()
		xxteaContainer.Show()
	}

	settingsLayout := container.NewVBox(
		widget.NewLabel(name),
		xxteaContainer,
		railFenceContainer,
	)

	settingsWindow.SetContent(settingsLayout)
	settingsWindow.Resize(fyne.NewSize(400, 200))

	return settingsWindow
}

func createCipherSelect() *widget.Select {
	return widget.NewSelect([]string{coders.RailFence.String(), coders.XXTEA.String()}, func(selected string) {
		switch selected {
		case coders.RailFence.String():
			coders.Cipher = coders.RailFence
		case coders.XXTEA.String():
			coders.Cipher = coders.XXTEA
		}
	})
}

func main() {
	targetDir := "./Target"
	xDir := "./X"

	go watchXDirectory()

	_ = os.MkdirAll(targetDir, os.ModePerm)
	_ = os.MkdirAll(xDir, os.ModePerm)

	a := app.New()
	w := a.NewWindow("Zastita informacija")
	statusLabel = widget.NewLabel("Status: Idle")
	statusLabelTCP = widget.NewLabel("TCP Status: Not Started")

	leftListScroll := container.NewScroll(leftListContainer)
	rightListScroll := container.NewScroll(rightListContainer)

	leftListScroll.SetMinSize(fyne.NewSize(1000, 500))
	rightListScroll.SetMinSize(fyne.NewSize(1000, 500))

	startButton = createStartButton()
	stopButton = createStopButton()

	stopButton.Disable()

	watcherContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), container.NewGridWrap(fyne.NewSize(150, 30), startButton), layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), container.NewGridWrap(fyne.NewSize(150, 30), stopButton), layout.NewSpacer()),
		layout.NewSpacer(),
	)

	settingsButton := widget.NewButton("Settings cipher", func() {
		settingsWindow := createSettingsWindow(a)
		settingsWindow.Show()
	})

	cipherSelect := createCipherSelect()
	cipherSelect.SetSelected(coders.Cipher.String())
	chiperLabel := widget.NewLabel("Coder/Decoder: ")
	cipherContainer := container.NewHBox(
		layout.NewSpacer(),
		chiperLabel,
		cipherSelect,
		layout.NewSpacer(),
	)

	uploadContainer := uploadContainerCreate(w)
	uploadAndEncodeContainer := uploadAndEncodeContainerCreate(w)

	tcpContainer := TCPContainer(w, a)
	decoderContainer := fileDecodeButton(w)

	buttonsContainer := container.NewVBox(
		watcherContainer,
		cipherContainer,
		uploadContainer,
		uploadAndEncodeContainer,
		decoderContainer,
		layout.NewSpacer(),
		tcpContainer,
		layout.NewSpacer(),
	)

	leftPanel := container.NewBorder(
		widget.NewLabel("Target Directory                         "), // Gore
		nil,                                    // Dole
		nil,                                    // Levo
		nil,                                    // Desno
		container.NewScroll(leftListContainer), // Centar
	)

	rightPanel := container.NewBorder(
		widget.NewLabel("X Directory                              "), // Gore
		nil,                                     // Dole
		nil,                                     // Levo
		nil,                                     // Desno
		container.NewScroll(rightListContainer), // Centar
	)

	mainLayout := container.NewBorder(
		container.NewHBox(statusLabel, layout.NewSpacer(), statusLabelTCP), // Gore
		settingsButton,   // Dole
		leftPanel,        // Levo
		rightPanel,       // Desno
		buttonsContainer, // Centar
	)

	w.SetContent(mainLayout)
	w.Resize(fyne.NewSize(1000, 600))
	w.ShowAndRun()
}

func readFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	return data, nil
}
