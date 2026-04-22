package model

import "strings"

// ModelRouter selects a provider for a given input.
type ModelRouter interface {
	SelectModel(input string) ModelProvider
}

type Router struct {
	providers    map[string]ModelProvider
	defaultName  string
	fallbackName string
}

func NewRouter(providers map[string]ModelProvider, defaultName string, fallbackName string) *Router {
	return &Router{providers: providers, defaultName: defaultName, fallbackName: fallbackName}
}

func (r *Router) SelectModel(input string) ModelProvider {
	if len(strings.Fields(input)) > 24 {
		if provider, ok := r.providers[r.defaultName]; ok {
			return provider
		}
	}
	if provider, ok := r.providers[r.fallbackName]; ok {
		return provider
	}
	return r.providers[r.defaultName]
}
