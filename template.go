package slimlastic

var clientTemplate = `// Code generated by slimlastic DO NOT EDIT.
//github.com/fvosberg/slimlastic

package {{.TargetPackage}}

import (
{{- range .Imports }}
	"{{.}}"
{{- end }}
)

{{- if .WithConstructor }}
// New{{.UppercaseClient}} instantiates a new elasticsearch client
// which is dedicated to the struct {{.SourcePackage}}.{{.Model}}
func new{{.UppercaseClient}}(url string) (*{{.LowercaseClient}}, error) {
	c := &{{.LowercaseClient}}{}
	c.Init(url)
	err := c.EnsureExistingIndex()
	if err != nil {
		return nil, err
	}
	return c, nil
}
{{- end}}

func (c *{{.LowercaseClient}}) Init(url string) {
	url = strings.TrimRight(url, "/")
	c.http = &http.Client{Timeout: 5 * time.Second}
	c.indexURL = fmt.Sprintf("%s/{{.IndexName}}", url)
	c.typeURL =  fmt.Sprintf("%s/{{.IndexName}}/{{.TypeName}}", url)
}

func (c *{{.LowercaseClient}}) EnsureExistingIndex() error {
	indexExists, err := c.IndexExists()
	if err != nil {
		return err
	}
	if indexExists {
		return nil
	}
	return c.CreateIndex()
}

type {{.LowercaseClient}} struct {
	http     *http.Client
	indexURL string
	typeURL  string
}

func (c *{{.LowercaseClient}}) Refresh() error {
	var result struct {
		Shards struct {
			Total      int ` + "`" + `json:"total"` + "`" + `
			Successful int ` + "`" + `json:"successful"` + "`" + `
			Failed     int ` + "`" + `json:"failed"` + "`" + `
		} ` + "`" + `json:"_shards"` + "`" + `
	}
	err := c.doRequest("POST", fmt.Sprintf("%s/_refresh", c.indexURL), nil, &result)
	if err != nil {
		return err
	}
	if result.Shards.Failed != 0 {
		return fmt.Errorf("Refreshing of %d shards failed (%d successful; %d in total)", result.Shards.Failed, result.Shards.Successful, result.Shards.Total)
	}
	return nil
}

func (c *{{.LowercaseClient}}) GetOneByID(ID string) (*{{.ModelWithPrefix}}, error) {
	var response struct {
		ID     string      ` + "`" + `json:"_id"` + "`" + `
		Source {{.ModelWithPrefix}} ` + "`" + `json:"_source"` + "`" + `
		Found  bool        ` + "`" + `json:"found"` + "`" + `
	}
	err := c.doRequest("GET", fmt.Sprintf("%s/%s", c.typeURL, ID), nil, &response)
	if err != nil {
		return nil, err
	}
	if !response.Found {
		return nil, errtypes.NewNotFoundf("{{.ModelWithPrefix}} with id %d not found", ID)
	}
	response.Source.ID = response.ID
	return &response.Source, nil
}

func {{.LowercaseModel}}FromElasticsearchHit(hit {{.LowercaseClient}}Hit) {{.ModelWithPrefix}} {
	hit.Source.ID = hit.ID
	return hit.Source
}

func {{.LowercaseModel}}sFromElasticsearchHits(hits []{{.LowercaseClient}}Hit) []{{.ModelWithPrefix}} {
	res := make([]{{.ModelWithPrefix}}, len(hits))
	for n, h := range hits {
		res[n] = {{.LowercaseModel}}FromElasticsearchHit(h)
	}
	return res
}

func (c *{{.LowercaseClient}}) GetList(offset, limit int) ([]{{.ModelWithPrefix}}, error) {
	return c.DoListRequest(strings.NewReader(fmt.Sprintf(` + "`" + `{"from":%d,"size":%d}` + "`" + `, offset, limit)))
}

func (c *{{.LowercaseClient}}) DoListRequest(body io.Reader, opts ...{{.LowercaseClient}}ListRequestOpt) ([]{{.ModelWithPrefix}}, error) {
	var cfg {{.LowercaseClient}}ListRequestOptions
	for _, o := range opts {
		o(&cfg)
	}
	var result {{.LowercaseClient}}Hits
	err := c.doRequest("GET", fmt.Sprintf("%s/_search", c.typeURL), body, &result)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("Error in elasticsearch: %s, caused by %#v", result.Error.Reason, result.Error.CausedBy)
	}
	if cfg.total != nil {
		*cfg.total = uint32(result.Hits.Total)
	}
	return {{.LowercaseModel}}sFromElasticsearchHits(result.Hits.Hits), nil
}

type {{.LowercaseClient}}ListRequestOptions struct {
	total *uint32
}

type {{.LowercaseClient}}ListRequestOpt func(*{{.LowercaseClient}}ListRequestOptions)

func {{.LowercaseClient}}WithTotal(t *uint32) {{.LowercaseClient}}ListRequestOpt {
	return func(o *{{.LowercaseClient}}ListRequestOptions) {
		o.total = t
	}
}


// Index creates a new {{.ModelWithPrefix}} in elasticsearch
// When the ID of the {{.Model}} is set, it updates the {{.Model}}
// The first return value indicates, whether a new records has been created or not
func (c *{{.LowercaseClient}}) Index(m *{{.ModelWithPrefix}}, opts ...{{.LowercaseModel}}ElasticsearchIndexOption) (bool, error) {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(m)
	if err != nil {
		return false, err
	}
	cfg := {{.LowercaseModel}}ElasticsearchIndexConfig{Refresh: "false"}
	for _, o := range opts {
		o(&cfg)
	}
	var response {{.LowercaseClient}}DocResponse
	err = c.doRequest("POST", fmt.Sprintf("%s/%s?refresh=%s", c.typeURL, m.ID, cfg.Refresh), body, &response)
	if err != nil {
		return false, err
	}
	if response.Error != nil {
		return false,fmt.Errorf("%#v", response.Error)
	}
	if response.ID == "" || (response.Result != "updated" && response.Result != "created") {
		// if this case happens, please report with furhter information to hello@frederikvosberg.de to implement a better error handling
		return false, errors.New("indexing of document in elasticsearch failed")
	}
	m.ID = response.ID
	return response.Result == "created", nil
}

type {{.LowercaseModel}}ElasticsearchIndexOption func(*{{.LowercaseModel}}ElasticsearchIndexConfig)

type {{.LowercaseModel}}ElasticsearchIndexConfig struct {
	Refresh string
}

// ForceReceiptIndexRefresh forces the immediate refresh after indexing
// it's used as an {{.LowercaseModel}}ElasticsearchIndexOption param to {{.LowercaseClient}}.Index
func Force{{.Model}}IndexRefresh(cfg *{{.LowercaseModel}}ElasticsearchIndexConfig) {
	cfg.Refresh = "true"
}

// DeleteOneByID deletes a {{.ModelWithPrefix}} in elasticsearch, given its ID
func (c *{{.LowercaseClient}}) DeleteOneByID(id string) error {
	var response {{.LowercaseClient}}DocResponse
	err := c.doRequest("DELETE", fmt.Sprintf("%s/%s", c.typeURL, id), nil, &response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return fmt.Errorf("%#v", response.Error)
	}
	if response.Result != "deleted" {
		// if this case happens, please report with furhter information to hello@frederikvosberg.de to implement a better error handling
		return fmt.Errorf("deletion of document in elasticsearch failed, result was %q", response.Result)
	}
	return nil
}

func (c *{{.LowercaseClient}}) RecreateIndex() error {
	_, err := c.DeleteIndex()
	if err != nil {
		return err
	}
	return c.CreateIndex()
}

func (c *{{.LowercaseClient}}) IndexExists() (bool, error) {
	req, err := c.newRequest("HEAD", c.indexURL, nil)
	if err != nil {
		return false, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	res.Body.Close()
	if res.StatusCode != 200 && res.StatusCode != 404 {
		return false, fmt.Errorf("status code should be 200 or 404 on checking existence of \"{{.ModelWithPrefix}}\" index, but was %d", res.StatusCode)
	}
	return res.StatusCode == 200, nil
}

func (c *{{.LowercaseClient}}) DeleteIndex() (bool, error) {
	var response {{.LowercaseClient}}IndexManipulationResponse
	err := c.doRequest("DELETE", c.indexURL, nil, response)
	if err != nil {
		return false, err
	}
	return response.Acknowledged, nil
}

func (c *{{.LowercaseClient}}) CreateIndex() error {
	var response {{.LowercaseClient}}IndexManipulationResponse
	err := c.doRequest("PUT", c.indexURL, strings.NewReader({{.LowercaseClient}}IndexDefinition), &response)
	if err != nil {
		return err
	}
	if !response.Acknowledged {
		return fmt.Errorf("creation of index not acknowledged: %#v", response.Error)
	}
	return nil
}

func (c *{{.LowercaseClient}}) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req, nil
}

func (c *{{.LowercaseClient}}) doRequest(method, url string, body io.Reader, response interface{}) error {
	req, err := c.newRequest(method, url, body)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return errors.Wrap(err, "couldn't decode JSON response")
	}
	return nil
}

var {{.LowercaseClient}}IndexDefinition = ` + "`{{.IndexDefinition}}`" + `

type {{.LowercaseClient}}IndexManipulationResponse struct {
	Acknowledged bool         ` + "`" + `json:"acknowledged"` + "`" + `
	Status       int          ` + "`" + `json:"status"` + "`" + `
	Error        elasticError ` + "`" + `json:"error"` + "`" + `
}

type {{.LowercaseClient}}Error struct {
	Type      string         ` + "`" + `json:"type"` + "`" + `
	Reason    string         ` + "`" + `json:"reason"` + "`" + `
	CausedBy  *elasticError  ` + "`" + `json:"caused_by"` + "`" + `
	RootCause []elasticError ` + "`" + `json:"root_cause"` + "`" + `
}

{{- if not .PreventCommonCode }}
type elasticError struct {
	Type      string         ` + "`" + `json:"type"` + "`" + `
	Reason    string         ` + "`" + `json:"reason"` + "`" + `
	CausedBy  *elasticError  ` + "`" + `json:"caused_by"` + "`" + `
	RootCause []elasticError ` + "`" + `json:"root_cause"` + "`" + `
}
{{- end }}

type {{.LowercaseClient}}DocResponse struct {
	ID      string ` + "`" + `json:"_id"` + "`" + `
	Result  string ` + "`" + `json:"result"` + "`" + `
	Error	*elasticError ` + "`" + `json:"error"` + "`" + `
}

type {{.LowercaseClient}}Hits struct {
	Took     int  ` + "`" + `json:"took"` + "`" + `
	TimedOut bool ` + "`" + `json:"timed_out"` + "`" + `
	Shards   struct {
		Total      int ` + "`" + `json:"total"` + "`" + `
		Successful int ` + "`" + `json:"successful"` + "`" + `
		Failed     int ` + "`" + `json:"failed"` + "`" + `
	} ` + "`" + `json:"_shards"` + "`" + `
	Hits struct {
		Total    int                             ` + "`" + `json:"total"` + "`" + `
		MaxScore float64                         ` + "`" + `json:"max_score"` + "`" + `
		Hits     []{{.LowercaseClient}}Hit ` + "`" + `json:"hits"` + "`" + `
	} ` + "`" + `json:"hits"` + "`" + `
	Error *elasticError ` + "`" + `json:"error"` + "`" + `
}

type {{.LowercaseClient}}Hit struct {
	ID     string  ` + "`" + `json:"_id"` + "`" + `
	Score  float64 ` + "`" + `json:"_score"` + "`" + `
	Source {{.ModelWithPrefix}} ` + "`" + `json:"_source"` + "`" + `
}`
