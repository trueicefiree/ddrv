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

// Chunk represents a portion of data with a URL for fetching the data,
type Chunk struct {
    URL   string // URL where the chunk is stored
    Size  int    // Size of the chunk
    Start int64  // Start position of the chunk in the overall data sequence
    End   int64  // End position of the chunk in the overall data sequence
}

// Discord is a structure that manages interactions with Discord webhooks,
// providing an interface to write and read attachments.
type Discord struct {
    chunkSize int              // Size of each chunk of data to be processed
    webhooks  []string         // List of webhook URLs to be used for data storing
    clients   []webhook.Client // List of webhook clients corresponding to the webhook URLs
    lastWbIdx int              // Index of the last used webhook client
    traceCtx  context.Context  // Context for HTTP client tracing
}

// New returns a new instance of Discord with specified chunk size and webhook URLs.
// It initializes a list of webhook clients for each webhook URL.
func New(chunkSize int, webhooks []string) (*Discord, error) {
    st := &Discord{
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

    // Initialize tracing context for HTTP requests
    clientTrace := &httptrace.ClientTrace{}
    st.traceCtx = httptrace.WithClientTrace(context.Background(), clientTrace)

    return st, nil
}

// NewWriter creates a new Writer instance with the chunk size of the Discord instance.
func (d *Discord) NewWriter() *Writer {
    return NewWriter(d.chunkSize, d)
}

// NewReader creates a new Reader instance with the provided chunks and position, and the Discord instance as the Archive.
func (d *Discord) NewReader(chunks []Chunk, pos int64) (*Reader, error) {
    return NewReader(chunks, pos, d)
}

// ReadAttachment fetches a range of data from the specified URL.
// The range is specified by the start and end positions.
func (d *Discord) ReadAttachment(url string, start, end int) (io.ReadCloser, error) {
    req, err := http.NewRequestWithContext(d.traceCtx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    // Set the Range header to specify the range of data to fetch
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }

    // Return the body of the response, which contains the requested data
    return res.Body, nil
}

// WriteAttachment writes the data read from the provided Reader as a new attachment, returning the URL and size of the attachment.
func (d *Discord) WriteAttachment(r io.Reader) (string, int, error) {
    // Select the next webhook client
    client := d.next()

    // Create a new Discord message with the data as an attachment
    m, err := client.CreateMessage(
        discord.
            NewWebhookMessageCreateBuilder().
            AddFile(uuid.New().String(), "", r).
            Build(),
    )
    if err != nil {
        return "", 0, err
    }
    // Ensure the message contains exactly one attachment
    if len(m.Attachments) != 1 {
        return "", 0, errors.New("invalid attachments len")
    }

    // Return the URL and size of the attachment
    return m.Attachments[0].URL, m.Attachments[0].Size, nil
}

// next returns the next webhook client in the list, cycling through the list in a round-robin manner.
func (d *Discord) next() webhook.Client {
    // Select the next client
    client := d.clients[d.lastWbIdx]
    // Update the index of the last used client, wrapping around to the start of the list if necessary
    d.lastWbIdx = (d.lastWbIdx + 1) % len(d.clients)
    return client
}
