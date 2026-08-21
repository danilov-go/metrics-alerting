package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type Audit struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Publisher interface {
	Register(observer)
	Deregister(observer)
	Notify(audit Audit)
}

type observer interface {
	update(audit Audit)
	getID() string
}

type Event struct {
	logger    log
	observers sync.Map
}

func NewEvent(l log) *Event {
	return &Event{
		logger: l,
	}
}

func (e *Event) Register(o observer) {
	e.observers.Store(o.getID(), o)
}

func (e *Event) Deregister(o observer) {
	e.observers.Delete(o.getID())
}

func (e *Event) Notify(audit Audit) {
	e.observers.Range(func(key, val any) bool {
		if obs, ok := val.(observer); ok {
			go obs.update(audit)
		}
		return true
	})
}

type FileSubscriber struct {
	id     string
	path   string
	logger log
	mu     sync.Mutex
}

func NewFileSubscriber(filePath string, l log) *FileSubscriber {
	return &FileSubscriber{
		id:     "file_audit",
		path:   filePath,
		logger: l,
	}
}

func (f *FileSubscriber) getID() string {
	return f.id
}

func (f *FileSubscriber) update(audit Audit) {
	jsonAudit, err := json.Marshal(audit)
	if err != nil {
		f.logger.Errorw("ошибка сериализации", "err", err)
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		f.logger.Errorw("ошибка открытия файла", "err", err)
		return
	}
	defer file.Close()
	if _, err = file.Write(jsonAudit); err != nil {
		f.logger.Errorw("ошибка записи в файл", "err", err)
		return
	}
	if _, err = file.Write([]byte{'\n'}); err != nil {
		f.logger.Errorw("ошибка записи в файл", "err", err)
		return
	}

}

type UrlSubscriber struct {
	id     string
	url    string
	logger log
	client *resty.Client
}

func NewURLSubscriber(url string, l log, client *resty.Client) *UrlSubscriber {
	return &UrlSubscriber{
		id:     "url_audit",
		url:    url,
		logger: l,
		client: client,
	}
}

func (u *UrlSubscriber) getID() string {
	return u.id
}

func (u *UrlSubscriber) update(audit Audit) {
	jsonAudit, err := json.Marshal(audit)
	if err != nil {
		u.logger.Errorw("ошибка сериализации", "err", err)
		return
	}
	resp, err := u.client.R().
		SetBody(jsonAudit).
		Post(u.url)
	if err != nil {
		u.logger.Errorw("ошибка отправки аудита", "err", err)
		return
	}
	if resp.StatusCode() >= 400 {
		u.logger.Errorw("сервер аудита вернул ошибку", "status", resp.Status())
		return
	}
}
