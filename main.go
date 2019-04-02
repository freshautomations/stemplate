package main

import (
	"github.com/freshautomations/stemplate/cmd"
	"github.com/freshautomations/stemplate/exit"
)

func main() {
	if err := cmd.Execute(); err != nil {
		exit.Fail(err)
	}
}
