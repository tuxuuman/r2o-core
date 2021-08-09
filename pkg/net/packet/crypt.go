package packet

import (
	"log"

	"github.com/tuxuuman/r2o-core/resources"
)

// Ключ шифрования
var dataCryptKey = resources.PACKET_CRYPT_KEY[6:]
var dataCryptKeyLength = len(dataCryptKey)
var headersCryptKey = resources.PACKET_CRYPT_KEY[:6]
var headersCryptKeyLength = len(dataCryptKey)

// Кодирует/декодирует данные. Первый вызов кодирует, второй декодирует, или наоборот.
//
// Алгоритм простейший, просто XOR-ит все байты ключем шифрования по порядку.
// В самой R2 это вероятно выполнено по какому-то алгоритму шифрования, я так и не понял по какому, поэтому пока так.
//
// Из-за ограниченной длинны ключа (2991 байт), пакеты могут быть расшифрованы/зашифрованы не полностью.
func dataCrypt(data []byte) {
	dLen := len(data)

	if dLen > dataCryptKeyLength {
		log.Printf("Размер пакета [%v] превышает размер ключа шифрования [%v]. Шифрование/расшифровка выполнена не полностью", dLen, dataCryptKeyLength)
		dLen = dataCryptKeyLength
	}

	for i := 0; i < dLen; i++ {
		data[i] ^= dataCryptKey[i]
	}
}

func headersCrypt(headers []byte) {
	hLen := len(headers)

	if hLen > headersCryptKeyLength {
		log.Printf("Размер пакета [%v] превышает размер ключа шифрования [%v]. Шифрование/расшифровка выполнена не полностью", hLen, headersCryptKeyLength)
		hLen = headersCryptKeyLength
	}

	for i := 0; i < hLen; i++ {
		headers[i] ^= headersCryptKey[i]
	}
}
