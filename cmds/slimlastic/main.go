package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/fvosberg/slimlastic"
)

func main() {
	var (
		outFile           = flag.String("out", "", "output file (default stdout)")
		pkgName           = flag.String("pkg", "", "package name (default will infer)")
		client            = flag.String("client", "", "client name (default modelElasticsearchClient)")
		httpTimeout       = flag.Int("timeout", 1, "timout for requests to elasticsearch")
		indexDefinition   = flag.String("indexDefinition", "", "path to the elasticsearch index definition")
		preventCommonCode = flag.Bool("preventCommon", false, "prevent the generation of common code") // TODO parse the package
		typeName          = flag.String("typeName", "", "custom name for the elasticsearch document type")
	)
	flag.Usage = func() {
		fmt.Println(`slimlastic [flags] model [indexDefinition]`)
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Not enough arguments")
		flag.Usage()
		os.Exit(1)
	}
	model := args[0]
	if *indexDefinition == "" {
		fmt.Fprintln(os.Stderr, "Path to elasticsearch index definition not set")
		flag.Usage()
		os.Exit(1)
	}
	var buf bytes.Buffer
	var out io.Writer
	out = os.Stdout
	if len(*outFile) > 0 {
		out = &buf
	}
	generator := slimlastic.ClientGenerator{
		Name:              *client,
		Model:             model,
		PkgName:           *pkgName,
		PreventCommonCode: *preventCommonCode,
		TypeName:          *typeName,
	}
	if *httpTimeout != 0 {
		generator.SetTimeout(time.Duration(*httpTimeout))
	}
	generator.SetIndexDefinitionPath(*indexDefinition)
	_, err := generator.WriteTo(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation of code failed: %s\n", err)
		os.Exit(1)
	}
	// create the file
	if len(*outFile) > 0 {
		err = ioutil.WriteFile(*outFile, buf.Bytes(), 0777)
	}
}
