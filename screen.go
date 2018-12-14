package main

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

func println(message string) {
	fmt.Println(message)
}

func printErrorln(message string) {
	fmt.Println(aurora.Red(message))
}
