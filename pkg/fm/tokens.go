package fm

import (
	"context"
	"strconv"
	"strings"
)

// CountRequest is the input to CountTokens.
type CountRequest struct {
	// Text is the text to count, passed with the text flag.
	Text string
}

// CountTokens runs fm count-tokens and returns the count. Output that is not
// exactly one non-negative integer returns ErrTokenCount with the raw stdout
// on the *Error.
func (c *Client) CountTokens(ctx context.Context, req CountRequest) (int, error) {
	out, err := c.run(ctx, []string{cmdCountTokens, opt(flagText, req.Text)})
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(out.stdout)
	if len(fields) == 1 {
		n, convErr := strconv.Atoi(fields[0])
		if convErr == nil && n >= 0 {
			return n, nil
		}
	}
	return 0, &Error{Op: cmdCountTokens, Kind: ErrTokenCount, Stdout: out.stdout, Stderr: out.stderr}
}
