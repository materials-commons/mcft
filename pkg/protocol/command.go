package protocol

import "time"

type RequestType int

const (
	AuthenticateReq RequestType = iota
	DownloadReq
	FileInfoReq
	FinishUploadReq
	ListDirectoryReq
	PauseUploadReq
	FileBlockReq
	ServerInfoReq
	UploadFileReq
)

type Version struct {
	Version string `json:"version"`
}

type IncomingRequestType struct {
	RequestType RequestType `json:"request_type"`
}

type AuthenticateRequest struct {
	APIToken string `json:"apitoken"`
	Version
}

type DownloadRequest struct {
	Path string `json:"path"`
	Version
}

type FileInfoRequest struct {
	Path string `json:"path"`
	Version
}

type FileInfoResponse struct {
	UploadOffset      int64     `json:"upload_offset"`
	CurrentChecksum   string    `json:"current_checksum"`
	ChecksumAlgorithm string    `json:"checksum_algorithm"`
	ExpiresAt         time.Time `json:"expires_at"`
	Version
}

type FinishUploadRequest struct {
	Path         string `json:"path"`
	FileChecksum string `json:"file_checksum"`
	Version
}

type FileInfo struct {
	Name              string    `json:"name"`
	IsDir             bool      `json:"is_dir"`
	Size              int64     `json:"size"`
	Checksum          string    `json:"checksum"`
	ChecksumAlgorithm string    `json:"checksum_algorithm"`
	UploadComplete    bool      `json:"upload_complete"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ListDirectoryResponse struct {
	Path   string     `json:"path"`
	Status string     `json:"status"`
	Files  []FileInfo `json:"files"`
	Version
}

type PauseUploadRequest struct {
	Path string `json:"path"`
	Version
}

type FileBlockRequest struct {
	Path              string `json:"path"`
	Block             []byte `json:"block"`
	ContentType       string `json:"content_type"`
	ContentLength     int64  `json:"content_length"`
	UploadOffset      int64  `json:"upload_offset"`
	Checksum          string `json:"checksum"`
	ChecksumAlgorithm string `json:"check_algorithm"`
	Version
}

type ServerInfoResponse struct {
	MaxSize                 int64    `json:"max_size"`
	ChecksumAlgorithms      []string `json:"checksum_algorithms"`
	BlockChecksumsSupported bool     `json:"block_checksums_supported"`
	UploadExpirationTime    int      `json:"upload_expiration_time"`
	Version
}

type StatusResponse struct {
	Path           string `json:"path"`
	ForRequestType string `json:"for_request_type"`
	Status         string `json:"status"`
	Version
}

type UploadFileRequest struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Version
}
