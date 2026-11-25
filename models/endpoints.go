package models

type APIEndpoint struct {
	Kind       string
	Path       string
	Namespaced bool
	Verbs      []string
}
