package main

import "flag"
import "fmt"

func main() {
	var prefix string
	flag.StringVar(&prefix, "p", "", "prefix for package map")
	flag.Parse()

	if prefix == "" {
		fmt.Println("need to provide prefix")
		return
	}

	ParseCompile(prefix)
}
