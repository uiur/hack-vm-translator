package generator

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// BootstrapCode returns code to setup vm
func BootstrapCode() string {
	bootstrap := os.Getenv("BOOTSTRAP") != "false"
	result := `
@256
D=A
@SP
M=D
`

	if bootstrap {
		result += call("Sys.init", "0")
	}

	return result
}

// GenerateCode returns hack asm code
func GenerateCode(source, filename string) string {
	result := ""

	for _, line := range strings.Split(source, "\n") {
		tokens := strings.Split(line, " ")
		operator := tokens[0]

		switch operator {
		case "push":
			segment := tokens[1]
			index := tokens[2]
			name := strings.Split(filepath.Base(filename), ".")[0]
			result += push(segment, index, name)
		case "pop":
			segment := tokens[1]
			index := tokens[2]
			name := strings.Split(filepath.Base(filename), ".")[0]
			result += pop(segment, index, name)
		case "add":
			result += binop("+")
		case "sub":
			result += binop("-")
		case "neg":
			result += uniop("-")
		case "eq":
			result += compare("EQ")
		case "gt":
			result += compare("GT")
		case "lt":
			result += compare("LT")
		case "and":
			result += binop("&")
		case "or":
			result += binop("|")
		case "not":
			result += uniop("!")
		case "label":
			name := tokens[1]
			result += "\n(" + name + ")\n"
		case "goto":
			labelName := tokens[1]
			result += gotoCode(labelName)
		case "if-goto":
			labelName := tokens[1]
			result += fmt.Sprintf(`
@SP
M=M-1

@SP
A=M
D=M

@%s
D;JNE
`, labelName)
		case "function":
			name := tokens[1]
			localVarNum, _ := strconv.Atoi(tokens[2])

			result += defineFunction(name, localVarNum)

		case "call":
			funcName := tokens[1]
			arg := tokens[2]
			result += call(funcName, arg)

		case "return":
			result += returnCode()
		}
	}

	return result
}

func pushSymbolAddress(symbol string) string {
	return fmt.Sprintf(`
@%s
D=M

@SP
A=M
M=D

@SP
M=M+1
`, symbol)
}

func gotoCode(name string) string {
	return fmt.Sprintf(`
@%s
0;JMP
`, name)
}

func returnCode() string {
	return `
@LCL
D=M
@R13
M=D

@5
D=D-A
A=D
D=M
@R14
M=D

@SP
M=M-1
A=M
D=M
@ARG
A=M
M=D

@ARG
D=M
D=D+1
@SP
M=D

@R13
D=M
@1
A=D-A
D=M
@THAT
M=D

@R13
D=M
@2
A=D-A
D=M
@THIS
M=D

@R13
D=M
@3
A=D-A
D=M
@ARG
M=D

@R13
D=M
@4
A=D-A
D=M
@LCL
M=D

@R14
A=M
0;JMP
`
}

func call(funcName, arg string) string {
	result := ""

	returnAddressLabel := "return-address." + strconv.Itoa(randNum())

	result += fmt.Sprintf(`
@%s
D=A

@SP
A=M
M=D

@SP
M=M+1
`, returnAddressLabel)

	symbols := []string{"LCL", "ARG", "THIS", "THAT"}
	for _, symbol := range symbols {
		result += pushSymbolAddress(symbol)
	}

	result += fmt.Sprintf(`
@SP
D=M
@5
D=D-A
@%s
D=D-A
@ARG
M=D
`, arg)

	result += `
@SP
D=M
@LCL
M=D
`
	result += gotoCode(funcName)
	result += "\n(" + returnAddressLabel + ")\n"

	return result
}

var segmentToSymbol = map[string]string{
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
	"pointer":  "THIS",
	"temp":     "R5",
}

func push(segment, index, filename string) string {
	var addressCode string
	switch segment {
	case "constant":
		addressCode = fmt.Sprintf(`
@%s
D=A
`, index)
	case "pointer", "temp":
		symbol := segmentToSymbol[segment]
		addressCode = fmt.Sprintf(`
@%s
D=A
@%s
A=D+A
D=M
`, index, symbol)
	case "static":
		addressCode = fmt.Sprintf(`
@%s.%s
D=M
`, filename, index)
	default:
		symbol := segmentToSymbol[segment]
		addressCode = fmt.Sprintf(`
@%s
D=A
@%s
A=D+M
D=M
`, index, symbol)
	}

	return addressCode + `
@SP
A=M
M=D

@SP
M=M+1
`
}

func pop(segment, index, filename string) string {
	var addressCode string
	symbol := segmentToSymbol[segment]

	switch segment {
	case "pointer", "temp":
		addressCode = fmt.Sprintf(`
@%s
D=A
@%s
D=A+D
`, index, symbol)
	case "static":
		addressCode = fmt.Sprintf(`
@%s.%s
D=A
`, filename, index)
	default:
		addressCode = fmt.Sprintf(`
@%s
D=A
@%s
D=M+D
`, index, symbol)
	}

	return addressCode + `
@R13
M=D

@SP
M=M-1

@SP
A=M
D=M

@R13
A=M
M=D
`
}

func defineFunction(name string, localVarNum int) string {
	result := fmt.Sprintf(`
(%s)
@SP
D=M
@LCL
M=D
`, name)

	for i := 0; i < localVarNum; i++ {
		result += push("constant", "0", "")
	}

	return result
}

func loadArgs() string {
	return `
@SP
M=M-1
@SP
A=M
D=M

@SP
M=M-1
@SP
A=M
`

}

func binop(operator string) string {
	return loadArgs() + fmt.Sprintf(`
M=M%sD
@SP
M=M+1
`, operator)
}

func uniop(operator string) string {
	return fmt.Sprintf(`
@SP
M=M-1
@SP
A=M

M=%sM
@SP
M=M+1
`, operator)
}

func randNum() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(1 * 1000 * 1000 * 1000)
}

func compare(operator string) string {
	label := fmt.Sprintf("IF%d", randNum())
	return loadArgs() + fmt.Sprintf(`
D=M-D
@SP
A=M
M=-1
@%s
D;J%s
@SP
A=M
M=0
(%s)
@SP
M=M+1
`, label, operator, label)
}
