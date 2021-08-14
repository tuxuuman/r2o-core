package events

import (
	"fmt"
	"sync"
	"time"

	"github.com/tuxuuman/r2o-core/internal/utils"
)

type EventCallback = func(args ...interface{})

// Слушатель события
type EventListener struct {
	cb   EventCallback
	id   string
	once bool
	off  func()
}

// Отменить подписку на событие
func (l *EventListener) Off() {
	l.off()
}

// Получить ID слушателя
func (l *EventListener) GetId() string {
	return l.id
}

// Потоко-безопасный испускатель событий
type Emitter struct {
	listeners map[interface{}]map[string]EventListener
	mu        sync.Mutex
}

// Добавлить слушателя события
//
// "evtType" - тип события
//
// "cb" - функция которая будет вызвана при срабатывании события
//
// "once" - должен ли слушатель отрабатывать только один раз, после чего будет выполнена автоматическая отписка от события
func (emitter *Emitter) AddEventHandler(evtType interface{}, cb EventCallback, once bool) EventListener {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	listeners, exists := emitter.listeners[evtType]

	if !exists {
		listeners = make(map[string]EventListener)
		emitter.listeners[evtType] = listeners
	}

	id := utils.StrToMd5(fmt.Sprintf("%v_%v", &cb, time.Now()))

	listener := EventListener{
		cb:   cb,
		id:   id,
		once: once,
		off: func() {
			emitter.mu.Lock()
			defer emitter.mu.Unlock()
			delete(listeners, id)
			if len(listeners) == 0 {
				delete(emitter.listeners, evtType)
			}
		},
	}

	listeners[id] = listener

	return listener
}

// Вызов всех обработчиков событий указанного типа и их последовательное синхронное выполнение
//
// "evtType" - тип события
//
// "evtData" - данные которые должны быть переданы в каждый обработчик события
func (emitter *Emitter) Emit(evtType interface{}, evtData ...interface{}) {
	if listeners, exists := emitter.listeners[evtType]; exists {
		for _, listener := range listeners {
			if listener.once {
				defer listener.off()
			}
			listener.cb()
		}
	}
}

// Вызов всех обработчиков событий указанного типа, их асинхронное выполнени в отдельных горутинах и ожидание завершения выполнения.
//
// "evtType" - тип события
//
// "evtData" - данные которые должны быть переданы в каждый обработчик события
func (emitter *Emitter) EmitAsyncAwait(evtType interface{}, evtData ...interface{}) {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	if listeners, exists := emitter.listeners[evtType]; exists {
		wg := sync.WaitGroup{}
		for _, listener := range listeners {
			wg.Add(1)
			if listener.once {
				listener.off()
			}
			cb := listener.cb
			go func() {
				defer wg.Done()
				cb(evtData...)
			}()
		}
		wg.Wait()
	}
}

// Полностью асинхронный вызов всех обработчиков событий указанного типа, без ожидания их выполнения
//
// "evtType" - тип события
//
// "evtData" - данные которые должны быть переданы в каждый обработчик события
func (emitter *Emitter) EmitAsync(evtType interface{}, evtData ...interface{}) {
	go func() {
		emitter.mu.Lock()
		defer emitter.mu.Unlock()
		if listeners, exists := emitter.listeners[evtType]; exists {
			for _, listener := range listeners {
				if listener.once {
					listener.off()
				}
				cb := listener.cb
				go func() {
					cb(evtData...)
				}()
			}
		}
	}()
}

// Постоянная подписка на событие. Аналогичен вызову "AddEventHandler" с параметром "once" = false
//	AddEventHandler("myEvent", func(){}, false)
func (emitter *Emitter) On(evtType interface{}, cb EventCallback) EventListener {
	return emitter.AddEventHandler(evtType, cb, false)
}

// Одноразовая подписка на событие. Аналогичен вызову "AddEventHandler" с параметром "once" = true
//	AddEventHandler("myEvent", func(){}, true)
func (emitter *Emitter) Once(evtType interface{}, cb EventCallback) EventListener {
	return emitter.AddEventHandler(evtType, cb, true)
}

// Удалить слушателей события данного типа с такими id или удалит все если массив id пуст
//
// "evtType" - тип события
//
// "ids" - массив идентификаторов слушателей, подписку которых нужно отменить
//
// Если массив "ids" будет пуст, то будет отменены подписки для все слушателей события "evtType"
func (emitter *Emitter) Off(evtType interface{}, ids ...string) {
	if _, hExists := emitter.listeners[evtType]; hExists {
		if len(ids) > 0 {
			for _, id := range ids {
				if listener, exists := emitter.listeners[evtType][id]; exists {
					listener.Off()
				}
			}
		} else {
			emitter.mu.Lock()
			defer emitter.mu.Unlock()
			delete(emitter.listeners, evtType)
		}
	}
}

func (emitter *Emitter) String() string {
	str := ""
	for evtType, listeners := range emitter.listeners {
		str += fmt.Sprintf("%v: ", evtType)

		for _, listener := range listeners {
			str += fmt.Sprintf("[%v]", listener.id)

		}
		str += "\n"
	}
	return str
}

func CreateEmitter() Emitter {
	return Emitter{
		listeners: make(map[interface{}]map[string]EventListener),
	}
}
