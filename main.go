package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const setupCode = `@256
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

		fmt.Print(setupCode)
		for _, line := range strings.Split(code, "\n") {
			tokens := strings.Split(line, " ")
			operator := tokens[0]

			if operator == "push" {
				segment := tokens[1]
				if segment == "constant" {
					value := tokens[2]
					fmt.Printf(`@%s
D=A
@SP
A=M
M=D
@SP
M=M+1
`, value)
				}
			} else if operator == "add" {
				fmt.Print(`
@SP
M=M-1
@SP
A=M
D=M
@SP
M=M-1
@SP
A=M
@SP
A=M
M=M+D
@SP
M=M+1
`)
			}
		}
	}
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
