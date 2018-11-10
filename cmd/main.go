package cmd

import (
	"lib-go-workload-node/pkg/vml"
)

func main() {
	vmID := "a774ddad-fca1-4670-86b2-605c88a16dab"
	imageID := "00448C61-46F2-E711-906E-001560A0406"
	hostHardwareUUID := "6ea6d824-d9b3-453f-9bf0-9167jba2fghj"
	imageEncrypted := true

	result,err := vml.GetVMManifest(vmID, imageID, hostHardwareUUID, imageEncrypted)
	if err != nil {
		println(err)
	}
	println(result)
}