package submitter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type FlagData struct {
	TargetTeam int
	Value      string
	Service    string
}

type SubmitData struct {
	Flags []FlagData
}

func WriteString(w io.Writer, s string) error {
	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, uint64(len(s)))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	_, err := w.Write([]byte(s))
	return err
}

func ReadString(r *bufio.Reader) (string, error) {
	l, err := binary.ReadUvarint(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, l)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func WriteSubmitData(w io.Writer, d SubmitData) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(d.Flags)))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	for _, f := range d.Flags {
		if err := WriteFlagData(w, f); err != nil {
			return err
		}
	}
	return nil
}

func ReadSubmitData(r *bufio.Reader) (SubmitData, error) {
	var d SubmitData
	count, err := binary.ReadUvarint(r)

	if err != nil {
		return d, err
	}

	d.Flags = make([]FlagData, count)
	for i := uint64(0); i < count; i++ {
		f, err := ReadFlagData(r)

		if err != nil {
			return d, err
		}

		d.Flags[i] = f
	}

	return d, nil
}

func WriteFlagData(w io.Writer, f FlagData) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(f.TargetTeam))

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}

	if err := WriteString(w, f.Value); err != nil {
		return err
	}

	return WriteString(w, f.Service)
}

func ReadFlagData(r *bufio.Reader) (FlagData, error) {
	var f FlagData
	targetTeam, err := binary.ReadUvarint(r)

	if err != nil {
		return f, err
	}

	f.TargetTeam = int(targetTeam)
	value, err := ReadString(r)

	if err != nil {
		return f, err
	}

	f.Value = value
	service, err := ReadString(r)

	if err != nil {
		return f, err
	}

	f.Service = service
	return f, nil
}

func WriteAuth(w io.Writer, token, submitter string) error {
	if err := WriteString(w, token); err != nil {
		return err
	}
	return WriteString(w, submitter)
}

func ReadAuth(r *bufio.Reader) (token, submitter string, err error) {
	token, err = ReadString(r)
	if err != nil {
		return "", "", err
	}
	submitter, err = ReadString(r)
	if err != nil {
		return "", "", err
	}
	return token, submitter, nil
}

func WriteAuthResponse(w io.Writer, ok bool, msg string) error {
	var status byte
	if ok {
		status = 1
	}
	if _, err := w.Write([]byte{status}); err != nil {
		return err
	}
	if !ok {
		return WriteString(w, msg)
	}
	return nil
}

func ReadAuthResponse(r *bufio.Reader) (ok bool, msg string, err error) {
	status := make([]byte, 1)
	if _, err := io.ReadFull(r, status); err != nil {
		return false, "", fmt.Errorf("read auth status: %w", err)
	}
	if status[0] != 1 {
		msg, err = ReadString(r)
		if err != nil {
			return false, "", fmt.Errorf("read auth error msg: %w", err)
		}
		return false, msg, nil
	}
	return true, "", nil
}
