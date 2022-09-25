package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/chainreactors/gogo/v2/pkg/fingers"
	"strings"
)

type Vulns []*fingers.Vuln

func (vs Vulns) ToString() string {
	var s string

	for _, vuln := range vs {
		var severity string
		if vuln.Severity == "" {
			severity = "high"
		} else {
			severity = vuln.Severity
		}
		s += fmt.Sprintf("[ %s: %s ] ", severity, vuln.ToString())
	}
	return s
}

type Frameworks []*fingers.Framework

func (fs Frameworks) ToString() string {
	frameworkStrs := make([]string, len(fs))
	for i, f := range fs {
		frameworkStrs[i] = f.ToString()
	}
	return strings.Join(frameworkStrs, "||")
}

func (fs Frameworks) GetNames() []string {
	var titles []string
	for _, f := range fs {
		if f.From != fingers.GUESS {
			titles = append(titles, f.Name)
		}
	}
	return titles
}

func (fs Frameworks) IsFocus() bool {
	for _, f := range fs {
		if f.IsFocus {
			return true
		}
	}
	return false
}

//func NewExtracted(name string, extractResult interface{}) *Extracted {
//	var e = &Extracted{
//		Name: name,
//	}
//	switch extractResult.(type) {
//	case string:
//		e.ExtractResult = append(e.ExtractResult, extractResult.(string))
//	case []byte:
//		e.ExtractResult = append(e.ExtractResult, string(extractResult.([]byte)))
//	case []string:
//		e.ExtractResult = append(e.ExtractResult, extractResult.([]string)...)
//	}
//	return e
//}

//type Extracted struct {
//	Name          string   `json:"name"`
//	ExtractResult []string `json:"extract_result"`
//}
//
//func (e *Extracted) ToString() string {
//	if len(e.ExtractResult) == 1 {
//		if len(e.ExtractResult[0]) > 30 {
//			return fmt.Sprintf("%s:%s ... %d bytes", e.Name, AsciiEncode(e.ExtractResult[0][:30]), len(e.ExtractResult[0]))
//		}
//		return fmt.Sprintf("%s:%s", e.Name, AsciiEncode(e.ExtractResult[0]))
//	} else {
//		return fmt.Sprintf("%s:%d items", e.Name, len(e.ExtractResult))
//	}
//}
//
type Extracts struct {
	Target       string               `json:"target"`
	MatchedNames []string             `json:"-"`
	Extractors   []*fingers.Extracted `json:"extracts"`
}

func (es *Extracts) ToResult() string {
	s, err := json.Marshal(es)
	if err != nil {
		return err.Error()
	}
	return string(s)
}

func (es *Extracts) ToString() string {
	var s string
	for _, e := range es.Extractors {
		s += fmt.Sprintf("[ Extract: %s ] ", e.ToString())
	}
	return s
}
