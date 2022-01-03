# r2o-core

Базовое ядро пиратских серверов **R2 Online**, реализующее протокол общения между клиентом игры и сервером.

Пример простейшего логин-сервера:
```go
package main

import (
	"log"
	"strings"

	"github.com/tuxuuman/r2o-core/pkg/net"
	"github.com/tuxuuman/r2o-core/pkg/net/packet"
)

type GameserverInfo struct {
	// доступен ли сервера
	Online bool
	Id     uint16
	Name   [SERVER_NAME_LENGTH]byte
	// загруженность сервера
	Workload uint8
	// Ip
	Ip [4]byte
	// Порт
	Port uint16
	// Id списка в котором отображен сервере (1, 2)
	ListId uint32
	// Скрыт ли сервер (0 - видно, 1 - скрыт)
	Hidden uint32
}

const (
	// хз зачем столько байт выделено под название сервера
	SERVER_NAME_LENGTH = 101
)

func makeServerName(name string) [SERVER_NAME_LENGTH]byte {
	bname := [SERVER_NAME_LENGTH]byte{}

	copy(bname[:], []byte(name))

	return bname
}

var (
	GAMESERVERS = []GameserverInfo{
		{
			Ip:       [4]byte{127, 0, 0, 1},
			Id:       1,
			Port:     11005,
			Online:   true,
			ListId:   1,
			Hidden:   0,
			Workload: 1,
			Name:     makeServerName("Server 1"),
		},
		{
			Ip:       [4]byte{127, 0, 0, 2},
			Id:       2,
			Port:     11005,
			Online:   true,
			ListId:   2,
			Hidden:   0,
			Workload: 2,
			Name:     makeServerName("Server 2"),
		},
		{
			Ip:       [4]byte{127, 0, 0, 3},
			Id:       3,
			Port:     11005,
			Online:   false,
			ListId:   1,
			Hidden:   0,
			Workload: 2,
			Name:     makeServerName("Server 3"),
		},
	}
)

func main() {
	server := net.CreateServer("127.0.0.1", 11004)

	server.Start(func(c *net.Client) {
		type AuthPacketStruct struct {
			// не рашифрованная часть
			_ [937]byte

			// P4 параметр передаваемый в параметрах запуска. сюда можно будет передать некий токен и по нему найти юзера в базе/
			// на самом деле этот параметр может быть размером гораздо больше (2000+ байт). для примера беру только 64
			P4 [64]byte
		}

		// первый пакет присылаемый клиентом игры на логин-сервер, после разрешения подключения
		// в нем есть некоторые параметры запуска и еще какая-то инфа
		c.SetPacketHandler(3100, func(p *packet.Packet, data interface{}) {
			authData := data.(*AuthPacketStruct)
			P4 := strings.Trim(string(authData.P4[:]), "\000") // далее надо удалить нулевые символы
			log.Printf("Клиент %s хочет авторизоваться %v", c.IP(), P4)

			// тут уже надо искать по токену юзера в бд, сравнивать ip, проверять блокировку и тд.

			// для теста просто сравниваем токен со стратическим значением
			if P4 != "qwerty" {
				// шлем ошибку "не верный идентификатор сессии"
				c.FatalError(1812061665)
			} else {
				// тут по всей видисмости id аккаунта
				accountId := uint32(1)
				// с этим идентификатором игрок будет подключаться к игровому серверу
				sessionId := uint32(123456)
				c.SendPacket(packet.CreatePacketOrPanic(3101, accountId, sessionId, uint8(len(GAMESERVERS)), GAMESERVERS))
			}
		}, &AuthPacketStruct{}, true)

		// запрос на обновление списка игровых серваков
		c.SetPacketHandler(3115, func(p *packet.Packet, data interface{}) {
			// отправляем пакет со списком игровых серверов
			c.SendPacket(packet.CreatePacketOrPanic(3116, uint8(len(GAMESERVERS)), GAMESERVERS))
		}, nil, false)

		type Packet3120Struct struct {
			// Id сэссии который мы передаем в пакете 3101
			SessionId uint32
			// P0 параметр запуска игра (обычно тут логин)
			P0 [20]byte
			// Id сервера
			ServerId uint16
		}

		// запрос на подключение к игровому серверу
		c.SetPacketHandler(3120, func(p *packet.Packet, data interface{}) {
			pdata := data.(*Packet3120Struct)
			log.Println("Игрок хочет подключиться к игровому серверу", pdata.SessionId, pdata.SessionId, string(pdata.P0[:]))
			// тут можно сделать какие-то доп. проверки, после чего разрешить или запретить подключение
			// разрешаем подключение. после этого игрок отключится от логин-сервера и начнет подключение к игровому
			c.SendPacket(packet.CreatePacketOrPanic(3121, uint32(0))) // 0 - хз за что отвечает, но он должен быть
		}, &Packet3120Struct{}, true)

		// разрешаем подключение
		c.Accept()
	})
}

```

Далее запускаем игру с такими параметрами:
```cmd
.\R2ClientRU.exe "P0=cXdlcnR5&P1=Q19SMg==&P2=NDYxMg==&P4=cXdlcnR5&PC1=Tg==&PC2=Tg=="
```
И не забываем поменять ip и порт в R2.cfg
```
channelserverip = 127.0.0.1
channelserverport = 11004
```
