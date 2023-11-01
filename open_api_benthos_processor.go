package open_api_benthos_processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/benthosdev/benthos/v4/public/service"
)

var openAiProcessorConfigSpec = service.NewConfigSpec().
	Summary("Creates a processor that sends requests to chat gpt").
	Field(service.NewStringField("source_field")).
	Field(service.NewStringField("target_field")).
	Field(service.NewStringField("prompt")).
	Field(service.NewStringField("api_url")).
	Field(service.NewStringField("api_key")).
	Field(service.NewStringField("model"))

type openAiProcessor struct {
	sourceField   string
	targetField   string
	apiUrl        string
	apiKey        string
	model         string
	prompt        string
	metricCounter *service.MetricCounter
}

func init() {
	err := service.RegisterProcessor(
		"open_ai",
		openAiProcessorConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return newOpenAiProcessor(conf, mgr.Metrics())
		},
	)

	if err != nil {
		panic(err)
	}
}

func newOpenAiProcessor(conf *service.ParsedConfig, metrics *service.Metrics) (*openAiProcessor, error) {
	var (
		sourceField string
		targetField string
		apiUrl      string
		apiKey      string
		model       string
		prompt      string
	)

	sourceField, err := conf.FieldString("source_field")

	if err != nil {
		return nil, err
	}

	targetField, err = conf.FieldString("target_field")

	if err != nil {
		return nil, err
	}

	apiUrl, err = conf.FieldString("api_url")

	if err != nil {
		return nil, err
	}

	apiKey, err = conf.FieldString("api_key")

	if err != nil {
		return nil, err
	}

	model, err = conf.FieldString("model")

	if err != nil {
		return nil, err
	}

	prompt, err = conf.FieldString("prompt")

	if err != nil {
		return nil, err
	}

	return &openAiProcessor{
		sourceField:   sourceField,
		targetField:   targetField,
		apiUrl:        apiUrl,
		apiKey:        apiKey,
		model:         model,
		prompt:        prompt,
		metricCounter: metrics.NewCounter("open_ai_request"),
	}, nil
}

func (o *openAiProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	content, err := m.AsStructuredMut()

	if err != nil {
		return nil, err
	}

	value, ok := content.(map[string]interface{})[o.sourceField]

	if !ok {
		return []*service.Message{m}, nil
	}

	prompt := fmt.Sprintf("Please take this value %s and do the following. %s", value, o.prompt)

	requestValues := make(map[string]interface{})
	requestValues["model"] = o.model
	requestValues["prompt"] = prompt
	requestValues["max_tokens"] = 5
	requestValues["temperature"] = 0

	body, _ := json.Marshal(requestValues)

	r, err := http.NewRequest(http.MethodPost, o.apiUrl, bytes.NewBuffer(body))

	if err != nil {
		return []*service.Message{m}, nil
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		return []*service.Message{m}, nil
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("open ai response ", res.Status)

		return []*service.Message{m}, nil
	}

	ioResponse, _ := io.ReadAll(res.Body)

	openAiResponse := make(map[string]interface{})

	err = json.Unmarshal(ioResponse, &openAiResponse)

	if err != nil {
		return []*service.Message{m}, nil
	}

	o.metricCounter.Incr(1)

	text := openAiResponse["choices"].([]interface{})[0].(map[string]interface{})["text"].(string)

	payload := make(map[string]interface{})

	payload[o.targetField] = text

	for k, v := range content.(map[string]interface{}) {
		payload[k] = v
	}

	m.SetStructured(payload)

	return []*service.Message{m}, nil
}

func (*openAiProcessor) Close(ctx context.Context) error {
	return nil
}
