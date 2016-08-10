package control

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"novm/utils"
	"os"
	filePath "path"
	"strings"
)

var (
	bind     bool
	sendRpc  bool = false
	listener net.Listener
	conn     net.Conn
)

func Setup(ctrlPath string) (err error) {
	dirName := filePath.Dir(ctrlPath)
	if !utils.PathExists(dirName) {
		err = os.MkdirAll(dirName, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if bind {
		//conn, err = net.ListenUnix("unix", &net.UnixAddr{ctrlPath, "unix"})
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

func rpc(name string, params map[string]string) (interface{}, error) {

	var (
		buf    [1000]byte
		msg    map[string]interface{}
		rawMsg interface{}
		err    error
	)

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
	"id": %s
}
`, strings.Title(name), utils.JsonDumps(params), rpcUuid)

	_, err = conn.Write([]byte(rpcCmd))
	if err != nil {
		panic(err)
	}

	n, err := conn.Read(buf[:])
	if err != nil {
		panic(err)
		return nil, err
	}
	err = json.Unmarshal(buf[:n], &rawMsg)
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
	if ok {
		return errorMsg, fmt.Errorf(errorMsg.(string))
	}

	result, ok := msg["id"]
	if !ok {
		err = fmt.Errorf("%+v", msg)
		panic(err)
		return msg, err
	}
	return result, nil
}

func Run(command string, env []string, cwd string, terminal bool) {
	if env == nil {
		env = os.Environ()
	}
	if cwd == "" {
		cwd = "/"
	}
	_, err := conn.Write([]byte("NOVM RPC\n"))
	if err != nil {
		panic(err)
	}

	startCmd := map[string]interface{}{
		"command":     command,
		"environment": env,
		"terminal":    terminal,
		"cwd":         cwd,
	}
	utils.JsonDump(startCmd, conn)

	// TODO: setting raw termios
	isTerminal := false

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

	for {
		break
	}
}
