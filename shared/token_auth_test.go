package shared

import "testing"

func TestAuthToken(t *testing.T) {
	ta := tokenAuth{}
	header := "Bearer TOKEN"
	result, err := ta.getToken(header)
	expected := "TOKEN"

	if result != expected || err != nil {
		t.Errorf("getToken(%q) == %q, want %q", header, result, expected)
	}
}

func TestBasicAuth(t *testing.T) {
	ta := tokenAuth{}
	header := "Basic VE9LRU46"
	result, err := ta.getToken(header)
	expected := "TOKEN"

	if result != expected || err != nil {
		t.Errorf("getToken(%q) == %q, want %q", header, result, expected)
	}
}

func TestAuthTypeNotSupported(t *testing.T) {
	ta := tokenAuth{}
	header := "Type TOKEN"
	_, err := ta.getToken(header)
	expected := "auth: Type not supported"

	if err.Error() != expected {
		t.Errorf("getToken(%q) == %q, want %q", header, err.Error(), expected)
	}
}

func TestIncorrectEncoding(t *testing.T) {
	ta := tokenAuth{}
	header := "Basic INVALID_ENCODING"
	_, err := ta.getToken(header)
	expected := "auth: Base64 encoding issue"

	if err.Error() != expected {
		t.Errorf("getToken(%q) == %q, want %q", header, err.Error(), expected)
	}
}
