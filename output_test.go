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

import "testing"

func TestStatusSincePreservesTrailingPartialRecord(t *testing.T) {
	complete := []byte(`{"message_type":"status","percent_done":0.5}`)
	partial := []byte("\n" + `{"message_type":"status","percent_done":0.9`)
	output := append(complete, partial...)

	cursor, statuses := statusSince(output, 0)

	if cursor != len(complete) {
		t.Fatalf("cursor = %d, want %d", cursor, len(complete))
	}
	if len(statuses) != 1 {
		t.Fatalf("len(statuses) = %d, want 1", len(statuses))
	}
	if statuses[0].PercentDone != 0.5 {
		t.Fatalf("PercentDone = %v, want 0.5", statuses[0].PercentDone)
	}
}

func TestStatusSinceKeepsCursorWhenNoCompleteRecord(t *testing.T) {
	output := []byte(`{"message_type":"status","percent_done":0.5`)

	cursor, statuses := statusSince(output, 0)

	if cursor != 0 {
		t.Fatalf("cursor = %d, want 0", cursor)
	}
	if len(statuses) != 0 {
		t.Fatalf("len(statuses) = %d, want 0", len(statuses))
	}
}

func TestStatusSinceClampsNegativeCursor(t *testing.T) {
	output := []byte(`{"message_type":"status","percent_done":0.5}`)

	cursor, statuses := statusSince(output, -10)

	if cursor != len(output) {
		t.Fatalf("cursor = %d, want %d", cursor, len(output))
	}
	if len(statuses) != 1 {
		t.Fatalf("len(statuses) = %d, want 1", len(statuses))
	}
}

func TestStatusSinceClampsCursorBeyondOutput(t *testing.T) {
	output := []byte(`{"message_type":"status","percent_done":0.5}`)

	cursor, statuses := statusSince(output, len(output)+10)

	if cursor != len(output) {
		t.Fatalf("cursor = %d, want %d", cursor, len(output))
	}
	if len(statuses) != 0 {
		t.Fatalf("len(statuses) = %d, want 0", len(statuses))
	}
}
