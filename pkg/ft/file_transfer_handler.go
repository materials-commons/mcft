package ft

import (
	"os"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
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
	return nil
}

func (h *FileTransferHandler) Close() {
	if h.f != nil {
		_ = h.f.Close()
	}
}
