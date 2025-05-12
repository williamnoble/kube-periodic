package main

import (
	"strings"
	"testing"
)

func TestStringFormatter(t *testing.T) {
	tc := []struct {
		input    string
		expected string
	}{
		{"prometheus-prometheus-kube-prometheus-prometheus", "Pp"},
		{"argocd-dex-server", "Ad"},
		{"traefik", "Tr"},
		{"nginx", "Ng"},
	}

	for _, tt := range tc {
		t.Run(tt.input, func(t *testing.T) {
			actual := stringFormatter(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestInsertNewline(t *testing.T) {
	expected := "f\no\no"
	actual := insertNewline("foo")
	if !strings.EqualFold(expected, *actual) {
		t.Errorf("expected %s, got %s", expected, *actual)
	}
}
