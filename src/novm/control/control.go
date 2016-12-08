package control

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	termios "github.com/creack/termios/raw"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

var (
	Debug bool

	sendRPC  = false
	listener *net.UnixListener
	conn     net.Conn
	reader   *bufio.Reader
)

// Read by chuncs the response until '\n' reached or EOF
func readLine() (line string, err error) {
	line, err = reader.ReadString('\n')
	if err == nil && line != "" {
		line = strings.TrimSpace(line)
	}
	return
}

// Bind creates and binds to listen on a Unix socket
func Bind(ctrlPath string) (fd uintptr, err error) {

	dirName := path.Dir(ctrlPath)
	if !utils.PathExists(dirName) {
		err = os.MkdirAll(dirName, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	utils.AddForCleaning(ctrlPath)

	listener, err = net.ListenUnix("unix", &net.UnixAddr{ctrlPath, "unix"})
	if err != nil {
		log.Fatal(err)
	}
	socketFile, err := listener.File()
	if err != nil {
		log.Fatal(err)
	}
	fd = socketFile.Fd()
	if err != nil {
		log.Fatal(err)
	}
	utils.ClearCloexec(int(fd))
	return
}

// Open opens a Unix socket if the socket doesn't exit it creates it
func Open(ctrlPath string) (err error) {

	dirName := path.Dir(ctrlPath)
	if !utils.PathExists(dirName) {
		err = os.MkdirAll(dirName, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	conn, err = net.Dial("unix", ctrlPath)
	if err != nil {
		log.Fatal(err)
	}
	reader = bufio.NewReader(conn)

	return
}

// Close attempts to close the Unix socket
func Close() error {
	return conn.Close()
}

// RPC calls RPC command over the Unix socket
func RPC(name string, params map[string]string) (interface{}, error) {

	var (
		msg    map[string]interface{}
		rawMsg interface{}
		err    error
		line   string
	)

	paramList := make(map[string]interface{})
	for n, v := range params {
		if utils.IsNumber(v) {
			if strings.Contains(v, ".") {
				paramList[n], err = strconv.ParseFloat(v, 64)
			} else {
				paramList[n], err = strconv.ParseInt(v, 10, 64)
			}
		} else if lv := strings.ToLower(v); lv == "true" || lv == "false" {
			paramList[n] = (lv == "true")
		} else {
			paramList[n] = v
		}
	}

	if !sendRPC {
		_, err = conn.Write([]byte("NOVM RPC\n"))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		sendRPC = true
	}

	rpcUUID := utils.UUID()

	rpcCmd := fmt.Sprintf(`{
		"method": "Rpc.%s",
		"params": [%s],
		"id": "%s"
	}`, strings.Title(name), utils.JSONDumps(paramList), rpcUUID)

	_, err = conn.Write([]byte(rpcCmd))
	if err != nil {
		log.Fatal(err)
	}

	line, err = readLine()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = json.Unmarshal([]byte(line), &rawMsg)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	msg = rawMsg.(map[string]interface{})
	respID, ok := msg["id"]
	if !ok {
		err = fmt.Errorf("%+v", msg)
		log.Fatal(err)
		return msg, err
	}
	if respID != rpcUUID {
		err = fmt.Errorf("Request ID and response ID mismatch.", rpcUUID, respID)
		log.Fatal(err)
		return msg, err
	}

	errorMsg, ok := msg["error"]
	if ok && errorMsg != nil {
		log.Fatal(errorMsg)
		return errorMsg, fmt.Errorf("%#v", errorMsg)
	}

	return msg, nil
}

// Run runs a remote command on VM
func Run(command []string, env []string, cwd string, terminal bool) (err error, exitCode int) {

	var (
		line        string
		origTcAttrs *termios.Termios
		isTerminal  bool
	)

	if env == nil || len(env) == 0 {
		env = os.Environ()
	}
	if cwd == "" {
		cwd = "/"
	}

	_, err = conn.Write([]byte("NOVM RUN\n"))
	if err != nil {
		log.Fatal(err)
	}

	startCmd := map[string]interface{}{
		"command":     command,
		"environment": env,
		"terminal":    terminal,
		"cwd":         cwd,
	}
	utils.JSONDump(startCmd, conn, false)

	if terminal {
		origTcAttrs, err = termios.MakeRaw(0)

		if err != nil {
			log.Print("Failed to enable RAW termios: ", err)
			isTerminal = false
		} else {
			isTerminal = true
			defer func() {
				if Debug {
					log.Print("Reverting termios settings...")
				}
				termios.TcSetAttr(0, origTcAttrs)
			}()
		}
	} else {
		isTerminal = false
	}

	// Read input form terminal and sends over to VM
	go func() {

		var (
			buf [4096]byte
			// We use a standard SSH-like escape.
			// Remember if we've seen an ~ already.
			seenTilde bool
		)

		for {
			n, err := os.Stdin.Read(buf[:])

			if err != nil {
				break
			}

			if isTerminal {
				if n == 1 && buf[0] == byte('~') {
					seenTilde = !seenTilde
				} else if seenTilde && n == 1 && buf[0] == byte('.') {
					break
				} else {
					seenTilde = false
				}
			}

			// Server expects JSON encoded value, eg, "ABC...", null, 1, etc.
			if n != 0 {
				conn.Write([]byte("\""))
				conn.Write([]byte(base64.StdEncoding.EncodeToString(buf[:n])))
				conn.Write([]byte("\""))
			} else {
				conn.Write([]byte("null"))
			}
			conn.Write([]byte("\n"))

			if n == 0 {
				break
			}
		}
	}()

	// Remember our exitcode.
	exitCode = 1

	// Expect our first object to be a None.
	// This indicates that we've started with
	// no errors. If it's a string -- error.
	started := false

	for {
		line, err = readLine()
		if Debug {
			log.Println("*** RAW LINE:", line)
		}

		if err != nil {
			log.Print(err)
			return
		}

		if !started {
			// See the note above about 'started'
			if line != "null" {
				err = fmt.Errorf("Received unexpected response %q executing %q", line, command)
				break
			}
			started = true
			continue
		}

		// Expected either: null, integer, or Base64 encoded string:
		if line == "null" {
			// Server is done.
			break
		} else if line[0] == '"' && line[len(line)-1] == '"' {
			data, err := base64.StdEncoding.DecodeString(line[1 : len(line)-1])
			if err != nil {
				if Debug {
					log.Print(err)
				}
				err = fmt.Errorf("Failed to decode Base64 message: %s", line)
				break
			} else {
				fmt.Print(string(data))
			}
		} else if rc, err := strconv.Atoi(line); err == nil {
			// remember the exit code
			exitCode = rc
		} else {
			err = fmt.Errorf("Received unexpected response %#v executing %q", line, command)
			break
		}

	}
	return
}
