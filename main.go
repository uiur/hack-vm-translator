package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

const setupCode = `
@256
D=A
@SP
M=D
`

func main() {
	if len(os.Args) > 1 {
		filename := os.Args[1]
		data, err := ioutil.ReadFile(filename)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		code := trimSpace(string(data))

		result := setupCode

		for _, line := range strings.Split(code, "\n") {
			tokens := strings.Split(line, " ")
			operator := tokens[0]

			switch operator {
			case "push":
				segment := tokens[1]
				if segment == "constant" {
					value := tokens[2]
					result += push(value)
				}
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
			}
		}

		fmt.Print(result)
	}
}

func push(value string) string {
	return fmt.Sprintf(`
@%s
D=A
@SP
A=M
M=D
@SP
M=M+1
`, value)
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
