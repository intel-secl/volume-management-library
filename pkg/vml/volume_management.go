package vml

import(

	"os/exec"
	"os"
	"log"
	"strings"
	"syscall"
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
	"encoding/pem"

)

/*
This method is used to create the sparse file if it doesn’t exist, associate the sparse file 
with the image and create the dm-crypt volume for an image or the instance.
 Input Parameters 
	sparseFilePath – This is the absolute path to the sparse file. All the directories leading up to 
					 the file must be created before running this test.
	deviceMapperLocation – This is the absolute path to the dm-crypt volume.
	keyFile – This is the absolute path to the key file.
	diskSize – This is the size of the sparse file to be created.
 Return Type : Boolean - Returns true if the volume was successfully created else returns false.
*/  

func IsVolumeCreated(sparseFilePath string, deviceMapperLocation string, keyFile string, diskSize string) bool {

	var formatDevice bool = false
	var args []string
	var deviceLoop string
	var cmdOutput string
	var err error

	// input validation
	if (sparseFilePath == "") || (len(sparseFilePath) <= 0) ||
		(deviceMapperLocation == "") || (len(deviceMapperLocation) <= 0) ||
        (keyFile == "") || (len(keyFile) <=0) {
			log.Fatal("Invalid input parameters")
	}

	// validate the disksize in KB/GB


	// check if the key file exists in the location
	_, err = os.Stat(keyFile)
	if(os.IsNotExist(err)) {
		log.Fatal("key file does not exist. ", err)
	}

	// 1. check if the sparse file exists
	log.Println("Checking if the sparse file exists in :", sparseFilePath)
	_, err = os.Stat(sparseFilePath)
	// sprase file does not exist, creating a new sparsefile
	if(os.IsNotExist(err)) {
		log.Println("Sparse file does not exist, creating a new file")
		//2. Create a sparse file
		args = []string{"-s", diskSize, sparseFilePath}
		cmdOutput, err = runCommand("truncate", args)
		if err != nil {
			log.Fatal("Error creating a sparse file. ", err)
		}
		formatDevice = true
	}
	log.Println("Sparse file exists in location", sparseFilePath)

	//3. find the loop device associated to the sparse file
	log.Println("Finding a loop device that is associated to the sparse file")
	args = []string{"-j", sparseFilePath}
	cmdOutput, err = runCommand("losetup", args)
		if err != nil {
			log.Fatal("Error trying to find a loop device associated with the sparse file. ", err)
		}
		// find the loop device and associate it with the sparse file
		if (cmdOutput == "") || (len(cmdOutput) <= 0) {
			log.Println("No loop device found that is assciated to the sparse file")
			log.Println("finding a loop device and associating it to the sparse file")
			//4. find the loop device
			args = []string{"-f", sparseFilePath}
			cmdOutput,err = runCommand("losetup", args)
			if err != nil {
				log.Fatal("Error trying to attach the sparse file to the loop device. ", err)
			}
		}

	//check if the loop device is associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput,err = runCommand("losetup", args)
		if (cmdOutput == "") || (len(cmdOutput) <= 0) {
			log.Fatal("Sparse file is not associated to the loop device correctly. ", err)
		} else {
			var modifiedOutput = strings.Split(cmdOutput, ":")
			deviceLoop = modifiedOutput[0]
			log.Println("The sparse file is associated to the loop device : ", deviceLoop)
		}

	// 6. format loop device
	log.Println("Formatting the loop device : ", deviceLoop)
	if (formatDevice) {
		log.Println("Format device value : ", formatDevice)
		args = []string{"-v", "--batch-mode", "luksFormat", deviceLoop, "-d", keyFile}
		cmdOutput, err = runCommand("cryptsetup", args)
		if err != nil {
			log.Fatal("Error trying to format the loop device. ", err)
		}
	}

	var deviceMapperString = strings.Split(deviceMapperLocation, "/")
    var deviceMapperName string = deviceMapperString[len(deviceMapperString)-1]
	

	// 7. check the status of the device device mapper
	log.Println("Checking the status of the device mapper ", deviceMapperLocation)
	log.Println("Checking the status of the device mapper name", deviceMapperName)
	args = []string{"status", deviceMapperLocation}
	cmdOutput, err = runCommand("cryptsetup", args)
		if (strings.Contains(cmdOutput, "inactive")) {
			log.Println("The device mapper is inactive, opening the luks volume")
			// 8. open the luks volume
			args = []string{"-v", "luksOpen", deviceLoop, deviceMapperName, "-d", keyFile}
			cmdOutput, err = runCommand("cryptsetup", args)
			if err != nil {
				log.Fatal("Error trying to open the luks volume. ", err)
			} else {
				formatDevice = true
			}

			//checking the status of the volume again
			log.Println("Checking the status of the device mapper ", deviceMapperLocation)
			args = []string{"status", deviceMapperLocation}
			cmdOutput, err = runCommand("cryptsetup", args)
			if (!strings.Contains(cmdOutput, "active")) {
				log.Fatal("Volume is not active for use. ", err)
			}
		}
	// 9. format the volume
	log.Println("Formatting the dm-crypt volume")
	if (formatDevice) {
		log.Println("The format device value is : ", formatDevice)
		args = []string{"-v", deviceMapperLocation}
		cmdOutput, err = runCommand("mkfs.ext4", args)
		if err != nil {
			log.Fatal("Error trying to format the luks volume. ", err)
		}
	}

	return true
}


func IsVolumeDeleted(deviceMapperLocation string) bool {
	if (deviceMapperLocation == "") || (len(deviceMapperLocation) <= 0) {
		log.Fatal("Invalid input parameters")
	}

	log.Println("Deleting the dm-crypt volume ", deviceMapperLocation)
	deleteVolumeCmd := "cryptsetup"
	args := []string{"luksClose", deviceMapperLocation}
	cmdOutput, err := runCommand(deleteVolumeCmd, args)
	if err != nil {
		log.Fatal("Error trying to close the device mapper. ", err)
	}	
	return true
}

func IsMount(deviceMapper string, mountLocation string) bool{
	if (deviceMapper == "") || (len(deviceMapper) <= 0) ||
	(mountLocation == "") || (len(mountLocation) <= 0) {
		log.Fatal("Invalid input parameters")
	}
	err := syscall.Mount(deviceMapper, mountLocation, "ext4", 0, "")
	if  err != nil {
		log.Fatal("Error while trying to mount. ", err)
	}
 return true
}

func IsUnmount(mountLocation string) bool{
	if (mountLocation == "") || (len(mountLocation) <= 0) {
		log.Fatal("Invalid input parameters")
	}
	err := syscall.Unmount(mountLocation, 0)
	if  err != nil {
		log.Fatal("Error while trying to unmount. ", err)
	}
	
	return true
}

func CreateVMManifest(vmId string, hostHardwareUuid string, imageId string, imageEncrypted bool) string {

	if (vmId == "") || (len(vmId) <= 0) ||
	 (hostHardwareUuid == "") || (len(hostHardwareUuid) <= 0) ||
	 (imageId == "") || (len(imageId) <= 0) {
		 log.Fatal("Invalid input parameters")
	 }
	
	manifest, err := GetVMManifest(vmId, hostHardwareUuid, imageId, imageEncrypted)
	if err != nil {
		log.Fatal("Error creating a manifest. ", err)
	}
	return manifest
}

func IsDecrypt(encImagePath, decPath, keyPath string) bool {

	// input validation
	if (encImagePath == "") || (len(encImagePath) <= 0) ||
	 (decPath == "") || (len(decPath) <= 0) ||
	 (keyPath == "") || (len(keyPath) <= 0) {
		 log.Fatal("Invalid input parameters")
	}

	//check if key file exists
	_, err := os.Stat(keyPath)
	if(os.IsNotExist(err)) {
		log.Fatal("Key does not exist. ", err)
	}

	// check if encrypted image file exists
	_, err = os.Stat(encImagePath)
	if(os.IsNotExist(err)) {
		log.Fatal("Encrypted file does not exist. ", err)
	}		

	// read the encrypted file
	data, _ := ioutil.ReadFile(encImagePath)
	plainText := decryptGCM(data, keyPath)

	if len(plainText) <= 0 {
		log.Fatal("Error during decryption of the file. ", err)
	}

	// write the decrypted output to file
	err = ioutil.WriteFile(decPath, plainText, 0644)
	if err != nil {
		log.Fatal("Error during writing to file. ", err)
	}
	return true
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
