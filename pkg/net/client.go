package net

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"

	"github.com/tuxuuman/r2o-core/pkg/net/packet"
	"github.com/tuxuuman/r2o-core/resources"
)

var acceptConnectionPacket packet.Packet = packet.CreatePacketFromBytesOrPanic(resources.ACP_PACKET)

type PacketHandler = func(p packet.Packet)

type Client struct {
	conn            net.Conn
	accepted        bool
	rejected        bool
	writePacketChan chan packet.Packet
	readPacketChan  chan packet.Packet
	isClosed        bool
	disconChan      bool
	ip              string
	id              uint16
}

func (this *Client) close() {
	if !this.isClosed {
		this.isClosed = true
		this.conn.Close()
		close(this.writePacketChan)
		close(this.readPacketChan)
		log.Printf("Клиент %v отключился", this.ip)
	}
}

func createFatalErrorPacket(erorrId uint32) packet.Packet {
	return packet.CreatePacketOrPanic(3102, uint32(erorrId))
}

func createErrorPacket(packetId uint16, erorrId uint32, code uint32) packet.Packet {
	return packet.CreatePacketOrPanic(1102, uint16(packetId), uint32(erorrId), uint32(code))
}

func (this *Client) sendPacket(p packet.Packet) {
	log.Print("\n\n->->->->->->->->->->->->->->->->\n\n", fmt.Sprintf("Исходящий пакет для [%v:%v]\n", this.ip, this.id), p.String(), "\n->->->->->->->->->->->->->->->->\n\n")
	_, err := this.conn.Write(p.Bytes())
	if err != nil {
		log.Printf(fmt.Sprintf("Не удалось отправить пакет [ID=%v]", p.Id), err)
	}
}

func (this *Client) startPacketWriter() {
	for p := range this.writePacketChan {
		this.sendPacket(p)
	}
}

func (this *Client) startPacketReader() {
	var err error

	for {
		bufLen := make([]byte, 2)
		_, err = io.ReadFull(this.conn, bufLen)

		if err != nil {
			break
		}

		bufPac := make([]byte, binary.LittleEndian.Uint16(bufLen)-2)
		_, err = io.ReadFull(this.conn, bufPac)

		if err != nil {
			break
		}

		var p packet.Packet
		p, err = packet.CreatePacketFromBytes(append(bufLen, bufPac...))

		if err != nil {
			break
		}

		if p.IsEncrypted() {
			p.Decrypt()
		}

		log.Print("\n\n<-<-<-<-<-<-<-<-<-<-<-<-<-<-<-<-\n\n", fmt.Sprintf("Входящий пакет от [%v:%v]", this.ip, this.id), p.String(), "\n<-<-<-<-<-<-<-<-<-<-<-<-<-<-<-<-\n\n")
		this.readPacketChan <- p
	}

	if err != nil && err != io.EOF {
		log.Printf("При обработки пакетов клиента [%v] произошла ошибка.", this.ip)
		log.Println(string(debug.Stack()))
	}

	this.close()
}

func (this *Client) ID() uint16 {
	return this.id
}

func (this *Client) IP() string {
	return this.ip
}

func (this *Client) SendPacket(p packet.Packet) {
	if !this.isClosed {
		this.writePacketChan <- p
	}
}

// Разрешить подключение клиента и начать принимать пакеты.
//
// После подключения клиента обязательно нужно вызвать этот метод или метод Reject, иначе Reject будет вызван автоматически, спустя некоторое время.
func (this *Client) Accept() {
	if this.rejected {
		panic(errors.New(fmt.Sprintf("Нельзя принять подключение которое уже отклонено. [ID = %v] [IP = %v]", this.id, this.ip)))
	} else if this.accepted {
		panic(errors.New(fmt.Sprintf("Подключение уже принято. [ID = %v] [IP = %v]", this.id, this.ip)))
	} else {
		this.accepted = true
	}

	go this.startPacketWriter()
	go this.startPacketReader()

	this.SendPacket(acceptConnectionPacket)
}

// Отклонить подключение клиента, послав ему ошибку и отключив его
//
// После подключения клиента обязательно нужно вызвать этот метод или метод Accept, иначе Reject будет вызван автоматически, спустя некоторое время.
func (this *Client) Reject(reason uint32) {
	if this.accepted {
		panic(errors.New(fmt.Sprintf("Нельзя отклонить подключение которое уже принято. [ID = %v] [IP = %v]", this.id, this.ip)))
	} else if this.rejected {
		panic(errors.New(fmt.Sprintf("Подключение уже отклонено. [ID = %v] [IP = %v]", this.id, this.ip)))
	} else {
		this.rejected = true
	}

	this.sendPacket(createFatalErrorPacket(reason))
	this.close()
}

// Отправить клиенту пакет с обычной ошибкой.
//
// Будет отображена в чате или диалоговом окне
//
// "packetId" - id пакета, в ответ на который возникла ошибка
//
// "errorId" - id ошибки. Можно посмотреть в LangPac.tsv файле, который находится в gui/gui.rfs в папке с игрой). Пример:
//	"2"	"10001"	"2208232205"	"eErrNoIpBlocked"	"Заблокированный IP."
//
// "code" - некий дополнительный код который будет указарн рядом с текстом ошибки
func (this *Client) Error(packetId uint16, errorId uint32, code uint32) {
	this.SendPacket(createErrorPacket(packetId, errorId, code))
}

// Отправить клиенту пакет с критической ошибкой, при получении которой клиент отключится от сервера
//
// Будет отображена в чате или диалоговом окне
func (this *Client) FatalError(errorId uint32) {
	this.SendPacket(createFatalErrorPacket(errorId))
}

func createClient(id uint16, conn net.Conn) Client {
	return Client{
		conn:            conn,
		writePacketChan: make(chan packet.Packet),
		readPacketChan:  make(chan packet.Packet),
		ip:              conn.RemoteAddr().(*net.TCPAddr).IP.String(),
		id:              id,
	}
}
