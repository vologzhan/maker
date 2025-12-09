package maker

import (
	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vologzhan/maker/source"
	"github.com/vologzhan/maker/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	tmpDir := mustCreateTmpDir(t)
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)
	service := root.mustCreateChild(t, "service", uuid.New(), map[string]string{
		"name": "notification",
	})
	_ = service.mustCreateChild(t, "sql", uuid.New(), map[string]string{
		"name": "20240109_init",
		"up":   "CREATE TABLE channel;",
		"down": "DROP TABLE channel;",
	})
	entity := service.mustCreateChild(t, "entity", uuid.New(), map[string]string{
		"name":        "channel",
		"name_db":     "channel",
		"plural_name": "channels",
	})
	_ = entity.mustCreateChild(t, "attribute", uuid.New(), map[string]string{
		"name":        "uuid",
		"type_go":     "uuid.UUID",
		"name_db":     "uuid",
		"primary_key": "1",
		"type_db":     "uuid",
		"default":     "uuid_generate_v4()",
	})
	_ = entity.mustCreateChild(t, "attribute", uuid.New(), map[string]string{
		"name":     "relation_uuid",
		"type_go":  "uuid.UUID",
		"name_db":  "relation_uuid",
		"type_db":  "uuid",
		"fk_table": "foreign_table",
		"fk_type":  "one-to-one",
	})
	_ = entity.mustCreateChild(t, "attribute", uuid.New(), map[string]string{
		"name":     "DeletedAt",
		"nullable": "1",
		"type_go":  "time.Time",
		"name_db":  "deleted_at",
		"type_db":  "timestamp(0)",
		"default":  "null",
	})
	root.mustFlush(t)

	compareDirectory("./test/read-create", tmpDir, "", t)
}

func TestRead(t *testing.T) {
	sourceDir := "test/read-create"

	root := newTestMaker(t, sourceDir)
	require.Equal(t, 1, len(root.entrypoints))
	assert.Equal(t, 1, len(root.values))
	assert.Equal(t, map[string]string{
		"path": sourceDir,
	}, root.values)

	services := root.mustChildren(t, "service")
	require.Equal(t, 1, len(services))

	service := services[0]
	assert.Equal(t, 1, len(service.values))
	assert.Equal(t, map[string]string{
		"name": "notification",
	}, service.values)

	entities := service.mustChildren(t, "entity")
	require.Equal(t, 1, len(entities))

	entity := entities[0]
	assert.Equal(t, 2, len(entity.values))
	assert.Equal(t, map[string]string{
		"name":    "channel",
		"name_db": "channel",
	}, entity.values)

	attributes := entity.mustChildren(t, "attribute")
	require.Equal(t, 3, len(attributes))

	assert.Equal(t, 9, len(attributes[0].values))
	assert.Equal(t, map[string]string{
		"name":        "Uuid",
		"nullable":    "",
		"type_go":     "uuid.UUID",
		"name_db":     "uuid",
		"primary_key": "1",
		"type_db":     "uuid",
		"default":     "uuid_generate_v4()",
		"fk_table":    "",
		"fk_type":     "",
	}, attributes[0].values)

	assert.Equal(t, 9, len(attributes[1].values))
	assert.Equal(t, map[string]string{
		"name":        "RelationUuid",
		"nullable":    "",
		"type_go":     "uuid.UUID",
		"name_db":     "relation_uuid",
		"primary_key": "",
		"type_db":     "uuid",
		"default":     "",
		"fk_table":    "foreign_table",
		"fk_type":     "one-to-one",
	}, attributes[1].values)

	assert.Equal(t, 9, len(attributes[2].values))
	assert.Equal(t, map[string]string{
		"name":        "DeletedAt",
		"nullable":    "1",
		"type_go":     "time.Time",
		"name_db":     "deleted_at",
		"primary_key": "",
		"type_db":     "timestamp(0)",
		"default":     "null",
		"fk_table":    "",
		"fk_type":     "",
	}, attributes[2].values)
}

func TestReadWithNextKeyPath(t *testing.T) {
	root := newTestMaker(t, "test/read-with-next-key-path")
	services := root.mustChildren(t, "service")

	assert.Equal(t, 1, len(services))
	assert.Equal(t, 2, len(root.entrypoints[0].(*source.Dir).Items))
	assert.Nil(t, root.entrypoints[0].(*source.Dir).Items[0].GetTemplate())
	assert.NotNil(t, root.entrypoints[0].(*source.Dir).Items[1].GetTemplate())
}

func TestEdit(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/edit/before")
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)

	service := root.mustChildren(t, "service")[0]
	service.mustSetValues(t, map[string]string{
		"name": "calendar",
	})

	entity := service.mustChildren(t, "entity")[0]
	entity.mustSetValues(t, map[string]string{
		"name":    "only_uuid",
		"name_db": "only_uuid",
	})

	attribute := entity.mustChildren(t, "attribute")[0]
	attribute.mustSetValues(t, map[string]string{
		"name":    "uuid",
		"type_go": "uuid.UUID",
		"name_db": "uuid",
		"type_db": "uuid",
		"default": "uuid_generate_v4()",
	})

	service.mustFlush(t)

	compareDirectory("./test/edit/after", tmpDir, "", t)
}

func TestCreateAttribute(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/create-attribute/before")
	defer os.RemoveAll(tmpDir)

	newTestMaker(t, tmpDir).
		mustChildren(t, "service")[0].
		mustChildren(t, "entity")[1].
		mustCreateChild(t, "attribute", uuid.New(), map[string]string{
			"name":     "employer_id",
			"type_go":  "int",
			"name_db":  "employer_id",
			"type_db":  "int",
			"fk_table": "employers",
			"fk_type":  "one-to-one",
		}).
		mustFlush(t)

	compareDirectory("./test/create-attribute/after", tmpDir, "", t)
}

func TestCreateEntityAfterFlushService(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/create-entity-after-flush-service/before")
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)

	service := root.mustChildren(t, "service")[0]
	service.mustFlush(t)

	newEntity := service.mustCreateChild(t, "entity", uuid.New(), map[string]string{
		"name":    "table",
		"name_db": "table",
	})
	newEntity.mustFlush(t)

	compareDirectory("./test/create-entity-after-flush-service/after", tmpDir, "", t)
}

func TestCreateEntityAndThenCreateAttributeInAnotherEntity(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/create-entity-and-then-create-attribute-in-another-entity/before")
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)

	service := root.mustChildren(t, "service")[0]
	_ = service.mustCreateChild(t, "entity", uuid.New(), map[string]string{
		"name":    "ho",
		"name_db": "ho",
	})

	entity := service.mustChildren(t, "entity")[0]
	_ = entity.mustCreateChild(t, "attribute", uuid.New(), map[string]string{
		"name":    "foo",
		"type_go": "int",
		"name_db": "foo",
		"type_db": "int",
	})

	service.mustFlush(t)

	compareDirectory("./test/create-entity-and-then-create-attribute-in-another-entity/after", tmpDir, "", t)
}

func TestDeleteEntity(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/delete/before")
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)
	service := root.mustChildren(t, "service")[0]
	entity := service.mustChildren(t, "entity")[0]
	entity.mustDelete(t)
	entity.mustFlush(t)

	assert.Equal(t, 1, len(service.mustChildren(t, "entity")))

	compareDirectory("./test/delete/after-entity", tmpDir, "", t)
}

func TestDeleteAttribute(t *testing.T) {
	tmpDir := mustCopyToTmp(t, "./test/delete/before")
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)

	service := root.mustChildren(t, "service")[0]
	entity := service.mustChildren(t, "entity")[1]
	attribute := entity.mustChildren(t, "attribute")[1]
	attribute.mustDelete(t)
	attribute.mustFlush(t)

	attributes := entity.mustChildren(t, "attribute")

	assert.Equal(t, 1, len(attributes))

	compareDirectory("./test/delete/after-attribute", tmpDir, "", t)
}

func TestUnableToDeleteRootNode(t *testing.T) {
	tmpDir := mustCreateTmpDir(t)
	defer os.RemoveAll(tmpDir)

	root := newTestMaker(t, tmpDir)
	err := root.Delete()

	assert.EqualError(t, err, "maker: Node.Delete: unable to delete root node")
}

func compareDirectory(expected, actual, relativePath string, t *testing.T) {
	fullPath1 := filepath.Join(expected, relativePath)
	files1, err := os.ReadDir(fullPath1)
	if err != nil {
		t.Errorf("Error reading directory %s: %v\n", fullPath1, err)
		return
	}

	fullPath2 := filepath.Join(actual, relativePath)
	files2, err := os.ReadDir(fullPath2)
	if err != nil {
		t.Errorf("Error reading directory %s: %v\n", fullPath2, err)
		return
	}

	fileMap1 := make(map[string]fs.DirEntry)
	for _, file := range files1 {
		filename := strings.TrimSuffix(file.Name(), ".e")
		fileMap1[filename] = file
	}

	fileMap2 := make(map[string]fs.DirEntry)
	for _, file := range files2 {
		filename := strings.TrimSuffix(file.Name(), ".e")
		fileMap2[filename] = file
	}

	for name, file1 := range fileMap1 {
		file2, exists := fileMap2[name]

		if !exists {
			t.Errorf("File or directory %s present in %s but not in %s\n", filepath.Join(relativePath, name), expected, actual)
			continue
		}

		if file1.IsDir() && file2.IsDir() {
			compareDirectory(expected, actual, filepath.Join(relativePath, name), t)
		} else if file1.IsDir() || file2.IsDir() {
			t.Errorf("One is a directory and other is not: %s\n", filepath.Join(relativePath, name))
		} else {
			filepath1 := filepath.Join(fullPath1, file1.Name())
			content1, err := os.ReadFile(filepath1)
			if err != nil {
				t.Errorf("Error reading file %s: %v\n", filepath1, err)
			}

			filepath2 := filepath.Join(fullPath2, file2.Name())
			content2, err := os.ReadFile(filepath2)
			if err != nil {
				t.Errorf("Error reading file %s: %v\n", filepath2, err)
			}

			assert.Equal(t, string(content1), string(content2))
		}
	}

	for name := range fileMap2 {
		if _, exists := fileMap1[name]; !exists {
			t.Errorf("File or directory %s present in %s but not in %s\n", filepath.Join(relativePath, name), actual, expected)
		}
	}
}

func newTestMaker(t *testing.T, srcDir string) *Node {
	source.Test = true

	tplDir := os.DirFS("test/template-go")
	tpl, err := template.New(tplDir, "")
	require.NoError(t, err)

	n, err := New(tpl, srcDir)
	require.NoError(t, err)

	return n
}

func mustCreateTmpDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	return tmpDir
}

func mustCopyToTmp(t *testing.T, src string) string {
	tmpDir := mustCreateTmpDir(t)

	err := copy.Copy(src, tmpDir)
	if err != nil {
		defer os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}

	return tmpDir
}

func (n *Node) mustCreateChild(t *testing.T, nspace string, id uuid.UUID, values map[string]string) *Node {
	node, err := n.CreateChild(nspace, id, values)
	require.NoError(t, err)
	return node
}

func (n *Node) mustChildren(t *testing.T, nspace string) []*Node {
	nodes, err := n.Children(nspace)
	require.NoError(t, err)
	return nodes
}

func (n *Node) mustSetValues(t *testing.T, values map[string]string) {
	err := n.SetValues(values)
	require.NoError(t, err)
}

func (n *Node) mustDelete(t *testing.T) {
	err := n.Delete()
	require.NoError(t, err)
}

func (n *Node) mustFlush(t *testing.T) {
	err := n.Flush()
	require.NoError(t, err)
}
