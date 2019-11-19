package internal

//const (
//	AddCmd          = "ADD"
//	DelCmd       = "DEL"
//	UpdateCmd       = "UPD"
//	GetCmd          = "GET"
//	SearchCmd       = "SRC"
//	CloseCmd        = "CLS"
//	CreateBucketCmd = "BKC"
//	DelBucketCmd = "BKD"
//	PingCmd         = "PNG"
//)

// Available commands
const (
	AddCmd = iota
	DelCmd
	UpdateCmd
	GetCmd
	SearchCmd
	CloseCmd
	CreateBucketCmd
	DelBucketCmd
	PingCmd
	ResultCmd
	ResultValueCmd
)
