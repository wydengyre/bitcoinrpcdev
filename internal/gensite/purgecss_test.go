package gensite

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_purgeStyleCss(t *testing.T) {
	const html = `<html>
<head><style>
body { margin: 0; }
.foo { color: red; }
</style></head>
<body><div class="app"></div></body></html>`
	const expected = `<html><head><style>
body { margin: 0; }

</style></head>
<body><div class="app"></div></body></html>`
	res, err := purgeStyleCss([]byte(html))
	require.NoError(t, err)
	assert.Equal(t, expected, string(res), "should remove unused css")
}

func Test_purgecss(t *testing.T) {
	const html = `<html><body><div class="app"></div></body></html>`
	const css = `body { margin: 0; }
.foo { color: red; }`
	const expected = "body { margin: 0; }\n"
	res, err := purgecss(html, css)
	require.NoError(t, err)
	assert.Equal(t, expected, string(res), "should remove unused css")
}
