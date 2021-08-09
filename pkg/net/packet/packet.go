package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
)

type Packet struct {
	// Id пакета
	Id uint16
	// Номер пакета
	Num       uint8
	encrypted bool
	data      []byte
	length    uint16
}

// Зашифровать пакет если он расшифрован
func (this *Packet) Encrypt() {
	if !this.encrypted {
		dataCrypt(this.data)
		this.encrypted = true
	}
}

// Расшифровать пакет если он зашифрован
func (this *Packet) Decrypt() {
	if this.encrypted {
		dataCrypt(this.data)
		this.encrypted = false
	}
}

// Зашифрован ли пакет
func (this *Packet) IsEncrypted() bool {
	return this.encrypted
}

// Длинна пакета
func (this *Packet) Length() uint16 {
	return this.length
}

// Представить пакет в виде hex-строки
func (this *Packet) Hex() string {
	return hex.EncodeToString(this.Bytes())
}

// Представить пакет в строковом виде
func (this *Packet) String() string {
	result := fmt.Sprintf("ID:       %v\n", this.Id)
	result += fmt.Sprintf("Length:   %v\n", this.length)
	result += fmt.Sprintf("Encrypt:  %v\n", this.encrypted)
	result += fmt.Sprintf("Num:      %v\n", this.Num)
	result += "\n"
	result += "Offset    01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16    ASCII\n\n"

	b := this.Bytes()
	bLen := len(b)
	maxRows := int(math.Ceil(float64(bLen) / float64(16)))

	for rNum := 0; rNum < maxRows; rNum++ {
		offset := rNum * 16
		bRow := b[offset : offset+16]
		result += fmt.Sprintf("%06d", rNum+1) + "    " + fmt.Sprintf("% x", bRow) + "    "
		for _, c := range bRow {
			if c >= 33 && c <= 126 {
				result += string(c)
			} else {
				result += "."
			}
		}
		if rNum != maxRows {
			result += "\n"
		}
	}
	return result
}

// Представить пакет в виде среза байт
func (this *Packet) Bytes() []byte {
	hb, err := encodePacketHeaders(packetHeaders{
		Id:        this.Id,
		Length:    this.length,
		Num:       this.Num,
		Encrypted: this.IsEncrypted(),
	})

	if err != nil {
		// ошибок быть не должно, поэтому если они все же появятся то паникуем
		panic(err)
	}

	if this.encrypted {
		headersCrypt(hb)
	}

	result := append(hb, this.data...)
	return result
}

// Считать данные из пакета в "data"
//
// "data" - то куда будет записан результат. Должны быть указателем на значение фиксированного размера или фрагмент значений фиксированного размера.
func (this *Packet) Read(data ...interface{}) error {
	r := bytes.NewReader(this.data)
	for _, d := range data {
		err := binary.Read(r, binary.LittleEndian, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Создает пакет с указаным "id" и записывает в него переданные "data"
//
// "id" - id пакета
// "data" - данные которые будут записаны в пакет. Должны быть значением фиксированного размера, фрагментом значений фиксированного размера или указателем на такие данные.
func CreatePacket(id uint16, data ...interface{}) (Packet, error) {
	packet := Packet{
		Id:     id,
		length: 6,
	}

	dBuf := new(bytes.Buffer)

	for _, d := range data {
		err := binary.Write(dBuf, binary.LittleEndian, d)
		if err != nil {
			return packet, err
		}
	}

	dBufBytes := dBuf.Bytes()
	dBufBytesLen := len(dBufBytes)

	if dBufBytesLen > 65529 {
		return packet, errors.New("Размер пакета не может превышать 65535 (uint16)")
	}

	packet.data = dBufBytes
	packet.length += uint16(dBufBytesLen)

	return packet, nil
}

// Обертка над CreatePacket, вызывающая панику в случае ошибки.
func CreatePacketOrPanic(id uint16, data ...interface{}) Packet {
	p, err := CreatePacket(id, data...)
	if err != nil {
		panic(err)
	}
	return p
}

// Создает пакет из среза байт
func CreatePacketFromBytes(b []byte) (Packet, error) {
	packet := Packet{}

	headers, err := decodePacketHeaders(b)

	if err != nil {
		return packet, err
	}

	packet.Id = headers.Id
	packet.Num = headers.Num

	packet.length = headers.Length
	packet.encrypted = headers.Encrypted
	packet.data = b[6:] // в data заголовки не нужны

	return packet, nil
}

// Обертка над CreatePacketFromBytes, вызывающая панику в случае ошибки.
func CreatePacketFromBytesOrPanic(b []byte) Packet {
	p, err := CreatePacketFromBytes(b)
	if err != nil {
		panic(err)
	}
	return p
}

// Создает пакет из hex-строки
func CreatePacketFromHexString(hexStr string) (Packet, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return Packet{}, err
	} else {
		return CreatePacketFromBytes(b)
	}
}

// Обертка над CreatePacketFromHexString, вызывающая панику в случае ошибки.
func CreatePacketFromHexStringOrPanic(hexStr string) Packet {
	packet, err := CreatePacketFromHexString(hexStr)
	if err != nil {
		panic(err)
	}
	return packet
}
