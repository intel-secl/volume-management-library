package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/pkg/vm"
	"intel/isecl/lib/vml"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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
		size, _ := strconv.Atoi(os.Args[5])
		key, err := hex.DecodeString(os.Args[4])
		if err != nil {
			log.Println("Invalid hex format for the key")
		}

		err = vml.CreateVolume(os.Args[2], os.Args[3], key, size)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Volume created successfully in %s\n", os.Args[3])
		}

	case "DeleteVolume":
		log.Printf("Delete volume method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management DeleteVolume deviceMapperLocation")
		}
		err = vml.DeleteVolume(os.Args[2])
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Volume %s deleted successfully\n", os.Args[2])
		}

	case "Mount":
		log.Printf("Mount method called")
		if len(os.Args[1:]) < 3 {
			log.Fatal("Usage :  ./lib-volume-management Mount deviceMapperLocation mountlocation")
		}
		err = vml.Mount(os.Args[2], os.Args[3])
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Mount Successful\n")
		}

	case "Unmount":
		log.Printf("Unmount method called")
		if len(os.Args[1:]) < 2 {
			log.Fatal("Usage :  ./lib-volume-management Unmount mountlocation")
		}
		err = vml.Unmount(os.Args[2])
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Unmount Successful\n")
		}

	case "Decrypt":
		log.Printf("Decrypt method called")

		// input parameters validation
		encImagePath := os.Args[2]
		decPath := os.Args[3]
		key, err := hex.DecodeString(os.Args[4])
		if err != nil {
			log.Println("Invalid hex format for the key")
		}
		if len(strings.TrimSpace(encImagePath)) <= 0 {
			log.Println("Encrypted file path is not given")
			log.Fatal("Usage :  ./lib-volume-management Decrypt encFilePath decFilePath keyPath")
		}

		if len(strings.TrimSpace(decPath)) <= 0 {
			log.Println("Path to save the decrypted file is not given")
			log.Fatal("Usage :  ./lib-volume-management Decrypt encFilePath decFilePath keyPath")
		}

		fmt.Println("enc image path: ", encImagePath)
		fmt.Println("decPath:", decPath)

		// check if encrypted image file exists
		_, err = os.Stat(encImagePath)
		if os.IsNotExist(err) {
			log.Fatal("encrypted file does not exist")
		}

		fmt.Println("enc image exists ", encImagePath)

		// read the encrypted file
		encryptedData, err := ioutil.ReadFile(encImagePath)
		if err != nil {
			log.Fatal("error while reading the image")
		}

		decryptedData, err := vml.Decrypt(encryptedData, key)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("File decrypted successfully")
		}
		//Save to file
		err = ioutil.WriteFile(decPath, decryptedData, 0600)
		if err != nil {
			log.Fatal("error during writing to file")
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
		} else {
			log.Println(manifestOutput)
		}

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
