# volume-management library (VML)
The Volume Management Library is used to perform some of the tasks during the VM launch as a part of VM integrity and confidentiality use case. Some of the tasks performed by VML are create the image and VM volumes, mount the image, decrypt the image, unmount an image, delete the dm-crypt volumes created and creation of VM manifest. 

pkg - contains the library source code
cmd - contains the main method which calls into the library

## Key features
- Create dm-crypt volume
- Delete dm-crypt volume
- Mount a device
- Unmount a device
- Decrypt a file
- Create VM manifest
- Create container manifest


## System Requirements
- RHEL 7.5/7.6
- Epel 7 Repo
- Proxy settings if applicable

## Software requirements
- git
- Go 11.4 or newer

# Step By Step Build Instructions

### Install `go 1.11.4` or newer
The `VML` requires Go version 11.4 that has support for `go modules`. The build was validated with version 11.4 version of `go`. It is recommended that you use a newer version of `go` - but please keep in mind that the product has been validated with 1.11.4 and newer versions of `go` may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
tar -xzf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

# Third Party Dependencies

## VML

### Direct dependencies

| Name                  | Repo URL           | Minimum Version Required           |
| ----------------------| -------------------| :--------------------------------: |
| system commands       | golang.org/x/sys   | v0.0.0-20181107165924-66b7b1311ac8 |


*Note: All dependencies are listed in go.mod*

# Links
https://01.org/intel-secl/