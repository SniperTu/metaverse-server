package message

import (
	"fmt"

	"github.com/bluenviron/mediamtx/internal/protocols/rtmp/rawmessage"
)

// ExtendedMetadata is a metadata extended message.
type ExtendedMetadata struct {
	FourCC FourCC
}

func (m *ExtendedMetadata) unmarshal(raw *rawmessage.Message) error {
	if len(raw.Body) != 5 {
		return fmt.Errorf("invalid body size")
	}

	m.FourCC = FourCC(raw.Body[1])<<24 | FourCC(raw.Body[2])<<16 | FourCC(raw.Body[3])<<8 | FourCC(raw.Body[4])

	return fmt.Errorf("ExtendedMetadata is not implemented yet")
}

func (m ExtendedMetadata) marshal() (*rawmessage.Message, error) {
	return nil, fmt.Errorf("TODO")
}
