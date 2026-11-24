package dockerutil

import (
	"bufio"
	"encoding/binary"
	"io"
)

// maxLogFrameSize is a safety cap (10MB) for multiplexed Docker log frames.
// This prevents corrupted streams or malformed headers from allocating gigabytes of RAM.
const maxLogFrameSize = 10 * 1024 * 1024

// StreamLogs safely reads Docker logs from reader, handling both TTY (raw text)
// and non-TTY (multiplexed 8-byte headers).
// If a non-TTY stream contains unmultiplexed raw text or malformed headers, it gracefully
// falls back to line-based parsing rather than misinterpreting ASCII text as giant uint32
// byte sizes (which previously caused 1.28GB memory spikes and 100% CPU starvation).
func StreamLogs(reader io.Reader, isTty bool, onPayload func(payload []byte) error) error {
	bufReader := bufio.NewReader(reader)

	if isTty {
		for {
			line, err := bufReader.ReadBytes('\n')
			if len(line) > 0 {
				if cbErr := onPayload(line); cbErr != nil {
					return cbErr
				}
			}
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}

	header := make([]byte, 8)
	for {
		_, err := io.ReadFull(bufReader, header)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Docker stdcopy multiplex header spec:
		// header[0] is stream type: 0 (stdin), 1 (stdout), 2 (stderr)
		// header[1], header[2], header[3] MUST be 0x00
		if header[0] <= 2 && header[1] == 0 && header[2] == 0 && header[3] == 0 {
			size := binary.BigEndian.Uint32(header[4:8])
			if size > maxLogFrameSize {
				// Corrupted frame size or huge single frame, abort to prevent OOM
				return nil
			}
			payload := make([]byte, size)
			if _, err := io.ReadFull(bufReader, payload); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			if cbErr := onPayload(payload); cbErr != nil {
				return cbErr
			}
		} else {
			// Not a multiplex header! The container output is raw/unmultiplexed text.
			// Treat the 8 bytes already consumed as text, and read until the next newline.
			rest, err := bufReader.ReadBytes('\n')
			line := append(header, rest...)
			if cbErr := onPayload(line); cbErr != nil {
				return cbErr
			}
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}

