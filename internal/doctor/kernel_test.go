package doctor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/doctor"
)

func TestParseKernelVersion(t *testing.T) {
	cases := []struct {
		in        string
		major     int
		minor     int
		expectErr bool
	}{
		{"6.13.5-arch1-1", 6, 13, false},
		{"6.12.0", 6, 12, false},
		{"5.15.150-generic", 5, 15, false},
		{"6.13", 6, 13, false},
		{"6.13-rc1", 6, 13, false},
		{"garbage", 0, 0, true},
		{"6", 0, 0, true},
		{"x.y.z", 0, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			major, minor, err := doctor.ParseKernelVersion(tc.in)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.major, major)
			assert.Equal(t, tc.minor, minor)
		})
	}
}

func TestSocketReachableReportsFalseForMissingPath(t *testing.T) {
	assert.False(t, doctor.SocketReachable("/this/path/does/not/exist/schedctl-doctor"))
}
