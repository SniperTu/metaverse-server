package formatprocessor //nolint:dupl

import (
	"errors"
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtpmpeg1audio"
	"github.com/bluenviron/gortsplib/v4/pkg/rtptime"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/unit"
)

type formatProcessorMPEG1Audio struct {
	udpMaxPayloadSize int
	format            *format.MPEG1Audio
	timeEncoder       *rtptime.Encoder
	encoder           *rtpmpeg1audio.Encoder
	decoder           *rtpmpeg1audio.Decoder
}

func newMPEG1Audio(
	udpMaxPayloadSize int,
	forma *format.MPEG1Audio,
	generateRTPPackets bool,
) (*formatProcessorMPEG1Audio, error) {
	t := &formatProcessorMPEG1Audio{
		udpMaxPayloadSize: udpMaxPayloadSize,
		format:            forma,
	}

	if generateRTPPackets {
		err := t.createEncoder()
		if err != nil {
			return nil, err
		}

		t.timeEncoder = &rtptime.Encoder{
			ClockRate: forma.ClockRate(),
		}
		err = t.timeEncoder.Initialize()
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (t *formatProcessorMPEG1Audio) createEncoder() error {
	t.encoder = &rtpmpeg1audio.Encoder{
		PayloadMaxSize: t.udpMaxPayloadSize - 12,
	}
	return t.encoder.Init()
}

func (t *formatProcessorMPEG1Audio) ProcessUnit(uu unit.Unit) error { //nolint:dupl
	u := uu.(*unit.MPEG1Audio)

	pkts, err := t.encoder.Encode(u.Frames)
	if err != nil {
		return err
	}
	u.RTPPackets = pkts

	ts := t.timeEncoder.Encode(u.PTS)
	for _, pkt := range u.RTPPackets {
		pkt.Timestamp += ts
	}

	return nil
}

func (t *formatProcessorMPEG1Audio) ProcessRTPPacket( //nolint:dupl
	pkt *rtp.Packet,
	ntp time.Time,
	pts time.Duration,
	hasNonRTSPReaders bool,
) (Unit, error) {
	u := &unit.MPEG1Audio{
		Base: unit.Base{
			RTPPackets: []*rtp.Packet{pkt},
			NTP:        ntp,
			PTS:        pts,
		},
	}

	// remove padding
	pkt.Header.Padding = false
	pkt.PaddingSize = 0

	if pkt.MarshalSize() > t.udpMaxPayloadSize {
		return nil, fmt.Errorf("payload size (%d) is greater than maximum allowed (%d)",
			pkt.MarshalSize(), t.udpMaxPayloadSize)
	}

	// decode from RTP
	if hasNonRTSPReaders || t.decoder != nil {
		if t.decoder == nil {
			var err error
			t.decoder, err = t.format.CreateDecoder()
			if err != nil {
				return nil, err
			}
		}

		frames, err := t.decoder.Decode(pkt)
		if err != nil {
			if errors.Is(err, rtpmpeg1audio.ErrNonStartingPacketAndNoPrevious) ||
				errors.Is(err, rtpmpeg1audio.ErrMorePacketsNeeded) {
				return u, nil
			}
			return nil, err
		}

		u.Frames = frames
	}

	// route packet as is
	return u, nil
}
