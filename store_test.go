package tplexpr

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestStore(t *testing.T) {
	store, err := BuildStore().
		AddFS(os.DirFS("testdata"), "*.template.txt").
		Build()
	if err != nil {
		t.Error(err)
		return
	}

	sb := strings.Builder{}
	err = store.Render(&sb, "true.template.txt", nil)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "true"
	found := sb.String()

	if expected != found {
		t.Errorf("expected '%s', found '%s'", expected, found)
	}
}

func TestStoreWatch(t *testing.T) {
	templateName := "watch-test.template.txt"
	templateFileName := path.Join("testdata", templateName)

	err := os.WriteFile(templateFileName, []byte("before"), 0644)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(templateFileName)

	store, err := BuildStore().
		AddFS(os.DirFS("testdata"), "*.template.txt").
		Watch(true).
		Build()
	if err != nil {
		t.Error(err)
		return
	}

	sb := strings.Builder{}
	err = store.Render(&sb, templateName, nil)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "before"
	found := sb.String()

	if expected != found {
		t.Errorf("expected '%s', found '%s'", expected, found)
	}

	sb.Reset()

	err = os.WriteFile(templateFileName, []byte("after"), 0644)
	if err != nil {
		t.Error(err)
		return
	}

	err = store.Render(&sb, templateName, nil)
	if err != nil {
		return
	}
	expected = "after"
	found = sb.String()

	if expected != found {
		t.Errorf("expected '%s', found '%s'", expected, found)
	}
}
