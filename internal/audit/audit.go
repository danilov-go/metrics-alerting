// Package audit реализует систему аудита событий на основе паттерна «Наблюдатель».
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

// Audit определяет информацию о событии аудита.
type Audit struct {
	// TS определяет временную метку события
	TS int64 `json:"ts"`
	// Metrics содержит список названий метрик.
	Metrics []string `json:"metrics"`
	// IPAddress содержит IP-адрес источника события.
	IPAddress string `json:"ip_address"`
}

// Publisher определяет интерфейс для управления подписками и уведомления наблюдателей.
type Publisher interface {
	Register(observer)
	Deregister(observer)
	Notify(audit Audit)
}

type observer interface {
	update(audit Audit)
	getID() string
}

// Event реализует паттерн Publisher для координации отправки событий подписчикам.
type Event struct {
	logger    log
	observers sync.Map
}

// NewEvent создает новый экземпляр Event.
func NewEvent(l log) *Event {
	return &Event{
		logger: l,
	}
}

// Register регистрирует нового наблюдателя в системе уведомлений.
func (e *Event) Register(o observer) {
	e.observers.Store(o.getID(), o)
}

// Deregister удаляет наблюдателя из системы уведомлений.
func (e *Event) Deregister(o observer) {
	e.observers.Delete(o.getID())
}

// Notify асинхронно отправляет событие аудита всем зарегистрированным подписчикам.
func (e *Event) Notify(audit Audit) {
	e.observers.Range(func(key, val any) bool {
		if obs, ok := val.(observer); ok {
			obs.update(audit)
		}
		return true
	})
}

// FileSubscriber записывает полученные события аудита в локальный файл.
type FileSubscriber struct {
	id     string
	file   *os.File
	logger log
	wg     sync.WaitGroup
	ch     chan Audit
}

// NewFileSubscriber создает нового подписчика для логирования аудита в файл.
func NewFileSubscriber(filePath string, l log) *FileSubscriber {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		l.Errorw("ошибка открытия файла", "err", err)
		return nil
	}
	f := &FileSubscriber{
		id:     "file_audit",
		file:   file,
		logger: l,
		ch:     make(chan Audit, 100),
	}
	f.wg.Add(1)
	go f.run()
	return f
}

func (f *FileSubscriber) getID() string {
	return f.id
}

func (f *FileSubscriber) update(audit Audit) {
	select {
	case f.ch <- audit:
	default:
	}
}

func (f *FileSubscriber) run() {
	defer f.wg.Done()
	for audit := range f.ch {
		f.write(audit)
	}
}

func (f *FileSubscriber) write(audit Audit) {
	jsonAudit, err := json.Marshal(audit)
	if err != nil {
		f.logger.Errorw("ошибка сериализации", "err", err)
		return
	}
	if _, err = f.file.Write(jsonAudit); err != nil {
		f.logger.Errorw("ошибка записи в файл", "err", err)
		return
	}
	if _, err = f.file.Write([]byte{'\n'}); err != nil {
		f.logger.Errorw("ошибка записи в файл", "err", err)
		return
	}
}

func (f *FileSubscriber) Close() error {
	close(f.ch)
	f.wg.Wait()
	return f.file.Close()
}

// URLSubscriber отправляет полученные события аудита на внешний HTTP-сервер.
type URLSubscriber struct {
	id     string
	url    string
	logger log
	client *resty.Client
	wg     sync.WaitGroup
	ch     chan Audit
}

// NewURLSubscriber создает нового подписчика для отправки аудита по HTTP-адресу.
func NewURLSubscriber(url string, l log, client *resty.Client) *URLSubscriber {
	u := &URLSubscriber{
		id:     "url_audit",
		url:    url,
		logger: l,
		client: client,
		ch:     make(chan Audit, 100),
	}
	u.wg.Add(1)
	go u.run()
	return u
}

func (u *URLSubscriber) getID() string {
	return u.id
}

func (u *URLSubscriber) update(audit Audit) {
	select {
	case u.ch <- audit:
	default:
	}
}
func (u *URLSubscriber) run() {
	defer u.wg.Done()
	for audit := range u.ch {
		u.write(audit)
	}
}

func (u *URLSubscriber) write(audit Audit) {
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

func (u *URLSubscriber) Close() {
	close(u.ch)
	u.wg.Wait()
}
