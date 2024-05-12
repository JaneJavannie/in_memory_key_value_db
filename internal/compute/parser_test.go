package compute

import (
	"errors"
	"testing"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"

	"github.com/stretchr/testify/assert"
)

func TestParser_parse(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "simple chars ok",
			args: args{
				text: "hello world",
			},
			want:    []string{"hello", "world"},
			wantErr: false,
		},

		{
			name: "punctuation ok",
			args: args{
				text: "SET H/e/l/l/o w_o_r_l_d ***"},
			want:    []string{"SET", "H/e/l/l/o", "w_o_r_l_d", "***"},
			wantErr: false,
		},
		{
			name: "whitespace ok",
			args: args{
				text: "SET \thello\tworld\n",
			},
			want:    []string{"SET", "hello", "world"},
			wantErr: false,
		},
		{
			name: "digits ok",
			args: args{
				text: "SET 123 456",
			},
			want:    []string{"SET", "123", "456"},
			wantErr: false,
		},

		{
			name: "error symbol: .",
			args: args{
				text: "GET /abc/zxc/123.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error symbol: +",
			args: args{
				text: "GET +++",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error symbol: п",
			args: args{
				text: "DEL привет",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}

			got, err := p.parse(tt.args.text)

			if err == nil && !tt.wantErr {
				assert.Equal(t, got, tt.want)
			}

			if tt.wantErr && !errors.Is(err, consts.ErrParseSymbol) {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'0', true},
		{'5', true},
		{'a', false},
	}

	for _, test := range tests {
		result := isDigit(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'3', false},
	}

	for _, test := range tests {
		result := isLetter(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}

func TestIsPunctuation(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'*', true},
		{'!', false},
		{'_', true},
	}

	for _, test := range tests {
		result := isPunctuation(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}

func TestIsWhiteSpace(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{' ', true},
		{'\t', true},
		{'\n', true},
		{'a', false},
	}

	for _, test := range tests {
		result := isWhiteSpace(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}
