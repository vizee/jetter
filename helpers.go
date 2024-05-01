package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/CloudyKit/jet/v6"
)

func quote(args jet.Arguments) reflect.Value {
	return reflect.ValueOf(strconv.Quote(args.Get(0).String()))
}

func file(root string) jet.Func {
	return func(args jet.Arguments) reflect.Value {
		fname := args.Get(0).String()
		data, err := os.ReadFile(filepath.Join(root, fname))
		if err != nil {
			fmt.Fprintf(os.Stderr, "file: read %s error %v\n", fname, err)
			return reflect.ValueOf("")
		}
		return reflect.ValueOf(string(data))
	}
}

func eval(set *jet.Set) jet.Func {
	return func(args jet.Arguments) reflect.Value {
		tmpl, err := set.Parse("eval", args.Get(0).String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "eval: parse template: %v\n", err)
			return reflect.ValueOf("")
		}
		var data any
		if args.IsSet(1) {
			data = args.Get(1).Interface()
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, globalVars, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "eval: execute template: %v\n", err)
			return reflect.ValueOf("")
		}
		return reflect.ValueOf(buf.String())
	}
}

func command(unsafe bool) jet.Func {
	return func(args jet.Arguments) reflect.Value {
		if !unsafe {
			fmt.Fprintf(os.Stderr, "command: need to set unsafe\n")
			return reflect.ValueOf("")
		}
		nargs := args.NumOfArguments()
		if nargs == 0 {
			return reflect.ValueOf("")
		}
		name := args.Get(0).String()
		cmdArgs := make([]string, 0, nargs)
		for i := 1; i < nargs; i++ {
			cmdArgs = append(cmdArgs, args.Get(i).String())
		}
		cmd := exec.Command(name, cmdArgs...)
		var cmdOut strings.Builder
		cmd.Stdout = &cmdOut
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "command: run error %v\n", err)
			return reflect.ValueOf("")
		}

		return reflect.ValueOf(cmdOut.String())
	}
}
