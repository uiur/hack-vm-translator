package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		filename := os.Args[1]
		data, err := ioutil.ReadFile(filename)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		code := trimSpace(string(data))

		result := ""

		for _, line := range strings.Split(code, "\n") {
			tokens := strings.Split(line, " ")
			operator := tokens[0]

			switch operator {
			case "push":
				segment := tokens[1]
				index := tokens[2]
				result += push(segment, index)
			case "pop":
				segment := tokens[1]
				index := tokens[2]
				result += pop(segment, index)
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
				result += fmt.Sprintf(`
@%s
0;JMP
`, labelName)
			case "if-goto":
				labelName := tokens[1]
				result += fmt.Sprintf(`
@SP
M=M-1

@SP
A=M
D=M

@%s
D;JGT
`, labelName)
			case "function":
				name := tokens[1]
				argNum, _ := strconv.Atoi(tokens[2])

				result += defineFunction(name, argNum)
			}
		}

		fmt.Print(result)
	}
}

var segmentToSymbol = map[string]string{
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
	"pointer":  "THIS",
	"temp":     "R5",
}

func push(segment, index string) string {
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
@a.%s
D=M
`, index)
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

func pop(segment, index string) string {
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
@a.%s
D=A
`, index)
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
		result += push("constant", "0")
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

func trimSpace(code string) string {
	var result string

	for _, line := range strings.Split(code, "\n") {
		trimmedLine := removeComment(strings.TrimSpace(line))
		if len(trimmedLine) > 0 {
			result += trimmedLine + "\n"
		}
	}

	return result
}

func removeComment(str string) string {
	return regexp.MustCompile(`\s*//.+$`).ReplaceAllString(str, "")
}
