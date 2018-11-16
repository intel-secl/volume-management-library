package vml

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

/*
CreateVolume is used to create the sparse file if it doesn’t exist, associate the sparse file
with the image and create the dm-crypt volume for an image or the instance.
Input Parameters:
	sparseFilePath – This is the absolute path to the sparse file. All the directories leading up to
					 the file must be created before running this test.
	deviceMapperLocation – This is the absolute path to the dm-crypt volume.
	keyPath – This is the absolute path to the key file.
	diskSize – This is the size of the sparse file to be created.
*/
func CreateVolume(sparseFilePath string, deviceMapperLocation string, keyPath string, diskSize string) {
	var formatDevice = false
	var args []string
	var deviceLoop string
	var cmdOutput string
	var err error

	// input validation
	if len(strings.TrimSpace(sparseFilePath)) <= 0 {
		log.Fatal("Sparse file path not given")
	}
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		log.Fatal("Device mapper location not given")
	}
	if len(strings.TrimSpace(keyPath)) <= 0 {
		log.Fatal("Key path not given")
	}
	if len(strings.TrimSpace(diskSize)) <= 0 {
		log.Fatal("Sparse file size not given")
	}

	// check if device mapper of the same name exists in the given location
	_, err = os.Stat(deviceMapperLocation)
	if !os.IsExist(err) {
		log.Fatal("A device mapper of the same already exists\n", err)
	}

	// check if the key file exists in the location
	_, err = os.Stat(keyPath)
	if os.IsNotExist(err) {
		log.Fatal("Key file does not exist\n", err)
	}

	// get loop device associated to the sparse file and format it
	deviceLoop = getLoopDevice(sparseFilePath, diskSize, keyPath, formatDevice)

	var deviceMapperString = strings.Split(deviceMapperLocation, "/")
	var deviceMapperName = deviceMapperString[len(deviceMapperString)-1]

	// check the status of the device device mapper
	log.Println("Checking the status of the device mapper ...")
	args = []string{"status", deviceMapperLocation}
	cmdOutput, err = runCommand("cryptsetup", args)
	if strings.Contains(cmdOutput, "inactive") {
		log.Println("The device mapper is inactive, opening the luks volume ...")
		// open the luks volume
		args = []string{"-v", "luksOpen", deviceLoop, deviceMapperName, "-d", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			log.Fatal("Error trying to open the luks volume\n", err)
		} else {
			formatDevice = true
		}

		//checking the status of the volume again
		log.Println("Checking the status of the device mapper ... ")
		args = []string{"status", deviceMapperLocation}
		cmdOutput, err = runCommand("cryptsetup", args)
		if !strings.Contains(cmdOutput, "active") {
			log.Fatal("Volume is not active for use\n", err)
		}
	}
	// 9. format the volume
	log.Println("Formatting the dm-crypt volume ...")
	if formatDevice {
		args = []string{"-v", deviceMapperLocation}
		cmdOutput, err = runCommand("mkfs.ext4", args)
		if err != nil {
			log.Fatal("Error trying to format the luks volume\n", err)
		}
	}
}

// This function is used to create a sparse file is it doesn't exist,
// find a loop device and associate the sparse file with it.
func getLoopDevice(sparseFilePath, diskSize, keyPath string, formatDevice bool) string {
	var err error
	var args []string
	var deviceLoop string

	// check if the sparse file exists
	log.Println("Checking if the sparse file exists ... ")
	_, err = os.Stat(sparseFilePath)
	// sparse file does not exist, creating a new sparsefile
	if os.IsNotExist(err) {
		log.Println("Sparse file does not exist, creating a new file")
		// create a sparse file
		args = []string{"-s", diskSize, sparseFilePath}
		_, err = runCommand("truncate", args)
		if err != nil {
			log.Fatal("Error creating a sparse file\n", err)
		}
		formatDevice = true
	}
	log.Println("Sparse file exists in location ", sparseFilePath)

	// find the loop device associated to the sparse file
	log.Println("Finding a loop device that is associated to the sparse file ...")
	args = []string{"-j", sparseFilePath}
	cmdOutput, err := runCommand("losetup", args)
	if err != nil {
		log.Fatal("Error trying to find a loop device associated with the sparse file\n", err)
	}
	// find the loop device and associate it with the sparse file
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		log.Println("Associating a loop device to the sparse file ...")
		// find the loop device
		args = []string{"-f", sparseFilePath}
		cmdOutput, err = runCommand("losetup", args)
		if err != nil {
			log.Fatal("Error trying to accociate a loop device to the sparse file\n", err)
		}
	}

	// check if the loop device is associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput, err = runCommand("losetup", args)
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		log.Fatal("Sparse file is not associated to the loop device\n", err)
	} else {
		var modifiedOutput = strings.Split(cmdOutput, ":")
		deviceLoop = modifiedOutput[0]
		log.Println("The sparse file is associated to the loop device ", deviceLoop)
	}

	// format loop device
	if formatDevice {
		log.Println("Formatting the loop device ...")
		args = []string{"-v", "--batch-mode", "luksFormat", deviceLoop, "-d", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			log.Fatal("Error trying to format the loop device\n", err)
		}
	}
	return deviceLoop
}

/*
DeleteVolume method is used to delete the given dm-crypt volume.
Input Parameter:
	deviceMapperLocation – This is the absolute path to the dm-crypt volume.
*/
func DeleteVolume(deviceMapperLocation string) {
	//validate input parameters
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		log.Fatal("Device mapper location not given")
	}

	// build and excute the cryptsetup luksClose command to close and delete the volume
	log.Println("Deleting the dm-crypt volume ...")
	deleteVolumeCmd := "cryptsetup"
	args := []string{"luksClose", deviceMapperLocation}
	_, err := runCommand(deleteVolumeCmd, args)
	if err != nil {
		log.Fatal("Error trying to close the device mapper\n", err)
	}
}

/*
Mount method is used to attach the filesystem on the device mapper of a specific type at the mount path.
Input Parameters:
	deviceMapperLocation – This is the absolute path to the dm-crypt volume.
	mountLocation – This is the mount point location where the device will be mounted
*/
func Mount(deviceMapperLocation string, mountLocation string) {
	//input parameters validation
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		log.Fatal("Device mapper location not given")
	}
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		log.Fatal("Mount location not given")
	}
	// call syscall to mount the file system
	err := syscall.Mount(deviceMapperLocation, mountLocation, "ext4", 0, "")
	if err != nil {
		log.Fatal("Error while trying to mount\n", err)
	}
}

/*
Unmount method is used to detach the filesystem from the mount path.
Input Parameter:
	mountLocation – This is the mount point location  where we want to unmount the device.
*/
func Unmount(mountLocation string) {
	//input parameters validation
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		log.Fatal("Unmount location not given")
	}

	// call syscall to unmount the file system from the mount location
	err := syscall.Unmount(mountLocation, 0)
	if err != nil {
		log.Fatal("Error while trying to unmount\n", err)
	}
}

/*
CreateVMManifest is used to create a VM manifest and return a manifest.
Input Parameters:
	vmId – This is the VM instance UUID.
	hostHardwareUuid – This is the hardware UUID of the host where the VM will be launched.
	imageId – This is the image ID of the image created by the cloud orchestrator.
	imageEncrypted – This is a boolean value indicating if the image downloaded on the host by the cloud orchestrator was encrypted.
*/
func CreateVMManifest(vmID string, hostHardwareUUID string, imageID string, imageEncrypted bool) string {

	if (vmID == "") || (len(vmID) <= 0) ||
		(hostHardwareUUID == "") || (len(hostHardwareUUID) <= 0) ||
		(imageID == "") || (len(imageID) <= 0) {
		log.Fatal("Invalid input parameters")
	}

	manifest, err := getVMManifest(vmID, hostHardwareUUID, imageID, imageEncrypted)
	if err != nil {
		log.Fatal("Error creating a manifest\n", err)
	}
	return manifest
}

/*
Decrypt is used to decrypt an encrypted file into the given decrypt location
with the key in pem format using AES 256 GCM algorithm.
Input Parameters:
	encImagePath – This is the absolute path to the encrypted image on disk.
	decPath – This is the absolute path of the file where the decrypted file will be saved.
	keyPath – This is the absolute path to the key file used to decrypt the image/file.
*/
func Decrypt(encImagePath, decPath, keyPath string) {

	// input parameters validation
	if len(strings.TrimSpace(encImagePath)) <= 0 {
		log.Fatal("Encrypted file path not given")
	}
	if len(strings.TrimSpace(decPath)) <= 0 {
		log.Fatal("Path to save the decrypted file is not given")
	}

	// check if key file exists
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		log.Fatal("Key does not exist\n", err)
	}

	// check if encrypted image file exists
	_, err = os.Stat(encImagePath)
	if os.IsNotExist(err) {
		log.Fatal("Encrypted file does not exist. ", err)
	}

	// read the encrypted file
	data, _ := ioutil.ReadFile(encImagePath)
	plainText := decryptGCM(data, keyPath)

	if len(plainText) <= 0 {
		log.Fatal("Error during decryption of the file. ", err)
	}

	// write the decrypted output to file
	err = ioutil.WriteFile(decPath, plainText, 0600)
	if err != nil {
		log.Fatal("Error during writing to file. ", err)
	}
}

func decryptGCM(data []byte, keyPath string) []byte {
	//read the key
	key, err := readKey(keyPath)
	if err != nil {
		log.Fatal("Error while reading th key. ", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("Error while creating the cipher. ", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal("Error while creating a cipher block. ", err)
	}
	nonce, ciphertext := data[:12], data[12:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatal("Error while decrypting the file. ", err)
	}
	return plaintext
}

func runCommand(cmd string, args []string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	return string(out), err
}

func readKey(filename string) ([]byte, error) {
	key, err := ioutil.ReadFile(filename)
	if err != nil {
		return key, err
	}
	block, _ := pem.Decode(key)
	return block.Bytes, nil
}
