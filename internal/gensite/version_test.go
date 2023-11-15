package gensite

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed version_test.html
var versionTestHtml string

func TestVersion_Html(t *testing.T) {
	sections := map[string][]string{
		"foo": {"first", "second"},
		"bar": {"third", "fourth"},
	}
	v := version{Sections: sections}

	result, err := v.html()
	assert.NoError(t, err)
	assert.Equal(t, versionTestHtml, string(result))
}
