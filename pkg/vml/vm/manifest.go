package vm

/**
 *
 * @author purvades
 */

type Manifest struct {
	VmInfo      	Info	`json:"vm_info"`
	ImageEncrypted  bool	`json:"image_encrypted"`
}

type Info struct {
	VmID         		string 		`json:"vm_id"`
	HostHardwareUUID   	string		`json:"host_hardware_uuid"`
	ImageID				string		`json:"image_id"`
}


