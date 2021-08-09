package packet

import (
	"encoding/hex"
	"testing"

	"github.com/tuxuuman/r2o-core/pkg/net/packet"
)

const ETALON_PACKET_HEX = "0a0000001e0cd489956e"

func decodeHexStringOrPanic(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	} else {
		return b
	}
}

func TestCreatePacket(t *testing.T) {
	p := packet.CreatePacketOrPanic(3102, uint32(1855293908))
	pHex := p.Hex()

	if pHex != ETALON_PACKET_HEX {
		t.Fatalf("Сгенерирован неправильный HEX. Ожидаемый результат: %v. Полученый результат: %v", ETALON_PACKET_HEX, pHex)
	}

	pBytes := hex.EncodeToString(p.Bytes())

	if pBytes != ETALON_PACKET_HEX {
		t.Fatalf("Сгенерирована неправильная последовательность байт. Ожидаемый результат: %v. Полученый результат: %v", ETALON_PACKET_HEX, pBytes)
	}
}

func TestCreatePacketFromHexString(t *testing.T) {
	p := packet.CreatePacketFromHexStringOrPanic(ETALON_PACKET_HEX)
	pHex := p.Hex()

	if pHex != ETALON_PACKET_HEX {
		t.Fail()
	}
}

func TestCreatePacketFromBytes(t *testing.T) {
	p := packet.CreatePacketFromBytesOrPanic(decodeHexStringOrPanic(ETALON_PACKET_HEX))
	pHex := p.Hex()

	if pHex != ETALON_PACKET_HEX {
		t.Fail()
	}
}

func TestReadPacket(t *testing.T) {
	p := packet.CreatePacketFromHexStringOrPanic("0a0000001e0cd489956edc05")

	var num1 uint32
	var num2 uint16

	err := p.Read(&num1, &num2)

	if err != nil {
		t.Fatal(err)
	}

	if num1 != 1855293908 || num2 != 1500 {
		t.Fatal("Считаны не правильные числа", num1, num2)
	}
}
