package codec

// Message used for processing all messages
type Message struct {
	Command string
	Bucket  string

	Key   string
	Value string

	Result bool
	Data   string
}

// AuthMessage contains credentials
type AuthMessage struct {
	Login, Password string
}

