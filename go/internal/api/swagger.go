package api

import (
	"encoding/json"
	"net/http"
)

var openAPIMethods = []string{"get", "put", "post", "delete", "patch", "options", "head"}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec, err := publicSpec()
	if err != nil {
		http.Error(w, "failed to load spec", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(spec)
}

func publicSpec() ([]byte, error) {
	raw, err := rawSpec()
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	stripInternalOperations(doc)
	return json.Marshal(doc)
}

func stripInternalOperations(doc map[string]any) {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return
	}
	for path, item := range paths {
		if pruneInternalMethods(item) {
			delete(paths, path)
		}
	}
}

func pruneInternalMethods(item any) bool {
	obj, ok := item.(map[string]any)
	if !ok {
		return false
	}
	for _, method := range openAPIMethods {
		if op, ok := obj[method].(map[string]any); ok && hasInternalTag(op) {
			delete(obj, method)
		}
	}
	return !hasAnyOperation(obj)
}

func hasInternalTag(op map[string]any) bool {
	tags, ok := op["tags"].([]any)
	if !ok {
		return false
	}
	for _, t := range tags {
		if s, ok := t.(string); ok && s == "internal" {
			return true
		}
	}
	return false
}

func hasAnyOperation(obj map[string]any) bool {
	for _, method := range openAPIMethods {
		if _, ok := obj[method]; ok {
			return true
		}
	}
	return false
}
