package protocol

type MsgType int

const (
	Login MsgType = iota
	SetProject
	SendStat
	SendChecksum
	StatInfo
	ChecksumInfo
	SetPosition
	Upload
	FileBlock
	FinishUpload
	Download
)

type UploadType int

const (
	FileType UploadType = iota
	DirType
)

// Command is the next command
type CommandMsg struct {
	MsgType MsgType `json:"msg_type"`
}

type LoginMsg struct {
	APIToken string `json:"api_token"`
}

type SetProjectMsg struct {
	ProjectID int `json:"project_id"`
}

type SendStatMsg struct {
	Path string `json:"path"`
}

type SendChecksumMsg struct {
	Path string `json:"path"`
}

type StatInfoMsg struct {
	Size int `json:"size"`
}

type UploadMsg struct {
	Path string     `json:"path"`
	Size int        `json:"size"`
	Type UploadType `json:"upload_type"`
}

type FileBlockMsg struct {
	Block    []byte `json:"block"`
	Checksum string `json:"checksum"`
}

type FinishUploadMsg struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

type DownloadMsg struct {
	Path string `json:"path"`
}
