package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/apenwarr/fixconsole"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	// Windows constats
	invalidHandleValue = ^windows.Handle(0)
	pageReadWrite      = 0x4
	fileMapWrite       = 0x2

	// ssh-agent/Pageant constants
	agentMaxMessageLength = 8192
	agentCopyDataID       = 0x804e50ba
)

var (
	verbose = flag.Bool("verbose", false, "Enable verbose logging")
	logFile = flag.String("logfile", "wsl2-gpg-ssh.log", "Path to logfile")
	gpg     = flag.String("gpg", "", "gpg mode")

	failureMessage = [...]byte{0, 0, 0, 1, 5}
)

// copyDataStruct is used to pass data in the WM_COPYDATA message.
// We directly pass a pointer to our copyDataStruct type, we need to be
// careful that it matches the Windows type exactly
type copyDataStruct struct {
	dwData uintptr
	cbData uint32
	lpData uintptr
}

func queryPageant(buf []byte) (result []byte, err error) {
	if len(buf) > agentMaxMessageLength {
		err = errors.New("Message too long")
		return
	}

	hwnd := win.FindWindow(syscall.StringToUTF16Ptr("Pageant"), syscall.StringToUTF16Ptr("Pageant"))

	// Launch gpg-connect-agent
	if hwnd == 0 {
		log.Println("launching gpg-connect-agent")
		exec.Command("gpg-connect-agent", "/bye").Run()
	}

	hwnd = win.FindWindow(syscall.StringToUTF16Ptr("Pageant"), syscall.StringToUTF16Ptr("Pageant"))
	if hwnd == 0 {
		err = errors.New("Could not find Pageant window")
		return
	}

	// Adding process id in order to support parrallel requests.
	requestName := "WSLPageantRequest" + strconv.Itoa(os.Getpid())
	mapName := fmt.Sprintf(requestName)

	fileMap, err := windows.CreateFileMapping(invalidHandleValue, nil, pageReadWrite, 0, agentMaxMessageLength, syscall.StringToUTF16Ptr(mapName))
	if err != nil {
		return
	}
	defer func() {
		windows.CloseHandle(fileMap)
	}()

	sharedMemory, err := windows.MapViewOfFile(fileMap, fileMapWrite, 0, 0, 0)
	if err != nil {
		return
	}
	defer windows.UnmapViewOfFile(sharedMemory)

	sharedMemoryArray := (*[agentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))
	copy(sharedMemoryArray[:], buf)

	mapNameWithNul := mapName + "\000"

	// We use our knowledge of Go strings to get the length and pointer to the
	// data and the length directly
	cds := copyDataStruct{
		dwData: agentCopyDataID,
		cbData: uint32(((*reflect.StringHeader)(unsafe.Pointer(&mapNameWithNul))).Len),
		lpData: ((*reflect.StringHeader)(unsafe.Pointer(&mapNameWithNul))).Data,
	}

	ret := win.SendMessage(hwnd, win.WM_COPYDATA, 0, uintptr(unsafe.Pointer(&cds)))
	if ret == 0 {
		err = errors.New("WM_COPYDATA failed")
		return
	}

	len := binary.BigEndian.Uint32(sharedMemoryArray[:4])
	len += 4

	if len > agentMaxMessageLength {
		err = errors.New("Return message too long")
		return
	}

	result = make([]byte, len)
	copy(result, sharedMemoryArray[:len])

	return
}

func main() {
	fixconsole.FixConsoleIfNeeded()
	flag.Parse()

	if *verbose {
		//Setting logput to file because we use stdout for communication
		f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
		log.Println("Starting exe")
	}

	if *gpg != "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("failed to find user home dir")
		}
		basePath := filepath.Join(homeDir, "AppData", "Roaming", "gnupg")
		handleGPG(filepath.Join(basePath, *gpg))
	} else {
		handleSSH()
	}

}

func handleGPG(path string) {
	var port int
	var nonce [16]byte

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	tmp, _, err := reader.ReadLine()
	port, err = strconv.Atoi(string(tmp))
	n, err := reader.Read(nonce[:])
	if err != nil {
		if *verbose {
			log.Printf("Could not read port from gpg nonce: %v\n", err)
		}
		return
	}

	if n != 16 {
		if *verbose {
			log.Printf("Could not connet gpg: incorrect number of bytes for nonceRead incorrect number of bytes for nonce\n")
		}
		return
	}

	gpgConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		if *verbose {
			log.Printf("Could not connet gpg: %v\n", err)
		}
		return
	}

	_, err = gpgConn.Write(nonce[:])
	if err != nil {
		if *verbose {
			log.Printf("Could not authenticate gpg: %v\n", err)
		}
		return
	}

	go func() {
		_, err := io.Copy(gpgConn, os.Stdin)
		if err != nil {
			if *verbose {
				log.Printf("Could not copy gpg data from assuan socket to socket: %v\n", err)
			}
			return
		}
	}()

	_, err = io.Copy(os.Stdout, gpgConn)
	if err != nil {
		if *verbose {
			log.Printf("Could not copy gpg data from socket to assuan socket: %v\n", err)
		}
		return
	}
}

func handleSSH() {
	reader := bufio.NewReader(os.Stdin)
	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lenBuf)
		if err != nil {
			if *verbose {
				log.Printf("io.ReadFull length error '%s'", err)
			}
			return
		}

		len := binary.BigEndian.Uint32(lenBuf)
		log.Printf("Reading length: %v", len)
		buf := make([]byte, len)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			if *verbose {
				log.Printf("io.ReadFull data error '%s'", err)
			}
			return
		}

		result, err := queryPageant(append(lenBuf, buf...))
		if err != nil {
			// If for some reason talking to Pageant fails we fall back to
			// sending an agent error to the client
			if *verbose {
				log.Printf("Pageant query error '%s'", err)
			}
			result = failureMessage[:]
		}

		_, err = os.Stdout.Write(result)
		if err != nil {
			if *verbose {
				log.Printf("net.Conn.Write error '%s'", err)
			}
			return
		}
	}
}
