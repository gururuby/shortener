package generator

import (
	"github.com/gururuby/shortener/pkg/generator/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestGenerator_UUID(t *testing.T) {
	type fields struct {
		aliasLength int
	}
	tests := []struct {
		want   *regexp.Regexp
		name   string
		fields fields
	}{
		{
			name:   "generate UUID",
			fields: fields{aliasLength: 8},
			want:   regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				aliasLength: tt.fields.aliasLength,
			}
			assert.Regexp(t, tt.want, g.UUID())
		})
	}
}

func TestGenerator_Alias(t *testing.T) {
	type fields struct {
		aliasLength int
	}
	tests := []struct {
		want   *regexp.Regexp
		name   string
		fields fields
	}{
		{
			name:   "generate alias",
			fields: fields{aliasLength: 8},
			want:   regexp.MustCompile(".{8}"),
		},
		{
			name:   "generate alias with 3 chars",
			fields: fields{aliasLength: 3},
			want:   regexp.MustCompile(".{3}"),
		},
		{
			name:   "when alias length is zero",
			fields: fields{aliasLength: 0},
			want:   regexp.MustCompile(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				aliasLength: tt.fields.aliasLength,
			}
			res, _ := g.Alias()
			assert.Regexp(t, tt.want, res)
		})
	}
}

func TestGenerator_Alias_Errors(t *testing.T) {
	type fields struct {
		aliasLength int
	}
	tests := []struct {
		want   error
		name   string
		fields fields
	}{
		{
			name:   "when alias length is zero",
			fields: fields{aliasLength: 0},
			want:   errors.ErrGeneratorEmptyAliasLength,
		},
		{
			name:   "when alias length is negative",
			fields: fields{aliasLength: -1},
			want:   errors.ErrGeneratorEmptyAliasLength,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				aliasLength: tt.fields.aliasLength,
			}
			_, err := g.Alias()
			require.Error(t, err)
		})
	}
}
