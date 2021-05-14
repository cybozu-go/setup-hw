package main

import (
	"testing"
)

func TestCheckRacadmOutput(t *testing.T) {
	t.Parallel()

	err := checkRacadmOutput(`Applying...
RAC987: Update initiated.
`, "filename.dat")
	if err != nil {
		t.Error(err)
	}

	err = checkRacadmOutput(`ERROR: Invalid File Type
`, "some_file_name.dat")
	if err == nil || err.Error() != "racadm update failed at file some_file_name.dat: msg: ERROR: Invalid File Type\n" {
		t.Fatal(err)
	}

	err = checkRacadmOutput(``, "some_file_name.dat")
	if err == nil || err.Error() != "racadm update failed at file some_file_name.dat: msg: " {
		t.Fatal(err)
	}
}
