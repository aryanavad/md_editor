package main

type Hub struct {
	documents  map[string]*Document
	register   chan *Client
	unregister chan *Client
}

type Document struct {
	id      string
	clients map[*Client]bool
}

func newHub() *Hub {
	return &Hub{
		documents:  make(map[string]*Document),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			doc, exists := h.documents[client.documentId]
			if !exists {
				doc = &Document{
					id:      client.documentId,
					clients: make(map[*Client]bool),
				}
				h.documents[client.documentId] = doc
			}
			doc.clients[client] = true

		case client := <-h.unregister:
			if doc, exists := h.documents[client.documentId]; exists {
				if _, ok := doc.clients[client]; ok {
					delete(doc.clients, client)
					close(client.send)
					
					if len(doc.clients) == 0 {
						delete(h.documents, client.documentId)
					}
				}
			}
		}
	}
}

func (h *Hub) broadcast(documentId string, message []byte, sender *Client) {
	if doc, exists := h.documents[documentId]; exists {
		for client := range doc.clients {
			if client != sender {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(doc.clients, client)
				}
			}
		}
	}
}
