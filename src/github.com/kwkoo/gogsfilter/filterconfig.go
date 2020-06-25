package gogsfilter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type rule struct {
	Ref      string `json:"ref"`
	Target   string `json:"target"`
	template *template.Template
}

// FilterConfig contains all configuration for gogsfilter.
type FilterConfig struct {
	client *http.Client
	rules  []rule
}

// InitFilterConfig initializes and returns a new FilterConfig.
func InitFilterConfig(rjson string) FilterConfig {
	fc := FilterConfig{
		client: &http.Client{},
		rules:  []rule{},
	}

	if len(rjson) == 0 {
		log.Print("No rules configured")
		return fc
	}
	dec := json.NewDecoder(strings.NewReader(rjson))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&(fc.rules)); err != nil {
		log.Fatalf("error parsing rules JSON: %v", err)
	}

	log.Printf("parsed %d rule(s)", len(fc.rules))
	for i, r := range fc.rules {
		if isTemplate(r.Target) {
			t, err := template.New(r.Ref).Parse(r.Target)
			if err != nil {
				log.Fatalf("error parsing rule template %s: %v", r.Target, err)
			}
			r.template = t
			fc.rules[i] = r
		}
		log.Print(r.String())
	}

	return fc
}

func (r rule) String() string {
	if r.template != nil {
		return fmt.Sprintf("ref: %s, template: %s", r.Ref, r.Target)
	}
	return fmt.Sprintf("ref: %s, target: %s", r.Ref, r.Target)
}

func (fc FilterConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// do not log requests from probes
	//
	if strings.HasPrefix(r.Header.Get("User-Agent"), "kube-probe") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "OK")
		return
	}

	path := r.URL.Path
	log.Printf("received request %s", path)

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %s", err)
		http.Error(w, "error reading body", http.StatusInternalServerError)
		return
	}

	parsed, err := parseJSON(body)
	if err != nil {
		log.Printf("error parsing request body JSON: %s", err)
		http.Error(w, "error parsing body JSON", http.StatusInternalServerError)
		return
	}

	ref, err := extractRefFromJSON(parsed)
	if err != nil {
		log.Printf("could not get ref from body: %v", err)
		http.Error(w, "could not get ref from body", http.StatusInternalServerError)
		return
	}

	target := fc.targetForRef(ref, parsed)
	if len(target) == 0 {
		log.Printf("could not get target for ref %s", ref)
		fmt.Fprint(w, "OK")
		return
	}

	req, err := http.NewRequest(r.Method, target, bytes.NewReader(body))
	if err != nil {
		log.Printf("error creating request to target server %s: %s", target, err)
		http.Error(w, fmt.Sprintf("error creating request to target server: %s", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// copy all headers starting with X-
	for k, v := range r.Header {
		if strings.HasPrefix(k, "X-") {
			req.Header[k] = v
		}
	}

	_, err = fc.client.Do(req)
	if err != nil {
		log.Printf("error while making request to target server %s: %s", target, err)
		http.Error(w, fmt.Sprintf("error while making request to target server: %s", err), http.StatusInternalServerError)
		return
	}
	log.Printf("successfully forwarded request to %s", target)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "OK")
}

func (fc FilterConfig) targetForRef(ref string, body map[string]interface{}) string {
	for _, r := range fc.rules {
		// empty refs in a rule will match all refs
		if len(r.Ref) == 0 || r.Ref == ref {
			if r.template == nil {
				return r.Target
			}
			var sb strings.Builder
			if err := r.template.Execute(&sb, body); err != nil {
				log.Printf("error processing template for ref %s: %v", ref, err)
				return ""
			}
			return sb.String()
		}
	}
	return ""
}

func parseJSON(data []byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	err := json.Unmarshal(data, &parsed)
	return parsed, err
}

func extractRefFromJSON(d map[string]interface{}) (string, error) {
	refint, ok := d["ref"]
	if !ok {
		return "", errors.New("ref key does not exist in JSON")
	}
	ref, ok := refint.(string)
	if !ok {
		return "", errors.New("ref is not a string")
	}
	return ref, nil
}

func isTemplate(s string) bool {
	left := strings.Index(s, "{{")
	if left == -1 {
		return false
	}
	left += 2
	if !strings.Contains(s[left:], "}}") {
		return false
	}
	return true
}
