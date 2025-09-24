package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
	"net/http"
	"strings"
)

func globalOutboxHandler(
	client *http.Client,
	bookURL string,
	authorURL string,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, bookURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, authorURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func outboxHandler(
	client *http.Client,
	url string,
	unmarshalFunc func(data []byte) (string, error),
) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		id, err := unmarshalFunc(data)
		if err != nil {
			return fmt.Errorf("can not deserialize data in outbox handler: %w", err)
		}

		resp, err := client.Post(url, "application/json", strings.NewReader(id))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
			return fmt.Errorf("request failed with status: %d", resp.StatusCode)
		}
		return nil
	}
}

func bookOutboxHandler(
	client *http.Client,
	url string,
) outbox.KindHandler {
	return outboxHandler(client, url, func(data []byte) (string, error) {
		book := entity.Book{}
		if err := json.Unmarshal(data, &book); err != nil {
			return "", err
		}
		return book.ID, nil
	})
}

func authorOutboxHandler(
	client *http.Client,
	url string,
) outbox.KindHandler {
	return outboxHandler(client, url, func(data []byte) (string, error) {
		author := entity.Author{}
		if err := json.Unmarshal(data, &author); err != nil {
			return "", err
		}
		return author.ID, nil
	})
}
