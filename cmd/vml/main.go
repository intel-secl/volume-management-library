/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/v3/pkg/instance"
	"intel/isecl/lib/common/v3/validation"
	"intel/isecl/lib/vml/v3"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type instanceManifest struct {
	Manifest instance.Manifest `json:"instance_manifest"`
}

func main() {

	if len(os.Args[0:]) < 2 {
		fmt.Println("Usage : ", os.Args[0], "<methodname> <parameters>")
		os.Exit(1)
	}

	inputValArr := []string{os.Args[0], os.Args[1]}
	if valErr := validation.ValidateStrings(inputValArr); valErr != nil {
		fmt.Println("Invalid string format")
		os.Exit(1)
	}

	var methodName = os.Args[1]
	var err error

	switch methodName {
	case "CreateVolume":
		fmt.Println("Creating dm-crypt volume...")
		if len(os.Args[1:]) < 5 {
			fmt.Println("Invalid arguments")
			fmt.Printf("Usage : %s CreateVolume sparseFilePath deviceMapperLocation key diskSize\n", os.Args[0])
			os.Exit(1)
		}

		inputArr := []string{os.Args[2], os.Args[3], os.Args[5]}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println("Invalid string format")
			os.Exit(1)
		}

		if validateHexStringErr := validation.ValidateHexString(os.Args[4]); validateHexStringErr != nil {
			fmt.Println("Invalid hex format for the key")
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
			fmt.Printf("Usage : %s DeleteVolume deviceMapperLocation\n", os.Args[0])
		}

		inputArr := []string{os.Args[2]}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println(err.Error())
			os.Exit(1)
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
			fmt.Printf("Usage : %s Mount deviceMapperLocation mountlocation\n", os.Args[0])
			os.Exit(1)
		}

		inputArr := []string{os.Args[2], os.Args[3]}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println("Invalid stting format")
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
			fmt.Printf("Usage : %s Unmount mountlocation\n", os.Args[0])
			os.Exit(1)
		}

		inputArr := []string{os.Args[2]}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println("Invalid mount location string format")
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
			fmt.Printf("Usage : %s Decrypt <encryptedImagePath> <decryptionOutputFilePath> <key>\n", os.Args[0])
			os.Exit(1)
		}
		// input parameters validation
		encImagePath := os.Args[2]
		decPath := os.Args[3]

		inputArr := []string{encImagePath, decPath}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println("Invalid string format")
			os.Exit(1)
		}

		if validateHexStringErr := validation.ValidateHexString(os.Args[4]); validateHexStringErr != nil {
			fmt.Println("Invalid hex format for the key")
			os.Exit(1)
		}

		key, err := hex.DecodeString(os.Args[4])
		if err != nil {
			fmt.Println("Error while decoding hex string")
			os.Exit(1)
		}
		if len(strings.TrimSpace(encImagePath)) <= 0 {
			fmt.Println("Encrypted file path is not given")
			fmt.Printf("Usage : %s Decrypt encFilePath decFilePath keyPath\n", os.Args[0])
			os.Exit(1)
		}

		if len(strings.TrimSpace(decPath)) <= 0 {
			fmt.Println("Path to save the decrypted file is not given")
			fmt.Printf("Usage : %s Decrypt encFilePath decFilePath keyPath\n", os.Args[0])
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
			fmt.Printf("Usage : %s CreateVMManifest vmID hostHardwareUUID imageID imageEncrypted\n", os.Args[0])
			os.Exit(1)
		}

		if err := validation.ValidateUUIDv4(os.Args[2]); err != nil {
			fmt.Println("Invalid VM UUID format")
			os.Exit(1)
		}

		if err = validation.ValidateHardwareUUID(os.Args[3]); err != nil {
			fmt.Println("Invalid VM UUID format")
			os.Exit(1)
		}

		if err = validation.ValidateUUIDv4(os.Args[4]); err != nil {
			fmt.Println("Invalid image UUID format")
			os.Exit(1)
		}

		isEncryptionRequiredValue, err := strconv.ParseBool(os.Args[5])
		if err != nil {
			fmt.Println("Invalid boolean value for EncryptionRequired")
			os.Exit(1)
		}

		createdManifest, err := vml.CreateVMManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequiredValue)
		var manifest instanceManifest
		manifest.Manifest = createdManifest
		if err != nil {
			fmt.Println(err)
		}
		manifestOutput, err := serialize(manifest)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(manifestOutput)
		}

	case "CreateContainerManifest":
		fmt.Printf("Manifest creation method called")
		if len(os.Args[1:]) < 6 {
			fmt.Printf("Usage :  %s CreateContainerManifest containerID hostHardwareUUID imageID imageEncrypted imageIntegrityEnforced\n", os.Args[0])
			os.Exit(1)
		}

		inputArr := []string{os.Args[2], os.Args[3], os.Args[4]}
		if validateInputErr := validation.ValidateStrings(inputArr); validateInputErr != nil {
			fmt.Println("Invalid string format")
			os.Exit(1)
		}

		isEncryptionRequired, err := strconv.ParseBool(os.Args[5])
		if err != nil {
			fmt.Printf("Enter value (true/false) for imageEncrypted : %s\n", err.Error())
			os.Exit(1)
		}
		isIntegrityEnforced, err := strconv.ParseBool(os.Args[6])
		if err != nil {
			fmt.Printf("Enter value (true/false) for imageIntegrityEnforced %s\n", err.Error())
			os.Exit(1)
		}
		createdManifest, err := vml.CreateContainerManifest(os.Args[2], os.Args[3], os.Args[4], isEncryptionRequired, isIntegrityEnforced)
		var manifest instanceManifest
		manifest.Manifest = createdManifest
		if err != nil {
			fmt.Printf("Error creating the container manifest: %s\n", err.Error())
			os.Exit(1)
		}
		if manifestOutput, err := serialize(manifest); err != nil {
			fmt.Printf("Error serializing manifest output\n")
			os.Exit(1)
		} else {
			fmt.Println(manifestOutput)
			os.Exit(0)
		}

	default:
		fmt.Println("Invalid method name \nExpected values: CreateVolume, DeleteVolume, Mount, Unmount, CreateVMManifest, Decrypt, CreateContainerManifest")
	}
}

func serialize(manifest instanceManifest) (string, error) {
	bytes, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
