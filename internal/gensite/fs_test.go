package gensite

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_site_write(t *testing.T) {
	tempDir := t.TempDir()

	s := newSite()
	s.add("foo.txt", []byte("foo"))
	s.add("bar/bar.txt", []byte("bar"))

	err := s.write(tempDir)
	assert.NoError(t, err)

	writtenFoo, err := os.ReadFile(tempDir + "/foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, "foo", string(writtenFoo))

	writtenBar, err := os.ReadFile(tempDir + "/bar/bar.txt")
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(writtenBar))
}
