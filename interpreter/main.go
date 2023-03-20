package main

import (
	"fmt"
	"os"
	"os/user"
	"waixg/interpreter/repl"
)

func main() {
	osUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! Welcome to the playground!\n\n", osUser.Name)
	repl.Start(os.Stdin, os.Stdout)
}
