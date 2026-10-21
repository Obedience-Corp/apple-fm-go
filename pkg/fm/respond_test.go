package fm

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func TestRespondArgv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prompt string
		opts   *RespondOptions
		want   []string
	}{
		{
			name:   "nil options sends only the positional prompt",
			prompt: "What does chmod do?",
			want:   []string{"respond", "--", "What does chmod do?"},
		},
		{
			name:   "empty model is omitted",
			prompt: "hi",
			opts:   &RespondOptions{Instructions: "Answer in one sentence."},
			want:   []string{"respond", "--instructions=Answer in one sentence.", "--", "hi"},
		},
		{
			name:   "every option in a stable order with repeated images",
			prompt: "What app is this?",
			opts: &RespondOptions{
				Instructions: "Be brief.",
				Model:        "system",
				ImagePaths:   []string{"one.png", "two.png"},
				SchemaPath:   "schema.json",
			},
			want: []string{
				"respond",
				"--instructions=Be brief.",
				"--model=system",
				"--image=one.png",
				"--image=two.png",
				"--schema=schema.json",
				"--",
				"What app is this?",
			},
		},
		{
			name: "text replaces the positional prompt",
			opts: &RespondOptions{Instructions: "Summarize.", Text: "- first\n- second"},
			want: []string{"respond", "--instructions=Summarize.", "--text=- first\n- second"},
		},
		{
			name:   "prompt that starts with a dash stays positional",
			prompt: "-v means verbose?",
			want:   []string{"respond", "--", "-v means verbose?"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stdout := "a reply\n"
			if tt.opts != nil && tt.opts.SchemaPath != "" {
				stdout = `{"app_names":["Xcode"]}`
			}
			client, argvPath := newFakeClient(t, fake{stdout: stdout})
			resp, err := client.Respond(t.Context(), tt.prompt, tt.opts)
			if err != nil {
				t.Fatalf("Respond: %v", err)
			}
			if got := readArgv(t, argvPath); !slices.Equal(got, tt.want) {
				t.Fatalf("argv\n got %q\nwant %q", got, tt.want)
			}
			if resp.Stdout != stdout {
				t.Fatalf("Stdout = %q, want %q", resp.Stdout, stdout)
			}
		})
	}
}

func TestRespondTextAndPromptAreExclusive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prompt string
		opts   *RespondOptions
		want   error
	}{
		{"both set", "task", &RespondOptions{Text: "doc"}, errPromptAndText},
		{"both empty", "", &RespondOptions{Instructions: "x"}, errEmptyPrompt},
		{"nil options and empty prompt", "", nil, errEmptyPrompt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, argvPath := newFakeClient(t, fake{stdout: "unused"})
			_, err := client.Respond(t.Context(), tt.prompt, tt.opts)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			assertNotRun(t, argvPath)
		})
	}
}

func TestRespondPlainReplyHasNoJSON(t *testing.T) {
	t.Parallel()
	client, _ := newFakeClient(t, fake{stdout: "not json at all", stderr: "note\n"})
	resp, err := client.Respond(t.Context(), "hi", nil)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if resp.JSON != nil {
		t.Fatalf("JSON = %q, want nil without a schema", resp.JSON)
	}
	if resp.Stderr != "note\n" {
		t.Fatalf("Stderr = %q", resp.Stderr)
	}
}

func TestRespondSchemaDecodesJSON(t *testing.T) {
	t.Parallel()
	client, _ := newFakeClient(t, fake{stdout: "\n{\"app_names\":[\"Xcode\",\"Terminal\"]}\n"})
	resp, err := client.Respond(t.Context(), "apps?", &RespondOptions{SchemaPath: "schema.json"})
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	var got struct {
		AppNames []string `json:"app_names"`
	}
	if err := json.Unmarshal(resp.JSON, &got); err != nil {
		t.Fatalf("decode JSON %q: %v", resp.JSON, err)
	}
	if !slices.Equal(got.AppNames, []string{"Xcode", "Terminal"}) {
		t.Fatalf("app_names = %q", got.AppNames)
	}
}

func TestRespondSchemaGarbageReturnsErrSchemaOutput(t *testing.T) {
	t.Parallel()
	const garbage = "Sure! The apps are Xcode and Terminal."
	client, _ := newFakeClient(t, fake{stdout: garbage})
	resp, err := client.Respond(t.Context(), "apps?", &RespondOptions{SchemaPath: "schema.json"})
	if !errors.Is(err, ErrSchemaOutput) {
		t.Fatalf("err = %v, want ErrSchemaOutput", err)
	}
	if resp != nil {
		t.Fatalf("resp = %+v, want nil", resp)
	}
	fmErr := asFMError(t, err)
	if fmErr.Stdout != garbage || fmErr.Op != "respond" || fmErr.ExitCode != 0 {
		t.Fatalf("error fields = %+v", fmErr)
	}
}

func TestRespondFailureCarriesExitCodeAndStderr(t *testing.T) {
	t.Parallel()
	const stderr = "Error: exceededContextWindowSize(4096)\n"
	client, _ := newFakeClient(t, fake{stderr: stderr, exit: 3})
	_, err := client.Respond(t.Context(), "a very long prompt", nil)
	if !errors.Is(err, ErrContextWindow) {
		t.Fatalf("err = %v, want ErrContextWindow", err)
	}
	fmErr := asFMError(t, err)
	if fmErr.ExitCode != 3 || fmErr.Stderr != stderr || fmErr.Op != "respond" {
		t.Fatalf("error fields = %+v", fmErr)
	}
}
