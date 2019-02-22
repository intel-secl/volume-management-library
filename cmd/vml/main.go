package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/pkg/image"
	"intel/isecl/lib/vml"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type imageManifest struct {
	Manifest image.Manifest `json:"image_manifest"`
}

func main() {

	if len(os.Args[0:]) < 2 {
		fmt.Println("Usage : ",os.Args[0], "<methodname> <parameters>")
		os.Exit(1)
	}
	var methodName = os.Args[1]
	var err error

	switch methodName {
	case "CreateVolume":
		fmt.Println("Creating dm-crypt volume...")
		if len(os.Args[1:]) < 5 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," CreateVolume sparseFilePath deviceMapperLocation key diskSize")
			os.Exit(1)
		}

		key, err := hex.DecodeString(os.Args[4])
		if err != nil {
			fmt.Println("Invalid hex format for the key")
			os.Exit(1)
		}

		size, _ := strconv.Atoi(os.Args[5])
		if err = vml.CreateVolume(os.Args[2], os.Args[3], key, size); err != nil {
			fmt.Printf("Error creating the dm-crypt volume: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Volume created successfully in %s\n", os.Args[3])
			os.Exit(0)
		}

	case "DeleteVolume":
		fmt.Println("Deleting dm-crypt volume...")
		if len(os.Args[1:]) < 2 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," DeleteVolume deviceMapperLocation")
		}
		if err = vml.DeleteVolume(os.Args[2]); err != nil {
			fmt.Printf("Error deleting the dm-crypt volume: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Successfully deleted dm-crypt volume: %s\n", os.Args[2])
			os.Exit(0)
		}

	case "Mount":
		fmt.Println("Mounting the device...")
		if len(os.Args[1:]) < 3 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," Mount deviceMapperLocation mountlocation")
			os.Exit(1)
		}
		if err = vml.Mount(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error mounting the device: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Device mounted successfully in %s\n", os.Args[3])
		os.Exit(0)

	case "Unmount":
		fmt.Println("Unmounting the device...")
		if len(os.Args[1:]) < 2 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," Unmount mountlocation")
			os.Exit(1)
		}
		err = vml.Unmount(os.Args[2])
		if err != nil {
			fmt.Printf("Error unmounting the device: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Unmounted %s successfully\n", os.Args[2])
		os.Exit(0)

	case "Decrypt":
		fmt.Println("Decrypting the image file...")
		if len(os.Args[1:]) < 4 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," Decrypt <encryptedImagePath> <decryptionOutputFilePath> <key>")
			os.Exit(1)
		}
		// input parameters validation
		encImagePath := os.Args[2]
		decPath := os.Args[3]
		key, err := hex.DecodeString(os.Args[4])
		if err != nil {
			fmt.Println("Invalid hex format for the key")
			os.Exit(1)
		}
		if len(strings.TrimSpace(encImagePath)) <= 0 {
			fmt.Println("Encrypted file path is not given")
			fmt.Println("Usage : ",os.Args[0]," Decrypt encFilePath decFilePath keyPath")
			os.Exit(1)
		}

		if len(strings.TrimSpace(decPath)) <= 0 {
			fmt.Println("Path to save the decrypted file is not given")
			fmt.Println("Usage : ",os.Args[0]," Decrypt encFilePath decFilePath keyPath")
			os.Exit(1)
		}

		// check if encrypted image file exists
		_, err = os.Stat(encImagePath)
		if os.IsNotExist(err) {
			fmt.Println("Encrypted file does not exist")
			os.Exit(1)
		}

		// read the encrypted file
		encryptedData, err := ioutil.ReadFile(encImagePath)
		if err != nil {
			fmt.Println("Error while reading the image file")
			os.Exit(1)
		}

		decryptedData, err := vml.Decrypt(encryptedData, key)
		if err != nil {
			fmt.Printf("Error decrypting the image: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("Image file decrypted successfully")
		//Save to file
		if err = ioutil.WriteFile(decPath, decryptedData, 0600); err != nil {
			fmt.Printf("Error during writing to file: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Decrypted image will be found in: %s\n", decPath)
		os.Exit(0)

	case "CreateVMManifest":
		fmt.Println("Creating VM manifest...")
		if len(os.Args[1:]) < 5 {
			fmt.Println("Invalid arguments")
			fmt.Println("Usage : ",os.Args[0]," vmID hostHardwareUUID imageID imageEncrypted")
			os.Exit(1)
		}
		isEncryptionRequiredValue, _ := strconv.ParseBool(os.Args[5])
		createdManifest, err := vml.CreateVMManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequiredValue)
		var manifest imageManifest
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

	case "CreateContainerManifest":
		log.Printf("Manifest creation method called")
		if len(os.Args[1:]) < 6 {
			log.Fatal("Usage :  ./lib-volume-management CreateContainerManifest containerID hostHardwareUUID imageID imageEncrypted imageIntegrityEnforced")
		}
		isEncryptionRequiredValue, err := strconv.ParseBool(os.Args[5])
		if err != nil {
			log.Fatal("Enter value (true/false) for imageEncrypted : " + err.Error())
		}
		isIntegrityEnforcedValue, err := strconv.ParseBool(os.Args[6])
		if err != nil {
			log.Fatal("Enter value (true/false) for imageIntegrityEnforced : " + err.Error())
		}
		createdManifest, err := vml.CreateContainerManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequiredValue, isIntegrityEnforcedValue)
		var manifest imageManifest
		manifest.Manifest = createdManifest
		if err != nil {
			fmt.Printf("Error creating the VM manifest: %s\n", err.Error())
			os.Exit(1)
		}
		if manifestOutput, err := serialize(manifest); err != nil {
			fmt.Printf("Error serializing manifest output")
			os.Exit(1)
		} else {
			fmt.Println(manifestOutput)
			os.Exit(0)
		}

	default:
		log.Println("Invalid method name. \nExpected values: CreateVolume, DeleteVolume, Mount, Unmount, CreateVMManifest, Decrypt, CreateContainerManifest")
	}
}

func serialize(manifest imageManifest) (string, error) {
	bytes, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Can't serialize", err)
		return "", err
	}
	return string(bytes), nil
}
