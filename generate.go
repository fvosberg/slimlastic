package slimlastic

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// ClientGenerator generates a new slimlastic client for a given struct (called model)
type ClientGenerator struct {
	Name              string // Name of the generated client
	Model             string // Name of the struct this client should handle
	PkgName           string // Name of the package the code is generated for
	PreventCommonCode bool

	timeout             time.Duration // Timeout for requests to elasticsearch for the generated client. Can be set with SetTimeout
	indexDefinitionPath string        // TODO to reader
}

type code struct {
	Model             string
	ModelWithPrefix   string
	LowercaseModel    string
	SourcePackage     string
	TargetPackage     string
	Imports           []string
	UppercaseClient   string
	LowercaseClient   string
	IndexName         string
	TypeName          string
	IndexDefinition   string
	WithConstructor   bool
	PreventCommonCode bool
}

// WriteTo writes the generated code to the given writer
func (g *ClientGenerator) WriteTo(w io.Writer) (int64, error) {
	clientName := g.Model + "ElasticsearchClient" // TODO configurable
	doc := code{
		Model:             g.Model,
		ModelWithPrefix:   g.Model,
		LowercaseModel:    strings.ToLower(string(g.Model[0])) + g.Model[1:],
		SourcePackage:     g.PkgName, // TODO
		TargetPackage:     g.PkgName, // TODO
		Imports:           []string{"bytes", "encoding/json", "fmt", "io", "net/http", "strings", "time", "github.com/fvosberg/errtypes", "github.com/pkg/errors"},
		UppercaseClient:   strings.ToUpper(string(clientName[0])) + clientName[1:],
		LowercaseClient:   strings.ToLower(string(clientName[0])) + clientName[1:],
		IndexName:         strings.ToLower(g.Model) + "s",
		TypeName:          strings.ToLower(g.Model),
		WithConstructor:   true,
		PreventCommonCode: g.PreventCommonCode,
	}
	if doc.SourcePackage != doc.TargetPackage {
		doc.Imports = append(doc.Imports, doc.SourcePackage)
		doc.ModelWithPrefix = fmt.Sprintf("%s.%s", doc.SourcePackage, doc.Model)
	}
	indexDef, err := ioutil.ReadFile(g.indexDefinitionPath)
	if err != nil {
		return 0, errors.Wrap(err, "reading index definition file failed")
	}
	doc.IndexDefinition = string(indexDef)
	tmpl, err := template.New("client").Parse(clientTemplate)
	if err != nil {
		return 0, errors.Wrap(err, "parsing template failed")
	}
	err = tmpl.Execute(w, doc)
	if err != nil {
		return 0, errors.Wrap(err, "executing template failed")
	}
	return 0, nil // TODO
}

// SetTimeout sets the timeout for requests to elasticsearch for the generated client
func (g *ClientGenerator) SetTimeout(d time.Duration) {
	g.timeout = d
}

// SetIndexDefinitionPath sets the path to the elasticsearch index definition JSON
func (g *ClientGenerator) SetIndexDefinitionPath(p string) {
	g.indexDefinitionPath = p
}
