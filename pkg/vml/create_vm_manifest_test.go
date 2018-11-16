package vml

import (
	"fmt"
	"testing"
)

func TestVMManifestCreation(t *testing.T) {
	manifest, err := getVMManifest("a774ddad-fca1-4670-86b2-605c88a16dab",
		"00448C61-46F2-E711-906E-001560A04062",
		"6ea6d824-d9b3-453f-9bf0-9167jba2fghj",
		true)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("VM Manifest:%s\n", manifest)
}
