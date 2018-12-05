package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"intel/isecl/lib/common/pkg/vm"
	"intel/isecl/lib/vml"
)

type vmManifest struct {
	Manifest vm.Manifest `json:"vm_manifest"`
}

func main() {

	if len(os.Args[0:]) < 2 {
		log.Fatal("Usage :  ./lib-volume-management <methodname> <parameters>")
	}
	var methodName = os.Args[1]
	var err error
	
	switch methodName {
	case "CreateVolume":
		log.Printf("Create volume method called")
		if len(os.Args[1:]) < 5 {
			log.Fatal("Usage :  ./lib-volume-management CreateVolume sparseFilePath deviceMapperLocation keyFile diskSize")
		}
		err = vml.CreateVolume(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
		if err != nil {
			log.Println(err)
		}

	case "DeleteVolume":
		log.Printf("Delete volume method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management DeleteVolume deviceMapperLocation")
		}
		err = vml.DeleteVolume(os.Args[2])
		if err != nil {
			log.Println(err)
		}

	case "Mount":
		log.Printf("Mount method called")
		if len(os.Args[1:]) < 3 {
			log.Fatal("Usage :  ./lib-volume-management Mount deviceMapperLocation mountlocation")
		}
		err = vml.Mount(os.Args[2], os.Args[3])
		if err != nil {
			log.Println(err)
		}

	case "Unmount":
		log.Printf("Unmount method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management Unmount mountlocation")
		}
		err = vml.Unmount(os.Args[2])
		if err != nil {
			log.Println(err)
		}

	case "Decrypt":
		log.Printf("Decrypt method called")
		if len(os.Args[1:]) < 4 {
			log.Fatal("Usage :  ./lib-volume-management Decrypt encFileLocation decFileLocation keyLocation")
		}
		err = vml.Decrypt(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Println(err)
		}

	case "CreateVMManifest":
		log.Printf("Manifest creation method called")
		if len(os.Args[1:]) < 5 {
			log.Fatal("Usage :  ./lib-volume-management CreateVMManifest vmID hostHardwareUUID imageID imageEncrypted")
		}
		isEncryptionRequiredValue, _ := strconv.ParseBool(os.Args[5])
		createdManifest, err := vml.CreateVMManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequiredValue)
		var manifest vmManifest
		manifest.Manifest = createdManifest
		if err != nil {
			log.Println(err)
		}
		manifestOutput, err := serialize(manifest)
		if err != nil {
			log.Println(err)
		}
		log.Println(manifestOutput)

	default:
		log.Println("Invalid method name. \nExpected values: CreateVolume, DeleteVolume, Mount, Unmount, CreateVMManifest, Decrypt")
	}
}

func serialize(manifest vmManifest) (string, error) {
	bytes, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Can't serislize", err)
		return "", err
	}
	return string(bytes), nil
}
