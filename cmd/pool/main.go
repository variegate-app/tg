package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"telegram/internal/roundtripper"
	"time"
)

type Pool struct {
	cli           *http.Client
	token         string
	poolTimeout   time.Duration
	reportTimeout time.Duration
	reportAddress string
	poolAddr      string
	lastUpdateID  int
}

// NewPool создает новый пул.
//
// ctx - контекст.
// token - токен.
// a - адрес.
// pd - интервал опроса telegram.
// r - таймаут запроса с сообщением.
//
// Возвращает новый пул.
func NewPool(token string, a string, pd, r time.Duration) *Pool {
	var rt http.RoundTripper
	rt = http.DefaultTransport
	rt = roundtripper.NewCompress(rt)

	return &Pool{
		cli: &http.Client{
			Transport: rt,
			Timeout:   pd,
		},
		token:         token,
		poolAddr:      fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", token),
		poolTimeout:   pd,
		reportTimeout: r,
		reportAddress: a,
		lastUpdateID:  0,
	}
}

type RequestGetUpdates struct {
	// Offset Identifier of the first update to be returned. Must be greater by one than the highest among the identifiers of previously received updates. By default, updates starting with the earliest unconfirmed update are returned. An update is considered confirmed as soon as getUpdates is called with an offset higher than its update_id. The negative offset can be specified to retrieve updates starting from -offset update from the end of the updates queue. All previous updates will be forgotten.
	Offset int `json:"offset,omitempty"`
	// Limit the number of updates to be retrieved. Values between 1-100 are accepted. Defaults to 100.
	Limit int `json:"limit,omitempty"`
	// Timeout in seconds for long polling. Defaults to 0, i.e. usual short polling. Should be positive, short polling should be used for testing purposes only.
	Timeout int `json:"timeout,omitempty"`
	/* Updates
	A JSON-serialized list of the update types you want your bot to receive. For example, specify ["message", "edited_channel_post", "callback_query"] to only receive updates of these types. See Update for a complete list of available update types. Specify an empty list to receive all update types except chat_member, message_reaction, and message_reaction_count (default). If not specified, the previous setting will be used.
	Please note that this parameter doesn't affect updates created before the call to getUpdates, so unwanted updates may be received for a short period of time.
	*/
	Updates []string `json:"allowed_updates,omitempty"`
}

type UpdatesMessage struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID        int    `json:"id"`
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Language  string `json:"language_code"`
	} `json:"from"`
	Chat struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date     int    `json:"date"`
	Text     string `json:"text"`
	Entities []struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Type   string `json:"type"`
	} `json:"entities,omitempty"`
}

type ResponseGetUpdates struct {
	OK     bool `json:"ok"`
	Result []struct {
		UpdateID int            `json:"update_id"`
		Message  UpdatesMessage `json:"message"`
	} `json:"result"`
}

func main() {
	var token string
	flag.StringVar(&token, "t", "", "server address")
	flag.Parse()
	if token == "" {
		log.Fatal("token is empty")
	}

	p := NewPool(token, "", 30*time.Second, 1*time.Second)

	r := make(chan UpdatesMessage)
	go p.Run(context.Background(), r)
	for msg := range r {
		fmt.Printf("%v\n", msg)
	}

	p.Run(context.TODO(), r)
}

func (p *Pool) Run(ctx context.Context, r chan<- UpdatesMessage) {
	defer close(r)
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
			u, err := p.request(ctx, offset)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			if len(u.Result) == 0 {
				continue
			}
			for _, res := range u.Result {
				offset = res.UpdateID + 1
				r <- res.Message
			}
		}
	}
}

func (p *Pool) request(ctx context.Context, offset int) (updates *ResponseGetUpdates, err error) {
	pl := &RequestGetUpdates{Timeout: 30, Offset: offset}
	b, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.poolAddr, bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := p.cli.Do(request)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	updates = new(ResponseGetUpdates)
	err = json.Unmarshal(b, updates)

	return
}
