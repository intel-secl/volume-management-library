module intel/isecl/lib/vml

replace intel/isecl/lib/vml => ./

require (
	golang.org/x/sys v0.0.0-20181107165924-66b7b1311ac8
	intel/isecl/lib/common v0.0.0
)

replace intel/isecl/lib/common => github.com/intel-secl/common v1.6
