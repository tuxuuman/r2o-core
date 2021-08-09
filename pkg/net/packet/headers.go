package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Заголовки пакета.
//
// Последовательность полей должна быть такая, нельзя менять, иначе могут быть проблемы с чтением/записью
type packetHeaders struct {
	Length    uint16
	Encrypted bool
	Num       uint8
	Id        uint16
}

func decodePacketHeaders(b []byte) (packetHeaders, error) {
	headers := packetHeaders{}

	hLen := len(b)
	if hLen < 6 {
		return headers, errors.New("Минимальный размер заголовков пакета 6 байт")
	} else if hLen > 6 {
		b = b[:6]
	}

	if b[2] != 0 {
		headersCrypt(b)
	}

	err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &headers)

	if err != nil {
		return headers, err
	}

	return headers, nil
}

func encodePacketHeaders(h packetHeaders) ([]byte, error) {
	hBuffer := bytes.NewBuffer(make([]byte, 0, 6))
	err := binary.Write(hBuffer, binary.LittleEndian, h)

	if err != nil {
		return []byte{}, err
	}

	return hBuffer.Bytes(), nil
}
