package cmd

import (
	"encoding/binary"
	"io"
)

type flagData struct {
	TargetTeam int
	Value      string
	Service    string
}

type submitData struct {
	Token     string
	Submitter string
	Flags     []flagData
}

func writeString(w io.Writer, s string) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(s)))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	_, err := w.Write([]byte(s))
	return err
}

func writeFlagData(w io.Writer, f flagData) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(f.TargetTeam))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	if err := writeString(w, f.Value); err != nil {
		return err
	}

	return writeString(w, f.Service)
}

func writeSubmitData(w io.Writer, d submitData) error {
	if err := writeString(w, d.Token); err != nil {
		return err
	}

	if err := writeString(w, d.Submitter); err != nil {
		return err
	}

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(d.Flags)))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	for _, f := range d.Flags {
		if err := writeFlagData(w, f); err != nil {
			return err
		}
	}

	return nil
}
