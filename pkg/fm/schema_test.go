package fm

import (
	"errors"
	"slices"
	"testing"
)

func TestSchemaObject(t *testing.T) {
	t.Parallel()
	const schemaJSON = "{\"title\":\"Person\",\"type\":\"object\"}\n"
	client, argvPath := newFakeClient(t, fake{stdout: schemaJSON})
	got, err := client.SchemaObject(t.Context(), ObjectSchema{
		Name: "Person",
		Fields: []Field{
			{Name: "name", Type: String},
			{Name: "age", Type: Int},
			{Name: "nicknames", Type: String, Repeated: true},
		},
	})
	if err != nil {
		t.Fatalf("SchemaObject: %v", err)
	}
	if string(got) != schemaJSON {
		t.Fatalf("schema = %q, want %q", got, schemaJSON)
	}
	want := []string{
		"schema", "object",
		"--name=Person",
		"--string=name",
		"--int=age",
		"--string=nicknames", "--array",
	}
	if args := readArgv(t, argvPath); !slices.Equal(args, want) {
		t.Fatalf("argv\n got %q\nwant %q", args, want)
	}
}

func TestSchemaObjectGarbageReturnsErrSchemaOutput(t *testing.T) {
	t.Parallel()
	client, _ := newFakeClient(t, fake{stdout: "wrote schema\n"})
	_, err := client.SchemaObject(t.Context(), ObjectSchema{Name: "P", Fields: []Field{{Name: "n", Type: String}}})
	if !errors.Is(err, ErrSchemaOutput) {
		t.Fatalf("err = %v, want ErrSchemaOutput", err)
	}
	if fmErr := asFMError(t, err); fmErr.Stdout != "wrote schema\n" || fmErr.Op != "schema object" {
		t.Fatalf("error fields = %+v", fmErr)
	}
}

func TestSchemaObjectRejectedByCLI(t *testing.T) {
	t.Parallel()
	client, _ := newFakeClient(t, fake{stderr: "Error: invalid schema name\n", exit: 64})
	_, err := client.SchemaObject(t.Context(), ObjectSchema{Name: "1bad", Fields: []Field{{Name: "n", Type: Int}}})
	if !errors.Is(err, ErrSchema) {
		t.Fatalf("err = %v, want ErrSchema", err)
	}
	if fmErr := asFMError(t, err); fmErr.ExitCode != 64 || fmErr.Op != "schema object" {
		t.Fatalf("error fields = %+v", fmErr)
	}
}

func TestSchemaObjectInvalidFieldsDoNotRun(t *testing.T) {
	t.Parallel()
	tests := map[string]Field{
		"unsupported type": {Name: "price", Type: FieldType("decimal")},
		"empty name":       {Type: String},
	}
	for name, field := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client, argvPath := newFakeClient(t, fake{stdout: "{}"})
			_, err := client.SchemaObject(t.Context(), ObjectSchema{Name: "P", Fields: []Field{field}})
			if !errors.Is(err, ErrSchema) {
				t.Fatalf("err = %v, want ErrSchema", err)
			}
			assertNotRun(t, argvPath)
		})
	}
}
