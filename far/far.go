package far

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Version string

const (
	Version1A Version = "1A"
	Version1B Version = "1B"
	Version3  Version = "3"
	VersionUnknown Version = "Unknown"
)

type ManifestEntry struct {
	DecompressedSize uint32
	CompressedSize   uint32
	FileOffset       uint32
	Filename         string
}

type Reader struct {
	rs      io.ReadSeeker
	closer  io.Closer
	Version Version
	Entries []ManifestEntry
}

func NewReader(rs io.ReadSeeker) (*Reader, error) {
	r := &Reader{rs: rs}
	if err := r.readHeader(); err != nil {
		return nil, err
	}
	return r, nil
}

func Open(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, err := NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	r.closer = f

	return r, nil
}

func (r *Reader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

func (r *Reader) readHeader() error {
	signature := make([]byte, 8)
	if _, err := io.ReadFull(r.rs, signature); err != nil {
		return err
	}
	if string(signature) != "FAR!byAZ" {
		return fmt.Errorf("invalid signature: %s", string(signature))
	}

	var version int32
	if err := binary.Read(r.rs, binary.LittleEndian, &version); err != nil {
		return err
	}

	var manifestOffset uint32
	if err := binary.Read(r.rs, binary.LittleEndian, &manifestOffset); err != nil {
		return err
	}

	if _, err := r.rs.Seek(int64(manifestOffset), io.SeekStart); err != nil {
		return err
	}

	var numFiles uint32
	if err := binary.Read(r.rs, binary.LittleEndian, &numFiles); err != nil {
		return err
	}

	r.Entries = make([]ManifestEntry, 0, numFiles)
	detected1B := false

	for i := uint32(0); i < numFiles; i++ {
		entry, is1B, err := r.readManifestEntry(version)
		if err != nil {
			return err
		}
		r.Entries = append(r.Entries, entry)
		if is1B {
			detected1B = true
		}
	}

	if version == 1 {
		if detected1B {
			r.Version = Version1B
		} else {
			r.Version = Version1A
		}
	} else if version == 3 {
		r.Version = Version3
	} else {
		r.Version = VersionUnknown
	}

	return nil
}

func (r *Reader) readManifestEntry(version int32) (ManifestEntry, bool, error) {
	entryStart, _ := r.rs.Seek(0, io.SeekCurrent)
	
	var entry ManifestEntry
	if err := binary.Read(r.rs, binary.LittleEndian, &entry.DecompressedSize); err != nil {
		return entry, false, err
	}
	if err := binary.Read(r.rs, binary.LittleEndian, &entry.CompressedSize); err != nil {
		return entry, false, err
	}
	if err := binary.Read(r.rs, binary.LittleEndian, &entry.FileOffset); err != nil {
		return entry, false, err
	}

	is1B := false
	var filenameLen uint32

	if version == 1 {
		currentPos, _ := r.rs.Seek(0, io.SeekCurrent)
		if _, err := r.rs.Seek(entryStart+14, io.SeekStart); err != nil {
			return entry, false, err
		}
		var val16 uint16
		if err := binary.Read(r.rs, binary.LittleEndian, &val16); err != nil {
			return entry, false, err
		}
		r.rs.Seek(currentPos, io.SeekStart)

		if val16 > 0 {
			is1B = true
			var len16 uint16
			if err := binary.Read(r.rs, binary.LittleEndian, &len16); err != nil {
				return entry, false, err
			}
			filenameLen = uint32(len16)
		} else {
			is1B = false
			if err := binary.Read(r.rs, binary.LittleEndian, &filenameLen); err != nil {
				return entry, false, err
			}
		}
	} else {
		// Default to 4-byte length for other versions
		if err := binary.Read(r.rs, binary.LittleEndian, &filenameLen); err != nil {
			return entry, false, err
		}
	}

	nameBuf := make([]byte, filenameLen)
	if _, err := io.ReadFull(r.rs, nameBuf); err != nil {
		return entry, false, err
	}
	entry.Filename = string(nameBuf)

	return entry, is1B, nil
}

func (r *Reader) Extract(entry ManifestEntry, destDir string) error {
	if _, err := r.rs.Seek(int64(entry.FileOffset), io.SeekStart); err != nil {
		return err
	}

	normalizedName := strings.ReplaceAll(entry.Filename, "\\", "/")
	destPath := filepath.Join(destDir, normalizedName)
	
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.CopyN(out, r.rs, int64(entry.CompressedSize))
	return err
}
