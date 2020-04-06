package notes

import (
	"fmt"
	"os"
	"time"

	"storj.io/uplink"
)

// Note is what is stored.
type Note struct {
	NoteMeta
	Message string
}

// NoteMeta is information without the message.
type NoteMeta struct {
	Identifier string
	Uploaded   time.Time
}

const (
	uploadTime = "notes:upload-time"
)

// ParseNote parses note from data and info.
func ParseNote(info *uplink.Object, data []byte) Note {
	var note Note
	note.NoteMeta = ParseNoteMeta(info)
	note.Message = string(data)
	return note
}

// ParseNoteMeta parses note metadata from info.
func ParseNoteMeta(info *uplink.Object) NoteMeta {
	var meta NoteMeta
	meta.Identifier = info.Key
	if info.Custom != nil {
		if s, ok := info.Custom[uploadTime]; ok {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse upload time %q\n", s)
			} else {
				meta.Uploaded = t
			}
		}
	}

	return meta
}
