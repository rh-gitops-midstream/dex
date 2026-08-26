package main

import (
	"crypto/tls"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		logger, err := newLogger(slog.LevelInfo, "json")
		require.NoError(t, err)
		require.NotEqual(t, (*slog.Logger)(nil), logger)
	})

	t.Run("Text", func(t *testing.T) {
		logger, err := newLogger(slog.LevelError, "text")
		require.NoError(t, err)
		require.NotEqual(t, (*slog.Logger)(nil), logger)
	})

	t.Run("Unknown", func(t *testing.T) {
		logger, err := newLogger(slog.LevelError, "gofmt")
		require.Error(t, err)
		require.Equal(t, "log format is not one of the supported values (json, text): gofmt", err.Error())
		require.Equal(t, (*slog.Logger)(nil), logger)
	})
}

func TestParseCipherSuites(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, got []uint16)
	}{
		{
			name:    "empty input",
			input:   []string{},
			wantErr: false,
			validate: func(t *testing.T, got []uint16) {
				assert.Empty(t, got)
			},
		},
		{
			name:    "single valid cipher",
			input:   []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
			wantErr: false,
			validate: func(t *testing.T, got []uint16) {
				require.Len(t, got, 1)
				assert.Equal(t, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, got[0])
			},
		},
		{
			name: "multiple valid ciphers",
			input: []string{
				"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			},
			wantErr: false,
			validate: func(t *testing.T, got []uint16) {
				require.Len(t, got, 2)
				assert.Equal(t, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, got[0])
				assert.Equal(t, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, got[1])
			},
		},
		{
			name:    "insecure cipher",
			input:   []string{"TLS_RSA_WITH_3DES_EDE_CBC_SHA"},
			wantErr: false,
			validate: func(t *testing.T, got []uint16) {
				require.Len(t, got, 1)
				assert.Equal(t, tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, got[0])
			},
		},
		{
			name:        "unsupported cipher",
			input:       []string{"TLS_FAKE_CIPHER"},
			wantErr:     true,
			errContains: `unsupported cipher suite "TLS_FAKE_CIPHER"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCipherSuites(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.errContains))
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			tt.validate(t, got)
		})
	}
}

func TestParseCurvePreferences(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		want        []tls.CurveID
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:  "single valid curve",
			input: []string{"X25519"},
			want:  []tls.CurveID{tls.X25519},
		},
		{
			name:  "multiple valid curves",
			input: []string{"X25519", "P256", "P384", "P521"},
			want: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
			},
		},
		{
			name:        "unknown curve",
			input:       []string{"X25519", "UnknownCurve"},
			want:        nil,
			wantErr:     true,
			errContains: `unknown curve: "UnknownCurve"`,
		},
		{
			name:        "unknown curve as first entry",
			input:       []string{"UnknownCurve"},
			want:        nil,
			wantErr:     true,
			errContains: `unknown curve: "UnknownCurve"`,
		},
		{
			name:        "unknown curve after valid curves",
			input:       []string{"P256", "UnknownCurve", "P384"},
			want:        nil,
			wantErr:     true,
			errContains: `unknown curve: "UnknownCurve"`,
		},
		{
			name:  "duplicate curves",
			input: []string{"P256", "P256", "X25519"},
			want: []tls.CurveID{
				tls.CurveP256,
				tls.CurveP256,
				tls.X25519,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCurvePreferences(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errContains)
				}

				if got != nil {
					t.Errorf("got curves = %v, want nil", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapKeys(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]tls.CurveID
		want  []string
	}{
		{
			name:  "empty map",
			input: map[string]tls.CurveID{},
			want:  []string{},
		},
		{
			name: "returns sorted keys",
			input: map[string]tls.CurveID{
				"P521":   tls.CurveP521,
				"X25519": tls.X25519,
				"P256":   tls.CurveP256,
				"P384":   tls.CurveP384,
			},
			want: []string{
				"P256",
				"P384",
				"P521",
				"X25519",
			},
		},
		{
			name: "single key",
			input: map[string]tls.CurveID{
				"X25519": tls.X25519,
			},
			want: []string{"X25519"},
		},
		{
			name:  "nil map",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapKeys(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
