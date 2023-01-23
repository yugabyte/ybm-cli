package cmd

import (
	"fmt"

	"github.com/hokaccha/go-prettyjson"
)

func prettyPrintJson(data interface{}) {
	//b, _ := json.MarshalIndent(data, "", "  ")
	b, _ := prettyjson.Marshal(data)
	fmt.Println(string(b))
}
