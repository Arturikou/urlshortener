package nopanic

import "fmt"

func bad() {
	panic("something went wrong") // want `use of built-in panic`
}

func good() {
	fmt.Println("all good")
}
