package main

import "fmt"

func logInfo(text string) {
	fmt.Println(text)
}

func logError(err error) {
	fmt.Println(err)
}

func logErrorString(err string) {
	fmt.Println(err)
}
