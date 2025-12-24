package kafka

import (
	"strconv"

	kafkago "github.com/segmentio/kafka-go"
)

const retryHeaderKey = "retry_count"

func GetRetryCount(msg kafkago.Message) int {
	for _, h := range msg.Headers {
		if h.Key == retryHeaderKey {
			if n, err := strconv.Atoi(string(h.Value)); err == nil {
				return n
			}
		}
	}
	return 0
}

func WithRetryCount(headers []kafkago.Header, count int) []kafkago.Header {
	out := make([]kafkago.Header, 0, len(headers)+1)
	for _, h := range headers {
		if h.Key != retryHeaderKey {
			out = append(out, h)
		}
	}
	out = append(out, kafkago.Header{
		Key:   retryHeaderKey,
		Value: []byte(strconv.Itoa(count)),
	})
	return out
}
