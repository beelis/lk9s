package ui //nolint:testpackage // tests unexported helpers

import (
	"strings"
	"testing"
)

func TestInfoTextFields(t *testing.T) {
	t.Parallel()

	h := headerInfo{context: "ctxname", url: "https://lk.example.com:7880", write: true}
	got := infoText(h, "v1.2.3", "v2.16.0", "2.19.4")

	for _, want := range []string{"ctxname", "lk.example.com:7880", "v1.2.3", "v2.16.0", "2.19.4", "RW"} {
		if !strings.Contains(got, want) {
			t.Errorf("header text missing %q:\n%s", want, got)
		}
	}

	if strings.Count(got, "ctxname") != 1 {
		t.Errorf("context name repeated:\n%s", got)
	}
}
