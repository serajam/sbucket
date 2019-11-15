package internal

//const (
//	AddCommand          = "ADD"
//	DeleteCommand       = "DEL"
//	UpdateCommand       = "UPD"
//	GetCommand          = "GET"
//	SearchCommand       = "SRC"
//	CloseCommand        = "CLS"
//	CreateBucketCommand = "BKC"
//	DeleteBucketCommand = "BKD"
//	PingCommand         = "PNG"
//)

// Available commands
const (
	AddCommand = iota
	DeleteCommand
	UpdateCommand
	GetCommand
	SearchCommand
	CloseCommand
	CreateBucketCommand
	DeleteBucketCommand
	PingCommand
	ResultCommand
)
