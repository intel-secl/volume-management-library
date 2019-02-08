// +build linux

package vml

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"intel/isecl/lib/common/pkg/vm"
	"intel/isecl/lib/common/crypt"
	"golang.org/x/sys/unix"
	"fmt"
	"strconv"
	"encoding/base64"
	"io/ioutil"
	"unsafe"
	"encoding/binary"
)

// CreateVolume is used to create the sparse file if it does not exist, associate the sparse file
// with the image and create the dm-crypt volume for an image or the instance.
//
// Input Parameters:
//
// 	sparseFilePath – Absolute path of the sparse file. All the directories leading up to
// 					 the file must be created before using this method.
//
// 	deviceMapperLocation – Absolute path of the dm-crypt volume.
//
// 	keyPath – Absolute path of the key file.
//
// 	diskSize – Size of the sparse file to be created.
func CreateVolume(sparseFilePath string, deviceMapperLocation string, key []byte, diskSize int) error {
	var formatDevice = false
	var args []string
	var deviceLoop string
	var cmdOutput string
	var err error

	// input validation
	if len(strings.TrimSpace(sparseFilePath)) <= 0 {
		return errors.New("sparse file path not given")
	}

	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		return errors.New("device mapper location not given")
	}

	if diskSize <= 0 {
		return errors.New("sparse file size should be greater than 0")
	}

	// check if device mapper of the same name exists in the given location
	_, err = os.Stat(deviceMapperLocation)
	if  !os.IsNotExist(err) {
		return errors.New("device mapper of the same already exists")
	}

	tmpKeyFile, err := ioutil.TempFile("/tmp", "volumeKey")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpKeyFile.Name()) // clean up

	if _, err := tmpKeyFile.Write(key); err != nil {
		fmt.Println("error while writing key to a temp file")
		return errors.New("error while writing key to a temp file")
	}
	
	if err := tmpKeyFile.Close(); err != nil {
		log.Fatal(err)
	}

	keyPath := tmpKeyFile.Name()
	
	// get loop device associated to the sparse file and format it
	deviceLoop, err = getLoopDevice(sparseFilePath, diskSize, keyPath, formatDevice)
	if err != nil {
		return errors.New("error while trying to get the device loop")
	}

	var deviceMapperString = strings.Split(deviceMapperLocation, "/")
	var deviceMapperName = deviceMapperString[len(deviceMapperString)-1]

	// check the status of the device device mapper
	fmt.Println("Checking the status of the device mapper ...")
	args = []string{"status", deviceMapperLocation}
	cmdOutput, err = runCommand("cryptsetup", args)
	if strings.Contains(cmdOutput, "inactive") {
		fmt.Println("The device mapper is inactive, opening the luks volume ...")
		// open the luks volume
		fmt.Println("dev loop name:", deviceLoop)
		fmt.Println("deviceMapperName:", deviceMapperName)
		fmt.Println("keyPath:", keyPath)
		args = []string{"-v", "luksOpen", deviceLoop, deviceMapperName, "--key-file", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			log.Println("Error: ", err)
			return errors.New("error trying to open the luks volume")
		} else {
			formatDevice = true
		}

		//checking the status of the volume again
		fmt.Println("Checking the status of the device mapper ... ")
		args = []string{"status", deviceMapperLocation}
		cmdOutput, err = runCommand("cryptsetup", args)
		if !strings.Contains(cmdOutput, "active") {
			log.Println("Error: ", err)
			return errors.New("volume is not active for use")
		}
	}
	// 9. format the volume
	fmt.Println("Formatting the dm-crypt volume ...")
	if formatDevice {
		args = []string{"-v", deviceMapperLocation}
		cmdOutput, err = runCommand("mkfs.ext4", args)
		if err != nil {
			log.Println("Error: ", err)
			return errors.New("error trying to format the luks volume")
		}
	}
	return nil
}

// This function is used to create a sparse file is it doesn't exist,
// find a loop device and associate the sparse file with it.
func getLoopDevice(sparseFilePath string, diskSize int, keyPath string, formatDevice bool) (string, error) {
	var err error
	var args []string
	var deviceLoop string

	// check if the sparse file exists
	fmt.Println("Checking if the sparse file exists ... ")
	fileInfo, err := os.Stat(sparseFilePath)
	var fileSizeMatches = false

	// if sparse file exists, check if the file size matches the given disk size
	if !os.IsNotExist(err) {
		diskSizeInBytes := diskSize * 1000000000
		fmt.Println("The file size %d:", fileInfo.Size())
		fmt.Println("The given disk size %d", diskSizeInBytes)
		if int64(diskSizeInBytes) == fileInfo.Size() {
			fileSizeMatches = true
		}
	}

	// sparse file does not exist, creating a new sparsefile
	if (os.IsNotExist(err)) || !fileSizeMatches {
		fmt.Println("Sparse file does not exist, creating a new file")
		// create a sparse file
		size := strconv.Itoa(diskSize) + "GB"
		args = []string{"-s", size, sparseFilePath}
		_, err = runCommand("truncate", args)
		if err != nil {
			log.Println("Error: ", err)
			return "", errors.New("error creating a sparse file")
		}
		formatDevice = true
	} 
	fmt.Println("Sparse file exists in location ", sparseFilePath)

	// find the loop device associated to the sparse file
	fmt.Println("Finding a loop device that is associated to the sparse file ...")
	args = []string{"-j", sparseFilePath}
	cmdOutput, err := runCommand("losetup", args)
	if err != nil {
		log.Println("Error: ", err)
		return "", errors.New("error trying to find a loop device associated with the sparse file")
	}
	// find the loop device and associate it with the sparse file
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		fmt.Println("Associating a loop device to the sparse file ...")
		// find the loop device
		args = []string{"-f", sparseFilePath}
		cmdOutput, err = runCommand("losetup", args)
		if err != nil {
			log.Println("Error: ", err)
			return "", errors.New("error trying to accociate a loop device to the sparse file")
		}
	}

	// check if the loop device is associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput, err = runCommand("losetup", args)
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		return "", errors.New("sparse file is not associated to the loop device")
	} else {
		var modifiedOutput = strings.Split(cmdOutput, ":")
		deviceLoop = modifiedOutput[0]
		fmt.Println("The sparse file is associated to the loop device ", deviceLoop)
	}

	// format loop device
	if formatDevice {
		fmt.Println("Formatting the loop device ...")
		args = []string{"-v", "--batch-mode", "luksFormat", deviceLoop, "--key-file", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			log.Println("Error: ", err)
			return "", errors.New("error trying to format the loop device")
		}
	}
	return deviceLoop, nil
}

// DeleteVolume method is used to delete the given dm-crypt volume.
//
// Input Parameter:
//
// deviceMapperLocation – Absolute path of the dm-crypt volume.
func DeleteVolume(deviceMapperLocation string) error {
	//validate input parameters
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		return errors.New("device mapper location not given")
	}

	// build and excute the cryptsetup luksClose command to close and delete the volume
	fmt.Println("Deleting the dm-crypt volume ...")
	deleteVolumeCmd := "cryptsetup"
	args := []string{"luksClose", deviceMapperLocation}
	_, err := runCommand(deleteVolumeCmd, args)
	if err != nil {
		log.Println("Error: ", err)
		return errors.New("error trying to close the device mapper")
	}
	return nil
}

// Mount method is used to attach the filesystem on the device mapper of a specific type at the mount path.
//
// Input Parameters:
//
// 	deviceMapperLocation – Absolute path of the dm-crypt volume.
//
// 	mountLocation – Mount point location where the device will be mounted
func Mount(deviceMapperLocation string, mountLocation string) error {
	//input parameters validation
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		return errors.New("device mapper location not given")
	}
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		return errors.New("mount location not given")
	}
	// call syscall to mount the file system
	err := unix.Mount(deviceMapperLocation, mountLocation, "ext4", 0, "")
	if err != nil {
		log.Println("Error: ", err)
		if strings.Contains(string(err.Error()),"device or resource busy") {
			return errors.New("device is already mounted")
		} else {
			return errors.New("error while trying to mount")
		}
	}
	return nil
}

// Unmount method is used to detach the filesystem from the mount path.
//
// Input Parameter:
//
// mountLocation – Mount point location  where we want to unmount the device.
func Unmount(mountLocation string) error {
	//input parameters validation
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		return errors.New("unmount location not given")
	}

	// call syscall to unmount the file system from the mount location
	err := unix.Unmount(mountLocation, 0)
	if err != nil {
		log.Println("Error: ", err)
		return errors.New("error while trying to unmount")
	}
	return nil
}

// CreateVMManifest is used to create a VM manifest and return a manifest.
//
// Input Parameters:
//
// 	vmId – VM instance UUID.
//
// 	hostHardwareUuid – Hardware UUID of the host where the VM will be launched.
//
// 	imageId – Image ID of the image created by the cloud orchestrator.
//
// 	imageEncrypted – A boolean value indicating if the image downloaded on the host by the cloud orchestrator was encrypted.
func CreateVMManifest(vmID string, hostHardwareUUID string, imageID string, imageEncrypted bool) (vm.Manifest, error) {
	err := validate(vmID, hostHardwareUUID, imageID)
	if err != nil {
		fmt.Println("Invalid input: \n", err)
		return vm.Manifest{}, err
	}

	vmInfo := vm.Info{}
	vmInfo.VmID = vmID
	vmInfo.HostHardwareUUID = hostHardwareUUID
	vmInfo.ImageID = imageID

	manifest := vm.Manifest{}
	manifest.ImageEncrypted = imageEncrypted
	manifest.VmInfo = vmInfo
	return manifest, nil
}

// Decrypt is used to decrypt an encrypted file with the key in 
// byte format using AES 256 GCM algorithm.
//
// Input Parameters:
//
// 	data – The encrypted data.
//
// 	key – The key file used to decrypt the image/file.
//
func Decrypt(data, key []byte) ([]byte, error) {

	var encryptionHeader crypt.EncryptionHeader
	fmt.Println("Key :", base64.StdEncoding.EncodeToString(key))
	fmt.Println("decryptGCM: creating a cipher block")
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error: ", err)
		return nil, errors.New("error while creating the cipher")
	}

	fmt.Println("decryptGCM: getting gcm object")
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("Error: ", err)
		return nil, errors.New("error while creating a cipher block")
	}
	
	iv := data[int(unsafe.Offsetof(encryptionHeader.IV)) : int(unsafe.Offsetof(encryptionHeader.IV))+int(unsafe.Sizeof(encryptionHeader.IV))]

	offsetSlice := data[int(unsafe.Offsetof(encryptionHeader.OffsetInLittleEndian)) : int(unsafe.Offsetof(encryptionHeader.OffsetInLittleEndian))+int(unsafe.Sizeof(encryptionHeader.OffsetInLittleEndian))]
	offsetValue := binary.LittleEndian.Uint32(offsetSlice)
	encryptedData := data[offsetValue:]

	plaintext, err := gcm.Open(nil, iv, encryptedData, nil)
	if err != nil {
		log.Println("Error: ", err)
		return nil, errors.New("error while decrypting the file")
	}
	return plaintext, nil
}

func runCommand(cmd string, args []string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	return string(out), err
}

func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func validate(vmID string, hostHardwareUUID string, imageID string) error {
	if !isValidUUID(vmID) {
		return errors.New("the VM ID provided is invalid")
	}
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	if !r.MatchString(hostHardwareUUID) {
		return errors.New("the host hardware UUID provided is invalid")
	}
	if !isValidUUID(imageID) {
		return errors.New("the image ID provided is invalid")
	}
	return nil
}
