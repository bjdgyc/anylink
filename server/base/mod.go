package base

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	procModulesPath = "/proc/modules"
	inContainerKey  = "ANYLINK_IN_CONTAINER"
)

var (
	inContainer = false
	modMap      = map[string]struct{}{}
)

func initMod() {
	container := os.Getenv(inContainerKey)
	if container == "true" {
		inContainer = true
	}
	log.Println("inContainer", inContainer)

	file, err := os.Open(procModulesPath)
	if err != nil {
		err = fmt.Errorf("[ERROR] Problem with open file: %s", err)
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		splited := strings.Split(scanner.Text(), " ")
		if len(splited[0]) > 0 {
			modMap[splited[0]] = struct{}{}
		}
	}
}

func CheckModOrLoad(mod string) {
	log.Println("CheckModOrLoad", mod)

	if _, ok := modMap[mod]; ok {
		return
	}

	if inContainer {
		err := fmt.Errorf("Linux modules %s is not loaded, please run `modprobe %s`", mod, mod)
		panic(err)
	}

	cmdstr := fmt.Sprintln("modprobe", mod)

	cmd := exec.Command("sh", "-c", cmdstr)
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(b))
		panic(err)
	}
}
