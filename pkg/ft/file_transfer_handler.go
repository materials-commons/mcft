package ft

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
	"github.com/materials-commons/gomcdb/mcmodel"
	"github.com/materials-commons/mcft/pkg/protocol"
	"gorm.io/gorm"
)

var ErrAlreadyAuthenticated = errors.New("already authenticated")
var ErrBadProtocolSequence = errors.New("bad protocol sequence")

var (
	basePath = "/home/gtarcea/uploads"
)

type FileTransferHandler struct {
	db *gorm.DB
	ws *websocket.Conn
	f  *os.File
}

func NewFileTransferHandler(ws *websocket.Conn, db *gorm.DB) *FileTransferHandler {
	return &FileTransferHandler{ws: ws, db: db}
}

func (h *FileTransferHandler) Run() error {
	defer h.close()

	if err := h.authenticate(); err != nil {
		return err
	}

	var incomingRequest protocol.IncomingRequestType

	for {
		if err := h.ws.ReadJSON(&incomingRequest); err != nil {
			//log.Errorf("Failed reading the incomingRequest: %s", err)
			break
		}

		var err error
		switch incomingRequest.RequestType {
		case protocol.AuthenticateReq:
			return ErrAlreadyAuthenticated
		case protocol.UploadFileReq:
			err = h.startUploadFile()
		case protocol.FileBlockReq:
			err = h.writeFileBlock()
		default:
			err = fmt.Errorf("unknown request type: %d", incomingRequest.RequestType)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (h *FileTransferHandler) close() {
	if h.f != nil {
		_ = h.f.Close()
	}
}

func (h *FileTransferHandler) authenticate() error {
	var incomingRequest protocol.IncomingRequestType
	if err := h.ws.ReadJSON(&incomingRequest); err != nil {
		return err
	}

	if incomingRequest.RequestType != protocol.AuthenticateReq {
		return errors.New("not authenticated")
	}

	var authReq protocol.AuthenticateRequest
	if err := h.ws.ReadJSON(&authReq); err != nil {
		return err
	}

	var user mcmodel.User

	return h.db.Where("api_token = ?", authReq.APIToken).First(&user).Error
}

func (h *FileTransferHandler) startUploadFile() error {
	var uploadReq protocol.UploadFileRequest
	if err := h.ws.ReadJSON(&uploadReq); err != nil {
		log.Errorf("Expected upload msg, got err: %s", err)
		return err
	}
	fullPath := filepath.Join(basePath, uploadReq.Path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0770); err != nil {
		log.Errorf("Unable to create directory: %s", err)
		return err
	}

	var err error

	h.f, err = os.Create(fullPath)
	if err != nil {
		log.Errorf("Unable to create file: %s", err)
	}

	return err
}

func (h *FileTransferHandler) writeFileBlock() error {
	if h.f == nil {
		return ErrBadProtocolSequence

	}

	var fileBlockReq protocol.FileBlockRequest

	if err := h.ws.ReadJSON(&fileBlockReq); err != nil {
		log.Errorf("Expected FileBlock msg, got err: %s", err)
		return err
	}

	// TODO: Put write into a loop to make sure we write all the blocks...
	n, err := h.f.Write(fileBlockReq.Block)
	if err != nil {
		log.Errorf("Failed writing to file: %s", err)
	}

	if n != len(fileBlockReq.Block) {
		log.Errorf("Did not write all of block, wrote %d, length %d", n, len(fileBlockReq.Block))
		err = errors.New("not all bytes written to file")
	}

	return err
}
