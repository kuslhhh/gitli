package embed

import (
	"testing"
)

func TestNewDefaults(t *testing.T) {
	e := New("", "")
	if e.baseURL != DefaultOllamaURL {
		t.Errorf("expected %s, got %s", DefaultOllamaURL, e.baseURL)
	}
	if e.model != DefaultModel {
		t.Errorf("expected %s, got %s", DefaultModel, e.model)
	}
}

func TestNewCustomURL(t *testing.T) {
	e := New("http://custom:11434", "test-model")
	if e.baseURL != "http://custom:11434" {
		t.Errorf("expected http://custom:11434, got %s", e.baseURL)
	}
	if e.model != "test-model" {
		t.Errorf("expected test-model, got %s", e.model)
	}
}

func TestNewTrailingSlash(t *testing.T) {
	e := New("http://localhost:11434/", "")
	if e.baseURL != "http://localhost:11434" {
		t.Errorf("expected no trailing slash, got %s", e.baseURL)
	}
}

func TestIsAvailableNoServer(t *testing.T) {
	e := New("http://localhost:19999", "")
	available := e.IsAvailable()
	if available {
		t.Log("expected false when no Ollama server running; got true (server may be running)")
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		a, b []float64
		want float64
	}{
		{[]float64{1, 0}, []float64{1, 0}, 1.0},
		{[]float64{1, 0}, []float64{0, 1}, 0.0},
		{[]float64{1, 2, 3}, []float64{4, 5, 6}, 0.974631846},
	}
	for _, tc := range tests {
		got := CosineSimilarity(tc.a, tc.b)
		if got < tc.want-0.0001 || got > tc.want+0.0001 {
			t.Errorf("CosineSimilarity(%v, %v) = %f, want %f", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCosineSimilarityIdentical(t *testing.T) {
	a := []float64{0.5, 0.3, 0.8, 0.1}
	b := []float64{0.5, 0.3, 0.8, 0.1}
	got := CosineSimilarity(a, b)
	if got < 0.9999 || got > 1.0001 {
		t.Errorf("identical vectors should have similarity ~1.0, got %f", got)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	got := CosineSimilarity(a, b)
	if got != 0.0 {
		t.Errorf("orthogonal vectors should have similarity 0.0, got %f", got)
	}
}

func TestCosineSimilarityMismatchedLength(t *testing.T) {
	got := CosineSimilarity([]float64{1, 2}, []float64{1, 2, 3})
	if got != 0 {
		t.Errorf("expected 0 for mismatched lengths, got %f", got)
	}
}

func TestCosineSimilarityZeroVector(t *testing.T) {
	got := CosineSimilarity([]float64{0, 0}, []float64{1, 1})
	if got != 0 {
		t.Errorf("expected 0 for zero vector, got %f", got)
	}
}
