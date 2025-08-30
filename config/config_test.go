package config

import (
	"os"
	"testing"

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
				"GRPC_PORT":         "9091",
				"GRPC_GATEWAY_PORT": "8082",
			},
			want: &Config{
				GRPC: GRPC{
					Port:        "9091",
					GatewayPort: "8082",
				},
			},
			wantErr: false,
		},
		{
			name: "MissingEnvVars",
			envVars: map[string]string{
				"GRPC_PORT":         "",
				"GRPC_GATEWAY_PORT": "",
			},
			want: &Config{
				GRPC: GRPC{
					Port:        "9090",
					GatewayPort: "8080",
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, value := range test.envVars {
				t.Setenv(key, value) // Установка переменных окружения ОС
			}

			cfg, err := NewConfig() // Запуск тестируемого метода

			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want, cfg)
			} else {
				require.Error(t, err)
			}

			for key := range test.envVars {
				os.Unsetenv(key) // Удаление переменных окружения
			}
		})
	}
}
