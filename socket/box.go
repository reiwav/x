package socket

import (
	"github.com/reiwav/x/mlog"
)

// Box handles ws request
//

var boxLog = mlog.NewTagLog("box")

type Box struct {
	ID       string
	Clients  *WsClientManager
	handlers map[string]IBoxHandler
	NotFound IBoxHandler
	Join     func(*WsClient) error
	Leave    func(*WsClient)
	Recover  func(*Request, interface{})
	history  map[string]HistoryAdapter
}

// NewBox create a new box
func NewBox(ID string) *Box {
	var b = &Box{
		ID:       ID,
		Clients:  NewWsClientManager(),
		handlers: make(map[string]IBoxHandler),
		history:  map[string]HistoryAdapter{},
	}
	b.Recover = b.defaultRecover
	// b.NotFound = b.notFound
	b.Join = b.join
	b.Leave = b.leave
	b.Handle("/echo", b.Echo)
	return b
}

// Handle add a handler
func (b *Box) Handle(uri string, handler IBoxHandler, adapters ...HistoryAdapter) {
	if len(adapters) < 1 || adapters[0] == nil {
		b.handlers[uri] = handler
		return
	}
	// enable history
	adapter := adapters[0]
	b.history[uri] = adapter
	b.handlers[uri] = func(r *Request) {
		adapter.Add(r.Data)
		handler(r)
	}
}

// Serve process the regggo getquest
func (b *Box) Serve(r *Request) {

	defer func() {
		if rc := recover(); rc != nil {
			if nil != b.Recover {
				b.Recover(r, rc)
			}
		}
	}()

	var handler = b.handlers[r.Path()]
	if handler == nil {
		if nil == b.NotFound {
			return
		}
		handler = b.NotFound
	}
	handler(r)
}

// Echo the default echo service
func (b *Box) Echo(r *Request) {
	r.Client.queueForSend(r.Payload)
}

func (b *Box) Broadcast(uri string, v interface{}) {
	b.Clients.SendJson(uri, v)
}

func (b *Box) Destroy() {
	b.Clients.Destroy()
}

func (b *Box) GetStatus() interface{} {
	return map[string]interface{}{
		"active_clients": b.Clients.Count(),
	}
}

func (b *Box) History() (map[string][]interface{}, error) {
	res := map[string][]interface{}{}
	for uri, adapter := range b.history {
		if v, err := adapter.Read(0, -1); err != nil {
			return nil, err
		} else {
			res[uri] = v
		}
	}
	return res, nil
}
