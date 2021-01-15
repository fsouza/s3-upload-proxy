package aws

import (
	"os"
	"testing"
)

func TestGetRegion(t *testing.T) {
	var tests = []struct {
		name           string
		envs           map[string]string
		expectedRegion string
	}{
		{
			name:           "with AWS_REGION",
			envs:           map[string]string{"AWS_REGION": "sa-east-1"},
			expectedRegion: "sa-east-1",
		},
		{
			name:           "with AWS_DEFAULT_REGION",
			envs:           map[string]string{"AWS_DEFAULT_REGION": "us-east-1"},
			expectedRegion: "us-east-1",
		},
		{
			name: "with both vars",
			envs: map[string]string{
				"AWS_REGION":         "us-west-2",
				"AWS_DEFAULT_REGION": "us-east-1",
			},
			expectedRegion: "us-west-2",
		},
		{
			name:           "with no vars",
			expectedRegion: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setEnv(t, test.envs)
			region := getRegion()
			if region != test.expectedRegion {
				t.Errorf("incorrect region returned\nwant %q\ngot  %q", test.expectedRegion, region)
			}
		})
	}
}

func setEnv(t *testing.T, envs map[string]string) {
	t.Helper()
	os.Clearenv()
	for k, v := range envs {
		os.Setenv(k, v)
	}
}
