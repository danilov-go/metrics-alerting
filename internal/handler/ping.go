package handler

import (
	"net/http"
)

// PingHandler возвращает обработчик для проверки соединения с базой данных.
func (h *MetricsHandler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := h.storage.Ping(ctx)
		if err != nil {
			h.logger.Errorw("ошибка при проверке соединения с базой данных", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
