package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"kpless"
)

func main() {
	kp := kpless.New()
	err := kp.LoadMarkDown("向火独行.md")
	if err != nil {
		panic(err)
	}
	_ = kp.SetGame("cli", "向火独行.md")
	inputs := bufio.NewScanner(os.Stdin)
	for inputs.Scan() {
		res, err := kp.Input("cli", inputs.Text())
		if errors.Is(err, io.EOF) {
			fmt.Println(res)
			os.Exit(1)
		}
		fmt.Println("-------------")
		if err != nil {
			fmt.Println(err.Error(), res)
		}
		fmt.Println(res)
	}
}
