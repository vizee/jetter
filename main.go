package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/CloudyKit/jet/v6"
	"github.com/spf13/pflag"
)

var globalVars = func() jet.VarMap {
	envList := os.Environ()
	envs := make(map[string]any, len(envList))
	for _, env := range envList {
		segs := strings.SplitN(env, "=", 2)
		envs[segs[0]] = segs[1]
	}

	return jet.VarMap{
		"global": reflect.ValueOf(map[string]any{}),
		"env":    reflect.ValueOf(envs),
	}
}()

func renderJets(jetsSet *jet.Set, jets []string, values any, wr *Writer) error {
	for _, jet := range jets {
		err := wr.SwtichFile(jet)
		if err != nil {
			return err
		}

		tmpl, err := jetsSet.GetTemplate(jet)
		if err != nil {
			return err
		}
		err = tmpl.Execute(wr, globalVars, values)
		if err != nil {
			return err
		}
	}

	return nil
}

func collectJetFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jet") {
			if strings.HasPrefix(path, dir) {
				files = append(files, path[len(dir)+1:])
			} else if dir == "." {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func searchJets(rootDir string, files []string) ([]string, error) {
	fi, err := os.Stat(rootDir)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		if len(files) > 0 {
			return files, nil
		}
		files, err := collectJetFiles(rootDir)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no jet files found in %s", rootDir)
		}
		return files, nil
	}
	return nil, fmt.Errorf("template root(%s) is not a directory", rootDir)
}

type appFlags struct {
	tmplBase    string
	assigns     []string
	valuesFile  string
	output      string
	sep         string
	safeWriter  string
	ext         string
	unsafe      bool
	debugValues bool
}

func runJetter(flags *appFlags, args []string) error {
	values, err := loadValues(flags.valuesFile, flags.assigns)
	if err != nil {
		return err
	}

	if flags.debugValues {
		printValues(os.Stderr, values)
	}

	if len(args) == 0 {
		return nil
	}
	setName := args[0]
	if strings.Contains(setName, "..") {
		return fmt.Errorf("Invalid name: %s", setName)
	}

	tmplRoot := filepath.Join(flags.tmplBase, setName)
	jets, err := searchJets(tmplRoot, args[1:])
	if err != nil {
		return err
	}

	var opts []jet.Option
	switch flags.safeWriter {
	case "html":
		opts = append(opts, jet.WithSafeWriter(template.HTMLEscape))
	case "js":
		opts = append(opts, jet.WithSafeWriter(template.JSEscape))
	default:
		opts = append(opts, jet.WithSafeWriter(nil))
	}
	jetsSet := jet.NewSet(jet.NewOSFileSystemLoader(tmplRoot), opts...)
	jetsSet.AddGlobalFunc("quote", quote)
	jetsSet.AddGlobalFunc("file", file(flags.tmplBase))
	jetsSet.AddGlobalFunc("eval", eval(jetsSet))
	jetsSet.AddGlobalFunc("command", command(flags.unsafe))
	jetsSet.AddGlobalFunc("loadcsv", loadcsv(flags.tmplBase))

	wr, err := newWriter(flags.output, flags.sep, flags.ext)
	if err != nil {
		return err
	}
	defer wr.Close()

	return renderJets(jetsSet, jets, values, wr)
}

func main() {
	var (
		flags appFlags
	)
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  jetter [flags] base-dir [files...]\n\nFlags:\n")
		pflag.PrintDefaults()
	}
	pflag.StringVarP(&flags.tmplBase, "dir", "d", ".", "templates base directory")
	pflag.StringArrayVar(&flags.assigns, "set", nil, "set value")
	pflag.StringVarP(&flags.valuesFile, "values", "v", "", "values file")
	pflag.StringVarP(&flags.output, "output", "o", "-", "output")
	pflag.StringVar(&flags.sep, "sep", "", "separator")
	pflag.StringVar(&flags.safeWriter, "safe-writer", "", "html/js")
	pflag.StringVarP(&flags.ext, "extension", "e", "", "rename extension")
	pflag.BoolVar(&flags.unsafe, "unsafe", false, "unsafe function")
	pflag.BoolVar(&flags.debugValues, "debug-values", false, "debug values")
	pflag.Parse()

	if pflag.NArg() == 0 && !flags.debugValues {
		pflag.Usage()
		os.Exit(1)
	}

	err := runJetter(&flags, pflag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
