package main

import (
	"log"
	"os"
	"strconv"

	//"gitlab.devtools.intel.com/sst/isecl/lib-go-volume-management/pkg/vml"
	"lib-go-volume-management/pkg/vml"
)

func main() {

	var methodName string = os.Args[1]

	switch methodName {
	case "CreateVolume":
		log.Printf("Create volume method called")
		if len(os.Args[1:]) < 5 {
			log.Fatal("Usage :  ./lib-volume-management CreateVolume sparseFilePath deviceMapperLocation keyFile diskSize")
		}
		vml.CreateVolume(os.Args[2], os.Args[3], os.Args[4], os.Args[5])

	case "DeleteVolume":
		log.Printf("Delete volume method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management DeleteVolume deviceMapperLocation")
		}
		vml.DeleteVolume(os.Args[2])

	case "Mount":
		log.Printf("Mount method called")
		if len(os.Args[1:]) < 3 {
			log.Fatal("Usage :  ./lib-volume-management Mount deviceMapperLocation mountlocation")
		}
		vml.Mount(os.Args[2], os.Args[3])

	case "Unmount":
		log.Printf("Unmount method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management Unmount mountlocation")
		}
		vml.Unmount(os.Args[2])

	case "Decrypt":
		log.Printf("Decrypt method called")
		if len(os.Args[1:]) < 4 {
			log.Fatal("Usage :  ./lib-volume-management Decrypt encFileLocation decFileLocation keyLocation")
		}
		vml.Decrypt(os.Args[2], os.Args[3], os.Args[4])

	case "CreateVMManifest":
		log.Printf("Manifest creation method called")
		if len(os.Args[1:]) < 5 {
			log.Fatal("Usage :  ./lib-volume-management CreateVMManifest vmID hostHardwareUUID imageID imageEncrypted")
		}
		isEncryptionRequiredValue, _ := strconv.ParseBool(os.Args[5])
		vml.CreateVMManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequiredValue)

	default:
		log.Printf("Invalid method name. \nExpected values: CreateVolume, DeleteVolume, Mount, Unmount, CreateVMManifest, Decrypt")
	}
}
