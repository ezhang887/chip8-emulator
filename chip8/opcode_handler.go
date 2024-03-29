package chip8

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func (cpu *Cpu) HandleOpcode() error {
	switch (cpu.Opcode & 0xF000) >> 12 {
	case 0x0:
		return cpu.Handle00Opcodes()
	case 0x1:
		return cpu.Handle1NNNOpcode()
	case 0x2:
		return cpu.Handle2NNNOpcode()
	case 0x3:
		return cpu.Handle3XNNOpcode()
	case 0x4:
		return cpu.Handle4XNNOpcode()
	case 0x5:
		return cpu.Handle5XY0Opcode()
	case 0x6:
		return cpu.Handle6XNNOpcode()
	case 0x7:
		return cpu.Handle7xNNOpcode()
	case 0x8:
		return cpu.Handle8XYOpcodes()
	case 0x9:
		return cpu.Handle9XY0Opcode()
	case 0xA:
		return cpu.HandleANNNOpcode()
	case 0xB:
		return cpu.HandleBNNNOpcode()
	case 0xC:
		return cpu.HandleCXNNOpcode(time.Now().UnixNano())
	case 0xD:
		return cpu.HandleDXYNOpcode()
	case 0xE:
		return cpu.HandleEXOpcodes()
	case 0xF:
		return cpu.HandleFXOpcodes()
	default:
		return errors.New(fmt.Sprintf("%d: Invalid Opcode in HandleOpcode()", cpu.Opcode))
	}
}

func (cpu *Cpu) Handle00Opcodes() error {
	switch cpu.Opcode {
	//clear screen
	case 0x00E0:
		for i := range cpu.Display {
			cpu.Display[i] = 0
		}
		cpu.ShouldDraw = true
	//return from function
	case 0x00EE:
		if cpu.SP == 0 {
			return errors.New("SP is 0! No functions to return from!")
		}
		cpu.SP--
		cpu.PC = cpu.Stack[cpu.SP]
	default:
		return errors.New(fmt.Sprintf("%d: Invalid Opcode in Handle00Opcodes()", cpu.Opcode))
	}
	cpu.PC += 2
	return nil
}

//jump to address NNN
func (cpu *Cpu) Handle1NNNOpcode() error {
	target := cpu.Opcode & 0x0FFF
	cpu.PC = target
	return nil
}

//call function (jump-and-link) at address NNN
func (cpu *Cpu) Handle2NNNOpcode() error {
	if cpu.SP >= STACK_SIZE {
		return errors.New("Cannot call function! Stack is full!")
	}
	cpu.Stack[cpu.SP] = cpu.PC //store current address in stack
	cpu.SP++
	target := cpu.Opcode & 0x0FFF
	cpu.PC = target
	return nil
}

//Skip next instruction if V[X] == NN
func (cpu *Cpu) Handle3XNNOpcode() error {
	nn := uint8(cpu.Opcode & 0x00FF)
	x := (cpu.Opcode & 0x0F00) >> 8
	if cpu.V[x] == nn {
		cpu.PC += 4
	} else {
		cpu.PC += 2
	}
	return nil
}

//Skip next instruction if V[X] != NN
func (cpu *Cpu) Handle4XNNOpcode() error {
	nn := uint8(cpu.Opcode & 0x00FF)
	x := (cpu.Opcode & 0x0F00) >> 8
	if cpu.V[x] != nn {
		cpu.PC += 4
	} else {
		cpu.PC += 2
	}
	return nil
}

//Skip next instruction if V[X] == V[Y]
func (cpu *Cpu) Handle5XY0Opcode() error {
	if cpu.Opcode&0x000F != 0 {
		return errors.New("Last word in 5XY0 instructions should be 0!")
	}
	x := (cpu.Opcode & 0x0F00) >> 8
	y := (cpu.Opcode & 0x00F0) >> 4
	if cpu.V[x] == cpu.V[y] {
		cpu.PC += 4
	} else {
		cpu.PC += 2
	}
	return nil
}

//V[X] = NN
func (cpu *Cpu) Handle6XNNOpcode() error {
	nn := uint8(cpu.Opcode & 0x00FF)
	x := (cpu.Opcode & 0x0F00) >> 8
	cpu.V[x] = nn
	cpu.PC += 2
	return nil
}

//V[X] += NN
func (cpu *Cpu) Handle7xNNOpcode() error {
	nn := uint8(cpu.Opcode & 0x00FF)
	x := (cpu.Opcode & 0x0F00) >> 8
	cpu.V[x] += nn
	cpu.PC += 2
	return nil
}

func (cpu *Cpu) Handle8XYOpcodes() error {
	x := (cpu.Opcode & 0x0F00) >> 8
	y := (cpu.Opcode & 0x00F0) >> 4
	switch cpu.Opcode & 0x000F {
	//V[x] = V[y]
	case 0x0:
		cpu.V[x] = cpu.V[y]
	//V[x] = V[x] | V[y]
	case 0x1:
		cpu.V[x] = cpu.V[x] | cpu.V[y]
	//V[x] = V[x] & V[y]
	case 0x2:
		cpu.V[x] = cpu.V[x] & cpu.V[y]
	//V[x] = V[x] ^ V[y]
	case 0x3:
		cpu.V[x] = cpu.V[x] ^ cpu.V[y]
	//V[x] += V[y]
	case 0x4:
		res := uint16(cpu.V[x]) + uint16(cpu.V[y])
		var carry uint8 = 0
		if res > 0xFF {
			carry = 1
		}
		cpu.V[x] = uint8(res)
		cpu.V[0xF] = carry
	//V[x] -= V[y]
	case 0x5:
		res := int16(cpu.V[x]) - int16(cpu.V[y])
		var not_borrow uint8 = 1
		if res < 0x00 {
			not_borrow = 0
		}
		cpu.V[x] = uint8(res)
		cpu.V[0xF] = not_borrow
	//V[x] >>= 1
	case 0x6:
		lsb := cpu.V[x] & 0x01
		cpu.V[0xF] = lsb
		cpu.V[x] >>= 1
	//V[x] = V[y] - V[x]
	case 0x7:
		res := int16(cpu.V[y]) - int16(cpu.V[x])
		var not_borrow uint8 = 1
		if res < 0x00 {
			not_borrow = 0
		}
		cpu.V[x] = uint8(res)
		cpu.V[0xF] = not_borrow
	//V[x] <<= 1
	case 0xE:
		lsb := cpu.V[x] & 0x01
		cpu.V[0xF] = lsb
		cpu.V[x] <<= 1
	default:
		return errors.New(fmt.Sprintf("%d: Invalid Opcode in Handle8XYOpcodes()", cpu.Opcode))
	}
	cpu.PC += 2
	return nil
}

//Skip next instruction if V[X] != V[Y]
func (cpu *Cpu) Handle9XY0Opcode() error {
	if cpu.Opcode&0x000F != 0 {
		return errors.New("Last word in 5XY0 instructions should be 0!")
	}
	x := (cpu.Opcode & 0x0F00) >> 8
	y := (cpu.Opcode & 0x00F0) >> 4
	if cpu.V[x] != cpu.V[y] {
		cpu.PC += 4
	} else {
		cpu.PC += 2
	}
	return nil
}

//I = NNN
func (cpu *Cpu) HandleANNNOpcode() error {
	nnn := cpu.Opcode & 0x0FFF
	cpu.I = nnn
	cpu.PC += 2
	return nil
}

//PC = V[0] + NNN
func (cpu *Cpu) HandleBNNNOpcode() error {
	nnn := cpu.Opcode & 0x0FFF
	cpu.PC = uint16(cpu.V[0]) + nnn
	return nil
}

//V[x] = rand() & NN
func (cpu *Cpu) HandleCXNNOpcode(seed int64) error {
	x := (cpu.Opcode & 0x0F00) >> 8
	nn := uint8(cpu.Opcode & 0x00FF)

	source := rand.NewSource(seed)
	rand_generator := rand.New(source)
	rand_num := uint8(rand_generator.Intn(256))
	cpu.V[x] = rand_num & nn
	cpu.PC += 2
	return nil
}

func (cpu *Cpu) HandleDXYNOpcode() error {
	x := uint16(cpu.V[(cpu.Opcode&0x0F00)>>8])
	y := uint16(cpu.V[(cpu.Opcode&0x00F0)>>4])
	n := cpu.Opcode & 0x000F

	cpu.V[0xF] = 0
	for curr_y := uint16(0); curr_y < n; curr_y++ {
		pixel := cpu.Memory[cpu.I+curr_y]
		for curr_x := uint16(0); curr_x < 8; curr_x++ {
			if (pixel & (0x80 >> curr_x)) != 0 {
				idx := x + curr_x + (y+curr_y)*NUM_COLS
				if cpu.Display[idx] == 1 {
					cpu.V[0xF] = 1
				}
				cpu.Display[idx] ^= 1
			}
		}
	}
	cpu.PC += 2
	cpu.ShouldDraw = true
	return nil
}

func (cpu *Cpu) HandleEXOpcodes() error {
	x := (cpu.Opcode & 0x0F00) >> 8
	switch cpu.Opcode & 0x00FF {
	//skip next instruction if V[x] is pressed
	case 0x9E:
		if cpu.Keypad[cpu.V[x]] != uint8(0) {
			cpu.PC += 2
		}
	//skip next instruction if V[x] is not pressed
	case 0xA1:
		if cpu.Keypad[cpu.V[x]] == uint8(0) {
			cpu.PC += 2
		}
	default:
		return errors.New(fmt.Sprintf("%d: Invalid Opcode in HandleEXOpcodes()", cpu.Opcode))
	}
	cpu.PC += 2
	return nil
}

func (cpu *Cpu) HandleFXOpcodes() error {
	x := (cpu.Opcode & 0x0F00) >> 8
	switch cpu.Opcode & 0x00FF {
	//V[x] = delay_timer
	case 0x07:
		cpu.V[x] = cpu.DelayTimer
	//V[x] = key_press() (blocking!)
	case 0x0A:
		key_pressed := false
		for i, elem := range cpu.Keypad {
			if elem != 0 {
				key_pressed = true
				cpu.V[x] = uint8(i)
			}
		}
		//keep on blocking (don't increment PC) if no keys have been pressed yet
		if !key_pressed {
			return nil
		}
	//delay_timer = V[x]
	case 0x15:
		cpu.DelayTimer = cpu.V[x]
	//sound_timer = V[x]
	case 0x18:
		cpu.SoundTimer = cpu.V[x]
	//I += V[x]
	case 0x1E:
		res := uint32(cpu.I) + uint32(cpu.V[x])
		var carry uint8 = 0
		if res > 0xFF {
			carry = 1
		}
		cpu.I = uint16(res)
		cpu.V[0xF] = carry
	//I = address of sprite for char in V[x]
	case 0x29:
		cpu.I = uint16(cpu.V[x]) * 5
	//Memory[I, I+1, I+2] stores binary-coded decimal of V[x]
	case 0x33:
		if cpu.I > MEMORY_SIZE-3 {
			return errors.New("I is too close to the memory limit for Opcode 0xFX33 to execute!")
		}
		cpu.Memory[cpu.I] = cpu.V[x] / 100
		cpu.Memory[cpu.I+1] = (cpu.V[x] % 100) / 10
		cpu.Memory[cpu.I+2] = cpu.V[x] % 10
	//Memory[I:I+x] = V[0:x] (inclusive on both ends)
	case 0x55:
		if cpu.I > MEMORY_SIZE-x {
			return errors.New("I is too close to the memory limit for Opcode 0xFX55 to execute!")
		}
		copy(cpu.Memory[cpu.I:], cpu.V[0:x+1])
	//V[0:x] = Memory[I:I+x] (inclusive on both ends)
	case 0x65:
		if cpu.I > MEMORY_SIZE-x {
			return errors.New("I is too close to the memory limit for Opcode 0xFX65 to execute!")
		}
		copy(cpu.V[0:], cpu.Memory[cpu.I:cpu.I+x+1])
	default:
		return errors.New(fmt.Sprintf("%d: Invalid Opcode in HandleFXOpcodes()", cpu.Opcode))
	}
	cpu.PC += 2
	return nil
}
