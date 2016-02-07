package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const emulatorPath = "/Users/zat/Downloads/nand2tetris/tools/CPUEmulator.sh"

func TestMain(t *testing.T) {
	matches, err := filepath.Glob("fixtures/*.vm")

	if err != nil {
		t.Error(err)
	}

	for _, input := range matches {
		name := strings.Split(filepath.Base(input), ".")[0]
		asmCode, err := exec.Command("go", "run", "main.go", input).Output()

		if err != nil {
			t.Error(err)
		}

		ioutil.WriteFile("fixtures/"+name+".asm", asmCode, os.ModePerm)

		out, err := exec.Command(emulatorPath, "fixtures/"+name+".tst").Output()

		if err != nil {
			t.Error(name+": ", err, string(out))
		}
	}
}
