package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	header, _, found := strings.Cut(string(data), "\r\n")
	if !found {
		return 0, false, nil
	} else if len(header) == 0 { // headers section ends with blank line
		return 2, true, nil // account for \r\n
	}

	name, value, found := strings.Cut(header, ":")
	if !found || name == "" {
		return 0, false, fmt.Errorf("Expected header to be a name:value pair): header='%v'", header)
	}

	err = h.Add(name, value)
	if err != nil {
		return 0, false, err
	}

	return len(header) + 2, false, nil // acount for \r\n
}

func (h Headers) Add(name string, value string) error {
	if strings.ContainsAny(name, " \t") {
		return fmt.Errorf("Header field name contains whitespace: name='%v'", name)
	}

	invalidToken := regexp.MustCompile("[^A-Za-z0-9!#$%&'*+.^_`|~-]")
	if invalidToken.MatchString(name) {
		return fmt.Errorf("Header field name contains invalid characters: name='%v'", name)
	}

	name = strings.ToLower(name)
	value = strings.TrimSpace(value)
	if existing, exists := h[name]; exists {
		h[name] = existing + ", " + value
	} else {
		h[name] = value
	}

	return nil
}

func (h Headers) Get(name string) (value string, found bool) {
	value, found = h[strings.ToLower(name)]
	return
}

func (h Headers) Remove(name string) {
	delete(h, strings.ToLower(name))
}
