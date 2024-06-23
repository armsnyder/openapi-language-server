package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

// Split is a bufio.SplitFunc that splits JSON-RPC messages.
func Split(data []byte, _ bool) (advance int, token []byte, err error) {
	const headerDelimiter = "\r\n\r\n"
	const contentLengthPrefix = "Content-Length: "

	header, payload, found := bytes.Cut(data, []byte(headerDelimiter))
	if !found {
		return 0, nil, nil
	}

	contentLengthIndex := bytes.Index(header, []byte(contentLengthPrefix))
	if contentLengthIndex == -1 {
		return 0, nil, errors.New("missing content length: header not found")
	}

	contentLengthValueStart := contentLengthIndex + len(contentLengthPrefix)
	contentLengthValueLength := bytes.IndexByte(header[contentLengthValueStart:], '\r')
	if contentLengthValueLength == -1 {
		contentLengthValueLength = len(header) - contentLengthValueStart
	}

	contentLength, err := strconv.Atoi(string(header[contentLengthValueStart : contentLengthValueStart+contentLengthValueLength]))
	if err != nil {
		return 0, nil, errors.New("invalid content length")
	}

	if len(payload) < contentLength {
		return 0, nil, nil
	}

	return len(header) + len(headerDelimiter) + contentLength, payload[:contentLength], nil
}

// Write writes the given JSON-encodable message using the JSON-RPC protocol.
func Write(w io.Writer, message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return WritePayload(w, payload)
}

// WritePayload writes the given JSON payload using the JSON-RPC protocol.
func WritePayload(w io.Writer, payload []byte) error {
	packet := append([]byte("Content-Length: "+strconv.Itoa(len(payload))+"\r\n\r\n"), payload...)

	_, err := w.Write(packet)
	return err
}
