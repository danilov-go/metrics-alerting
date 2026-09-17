package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var metricsExp = []string{
	"Alloc",
	"CPUutilization1",
	"CPUutilization2",
	"CPUutilization3",
	"CPUutilization4",
	"CPUutilization5",
	"CPUutilization6",
	"CPUutilization7",
	"CPUutilization8",
	"PollCount",
}

func TestFileSubscriber(t *testing.T) {
	tests := []struct {
		name       string
		expAudit   Audit
		countSub   int
		deregFirst bool
	}{
		{
			name:     "положительный тест (один подписчик)",
			expAudit: Audit{TS: 1234567891, IPAddress: "192.168.1.1", Metrics: metricsExp},
			countSub: 1,
		},
		{
			name:     "положительный тест (три подписчика)",
			expAudit: Audit{TS: 1234567892, IPAddress: "192.168.1.2", Metrics: metricsExp},
			countSub: 3,
		},
		{
			name:       "удаленный подписчик",
			expAudit:   Audit{TS: 1234567893, IPAddress: "192.168.1.3", Metrics: metricsExp},
			countSub:   1,
			deregFirst: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			event := NewEvent(logger.Sugar())
			testDir := t.TempDir()
			subs := make([]*FileSubscriber, tt.countSub)
			filePaths := make([]string, tt.countSub)
			for i := 0; i < tt.countSub; i++ {
				path := filepath.Join(testDir, fmt.Sprintf("%d_audit.log", i))
				filePaths[i] = path
				sub := NewFileSubscriber(path, logger.Sugar())
				require.NotNil(t, sub)
				sub.id = filepath.Base(path)
				subs[i] = sub
				event.Register(sub)
			}
			if tt.deregFirst && len(subs) > 0 {
				event.Deregister(subs[0])
			}
			event.Notify(tt.expAudit)
			for _, sub := range subs {
				err := sub.Close()
				assert.NoError(t, err)
			}
			for i, path := range filePaths {
				if tt.deregFirst && i == 0 {
					info, err := os.Stat(path)
					if err == nil {
						assert.Zero(t, info.Size())
					}
					continue
				}
				file, err := os.ReadFile(path)
				require.NoError(t, err)
				require.NotEmpty(t, file)
				var audit Audit
				err = json.Unmarshal(file, &audit)
				assert.NoError(t, err)
				assert.Equal(t, tt.expAudit, audit)
			}
		})
	}
}
func TestURLSubscriber(t *testing.T) {
	tests := []struct {
		name       string
		expAudit   Audit
		statusCode int
		expError   bool
	}{
		{
			name:       "положительный тест",
			expAudit:   Audit{TS: 1234567891, IPAddress: "192.168.1.1", Metrics: metricsExp},
			statusCode: http.StatusOK,
		},
		{
			name:       "ошибка сервера",
			expAudit:   Audit{TS: 1234567892, IPAddress: "192.168.1.2", Metrics: metricsExp},
			statusCode: http.StatusBadRequest,
			expError:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			var reqBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				reqBody = body
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			sub := NewURLSubscriber(server.URL, logger, resty.New())
			require.NotNil(t, sub)
			sub.update(tt.expAudit)
			sub.Close()
			require.NotEmpty(t, reqBody)
			if !tt.expError {
				var audit Audit
				err := json.Unmarshal(reqBody, &audit)
				assert.NoError(t, err)
				assert.Equal(t, tt.expAudit, audit)
			}
		})
	}
}
