# lib-go-volume-management
The Volume Measurement Library is used to perform some of the tasks during the VM launch as a part of VM integrity and confidentiality use case. Some of the tasks performed by VML are create the image and VM volumes, mount the image, decrypt the image, unmount an image, delete the dm-crypt volumes created and creation of VM manifest. 

pkg - contains the library source code
cmd - contains the main method which calls into the library