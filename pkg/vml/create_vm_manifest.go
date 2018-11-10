package vml

import (
	"encoding/json"
	"lib-go-workload-node/pkg/vml/vm"
)
/**
 *
 * @author purvades
 */

type vmManifest struct {
	Manifest vm.Manifest	`json:"vm_manifest"`
}

func GetVMManifest(vmID string, hostHardwareUUID string, imageID string, imageEncrypted bool) (string,error) {
	var info vm.Info
	info.VmID = vmID
	info.HostHardwareUUID = hostHardwareUUID
	info.ImageID = imageID

	var manifest vmManifest
	manifest.Manifest.VmInfo = info
	manifest.Manifest.ImageEncrypted = imageEncrypted
	return serialize(manifest)
}

func serialize(manifest vmManifest) (string,error) {
	bytes, err := json.Marshal(manifest)
	if err != nil {
		//fmt.Println("Can't serislize", manifest)
		return "",err
	}
	return string(bytes),nil
}

func deserialize(manifestJson string) (vmManifest,error) {
	var vmManifest vmManifest
	err := json.Unmarshal([]byte(manifestJson), &vmManifest)
	if err != nil {
		//fmt.Println("Can't deserislize", manifestJson)
		return vmManifest,err
	}
	return vmManifest,err
}
