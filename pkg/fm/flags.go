package fm

const (
	cmdAvailable    = "available"
	cmdRespond      = "respond"
	cmdCountTokens  = "count-tokens"
	cmdSchema       = "schema"
	cmdSchemaObject = "object"

	flagInstructions = "--instructions"
	flagText         = "--text"
	flagModel        = "--model"
	flagImage        = "--image"
	flagSchema       = "--schema"
	flagSchemaName   = "--name"
	flagFieldString  = "--string"
	flagFieldInt     = "--int"
	flagFieldArray   = "--array"
	endOfOptions     = "--"
)

func opt(flag, value string) string {
	return flag + "=" + value
}

var fieldTypeFlags = map[FieldType]string{
	String: flagFieldString,
	Int:    flagFieldInt,
}

var exitCodeKinds = map[int]error{}

var stderrLeadIns = []string{"error:", "fm:"}

var stderrPrefixes = []struct {
	kind     error
	prefixes []string
}{
	{ErrLicense, []string{
		"license",
		"you must accept",
		"terms not accepted",
	}},
	{ErrUnavailable, []string{
		"modelnotready",
		"model not ready",
		"model is not ready",
		"model unavailable",
		"model is unavailable",
		"model not available",
		"model is not available",
		"appleintelligencenotenabled",
		"apple intelligence is not enabled",
		"apple intelligence is off",
		"devicenoteligible",
		"device not eligible",
		"assetsunavailable",
		"assets unavailable",
	}},
	{ErrUnknownModel, []string{
		"unknown model",
		"unsupported model",
		"invalid model",
		"the only model",
		"the only supported model",
	}},
	{ErrContextWindow, []string{
		"exceededcontextwindowsize",
		"exceeded context window",
		"context window",
	}},
	{ErrGuardrail, []string{
		"guardrailviolation",
		"guardrail",
		"refusal",
	}},
	{ErrSchema, []string{
		"invalid schema",
		"unsupportedguide",
		"schema",
		"could not load schema",
		"failed to load schema",
	}},
}
