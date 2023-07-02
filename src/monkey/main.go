package main

import (
	"fmt"
	"os"
	"os/user"
	"monkey/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! Ready to monkey around!\n", user.Username)
	fmt.Printf("Plz type your commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
