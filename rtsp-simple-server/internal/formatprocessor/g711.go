package formatprocessor //nolint:dupl

import (
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtplpcm"
	"github.com/bluenviron/gortsplib/v4/pkg/rtptime"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/unit"
)

type formatProcessorG711 struct {
	udpMaxPayloadSize int
	format            *format.G711
	timeEncoder       *rtptime.Encoder
	encoder           *rtplpcm.Encoder
	decoder           *rtplpcm.Decoder
}

func newG711(
	udpMaxPayloadSize int,
	forma *format.G711,
	generateRTPPackets bool,
) (*formatProcessorG711, error) {
	t := &formatProcessorG711{
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

func (t *formatProcessorG711) createEncoder() error {
	t.encoder = &rtplpcm.Encoder{
		PayloadMaxSize: t.udpMaxPayloadSize - 12,
		PayloadType:    t.format.PayloadType(),
		BitDepth:       8,
		ChannelCount:   t.format.ChannelCount,
	}
	return t.encoder.Init()
}

func (t *formatProcessorG711) ProcessUnit(uu unit.Unit) error { //nolint:dupl
	u := uu.(*unit.G711)

	pkts, err := t.encoder.Encode(u.Samples)
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

func (t *formatProcessorG711) ProcessRTPPacket( //nolint:dupl
	pkt *rtp.Packet,
	ntp time.Time,
	pts time.Duration,
	hasNonRTSPReaders bool,
) (Unit, error) {
	u := &unit.G711{
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

		samples, err := t.decoder.Decode(pkt)
		if err != nil {
			return nil, err
		}

		u.Samples = samples
	}

	// route packet as is
	return u, nil
}
