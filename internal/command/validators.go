// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func GlobalFlagsValidator(ctx context.Context, c *cli.Command) error {
	return nil
}

type FlagValidatorType func(any) error

func FlagValidators(value any, validators ...FlagValidatorType) error {
	for _, v := range validators {
		if err := v(value); err != nil {
			return err
		}
	}
	return nil
}

// JammedFlagValidator verifies that the arg following a flag does not begin
// with '--'.  urfave/cli allows this and I don't see how to turn it off.
func JammedFlagValidator(value any) error {
	if strings.HasPrefix(value.(string), "--") {
		return errors.New("must not begin with '--'")
	}
	return nil
}

func MustBeTrueValidator(value any) error {
	if !value.(bool) {
		return errors.New("must be true")
	}
	return nil
}

func OutputValidator(value any) error {
	var validOutputFlagValues = []string{"text", "json", "raw", "yaml"}
	valid := false
	for _, v := range validOutputFlagValues {
		if v == value {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("must be one of %v", validOutputFlagValues)
	}
	return nil
}
