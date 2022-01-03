package net

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tuxuuman/r2o-core/pkg/net/packet"
)

const (
	ERROR_SERVER_IS_FULL         uint32 = 1855293908
	ERROR_IDENTIFICATION_TIMEOUT uint32 = 801713924
)

type ClientPacket struct {
	Client *Client
	Packet *packet.Packet
}

type Server struct {
	host         string
	port         uint16
	listener     net.Listener
	clients      map[uint16]*Client
	clientsCount uint16
	// Максимальное кол-во клиентов. при достижении лимита которого, все новые подключения будут автоматически оклонятся.
	MaxClientsCount uint16
	// Максимальное время ожидания подтверждения подключения клиента (По умолчанию: 10 сек).
	MaxClientAcceptTimeout uint16
}

func (this *Server) genNewClientId() uint16 {
	minId := this.clientsCount + 1

	for {
		if _, exists := this.clients[minId]; exists {
			minId += 1
		} else {
			return minId
		}
	}
}

// Получить кол-во подключенных клиентов
func (this *Server) GetClientsCount() uint16 {
	return this.clientsCount
}

// Запущен ли сервер
func (this *Server) IsStarted() bool {
	return this.listener != nil
}

// ЗАпустить сервер
//
// "onConnection" - коллбэк который будет вызван при подключении клиента
func (this *Server) Start(onConnection func(c *Client)) {
	if this.IsStarted() {
		panic(errors.New("Сервер уже запущен"))
	}

	address := this.host + ":" + fmt.Sprint(this.port)
	ln, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	this.listener = ln
	log.Printf("Сервер запущен: %v", address)
	taskChan := make(chan func())

	go func() {
		for {
			conn, err := ln.Accept()

			if err != nil {
				log.Println("Не удалось обработать подключение клиента", err)
				continue
			}

			func() {
				clId := this.genNewClientId()
				cl := createClient(clId, conn)

				if this.clientsCount >= this.MaxClientsCount {
					cl.Reject(ERROR_SERVER_IS_FULL)
					return
				}

				this.clients[clId] = &cl
				this.clientsCount += 1

				log.Printf("Подключился новый клиент %v", cl.ip)

				cl.OnDisconnect(func() {
					taskChan <- func() {
						delete(this.clients, clId)
						this.clientsCount -= 1
					}
				}, true)

				go func() {
					time.Sleep(time.Second * time.Duration(this.MaxClientAcceptTimeout))
					taskChan <- func() {
						if cl.accepted == false && cl.rejected == false {
							log.Printf("Превышено время ожидания подтверждения подключения [%v][%v]", clId, cl.ip)
							cl.Reject(ERROR_IDENTIFICATION_TIMEOUT)
						}
					}
				}()

				taskChan <- func() {
					onConnection(&cl)
				}
			}()
		}
	}()

	for task := range taskChan {
		task()
	}
}

func CreateServer(host string, port uint16) Server {
	return Server{
		host:                   host,
		port:                   port,
		clients:                make(map[uint16]*Client, 1024),
		MaxClientsCount:        1000,
		MaxClientAcceptTimeout: 10,
	}
}
