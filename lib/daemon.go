package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var (
	ErrNotRunning    = errors.New("Process is not running")
	ErrUnableToParse = errors.New("Unable to read and parse process id")
	ErrUnableToKill  = errors.New("Unable to kill process")
)

func getPidFilePath() string {
	return os.Getenv("HOME") + "/daemon.pid"
}

func savePID(pid int) {
	file, err := os.Create(getPidFilePath())
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))

	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk
}

func runDaemon() error {
	// check if daemon already running.
	if _, err := os.Stat(getPidFilePath()); err == nil {
		fmt.Println("Already running or /tmp/daemonize.pid file exist.")
		os.Exit(1)
		return nil
	}

	cmd := exec.Command(os.Args[0], getArgsD(os.Args[1:])...)
	cmd.Start()
	fmt.Println("Daemon process ID is : ", cmd.Process.Pid)
	savePID(cmd.Process.Pid)
	os.Exit(0)

	return nil
}

func getArgsD(osArgs []string) []string {
	for i, n := range osArgs {
		if n == "-d" {
			a := append(osArgs[:i], osArgs[i+1:]...)
			return a
		}
	}
	return osArgs
}
