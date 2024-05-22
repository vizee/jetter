package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
			files = append(files, path[len(dir)+1:])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func printYaml(w io.Writer, obj any) {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	enc.SetIndent(2)
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		tmplRoot    string
		assigns     []string
		valuesFile  string
		output      string
		sep         string
		safeWriter  string
		ext         string
		unsafe      bool
		debugValues bool
	)

	appCmd := &cobra.Command{
		Use:               "jetter [flags] name [file...]",
		Short:             "Jet Templates Renderer",
		DisableAutoGenTag: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !debugValues {
				return fmt.Errorf("Specify name")
			}

			values, err := loadValues(valuesFile, assigns)
			if err != nil {
				return err
			}

			if debugValues {
				printYaml(os.Stderr, values)
			}

			if len(args) == 0 {
				return nil
			}
			setName := args[0]
			if strings.Contains(setName, "..") {
				return fmt.Errorf("Invalid name: %s", setName)
			}

			tmplDir := filepath.Join(tmplRoot, setName)

			var jets []string
			if len(args) > 1 {
				jets = args[1:]
			} else {
				files, err := collectJetFiles(tmplDir)
				if err != nil {
					return err
				}
				if len(files) == 0 {
					return fmt.Errorf("No jet files found in %s", tmplDir)
				}
				jets = files
			}

			var opts []jet.Option
			switch safeWriter {
			case "html":
				opts = append(opts, jet.WithSafeWriter(template.HTMLEscape))
			case "js":
				opts = append(opts, jet.WithSafeWriter(template.JSEscape))
			default:
				opts = append(opts, jet.WithSafeWriter(nil))
			}
			jetsSet := jet.NewSet(jet.NewOSFileSystemLoader(tmplDir), opts...)
			jetsSet.AddGlobalFunc("quote", quote)
			jetsSet.AddGlobalFunc("file", file(tmplRoot))
			jetsSet.AddGlobalFunc("eval", eval(jetsSet))
			jetsSet.AddGlobalFunc("command", command(unsafe))
			jetsSet.AddGlobalFunc("loadcsv", loadcsv(tmplRoot))

			wr, err := newWriter(output, sep, ext)
			if err != nil {
				return err
			}
			defer wr.Close()

			return renderJets(jetsSet, jets, values, wr)
		},
	}
	appCmd.Flags().StringVarP(&tmplRoot, "dir", "d", ".", "templates directory")
	appCmd.Flags().StringArrayVar(&assigns, "set", nil, "set value")
	appCmd.Flags().StringVarP(&valuesFile, "values", "v", "./values.yaml", "values file")
	appCmd.Flags().StringVarP(&output, "output", "o", "-", "output")
	appCmd.Flags().StringVar(&sep, "sep", "", "separator")
	appCmd.Flags().StringVar(&safeWriter, "safe-writer", "", "html/js")
	appCmd.Flags().StringVarP(&ext, "extension", "e", "", "rename extension")
	appCmd.Flags().BoolVar(&unsafe, "unsafe", false, "unsafe function")
	appCmd.Flags().BoolVar(&debugValues, "debug-values", false, "debug values")
	err := appCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
