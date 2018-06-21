package slimlastic

import (
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
	TypeName            string        // Name of the elasticsearch document type, default to lowercase model
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
	model := g.Model
	modelWithPrefix := model
	sourcePackage := g.PkgName
	if i := strings.LastIndex(model, "."); i > -1 {
		model = g.Model[i+1:]
		sourcePackage = g.Model[:i]
		if i := strings.LastIndex(sourcePackage, "/"); i > -1 {
			modelWithPrefix = sourcePackage[i+1:] + "." + model
		}
	}
	clientName := model + "ElasticsearchClient" // TODO configurable
	typeName := g.TypeName
	if typeName == "" {
		typeName = strings.ToLower(model)
	}

	doc := code{
		Model:             model,
		ModelWithPrefix:   modelWithPrefix,
		LowercaseModel:    strings.ToLower(string(model[0])) + model[1:],
		SourcePackage:     sourcePackage,
		TargetPackage:     g.PkgName,
		Imports:           []string{"bytes", "encoding/json", "fmt", "io", "net/http", "strings", "time", "github.com/fvosberg/errtypes", "github.com/pkg/errors"},
		UppercaseClient:   strings.ToUpper(string(clientName[0])) + clientName[1:],
		LowercaseClient:   strings.ToLower(string(clientName[0])) + clientName[1:],
		IndexName:         typeName + "s",
		TypeName:          typeName,
		WithConstructor:   true,
		PreventCommonCode: g.PreventCommonCode,
	}
	if doc.SourcePackage != doc.TargetPackage {
		doc.Imports = append(doc.Imports, doc.SourcePackage)
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
