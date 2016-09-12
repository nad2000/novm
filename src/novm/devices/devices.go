package devices

import "fmt"

// State is a shorthand equivalent for Python dictionary used as "state"
type State map[string]interface{}

func createCom(comNo int) State {
	var (
		base uint16 // I/O port
		irq  uint16 // interrupt
	)
	switch comNo {
	case 1:
		base, irq = 0x3f8, 4
	case 2:
		base, irq = 0x2f8, 3
	case 3:
		base, irq = 0x3e8, 4
	case 4:
		base, irq = 0x2e8, 3
	}

	return State{
		"driver": "uart",
		"data": State{
			"base":      base,
			"interrupt": irq,
		},
		"cmdline": fmt.Sprintf("console=uart,io,0x%03x", base),
	}
}

// CreateCom1 creats serial port COM1
func CreateCom1() State {
	return createCom(1)
}

// CreateCom2 creats serial port COM2
func CreateCom2() State {
	return createCom(2)
}
