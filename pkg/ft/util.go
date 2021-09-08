package ft

import (
	"os"

	"github.com/materials-commons/mcft/pkg/protocol"
)

const McfsDefault = "/mcfs/data/materialscommons"

func Error2Status(err error) protocol.StatusResponse {
	return protocol.StatusResponse{}
}

func GetMCFSRoot() string {
	root := os.Getenv("MCFS_DIR")
	if root == "" {
		return McfsDefault
	}

	return root
}
