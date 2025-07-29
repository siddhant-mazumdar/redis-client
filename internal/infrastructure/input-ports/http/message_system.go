package http

import (
	"context"
	"sync"
	"time"
)

type MessageType int

const (
	RequestMessage MessageType = iota
	ResponseMessage
	ErrorMessage
	ShutdownMessage
)

type Message struct {
	ID        string
	Type      MessageType
	Data      interface{}
	Timestamp time.Time
	Context   context.Context
	Response  chan interface{}
}

type MessageBus struct {
	requestChan  chan Message
	responseChan chan Message
	errorChan    chan Message
	shutdown     chan bool
	workers      int
	wg           sync.WaitGroup
	mu           sync.RWMutex
	activeReqs   map[string]*Message
}

func NewMessageBus(workers int) *MessageBus {
	return &MessageBus{
		requestChan:  make(chan Message, 100),
		responseChan: make(chan Message, 100),
		errorChan:    make(chan Message, 100),
		shutdown:     make(chan bool, 1),
		workers:      workers,
		activeReqs:   make(map[string]*Message),
	}
}

func (mb *MessageBus) Start() {
	for i := 0; i < mb.workers; i++ {
		mb.wg.Add(1)
		go mb.worker(i)
	}

	// Start message dispatcher
	go mb.dispatcher()
}

func (mb *MessageBus) Stop() {
	close(mb.shutdown)
	mb.wg.Wait()
}

func (mb *MessageBus) SendRequest(msg Message) {
	mb.mu.Lock()
	mb.activeReqs[msg.ID] = &msg
	mb.mu.Unlock()

	select {
	case mb.requestChan <- msg:
	case <-time.After(5 * time.Second):
		mb.SendError(Message{
			ID:   msg.ID,
			Type: ErrorMessage,
			Data: "Request timeout",
		})
	}
}

func (mb *MessageBus) SendResponse(msg Message) {
	mb.responseChan <- msg
}

func (mb *MessageBus) SendError(msg Message) {
	mb.errorChan <- msg
}

func (mb *MessageBus) worker(id int) {
	defer mb.wg.Done()

	for {
		select {
		case msg := <-mb.requestChan:
			// Process request in goroutine
			go mb.processRequest(msg, id)
		case <-mb.shutdown:
			return
		}
	}
}

func (mb *MessageBus) processRequest(msg Message, workerID int) {
	// This will be implemented by specific handlers
	// For now, just echo back
	response := Message{
		ID:        msg.ID,
		Type:      ResponseMessage,
		Data:      map[string]interface{}{"worker": workerID, "processed": true},
		Timestamp: time.Now(),
	}

	if msg.Response != nil {
		msg.Response <- response.Data
		close(msg.Response)
	}

	mb.SendResponse(response)
}

func (mb *MessageBus) dispatcher() {
	for {
		select {
		case response := <-mb.responseChan:
			mb.mu.Lock()
			delete(mb.activeReqs, response.ID)
			mb.mu.Unlock()

		case err := <-mb.errorChan:
			mb.mu.Lock()
			if req, exists := mb.activeReqs[err.ID]; exists {
				if req.Response != nil {
					req.Response <- map[string]interface{}{"error": err.Data}
					close(req.Response)
				}
				delete(mb.activeReqs, err.ID)
			}
			mb.mu.Unlock()

		case <-mb.shutdown:
			return
		}
	}
}
