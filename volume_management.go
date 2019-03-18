package vml

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"intel/isecl/lib/common/crypt"
	"intel/isecl/lib/common/pkg/instance"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
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
	if !os.IsNotExist(err) {
		return errors.New("device mapper of the same already exists")
	}

	tmpKeyFile, err := ioutil.TempFile("/tmp", "volumeKey")
	if err != nil {
		return errors.New("error creating a temp key file")
	}

	defer os.Remove(tmpKeyFile.Name()) // clean up

	if _, err := tmpKeyFile.Write(key); err != nil {
		return errors.New("error while writing key to a temp file")
	}

	if err := tmpKeyFile.Close(); err != nil {
		return errors.New("error closing the temp key file")
	}

	keyPath := tmpKeyFile.Name()

	// get loop device associated to the sparse file and format it
	deviceLoop, formatDevice, err := getLoopDevice(sparseFilePath, diskSize, keyPath)
	if err != nil {
		return fmt.Errorf("error while trying to get the device loop: %s", err.Error())
	}

	var deviceMapperString = strings.Split(deviceMapperLocation, "/")
	var deviceMapperName = deviceMapperString[len(deviceMapperString)-1]

	// check the status of the device device mapper
	args = []string{"status", deviceMapperLocation}
	cmdOutput, err = runCommand("cryptsetup", args)
	if strings.Contains(cmdOutput, "inactive") {
		args = []string{"-v", "luksOpen", deviceLoop, deviceMapperName, "--key-file", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			return errors.New("error trying to open the luks volume")
		}

		//checking the status of the volume again
		args = []string{"status", deviceMapperLocation}
		cmdOutput, err = runCommand("cryptsetup", args)
		if !strings.Contains(cmdOutput, "active") {
			return errors.New("volume is not active for use")
		}
	}
	// 9. format the volume
	if formatDevice {
		args = []string{"-v", deviceMapperLocation}
		cmdOutput, err = runCommand("mkfs.ext4", args)
		if err != nil {
			return errors.New("error trying to format the luks volume")
		}
	}
	return nil
}

// This function is used to create a sparse file is it doesn't exist,
// find a loop device and associate the sparse file with it.
func getLoopDevice(sparseFilePath string, diskSize int, keyPath string) (string, bool, error) {
	var err error
	var args []string
	var deviceLoop string
	var formatDevice = false

	// check if the sparse file exists
	fileInfo, err := os.Stat(sparseFilePath)
	var fileSizeMatches = false

	// if sparse file exists, check if the file size matches the given disk size
	if !os.IsNotExist(err) {
		diskSizeInBytes := diskSize * 1000000000
		if int64(diskSizeInBytes) == fileInfo.Size() {
			fileSizeMatches = true
		}
	}

	// sparse file does not exist, creating a new sparsefile
	if (os.IsNotExist(err)) || !fileSizeMatches {
		// create a sparse file
		size := strconv.Itoa(diskSize) + "GB"
		args = []string{"-s", size, sparseFilePath}
		_, err = runCommand("truncate", args)
		if err != nil {
			return "", false, fmt.Errorf("error creating a sparse file: ", err.Error())
		}
		formatDevice = true
	}

	// find the loop device associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput, err := runCommand("losetup", args)
	if err != nil {
		return "", false, errors.New("error trying to find a loop device associated with the sparse file")
	}
	// find the loop device and associate it with the sparse file
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		// find the loop device
		args = []string{"-f", sparseFilePath}
		cmdOutput, err = runCommand("losetup", args)
		if err != nil {
			return "", false, errors.New("error trying to accociate a loop device to the sparse file")
		}
	}

	// check if the loop device is associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput, err = runCommand("losetup", args)
	if (cmdOutput == "") || (len(cmdOutput) <= 0) {
		return "", false, errors.New("sparse file is not associated to the loop device")
	} else {
		var modifiedOutput = strings.Split(cmdOutput, ":")
		deviceLoop = modifiedOutput[0]
	}

	// format loop device
	if formatDevice {
		args = []string{"-v", "--batch-mode", "luksFormat", deviceLoop, "--key-file", keyPath}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			return "", false, fmt.Errorf("error trying to format the loop device: %s", err.Error())
		}
	}
	return deviceLoop, formatDevice, nil
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
	deleteVolumeCmd := "cryptsetup"
	args := []string{"luksClose", deviceMapperLocation}
	_, err := runCommand(deleteVolumeCmd, args)
	if err != nil {
		return fmt.Errorf("error trying to close the device mapper: %s", err.Error())
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
func CreateVMManifest(vmID string, hostHardwareUUID string, imageID string, imageEncrypted bool) (instance.Manifest, error) {
	err := validate(vmID, hostHardwareUUID, imageID)
	if err != nil {
		return instance.Manifest{}, fmt.Errorf("Invalid input: %s", err.Error())
	}

	vmInfo := instance.Info{}
	vmInfo.InstanceID = vmID
	vmInfo.HostHardwareUUID = hostHardwareUUID
	vmInfo.ImageID = imageID

	manifest := instance.Manifest{}
	manifest.ImageEncrypted = imageEncrypted
	manifest.InstanceInfo = vmInfo
	return manifest, nil
}

// CreateContainerManifest is used to create a container manifest and return a manifest.
//
// Input Parameters:
//
// 	containerId – Container instance UUID.
//
// 	hostHardwareUuid – Hardware UUID of the host where the container will be launched.
//
// 	imageId – Image ID of the container image.
//
// 	imageEncrypted – A boolean value indicating if the image built is encrypted.
//
// 	integrityEnforced - A boolean value indicating if the image is signed.
func CreateContainerManifest(containerID, hostHardwareUUID, imageID string, imageEncrypted, imageIntegrityEnforced bool) (instance.Manifest, error) {
	err := validate(containerID, hostHardwareUUID, imageID)
	if err != nil {
		return instance.Manifest{}, fmt.Errorf("Invalid input: %s", err.Error())
	}

	containerInfo := instance.Info{}
	containerInfo.InstanceID = containerID
	containerInfo.HostHardwareUUID = hostHardwareUUID
	containerInfo.ImageID = imageID

	manifest := instance.Manifest{}
	manifest.ImageEncrypted = imageEncrypted
	manifest.ImageIntegrityEnforced = imageIntegrityEnforced
	manifest.InstanceInfo = containerInfo
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
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error while creating the cipher: %s", err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error while creating a cipher block: %s", err.Error())
	}

	iv := data[int(unsafe.Offsetof(encryptionHeader.IV)) : int(unsafe.Offsetof(encryptionHeader.IV))+int(unsafe.Sizeof(encryptionHeader.IV))]

	offsetSlice := data[int(unsafe.Offsetof(encryptionHeader.OffsetInLittleEndian)) : int(unsafe.Offsetof(encryptionHeader.OffsetInLittleEndian))+int(unsafe.Sizeof(encryptionHeader.OffsetInLittleEndian))]
	offsetValue := binary.LittleEndian.Uint32(offsetSlice)
	encryptedData := data[offsetValue:]

	plaintext, err := gcm.Open(nil, iv, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("error while decrypting the file: %s", err.Error())
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
