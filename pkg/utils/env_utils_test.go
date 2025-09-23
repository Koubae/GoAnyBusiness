package utils

import (
	"os"
	"reflect"
	"testing"

	_ "github.com/Koubae/GoAnyBusiness/pkg/testings"
)

func TestGetEnv(t *testing.T) {
	cleanup := func(keys ...string) {
		for _, key := range keys {
			err := os.Unsetenv(key)
			if err != nil {
				return
			}
		}
	}

	t.Run(
		"string tests", func(t *testing.T) {
			tests := []struct {
				name       string
				envKey     string
				envValue   string
				defaultVal string
				want       string
			}{
				{
					name:       "existing string value",
					envKey:     "TEST_STRING",
					envValue:   "hello",
					defaultVal: "default",
					want:       "hello",
				},
				{
					name:       "missing env returns default",
					envKey:     "MISSING_STRING",
					envValue:   "",
					defaultVal: "default",
					want:       "default",
				},
				{
					name:       "missing env returns default",
					envKey:     "MISSING_ENV",
					envValue:   "",
					defaultVal: "default",
					want:       "default",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						if tt.envValue != "MISSING_ENV" {
							err := os.Setenv(tt.envKey, tt.envValue)
							if err != nil {
								return
							}
						}
						got := GetEnvString(tt.envKey, tt.defaultVal)
						if got != tt.want {
							t.Errorf("getEnv() = %v, want %v", got, tt.want)
						}
						cleanup(tt.envKey)
					},
				)
			}
		},
	)

	t.Run(
		"int tests", func(t *testing.T) {
			tests := []struct {
				name       string
				envKey     string
				envValue   string
				defaultVal int
				want       int
				panics     bool
			}{
				{
					name:       "valid integer",
					envKey:     "TEST_INT",
					envValue:   "123",
					defaultVal: 0,
					want:       123,
					panics:     false,
				},
				{
					name:       "invalid integer returns default",
					envKey:     "TEST_INT",
					envValue:   "not_a_number",
					defaultVal: 42,
					want:       42,
					panics:     true,
				},
				{
					name:       "missing env returns default",
					envKey:     "MISSING_INT",
					envValue:   "",
					defaultVal: 42,
					want:       42,
					panics:     false,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						defer func() {
							if r := recover(); r != nil {
								if !tt.panics {
									t.Errorf("getEnv() panicked: %v", r)
								}
							}
						}()

						if tt.envValue != "" {
							err := os.Setenv(tt.envKey, tt.envValue)
							if err != nil {
								return
							}
						}

						got := GetEnvInt(tt.envKey, tt.defaultVal)

						if tt.panics {
							t.Errorf("GetEnvBool() should have panicked but didn't")
						}
						if got != tt.want {
							t.Errorf("getEnv() = %v, want %v", got, tt.want)
						}
						cleanup(tt.envKey)
					},
				)
			}
		},
	)

	t.Run(
		"bool tests", func(t *testing.T) {
			tests := []struct {
				name       string
				envKey     string
				envValue   string
				defaultVal bool
				want       bool
				panics     bool
			}{
				{
					name:       "true value",
					envKey:     "TEST_BOOL",
					envValue:   "true",
					defaultVal: false,
					want:       true,
					panics:     false,
				},
				{
					name:       "false value",
					envKey:     "TEST_BOOL",
					envValue:   "false",
					defaultVal: true,
					want:       false,
					panics:     false,
				},
				{
					name:       "invalid bool returns default",
					envKey:     "TEST_BOOL",
					envValue:   "not_a_bool",
					defaultVal: true,
					want:       true,
					panics:     true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						defer func() {
							if r := recover(); r != nil {
								if !tt.panics {
									t.Errorf("GetEnvBool() panicked: %v", r)
								}
							}
						}()

						if tt.envValue != "" {
							err := os.Setenv(tt.envKey, tt.envValue)
							if err != nil {
								return
							}
						}

						got := GetEnvBool(tt.envKey, tt.defaultVal)

						if tt.panics {
							t.Errorf("GetEnvBool() should have panicked but didn't")
						}
						if got != tt.want {
							t.Errorf("GetEnvBool() = %v, want %v", got, tt.want)
						}
						cleanup(tt.envKey)
					},
				)
			}
		},
	)

	t.Run(
		"string slice tests", func(t *testing.T) {
			tests := []struct {
				name       string
				envKey     string
				envValue   string
				defaultVal []string
				want       []string
			}{
				{
					name:       "valid string array",
					envKey:     "TEST_STRING_ARRAY",
					envValue:   "a,b,c",
					defaultVal: []string{},
					want:       []string{"a", "b", "c"},
				},
				{
					name:       "empty string array",
					envKey:     "TEST_STRING_ARRAY",
					envValue:   "",
					defaultVal: []string{"default"},
					want:       []string{"default"},
				},
				{
					name:       "empty string array with spaces",
					envKey:     "TEST_STRING_ARRAY",
					envValue:   "  ",
					defaultVal: []string{"default"},
					want:       []string{"default"},
				},
				{
					name:       "missing env returns default",
					envKey:     "MISSING_ENV",
					envValue:   "",
					defaultVal: []string{"default"},
					want:       []string{"default"},
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						if tt.envValue != "MISSING_ENV" {
							err := os.Setenv(tt.envKey, tt.envValue)
							if err != nil {
								return
							}
						}
						got := GetEnvStringSlice(tt.envKey, tt.defaultVal)
						if !reflect.DeepEqual(got, tt.want) {
							t.Errorf("getEnv() = %v, want %v", got, tt.want)
						}
						cleanup(tt.envKey)
					},
				)
			}
		},
	)

	t.Run(
		"int slice tests", func(t *testing.T) {
			tests := []struct {
				name       string
				envKey     string
				envValue   string
				defaultVal []int
				want       []int
				panics     bool
			}{
				{
					name:       "valid int array",
					envKey:     "TEST_INT_ARRAY",
					envValue:   "1,2,3",
					defaultVal: []int{},
					want:       []int{1, 2, 3},
					panics:     false,
				},
				{
					name:       "partially valid int array",
					envKey:     "TEST_INT_ARRAY",
					envValue:   "1,bad,3",
					defaultVal: []int{},
					want:       []int{1, 3},
					panics:     true,
				},
				{
					name:       "empty int array",
					envKey:     "TEST_INT_ARRAY",
					envValue:   "",
					defaultVal: []int{42},
					want:       []int{42},
					panics:     false,
				},
				{
					name:       "empty spaces int array",
					envKey:     "TEST_INT_ARRAY",
					envValue:   "     ",
					defaultVal: []int{42},
					want:       []int{42},
					panics:     false,
				},
				{
					name:       "missing env returns default",
					envKey:     "MISSING_ENV",
					envValue:   "",
					defaultVal: []int{42},
					want:       []int{42},
					panics:     false,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						defer func() {
							if r := recover(); r != nil {
								if !tt.panics {
									t.Errorf("getEnv() panicked: %v", r)
								}
							}
						}()

						if tt.envValue != "MISSING_ENV" {
							err := os.Setenv(tt.envKey, tt.envValue)
							if err != nil {
								return
							}
						}
						got := GetEnvIntSlice(tt.envKey, tt.defaultVal)

						if tt.panics {
							t.Errorf("GetEnvBool() should have panicked but didn't")
						}
						if !reflect.DeepEqual(got, tt.want) {
							t.Errorf("getEnv() = %v, want %v", got, tt.want)
						}
						cleanup(tt.envKey)
					},
				)
			}
		},
	)
}
