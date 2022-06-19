package main

import (
	"fmt"
	"os"
)

func main() {
	_ = 1
	Exit := "Exit"
	fmt.Println(Exit)
	os.Exit(1) // want `there is os exit in main`
}
