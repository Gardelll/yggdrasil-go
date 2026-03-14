package util

import (
	"testing"
)

func TestOfflineUUIDFromName(t *testing.T) {
	// Expected value: MD5("OfflinePlayer:Notch") with UUID v3 version and IETF variant bits applied.
	// Compatible with Java's UUID.nameUUIDFromBytes(("OfflinePlayer:Notch").getBytes(UTF_8)).
	got := OfflineUUIDFromName("Notch")
	expected := "b50ad385-829d-3141-a216-7e7d7539ba7f"
	if got.String() != expected {
		t.Errorf("OfflineUUIDFromName(\"Notch\") = %s, want %s", got.String(), expected)
	}

	unsigned := UnsignedString(got)
	expectedUnsigned := "b50ad385829d3141a2167e7d7539ba7f"
	if unsigned != expectedUnsigned {
		t.Errorf("UnsignedString = %s, want %s", unsigned, expectedUnsigned)
	}
}
