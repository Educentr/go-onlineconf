package main

import (
	"context"
	"fmt"

	"gitlab.educentr.info/godev/onlinecof-test/pkg/onlineconf"
)

func main() {
	ctx, err := onlineconf.Initialize(context.Background())
	if err != nil {
		fmt.Printf("Error initialize onlineconf: %s", err)
		return
	}

	v, ex, err := onlineconf.GetStringIfExists(ctx, "/testapp/bla")
	if err != nil {
		fmt.Printf("Error while geting param: %s\n", err)
		return
	}

	if !ex {
		fmt.Printf("String does not exists\n")
		return
	}

	fmt.Printf("Value %+v\n", v)
}
