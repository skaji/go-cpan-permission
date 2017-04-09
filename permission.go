package permission

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
)

type Permission struct {
	CPANMeta string
	MetaCPAN string
	Client   *http.Client
}

func New() *Permission {
	return &Permission{
		CPANMeta: "http://cpanmetadb.plackperl.org/v1.0/package",
		MetaCPAN: "https://fastapi.metacpan.org/v1/permission",
		Client:   http.DefaultClient,
	}
}

func (p *Permission) Get(module string) (string, []PermissionResult, error) {
	distfile, modules, err := p.getMeta(module)
	if err != nil {
		return "", nil, err
	}
	res, err := p.get(modules)
	if err != nil {
		return "", nil, err
	}
	return distfile, res, nil
}

type cpanmetaResult struct {
	Distfile string            `yaml:"distfile"`
	Provides map[string]string `yaml:"provides"`
	Version  string            `yaml:"version"`
}

func (p *Permission) getMeta(module string) (string, []string, error) {
	url := p.CPANMeta + "/" + module
	res, err := p.Client.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err := errors.New(res.Status + ", " + url)
		return "", nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}
	var result cpanmetaResult
	if err := yaml.Unmarshal(b, &result); err != nil {
		return "", nil, err
	}
	modules := make([]string, 0, len(result.Provides))
	for module, _ := range result.Provides {
		modules = append(modules, module)
	}
	sort.Strings(modules)

	return result.Distfile, modules, nil
}

type metacpanQuery struct {
	Query map[string]map[string][]metacpanQueryTerm `json:"query"`
	Size  int                                       `json:"size"`
}

type metacpanQueryTerm struct {
	Term map[string]string `json:"term"`
}

type metacpanResult struct {
	Shards struct {
		Successful int `json:"successful"`
		Total      int `json:"total"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Source struct {
				ModuleName    string   `json:"module_name"`
				Owner         string   `json:"owner"`
				CoMaintainers []string `json:"co_maintainers"`
			} `json:"_source"`
			ID    string  `json:"_id"`
			Type  string  `json:"_type"`
			Index string  `json:"_index"`
			Score float64 `json:"_score"`
		} `json:"hits"`
		MaxScore float64 `json:"max_score"`
	} `json:"hits"`
	TimedOut bool `json:"timed_out"`
	Took     int  `json:"took"`
}

func makeMetaCPANQuery(modules []string) metacpanQuery {
	query := metacpanQuery{
		Query: make(map[string]map[string][]metacpanQueryTerm),
		Size:  len(modules),
	}
	query.Query["bool"] = make(map[string][]metacpanQueryTerm)
	terms := make([]metacpanQueryTerm, 0, len(modules))
	for _, module := range modules {
		terms = append(terms, metacpanQueryTerm{Term: map[string]string{"module_name": module}})
	}
	query.Query["bool"]["should"] = terms
	return query
}

func (p *Permission) get(modules []string) ([]PermissionResult, error) {
	url := p.MetaCPAN + "/_search"
	query := makeMetaCPANQuery(modules)
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	res, err := p.Client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err := errors.New(res.Status + ", " + url)
		return nil, err
	}
	var result metacpanResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result.Hits.Hits
	results := make(PermissionResults, 0, len(hits))
	for _, hit := range hits {
		co := hit.Source.CoMaintainers
		sort.Strings(co)
		p := PermissionResult{
			Owner:         hit.Source.Owner,
			ModuleName:    hit.Source.ModuleName,
			CoMaintainers: co,
		}
		results = append(results, p)
	}
	sort.Sort(results)
	return results, nil
}
