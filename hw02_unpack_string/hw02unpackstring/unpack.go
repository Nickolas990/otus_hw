package hw02unpackstring

import (
	"errors"
	"strconv"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var result []rune
	var prev rune
	var escaped bool

	for _, r := range input {
		if escaped {
			if err := handleEscapedCharacter(r, &prev); err != nil {
				return "", err
			}
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			appendPreviousCharacter(&result, &prev)
			continue
		}

		if unicode.IsDigit(r) {
			if err := handleDigit(r, &result, &prev); err != nil {
				return "", err
			}
		} else {
			appendPreviousCharacter(&result, &prev)
			prev = r
		}
	}

	appendPreviousCharacter(&result, &prev)
	return string(result), nil
}

func handleEscapedCharacter(r rune, prev *rune) error {

	if !unicode.IsDigit(r) && r != '\\' {
		return ErrInvalidString
	}
	*prev = r
	return nil
}

func handleDigit(r rune, result *[]rune, prev *rune) error {
	if *prev == 0 {
		return ErrInvalidString
	}
	count, err := strconv.Atoi(string(r))
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		*result = append(*result, *prev)
	}
	*prev = 0
	return nil
}

func appendPreviousCharacter(result *[]rune, prev *rune) {
	if *prev != 0 {
		*result = append(*result, *prev)
		*prev = 0
	}
}
