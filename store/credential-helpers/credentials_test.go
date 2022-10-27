package credhelpers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	ch "github.com/DelineaXPM/dsv-cli/store/credential-helpers"
)

type memoryStore struct {
	creds map[string]*ch.Credentials
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		creds: make(map[string]*ch.Credentials),
	}
}

func (m *memoryStore) Add(creds *ch.Credentials) error {
	m.creds[creds.ServerURL] = creds
	return nil
}

func (m *memoryStore) Delete(serverURL string) error {
	delete(m.creds, serverURL)
	return nil
}

func (m *memoryStore) Get(serverURL string) (string, string, error) {
	c, ok := m.creds[serverURL]
	if !ok {
		return "", "", fmt.Errorf("creds not found for %s", serverURL)
	}
	return c.Username, c.Secret, nil
}

func (m *memoryStore) List(prefix string) (map[string]string, error) {
	//Simply a placeholder to let memoryStore be a valid implementation of Helper interface
	return nil, nil
}

func (m *memoryStore) GetName() string {
	return "in-memory store"
}

func TestStore(t *testing.T) {
	serverURL := "https://index.docker.io/v1/"
	creds := &ch.Credentials{
		ServerURL: serverURL,
		Username:  "foo",
		Secret:    "bar",
	}
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()
	if err := ch.Store(h, in); err != nil {
		t.Fatal(err)
	}

	c, ok := h.creds[serverURL]
	if !ok {
		t.Fatalf("creds not found for %s\n", serverURL)
	}

	if c.Username != "foo" {
		t.Fatalf("expected username foo, got %s\n", c.Username)
	}

	if c.Secret != "bar" {
		t.Fatalf("expected username bar, got %s\n", c.Secret)
	}
}

func TestStoreMissingServerURL(t *testing.T) {
	creds := &ch.Credentials{
		ServerURL: "",
		Username:  "foo",
		Secret:    "bar",
	}

	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()

	if err := ch.Store(h, in); ch.IsCredentialsMissingServerURL(err) == false {
		t.Fatal(err)
	}
}

func TestStoreMissingUsername(t *testing.T) {
	creds := &ch.Credentials{
		ServerURL: "https://index.docker.io/v1/",
		Username:  "",
		Secret:    "bar",
	}

	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()

	if err := ch.Store(h, in); ch.IsCredentialsMissingUsername(err) == false {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	serverURL := "https://index.docker.io/v1/"
	creds := &ch.Credentials{
		ServerURL: serverURL,
		Username:  "foo",
		Secret:    "bar",
	}
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()
	if err := ch.Store(h, in); err != nil {
		t.Fatal(err)
	}

	buf := strings.NewReader(serverURL)
	w := new(bytes.Buffer)
	if err := ch.Get(h, buf, w); err != nil {
		t.Fatal(err)
	}

	if w.Len() == 0 {
		t.Fatalf("expected output in the writer, got %d", w.Len())
	}

	var c ch.Credentials
	if err := json.NewDecoder(w).Decode(&c); err != nil {
		t.Fatal(err)
	}

	if c.Username != "foo" {
		t.Fatalf("expected username foo, got %s\n", c.Username)
	}

	if c.Secret != "bar" {
		t.Fatalf("expected username bar, got %s\n", c.Secret)
	}
}

func TestGetMissingServerURL(t *testing.T) {
	serverURL := "https://index.docker.io/v1/"
	creds := &ch.Credentials{
		ServerURL: serverURL,
		Username:  "foo",
		Secret:    "bar",
	}
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()
	if err := ch.Store(h, in); err != nil {
		t.Fatal(err)
	}

	buf := strings.NewReader("")
	w := new(bytes.Buffer)

	if err := ch.Get(h, buf, w); ch.IsCredentialsMissingServerURL(err) == false {
		t.Fatal(err)
	}
}

func TestErase(t *testing.T) {
	serverURL := "https://index.docker.io/v1/"
	creds := &ch.Credentials{
		ServerURL: serverURL,
		Username:  "foo",
		Secret:    "bar",
	}
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()
	if err := ch.Store(h, in); err != nil {
		t.Fatal(err)
	}

	buf := strings.NewReader(serverURL)
	if err := ch.Erase(h, buf); err != nil {
		t.Fatal(err)
	}

	w := new(bytes.Buffer)
	if err := ch.Get(h, buf, w); err == nil {
		t.Fatal("expected error getting missing creds, got empty")
	}
}

func TestEraseMissingServerURL(t *testing.T) {
	serverURL := "https://index.docker.io/v1/"
	creds := &ch.Credentials{
		ServerURL: serverURL,
		Username:  "foo",
		Secret:    "bar",
	}
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	in := bytes.NewReader(b)

	h := newMemoryStore()
	if err := ch.Store(h, in); err != nil {
		t.Fatal(err)
	}

	buf := strings.NewReader("")
	if err := ch.Erase(h, buf); ch.IsCredentialsMissingServerURL(err) == false {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
	//This tests that there is proper input an output into the byte stream
	//Individual stores are very OS specific and have been tested in osxkeychain and secretservice respectively
	out := new(bytes.Buffer)
	h := newMemoryStore()
	if err := ch.List(h, out); err != nil {
		t.Fatal(err)
	}
	//testing that there is an output
	if out.Len() == 0 {
		t.Fatalf("expected output in the writer, got %d", 0)
	}
}
