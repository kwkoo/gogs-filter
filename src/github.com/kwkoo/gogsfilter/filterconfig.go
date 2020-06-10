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
)

type rule struct {
	Ref    string `json:"ref"`
	Target string `json:"target"`
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
	for _, r := range fc.rules {
		log.Print(r.String())
	}

	return fc
}

func (r rule) String() string {
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

	ref, err := extractRefFromJSON(body)
	if err != nil {
		log.Printf("could not get ref from body: %v", err)
		http.Error(w, "could not get ref from body", http.StatusInternalServerError)
		return
	}

	target := fc.targetForRef(ref)
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

func (fc FilterConfig) targetForRef(ref string) string {
	for _, r := range fc.rules {
		// empty refs in a rule will match all refs
		if len(r.Ref) == 0 || r.Ref == ref {
			return r.Target
		}
	}
	return ""
}

func extractRefFromJSON(data []byte) (string, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return "", fmt.Errorf("could not umarshal JSON: %v", err)
	}

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
