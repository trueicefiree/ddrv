package discord

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/google/uuid"
)

// Archive provides configuration for discord file system
type Archive struct {
	chunkSize int
	webhooks  []string
	clients   []webhook.Client
	lastWbIdx int
	traceCtx  context.Context
}

// NewArchive returns a new instance of Archive
func NewArchive(chunkSize int, webhooks []string) (*Archive, error) {
	st := &Archive{
		chunkSize: chunkSize,
		webhooks:  webhooks,
		clients:   make([]webhook.Client, 0),
	}
	for _, url := range webhooks {
		client, err := webhook.NewWithURL(url)
		if err != nil {
			return nil, err
		}
		st.clients = append(st.clients, client)
	}

	clientTrace := &httptrace.ClientTrace{}
	st.traceCtx = httptrace.WithClientTrace(context.Background(), clientTrace)

	return st, nil
}

func (a *Archive) StreamWriter() *StreamWriter {
	return NewStreamWriter(a.chunkSize, a)
}

// next returns the next webhook client in the list, cycling through the list
func (a *Archive) next() webhook.Client {
	webhookClient := a.clients[a.lastWbIdx]
	a.lastWbIdx = (a.lastWbIdx + 1) % len(a.clients)
	return webhookClient
}

// ReadAttachment reads data from the Archive
func (a *Archive) ReadAttachment(url string, start, end int) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(a.traceCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("%s : unexpected status code: %a", url, res.StatusCode)
	}
	return res.Body, nil
}

func (a *Archive) WriteAttachment(r io.Reader) (string, int, error) {
	client := a.next()

	m, err := client.CreateMessage(
		discord.
			NewWebhookMessageCreateBuilder().
			AddFile(uuid.New().String(), "", r).
			Build(),
	)
	if err != nil {
		return "", 0, err
	}
	if len(m.Attachments) != 1 {
		return "", 0, errors.New("invalid attachments len")
	}

	return m.Attachments[0].URL, m.Attachments[0].Size, nil
}
