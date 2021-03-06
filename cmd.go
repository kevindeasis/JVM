package main

import "flag"
import "fmt"
import "os"

type Cmd struct {
	helpFlag         bool
	versionFlag      bool
	verboseClassFlag bool
	verboseInstFlag  bool
	cpOption         string
	class            string
	XjreOption       string
	args             []string
}

func parseCmd() *Cmd {
	cmd := &Cmd{}

	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help messages")
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help messages")
	flag.BoolVar(&cmd.versionFlag, "version", false, "print version messages")
	flag.BoolVar(&cmd.verboseClassFlag, "logKls", false, "print version messages")
	flag.BoolVar(&cmd.verboseInstFlag, "logInst", false, "print version messages")
	flag.StringVar(&cmd.cpOption, "classpath", "", "classpath")
	flag.StringVar(&cmd.cpOption, "cp", "", "classpath")
	flag.StringVar(&cmd.XjreOption, "Xjre", "", "path to jre")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		cmd.class = args[0]
		cmd.args = args[1:]
	}

	return cmd
}

func printUsage() {
	fmt.Printf("Usage: %s [-options] class [args...]\n", os.Args[0])
}
