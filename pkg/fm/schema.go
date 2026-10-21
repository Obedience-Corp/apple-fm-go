package fm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// FieldType is the type of one field in an ObjectSchema.
type FieldType string

// Field types that SchemaObject can express as fm flags.
const (
	String FieldType = "string"
	Int    FieldType = "int"
)

// Field is one property of an ObjectSchema.
type Field struct {
	// Name is the JSON property name.
	Name string
	// Type is the element type.
	Type FieldType
	// Repeated makes the field an array of Type.
	Repeated bool
}

// ObjectSchema describes the object that fm schema object should produce.
// It is translated to fm flags, not to a separate schema language.
type ObjectSchema struct {
	// Name is the schema's type name.
	Name string
	// Fields are the object's properties, sent in order.
	Fields []Field
}

// SchemaObject runs fm schema object and returns the schema JSON exactly as
// fm printed it. Write the bytes to a file and pass that path as
// RespondOptions.SchemaPath. A field with an unsupported type or an empty name
// returns ErrSchema without running fm. Output that is not JSON returns
// ErrSchemaOutput.
func (c *Client) SchemaObject(ctx context.Context, schema ObjectSchema) ([]byte, error) {
	args, err := schemaObjectArgs(schema)
	if err != nil {
		return nil, err
	}
	out, err := c.run(ctx, args)
	if err != nil {
		return nil, err
	}
	if !json.Valid(bytes.TrimSpace([]byte(out.stdout))) {
		return nil, &Error{Op: cmdSchema + " " + cmdSchemaObject, Kind: ErrSchemaOutput, Stdout: out.stdout, Stderr: out.stderr}
	}
	return []byte(out.stdout), nil
}

func schemaObjectArgs(schema ObjectSchema) ([]string, error) {
	args := []string{cmdSchema, cmdSchemaObject}
	if schema.Name != "" {
		args = append(args, opt(flagSchemaName, schema.Name))
	}
	for i, field := range schema.Fields {
		flag, ok := fieldTypeFlags[field.Type]
		switch {
		case field.Name == "":
			return nil, fmt.Errorf("fm schema object: field %d: %w: empty name", i, ErrSchema)
		case !ok:
			return nil, fmt.Errorf("fm schema object: field %q: %w: unsupported type %q", field.Name, ErrSchema, field.Type)
		}
		args = append(args, opt(flag, field.Name))
		if field.Repeated {
			args = append(args, flagFieldArray)
		}
	}
	return args, nil
}
