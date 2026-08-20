package pathz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func TestUnitParse(t *testing.T) {
	tests := []struct {
		name          string
		inputPath     string
		expectedParts []*pathz.Part
	}{
		{
			name:      "simple path",
			inputPath: "/users/profile",
			expectedParts: []*pathz.Part{
				{Name: "users", Placeholder: false},
				{Name: "profile", Placeholder: false},
			},
		},
		{
			name:      "path with placeholder",
			inputPath: "/users/{id}/profile",
			expectedParts: []*pathz.Part{
				{Name: "users", Placeholder: false},
				{Name: "id", Placeholder: true},
				{Name: "profile", Placeholder: false},
			},
		},
		{
			name:      "path with multiple placeholders",
			inputPath: "/orgs/{org_id}/users/{user_id}",
			expectedParts: []*pathz.Part{
				{Name: "orgs", Placeholder: false},
				{Name: "org_id", Placeholder: true},
				{Name: "users", Placeholder: false},
				{Name: "user_id", Placeholder: true},
			},
		},
		{
			name:      "empty path",
			inputPath: "",
			expectedParts: []*pathz.Part{},
		},
		{
			name:      "root path",
			inputPath: "/",
			expectedParts: []*pathz.Part{},
		},
		{
			name:      "path with trailing slash",
			inputPath: "/users/",
			expectedParts: []*pathz.Part{
				{Name: "users", Placeholder: false},
			},
		},
		{
			name:      "path with leading and trailing slashes",
			inputPath: "/users/profile/",
			expectedParts: []*pathz.Part{
				{Name: "users", Placeholder: false},
				{Name: "profile", Placeholder: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathz.Parse(tt.inputPath)
			require.NotNil(t, result)
			require.NotNil(t, result.Pattern)
			require.Equal(t, tt.expectedParts, result.Pattern.Parts)
			require.NotNil(t, result.Values)
			require.Empty(t, result.Values)
		})
	}
}

func TestUnitString(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    *pathz.Path
		expectedPath string
	}{
		{
			name: "simple path",
			inputPath: &pathz.Path{
				Pattern: &pathz.Pattern{
					Parts: []*pathz.Part{
						{Name: "users", Placeholder: false},
						{Name: "profile", Placeholder: false},
					},
				},
				Values: make(map[string]string),
			},
			expectedPath: "/users/profile",
		},
		{
			name: "path with placeholder and values",
			inputPath: &pathz.Path{
				Pattern: &pathz.Pattern{
					Parts: []*pathz.Part{
						{Name: "users", Placeholder: false},
						{Name: "id", Placeholder: true},
						{Name: "profile", Placeholder: false},
					},
				},
				Values: map[string]string{"id": "123"},
			},
			expectedPath: "/users/123/profile",
		},
		{
			name: "path with placeholder, no value",
			inputPath: &pathz.Path{
				Pattern: &pathz.Pattern{
					Parts: []*pathz.Part{
						{Name: "users", Placeholder: false},
						{Name: "id", Placeholder: true},
					},
				},
				Values: make(map[string]string),
			},
			expectedPath: "/users/{id}",
		},
		{
			name: "path with multiple placeholders and values",
			inputPath: &pathz.Path{
				Pattern: &pathz.Pattern{
					Parts: []*pathz.Part{
						{Name: "orgs", Placeholder: false},
						{Name: "org_id", Placeholder: true},
						{Name: "users", Placeholder: false},
						{Name: "user_id", Placeholder: true},
					},
				},
				Values: map[string]string{"org_id": "google", "user_id": "john.doe"},
			},
			expectedPath: "/orgs/google/users/john.doe",
		},
		{
			name: "empty path",
			inputPath: &pathz.Path{
				Pattern: &pathz.Pattern{Parts: []*pathz.Part{}},
				Values:  make(map[string]string),
			},
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedPath, tt.inputPath.String())
		})
	}
}
