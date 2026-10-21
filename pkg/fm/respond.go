package fm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

// RespondOptions configures one fm respond call. The zero value sends only
// the prompt.
type RespondOptions struct {
	// Instructions is the system-style instruction for the model.
	Instructions string
	// Model selects the model. Leave it empty for the CLI default, or pass
	// "system". Other names are sent as given.
	Model string
	// ImagePaths are image files attached to the prompt, sent in order.
	ImagePaths []string
	// SchemaPath is an existing guided-generation schema file. When it is set,
	// the reply must be JSON and is returned in Response.JSON. Use
	// Client.SchemaObject to produce the file's contents.
	SchemaPath string
	// Text is a document passed with the text flag instead of a positional
	// prompt. When Text is set the prompt argument to Respond must be empty;
	// put the task in Instructions.
	Text string
}

// Response is the result of one fm respond call.
type Response struct {
	// Stdout is the reply exactly as fm printed it.
	Stdout string
	// Stderr is anything fm wrote to standard error on success.
	Stderr string
	// JSON is the reply as JSON. It is set only when RespondOptions.SchemaPath
	// was set.
	JSON json.RawMessage
}

var (
	errEmptyPrompt   = errors.New("fm respond: prompt and RespondOptions.Text are both empty")
	errPromptAndText = errors.New("fm respond: prompt and RespondOptions.Text are mutually exclusive; put the task in Instructions")
)

// Respond runs fm respond once and returns its reply. A nil opts sends only
// the prompt. With a schema set, a reply that is not JSON returns
// ErrSchemaOutput with the raw stdout on the *Error.
func (c *Client) Respond(ctx context.Context, prompt string, opts *RespondOptions) (*Response, error) {
	var o RespondOptions
	if opts != nil {
		o = *opts
	}
	args, err := respondArgs(prompt, o)
	if err != nil {
		return nil, err
	}

	out, err := c.run(ctx, args)
	if err != nil {
		return nil, err
	}
	resp := &Response{Stdout: out.stdout, Stderr: out.stderr}
	if o.SchemaPath == "" {
		return resp, nil
	}
	body := bytes.TrimSpace([]byte(out.stdout))
	if !json.Valid(body) {
		return nil, &Error{Op: cmdRespond, Kind: ErrSchemaOutput, Stdout: out.stdout, Stderr: out.stderr}
	}
	resp.JSON = json.RawMessage(body)
	return resp, nil
}

func respondArgs(prompt string, o RespondOptions) ([]string, error) {
	switch {
	case prompt == "" && o.Text == "":
		return nil, errEmptyPrompt
	case prompt != "" && o.Text != "":
		return nil, errPromptAndText
	}

	args := []string{cmdRespond}
	if o.Instructions != "" {
		args = append(args, opt(flagInstructions, o.Instructions))
	}
	if o.Model != "" {
		args = append(args, opt(flagModel, o.Model))
	}
	for _, path := range o.ImagePaths {
		args = append(args, opt(flagImage, path))
	}
	if o.SchemaPath != "" {
		args = append(args, opt(flagSchema, o.SchemaPath))
	}
	if o.Text != "" {
		return append(args, opt(flagText, o.Text)), nil
	}
	return append(args, endOfOptions, prompt), nil
}
