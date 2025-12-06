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

func TestWriter(t *testing.T) {
	tmpSourceDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSourceDir)

	root, err := newTestNamespace(tmpSourceDir)
	require.NoError(t, err)

	service, err := root.CreateChild("service", uuid.New(), map[string]string{
		"name": "notification",
	})
	require.NoError(t, err)

	_, err = service.CreateChild("sql", uuid.New(), map[string]string{
		"name": "20240109_init",
		"up":   "CREATE TABLE channel;",
		"down": "DROP TABLE channel;",
	})
	require.NoError(t, err)

	entity, err := service.CreateChild("entity", uuid.New(), map[string]string{
		"name":    "Channel",
		"name_db": "channel",
	})
	require.NoError(t, err)

	_, err = entity.CreateChild("attribute", uuid.New(), map[string]string{
		"name":        "uuid",
		"type_go":     "uuid.UUID",
		"name_db":     "uuid",
		"primary_key": "1",
		"type_db":     "uuid",
		"default":     "uuid_generate_v4()",
	})
	require.NoError(t, err)

	_, err = entity.CreateChild("attribute", uuid.New(), map[string]string{
		"name":     "relation_uuid",
		"type_go":  "uuid.UUID",
		"name_db":  "relation_uuid",
		"type_db":  "uuid",
		"fk_table": "foreign_table",
		"fk_type":  "one-to-one",
	})
	require.NoError(t, err)

	_, err = entity.CreateChild("attribute", uuid.New(), map[string]string{
		"name":     "DeletedAt",
		"nullable": "1",
		"type_go":  "time.Time",
		"name_db":  "deleted_at",
		"type_db":  "timestamp(0)",
		"default":  "null",
	})
	require.NoError(t, err)

	err = root.Flush()
	require.NoError(t, err)

	compareDirectory("./test/read-write", tmpSourceDir, "", t)
}

func TestReader(t *testing.T) {
	sourceDir := "test/read-write"

	root, err := newTestNamespace(sourceDir)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"path": sourceDir,
	}, root.Values())
	require.Equal(t, 1, len(root.entrypoints))

	services, err := root.Children("service")
	require.NoError(t, err)
	require.Equal(t, 1, len(services))
	require.Equal(t, map[string]string{
		"name": "notification",
	}, services[0].Values())

	entities, err := services[0].Children("entity")
	require.NoError(t, err)
	require.Equal(t, 1, len(entities))
	require.Equal(t, map[string]string{
		"name":    "channel",
		"name_db": "channel",
	}, entities[0].Values())

	attributes, err := entities[0].Children("attribute")
	require.NoError(t, err)
	require.Equal(t, 3, len(attributes))

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
	}, attributes[0].Values())

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
	}, attributes[1].Values())

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
	}, attributes[2].Values())
}

func TestReadWithNextKeyPath(t *testing.T) {
	root, err := newTestNamespace("test/read-with-next-key-path")
	require.NoError(t, err)

	services, err := root.Children("service")
	require.NoError(t, err)

	assert.Equal(t, 1, len(services))

	assert.Equal(t, 2, len(root.entrypoints[0].(*source.Dir).Items))
	assert.Nil(t, root.entrypoints[0].(*source.Dir).Items[0].GetTemplate())
	assert.NotNil(t, root.entrypoints[0].(*source.Dir).Items[1].GetTemplate())
}

func TestEdit(t *testing.T) {
	tmpSourceDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSourceDir)

	err = copy.Copy("./test/edit/before", tmpSourceDir)
	require.NoError(t, err)

	root, err := newTestNamespace(tmpSourceDir)
	require.NoError(t, err)

	services, err := root.Children("service")
	require.NoError(t, err)

	service := services[0]
	err = service.SetValues(map[string]string{
		"name": "calendar",
	})
	require.NoError(t, err)

	entities, err := service.Children("entity")
	require.NoError(t, err)

	entity := entities[0]
	err = entity.SetValues(map[string]string{
		"name":    "only_uuid",
		"name_db": "only_uuid",
	})
	require.NoError(t, err)

	attributes, err := entity.Children("attribute")
	require.NoError(t, err)

	attribute := attributes[0]
	err = attribute.SetValues(map[string]string{
		"name":    "uuid",
		"type_go": "uuid.UUID",
		"name_db": "uuid",
		"type_db": "uuid",
		"default": "uuid_generate_v4()",
	})
	require.NoError(t, err)

	err = root.Flush()
	require.NoError(t, err)

	compareDirectory("./test/edit/after", tmpSourceDir, "", t)
}

func TestCreateAttribute(t *testing.T) {
	tmpSourceDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSourceDir)

	err = copy.Copy("./test/create-fk-attribute/before", tmpSourceDir)
	require.NoError(t, err)

	root, err := newTestNamespace(tmpSourceDir)
	require.NoError(t, err)

	services, err := root.Children("service")
	require.NoError(t, err)

	entities, err := services[0].Children("entity")
	require.NoError(t, err)

	attr, err := entities[1].CreateChild("attribute", uuid.New(), map[string]string{
		"name":     "employer_id",
		"type_go":  "int",
		"name_db":  "employer_id",
		"type_db":  "int",
		"fk_table": "employers",
		"fk_type":  "one-to-one",
	})
	require.NoError(t, err)

	err = attr.Flush()
	require.NoError(t, err)

	compareDirectory("./test/create-fk-attribute/after", tmpSourceDir, "", t)
}

func TestCreateEntityAfterFlushService(t *testing.T) {
	tmpSourceDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSourceDir)

	err = copy.Copy("./test/create-entity-after-flush-service/before", tmpSourceDir)
	require.NoError(t, err)

	root, err := newTestNamespace(tmpSourceDir)
	require.NoError(t, err)

	services, err := root.Children("service")
	require.NoError(t, err)

	err = services[0].Flush()
	require.NoError(t, err)

	newEntity, err := services[0].CreateChild("entity", uuid.New(), map[string]string{
		"name":    "table",
		"name_db": "table",
	})
	require.NoError(t, err)

	err = newEntity.Flush()
	require.NoError(t, err)

	compareDirectory("./test/create-entity-after-flush-service/after", tmpSourceDir, "", t)
}

func TestCreateEntityAndThenCreateAttributeInAnotherEntity(t *testing.T) {
	tmpSourceDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSourceDir)

	err = copy.Copy("./test/create-entity-and-then-create-attribute-in-another-entity/before", tmpSourceDir)
	require.NoError(t, err)

	root, err := newTestNamespace(tmpSourceDir)
	require.NoError(t, err)

	services, err := root.Children("service")
	require.NoError(t, err)

	_, err = services[0].CreateChild("entity", uuid.New(), map[string]string{
		"name":    "ho",
		"name_db": "ho",
	})
	require.NoError(t, err)

	entities, err := services[0].Children("entity")
	require.NoError(t, err)

	_, err = entities[0].CreateChild("attribute", uuid.New(), map[string]string{
		"name":    "foo",
		"type_go": "int",
		"name_db": "foo",
		"type_db": "int",
	})
	require.NoError(t, err)

	err = root.Flush()
	require.NoError(t, err)

	compareDirectory("./test/create-entity-and-then-create-attribute-in-another-entity/after", tmpSourceDir, "", t)
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

func newTestNamespace(srcDir string) (*Node, error) {
	source.Test = true

	tplDir := os.DirFS("test/template-go")

	tpl, err := template.New(tplDir, "")
	if err != nil {
		return nil, err
	}

	return New(tpl, srcDir)
}
