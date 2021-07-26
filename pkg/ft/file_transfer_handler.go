package ft

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
	"github.com/materials-commons/gomcdb/mcmodel"
	"github.com/materials-commons/gomcdb/store"
	"github.com/materials-commons/mcft/pkg/protocol"
	"gorm.io/gorm"
)

var ErrAlreadyAuthenticated = errors.New("already authenticated")
var ErrBadProtocolSequence = errors.New("bad protocol sequence")
var ErrNotAuthenticated = errors.New("not authenticated")

type FileTransferHandler struct {
	db           *gorm.DB
	ws           *websocket.Conn
	f            *os.File
	ProjectID    int
	User         mcmodel.User
	projectStore *store.ProjectStore
	fileStore    *store.FileStore
	mcfsRoot     string
}

func NewFileTransferHandler(ws *websocket.Conn, db *gorm.DB) *FileTransferHandler {
	return &FileTransferHandler{
		ws:           ws,
		db:           db,
		projectStore: store.NewProjectStore(db),
		fileStore:    store.NewFileStore(db, GetMCFSRoot()),
		mcfsRoot:     GetMCFSRoot(),
	}
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
		// Here is where we need to finish closing the file in the database by:
		//   1. Setting it to the current file
		//   2. Updating its size info
		//   3. Kicking off a job to do a conversion (if appropriate)
		//   4. Updating counts (by asking Laravel to do that in an update request)
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

	if err := h.db.Where("api_token = ?", authReq.APIToken).First(&user).Error; err != nil {
		return err
	}

	h.User = user

	if !h.projectStore.UserCanAccessProject(h.User.ID, authReq.ProjectID) {
		return ErrNotAuthenticated
	}

	h.ProjectID = authReq.ProjectID

	return nil
}

func (h *FileTransferHandler) startUploadFile() error {
	var (
		uploadReq protocol.UploadFileRequest
		file      *mcmodel.File
	)
	if err := h.ws.ReadJSON(&uploadReq); err != nil {
		log.Errorf("Expected upload msg, got err: %s", err)
		return err
	}

	dir, err := h.fileStore.FindDirByPath(h.ProjectID, filepath.Dir(uploadReq.Path))
	if err != nil {
		dir, err = h.CreateDirectoryAll(filepath.Dir(uploadReq.Path))
		if err != nil {
			return err
		}

		_ = os.MkdirAll(dir.ToUnderlyingDirPath(h.mcfsRoot), 0770)
	}

	name := filepath.Base(uploadReq.Path)
	file, err = h.fileStore.CreateFile(name, h.ProjectID, dir.ID, h.User.ID, getMimeType(name))
	if err != nil {
		return err
	}

	h.f, err = os.Create(file.ToUnderlyingFilePath(h.mcfsRoot))
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

func (h *FileTransferHandler) CreateDirectoryAll(dir string) (*mcmodel.File, error) {
	return nil, nil
}

// getMimeType will determine the type of a file from its extension. It strips out the extra information
// such as the charset and just returns the underlying type. It returns "unknown" for the mime type if
// the mime package is unable to determine the type.
func getMimeType(name string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	if mimeType == "" {
		return "unknown"
	}

	if mediaType, _, err := mime.ParseMediaType(mimeType); err == nil {
		// If err is nil then we can returned the parsed mediaType
		return mediaType
	}

	// If we are here then ParseMediaType returned an error, so brute force separating
	// the string to get the media type
	semicolon := strings.Index(mimeType, ";")
	if semicolon == -1 {
		return strings.TrimSpace(mimeType)
	}

	return strings.TrimSpace(mimeType[:semicolon])
}
