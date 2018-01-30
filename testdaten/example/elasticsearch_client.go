package example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fvosberg/errtypes"
	"github.com/pkg/errors"
)

// NewExampleElasticsearchClient instantiates a new elasticsearch client
// which is dedicated to the struct example.Example
func NewExampleElasticsearchClient(url string) (*exampleElasticsearchClient, error) {
	url = strings.TrimRight(url, "/")
	c := &exampleElasticsearchClient{
		http:     &http.Client{Timeout: 1 * time.Second},
		indexURL: fmt.Sprintf("%s/examples", url),
		typeURL:  fmt.Sprintf("%s/examples/example", url),
	}
	indexExists, err := c.IndexExists()
	if err != nil {
		return nil, err
	}
	if indexExists {
		return c, nil
	}
	err = c.CreateIndex()
	if err != nil {
		return nil, err
	}
	return c, nil
}

type exampleElasticsearchClient struct {
	http     *http.Client
	indexURL string
	typeURL  string
}

func (c *exampleElasticsearchClient) Refresh() error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/_refresh", c.indexURL), nil)
	if err != nil {
		return errors.Wrap(err, "creating refresh request failed")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return errors.Wrap(err, "refreshing examples index failed")
	}
	defer res.Body.Close()
	var result struct {
		Shards struct {
			Total      int `json:"total"`
			Successful int `json:"successful"`
			Failed     int `json:"failed"`
		} `json:"_shards"`
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return errors.Wrap(err, "decoding of refresh response failed")
	}
	if result.Shards.Failed != 0 {
		return fmt.Errorf("Refreshing of %d shards failed (%d successful; %d in total)", result.Shards.Failed, result.Shards.Successful, result.Shards.Total)
	}
	return nil
}

func (c *exampleElasticsearchClient) GetOneByID(ID string) (*Example, error) {
	res, err := ds.http.Get(fmt.Sprintf("%s/%s", ds.typeURL, ID))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response struct {
		ID     string      `json:"_id"`
		Source Transaction `json:"_source"`
		Found  bool        `json:"found"`
	}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, errors.Wrap(err, "decoding of by ID response failed")
	}
	if !response.Found {
		return nil, errtypes.NewNotFoundf("transaction with id %d not found", ID)
	}
	response.Source.ID = response.ID
	return &response.Source, nil
}

func exampleFromElasticsearchHit(hit exampleElasticsearchClientHit) Example {
	hit.Example.ID = hit.ID
	return hit.Example
}

func examplesFromElasticsearchHits(hits []exampleElasticsearchClientHits) []Example {
	res := make([]Example, len(hits))
	for n, h := range hits {
		res[n] = exampleFromElasticsearchHit(h)
	}
	return res
}

func (c *exampleElasticsearchClient) GetList(offset, limit int) ([]Example, error) {
	res, err := c.http.Get(fmt.Sprintf("%s/_search?size=%d&from=%d", ds.typeURL, limit, offset))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var result exampleElasticsearchClientHits
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "decoding of search response failed")
	}
	return examplesFromElasticsearchHits(result.Hits.Hits), nil
}

// Index creates a new Example in elasticsearch
// When the ID of the example is set, it updates the example
// The first return value indicates, whether a new records has been created or not
func (c *exampleElasticsearchClient) Index(m *Example) (bool, error) {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(t)
	if err != nil {
		return false, errors.Wrap(err, "JSON encoding of example failed")
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", c.typeURL, m.ID), body)
	if err != nil {
		return false, errors.Wrap(err, "creating index request failed")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "request to elasticsearch failed")
	}
	defer res.Body.Close()
	var response exampleElasticsearchClientIndexResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return false, errors.Wrap(err, "decoding of elasticsearch response failed")
	}
	if response.ID == "" || (response.Result != "updated" && !response.Created) {
		// if this case happens, please report with furhter information to hello@frederikvosberg.de to implement a better error handling
		return false, errors.New("indexing of document in elasticsearch failed")
	}
	t.ID = response.ID
	return response.Created, nil
}

func (c *exampleElasticsearchClient) RecreateIndex() error {
	err := ds.DeleteIndex()
	if err != nil {
		return err
	}
	return ds.CreateIndex()
}

func (c *exampleElasticsearchClient) IndexExists() (bool, error) {
	req, err := http.NewRequest("HEAD", c.indexURL, nil)
	if err != nil {
		return false, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	res.Body.Close()
	if res.StatusCode != 200 && res.StatusCode != 404 {
		return false, fmt.Errorf("status code should be 200 or 404 on checking existence of \"example\" index, but was %d", res.StatusCode)
	}
	return res.StatusCode == 200, nil
}

func (c *exampleElasticsearchClient) DeleteIndex() error {
	req, err := http.NewRequest("DELETE", c.indexURL, nil)
	if err != nil {
		return err
	}
	res, err := ds.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var response exampleElasticsearchClientIndexResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return errors.Wrap(err, "couldn't decode JSON response")
	}
	if !response.Acknowledged {
		return fmt.Errorf("deletion of index not acknowledged: %#v", response.Error)
	}
	return nil
}

func (c *exampleElasticsearchClient) CreateIndex() error {
	req, err := http.NewRequest("PUT", c.indexURL, strings.NewReader(exampleElasticsearchClientIndexDefinition))
	if err != nil {
		return err
	}
	res, err := ds.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var response exampleElasticsearchClientIndexResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return errors.Wrap(err, "couldn't decode JSON response")
	}
	if !response.Acknowledged {
		return fmt.Errorf("creation of index not acknowledged: %#v", response.Error)
	}
	return nil
}

var exampleElasticsearchClientIndexDefinition = `{
    "settings" : {
        "number_of_shards" : 1
    },
    "mappings" : {
        "transaction" : {
            "properties" : {
                "entity" : {
					"type" : "nested",
					"properties": {
						"id": {"type": "keyword"},
						"type": {"type": "keyword"}
					}
				},
				"status": {"type": "keyword"},
				"status_reason": {"type": "keyword"},
				"partner": {"type": "keyword"},
				"customer": {"type": "keyword"},
				"created": {"type": "date"},
				"points": {"type": "integer"}
            }
        }
    }
}`

type exampleElasticsearchClientIndexManipulationResponse struct {
	Acknowledged bool         `json:"acknowledged"`
	Status       int          `json:"status"`
	Error        elasticError `json:"error"`
}

type exampleElasticsearchClientError struct {
	Type      string         `json:"type"`
	Reason    string         `json:"reason"`
	CausedBy  *elasticError  `json:"caused_by"`
	RootCause []elasticError `json:"root_cause"`
}

type exampleElasticsearchClientIndexExampleResponse struct {
	ID      string `json:"_id"`
	Created bool   `json:"created"`
	Result  string `json:"result"`
}

type exampleElasticsearchClientHits struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int                             `json:"total"`
		MaxScore float64                         `json:"max_score"`
		Hits     []exampleElasticsearchClientHit `json:"hits"`
	} `json:"hits"`
}

type exampleElasticsearchClientHit struct {
	ID     string  `json:"_id"`
	Score  float64 `json:"_score"`
	Source Example `json:"_source"`
}
