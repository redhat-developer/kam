package main

import (
	"log"

	"github.com/redhat-developer/kam/pkg/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.MakeRootCmd(), "./docs/commands")
	if err != nil {
		log.Fatal(err)
	}
}
