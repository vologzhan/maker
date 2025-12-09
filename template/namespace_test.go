package template

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	fsys := os.DirFS("../_test/template-go")

	root, err := New(fsys, "")
	require.NoError(t, err)
	assert.Equal(t, "", root.Name)
	assert.Equal(t, 1, len(root.Entrypoints))
	assert.Equal(t, 1, len(root.Keys))
	assert.Equal(t, 1, len(root.Paths))
	require.Equal(t, 1, len(root.Children))

	service := root.Children["service"]
	assert.Equal(t, "service", service.Name)
	assert.Equal(t, 1, len(service.Entrypoints))
	assert.Equal(t, 1, len(service.Keys))
	assert.NotNil(t, service.Keys[0].(*Dir).NextInKeyPath)
	assert.Equal(t, 12, len(service.Paths))
	require.Equal(t, 2, len(service.Children))

	sql := service.Children["sql"]
	assert.Equal(t, "sql", sql.Name)
	assert.Equal(t, 2, len(sql.Entrypoints))
	assert.Equal(t, 0, len(sql.Keys)) // unreadable
	assert.Equal(t, 2, len(sql.Paths))
	assert.Equal(t, 0, len(sql.Children))

	entity := service.Children["entity"]
	assert.Equal(t, "entity", entity.Name)
	assert.Equal(t, 6, len(entity.Entrypoints))
	assert.Equal(t, 1, len(entity.Keys))
	assert.Equal(t, 4, len(entity.Paths))
	require.Equal(t, 2, len(entity.Children))

	attribute := entity.Children["attribute"]
	assert.Equal(t, "attribute", attribute.Name)
	assert.Equal(t, 3, len(attribute.Entrypoints))
	assert.Equal(t, 1, len(attribute.Keys))
	assert.Equal(t, 2, len(attribute.Paths))
	assert.Equal(t, 0, len(attribute.Children))

	relation := entity.Children["relation"]
	assert.Equal(t, "relation", relation.Name)
	assert.Equal(t, 2, len(relation.Entrypoints))
	assert.Equal(t, 1, len(relation.Keys))
	assert.Equal(t, 2, len(relation.Paths))
	assert.Equal(t, 0, len(relation.Children))

}
