package ddrv

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/google/uuid"
)

const (
    rateRemainingHeader = "X-RateLimit-Remaining"
    rateResetHeader     = "X-RateLimit-Reset"
)

// WebhookURLRegex is the regular expression pattern to validate a webhook URL.
var WebhookURLRegex = regexp.MustCompile(`^https://(?:[a-zA-Z]+\.)?discord\.com/api/webhooks/\d+/[a-zA-Z0-9_-]+$`)

// Message represents a Discord message and contains attachments (files uploaded within the message).
type Message struct {
    Attachments []Attachment `json:"attachments"`
}

// Rest represents the Discord webhook REST client.
type Rest struct {
    url       string
    resetAt   time.Time
    remaining int
    mu        sync.RWMutex
}

// NewRest creates a new Rest instance with the provided webhook URL.
func NewRest(webhookURL string) (*Rest, error) {
    // Validate the webhook URL using the regex pattern
    if ok := WebhookURLRegex.MatchString(webhookURL); !ok {
        return nil, ErrInvalidWebhookURL
    }

    return &Rest{url: webhookURL}, nil
}

// CreateAttachment uploads a file to the Discord channel using the webhook.
func (r *Rest) CreateAttachment(reader io.Reader) (*Attachment, error) {
    r.mu.RLock()
    // Sleep if we hit the rate limit, and it's not reset yet
    if r.remaining == 0 && time.Now().Before(r.resetAt) {
        sleepDuration := r.resetAt.Sub(time.Now())
        time.Sleep(sleepDuration)
    }
    r.mu.RUnlock()

    // Make the HTTP POST request to the webhook URL
    // using multipart body
    ctype, body := mbody(reader)
    resp, err := http.Post(r.url, ctype, body)
    if err != nil {
        return nil, err
    }

    // Update rate limit headers
    r.mu.Lock()
    r.remaining, _ = strconv.Atoi(resp.Header.Get(rateRemainingHeader))
    resetUnix, _ := strconv.ParseInt(resp.Header.Get(rateResetHeader), 10, 64)
    r.resetAt = time.Unix(resetUnix, 0)
    r.mu.Unlock()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("create attachment : expected status code %d but recevied %d", http.StatusOK, resp.StatusCode)
    }
    // Read and parse the response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var m Message
    if err := json.Unmarshal(respBody, &m); err != nil {
        return nil, err
    }

    // Return the first attachment from the response
    return &m.Attachments[0], nil
}

// mbody creates the multipart form-data body to upload a file to the Discord channel using the webhook.
func mbody(reader io.Reader) (string, io.Reader) {
    boundary := "disgosucks"
    // Set the content type including the boundary
    contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)

    CRLF := "\r\n"
    fname := uuid.New().String()

    // Assemble all the parts of the multipart form-data
    // This includes the boundary, content disposition with the file name, content type,
    // a blank line to end headers, the actual content (reader), end of content,
    // and end of multipart data
    parts := []io.Reader{
        strings.NewReader("--" + boundary + CRLF),
        strings.NewReader(fmt.Sprintf(`Content-Disposition: form-data; name="%s"; filename="%s"`, fname, fname) + CRLF),
        strings.NewReader(fmt.Sprintf(`Content-Type: %s`, "application/octet-stream") + CRLF),
        strings.NewReader(CRLF),
        reader,
        strings.NewReader(CRLF),
        strings.NewReader("--" + boundary + "--" + CRLF),
    }

    // Return the content type and the combined reader of all parts
    return contentType, io.MultiReader(parts...)
}
