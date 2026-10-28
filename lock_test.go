/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restic

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLockError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-exclusive lock",
			err:      errors.New("unable to create lock in backend: repository is already locked by PID 68 on worm-restic-shell by nobody (UID 65534, GID 65534)"),
			expected: true,
		},
		{
			name:     "exclusive lock",
			err:      errors.New("repository is already locked exclusively by PID 51 on retention-policy-x\nthe `unlock` command can be used to remove stale locks"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("Fatal: wrong password or no key found"),
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isLockError(tt.err))
		})
	}
}
