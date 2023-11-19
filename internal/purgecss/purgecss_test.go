package purgecss

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_purgecss(t *testing.T) {
	const html = `<html><body><div class="app"></div></body></html>`
	const css = `body { margin: 0; }
.foo { color: red; }`
	const expected = "body { margin: 0; }\n"
	res, err := purgecss(html, css)
	require.NoError(t, err)
	assert.Equal(t, expected, string(res), "should remove unused css")
}
