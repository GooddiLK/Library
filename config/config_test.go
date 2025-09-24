package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:parallel // Изменение переменных окружения
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string // Переменные окружения для теста
		want    *Config           // Ожидаемый результат
		wantErr bool              // Ожидаема ли ошибка
	}{
		{
			name: "ValidConfig",
			envVars: map[string]string{
				"GRPC_PORT":                 "50051",
				"GRPC_GATEWAY_PORT":         "8080",
				"POSTGRES_HOST":             "localhost",
				"POSTGRES_PORT":             "5432",
				"POSTGRES_DB":               "testdb",
				"POSTGRES_USER":             "testuser",
				"POSTGRES_PASSWORD":         "testpassword",
				"POSTGRES_MAX_CONN":         "10",
				"OUTBOX_ENABLED":            "true",
				"OUTBOX_WORKERS":            "5",
				"OUTBOX_BATCH_SIZE":         "100",
				"OUTBOX_WAIT_TIME_MS":       "500",
				"OUTBOX_IN_PROGRESS_TTL_MS": "1000",
				"OUTBOX_BOOK_SEND_URL":      "http://book-service/send",
				"OUTBOX_AUTHOR_SEND_URL":    "http://author-service/send",
			},
			want: &Config{
				GRPC: GRPC{
					Port:        "50051",
					GatewayPort: "8080",
				},
				PG: PG{
					URL:      "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable&pool_max_conns=10",
					Host:     "localhost",
					Port:     "5432",
					DB:       "testdb",
					User:     "testuser",
					Password: "testpassword",
					MaxConn:  "10",
				},
				Outbox: Outbox{
					Enabled:         true,
					Workers:         5,
					BatchSize:       100,
					WaitTimeMS:      500 * time.Millisecond,
					InProgressTTLMS: 1000 * time.Millisecond,
					BookSendURL:     "http://book-service/send",
					AuthorSendURL:   "http://author-service/send",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid outbox enabled",
			envVars: map[string]string{
				"OUTBOX_ENABLED": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid outbox workers",
			envVars: map[string]string{
				"OUTBOX_ENABLED": "true",
				"OUTBOX_WORKERS": "invalid workers",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid outbox workers",
			envVars: map[string]string{
				"OUTBOX_ENABLED":    "true",
				"OUTBOX_WORKERS":    "5",
				"OUTBOX_BATCH_SIZE": "invalid batch size",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid outbox wait time",
			envVars: map[string]string{
				"OUTBOX_ENABLED":      "true",
				"OUTBOX_WORKERS":      "5",
				"OUTBOX_BATCH_SIZE":   "100",
				"OUTBOX_WAIT_TIME_MS": "invalid wait time",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid outbox progress TTL",
			envVars: map[string]string{
				"OUTBOX_ENABLED":            "true",
				"OUTBOX_WORKERS":            "5",
				"OUTBOX_BATCH_SIZE":         "100",
				"OUTBOX_WAIT_TIME_MS":       "1000",
				"OUTBOX_IN_PROGRESS_TTL_MS": "invalid progress TTl",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, value := range test.envVars {
				t.Setenv(key, value) // Установка переменных окружения ОС
			}

			cfg, err := New() // Запуск тестируемого метода

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, cfg)
			}

			for key := range test.envVars {
				os.Unsetenv(key) // Удаление переменных окружения
			}
		})
	}
}
