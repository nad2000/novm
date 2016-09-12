package control

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/rocksolidlabs/novm/src/novm/utils"

	termios "github.com/creack/termios/raw"
)

type CommandMessage struct {
	method string
	params []string
	id     string
}

var (
	bind     bool
	sendRpc  bool = false
	listener net.Listener
	conn     net.Conn
)

// Read by chuncs the response until '\n' reached or EOF
func readLine() (line []byte, err error) {

	var buf [bytes.MinRead]byte

	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			panic(err)
			if n > 0 {
				return buf[:n], err
			} else {
				return nil, err
			}
		}
		if line == nil {
			if n < bytes.MinRead {
				line = make([]byte, n)
				copy(line, buf[:n])

				break
			} else {
				line = make([]byte, 0, 2*bytes.MinRead)
			}
		}
		if n > 0 {
			line = append(line, buf[:n]...)
		}
		if n == 0 || buf[n-1] == '\n' {
			break
		}
	}
	return line, nil
}

func Open(ctrlPath string) (err error) {
	dirName := path.Dir(ctrlPath)
	if !utils.PathExists(dirName) {
		err = os.MkdirAll(dirName, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if bind {
		listener, err = net.Listen("unix", ctrlPath)
		if err != nil {
			panic(err)
		}
	} else {
		conn, err = net.Dial("unix", ctrlPath)
		if err != nil {
			panic(err)
		}
	}
	return
}

func Rpc(name string, params map[string]string) (interface{}, error) {

	var (
		msg     map[string]interface{}
		rawMsg  interface{}
		err     error
		rawResp []byte
	)

	param_list := make(map[string]interface{})
	for n, v := range params {
		if utils.IsNumber(v) {
			if strings.Contains(v, ".") {
				param_list[n], err = strconv.ParseFloat(v, 64)
			} else {
				param_list[n], err = strconv.ParseInt(v, 10, 64)
			}
		} else if lv := strings.ToLower(v); lv == "true" || lv == "false" {
			param_list[n] = (lv == "true")
		} else {
			param_list[n] = v
		}
	}

	if !sendRpc {
		_, err = conn.Write([]byte("NOVM RPC\n"))
		if err != nil {
			panic(err)
			return nil, err
		}
		sendRpc = true
	}

	rpcUuid := utils.Uuid()

	rpcCmd := fmt.Sprintf(`{
		"method": "Rpc.%s",
		"params": [%s],
		"id": "%s"
	}`, strings.Title(name), utils.JsonDumps(param_list), rpcUuid)

	_, err = conn.Write([]byte(rpcCmd))
	if err != nil {
		panic(err)
	}

	rawResp, err = readLine()
	if err != nil {
		panic(err)
		return nil, err
	}

	err = json.Unmarshal(rawResp, &rawMsg)
	if err != nil {
		panic(err)
		return nil, err
	}

	msg = rawMsg.(map[string]interface{})
	respId, ok := msg["id"]
	if !ok {
		err = fmt.Errorf("%+v", msg)
		panic(err)
		return msg, err
	}
	if respId != rpcUuid {
		err = fmt.Errorf("Request ID and response ID mismatch.", rpcUuid, respId)
		panic(err)
		return msg, err
	}

	errorMsg, ok := msg["error"]
	if ok && errorMsg != nil {
		log.Fatal(errorMsg)
		return errorMsg, fmt.Errorf("%#v", errorMsg)
	}

	return msg, nil
}

func Run(command []string, env []string, cwd string, terminal bool) (interface{}, error) {

	var (
		rawMsg      interface{}
		rawResp     []byte
		origTcAttrs *termios.Termios
		isTerminal  bool
	)

	if env == nil || len(env) == 0 {
		env = os.Environ()
	}
	if cwd == "" {
		cwd = "/"
	}

	_, err := conn.Write([]byte("NOVM RUN\n"))
	if err != nil {
		panic(err)
		return nil, err
	}

	startCmd := map[string]interface{}{
		"command":     command,
		"environment": env,
		"terminal":    terminal,
		"cwd":         cwd,
	}
	utils.JsonDump(startCmd, conn)
	conn.Write([]byte("\n"))

	if terminal {
		origTcAttrs, err = termios.MakeRaw(0)
		if err != nil {
			isTerminal = false
		} else {
			isTerminal = true
			defer termios.TcSetAttr(0, origTcAttrs)
		}
	} else {
		isTerminal = false
	}

	// Remember our exitcode.
	exitCode := 1

	// Expect our first object to be a None.
	// This indicates that we've started with
	// no errors. If it's a string -- error.
	started := false

	// We use a standard SSH-like escape.
	// Remember if we've seen an ~ already.
	seenTitle := false

	// Poll and transform the event stream.
	// This will basically turn this process
	// into a proxy for the remote process.
	readSet := []interface{}{conn, os.Stdin}
	_ = isTerminal
	_ = exitCode
	_ = started
	_ = seenTitle
	_ = readSet

	// Read input form terminal
	go func() {

		var (
			buf       [4096]byte
			seenTilde bool
		)

		for {
			n, err := os.Stdin.Read(buf[:])
			if err == io.EOF {
				break
			} else if err != nil {
				log.Panic(err)
			}
			if isTerminal {
				if n == 1 && buf[0] == byte('~') {
					seenTilde = !seenTitle
				} else if seenTilde && n == 1 && buf[0] == byte('.') {
					break // TODO: should break from the loop
				} else {
					seenTitle = false
				}
			}

			// Server expects JSON encoded value, eg, "ABC...", null, 1, etc.
			conn.Write([]byte("\""))
			if n != 0 {
				conn.Write([]byte(base64.StdEncoding.EncodeToString(buf[:n])))
			} else {

				conn.Write([]byte("null"))
			}
			conn.Write([]byte("\"\n"))

			if n == 0 {
				exitCode = 0 // in order to gracefully exit
				break
			}
		}
	}()

	for {
		rawResp, err = readLine()

		err = json.Unmarshal(rawResp, &rawMsg)

		if err != nil {
			panic(err)
			return nil, err
		}

		if !started {
			// See the note above about 'started'
			if rawMsg != nil {
				err = fmt.Errorf("%#v: %#v", command)
				panic(err)
				return nil, err
			}
			started = true
			continue
		}

		if rawMsg == nil {
			// Server is done.
			os.Exit(exitCode)
		}

		switch rawMsg.(type) {
		case string:
			data, err := base64.StdEncoding.DecodeString(rawMsg.(string))
			if err != nil {
				log.Panicf("Failed to decode Base64 message: %v", rawMsg)
				panic(err)
			}
			fmt.Print(string(data))
		case float64:
			// remember the exit code
			exitCode = int(rawMsg.(float64))
		}

	}
	return rawMsg, nil
}
