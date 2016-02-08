package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/uiureo/hack-vm-translator/generator"
)

func main() {
	if len(os.Args) > 1 {
		result := generator.BootstrapCode()
		for _, filename := range os.Args[1:] {
			data, err := ioutil.ReadFile(filename)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			code := trimSpace(string(data))
			result += generator.GenerateCode(code, filename)
		}
		fmt.Print(result)
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
