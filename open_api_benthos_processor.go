package open_api_benthos_processor

import (
	"context"
	"fmt"
	"github.com/benthosdev/benthos/v4/public/service"
)

var openAiProcessorConfigSpec = service.NewConfigSpec().
	Summary("Creates a processor that sends requests to chat gpt").
	Field(service.NewStringField("source_field")).
	Field(service.NewStringField("target_field")).
	Field(service.NewStringField("prompt")).
	Field(service.NewStringField("api_key")).
	Field(service.NewStringField("model")).
	Field(service.NewStringField("api_endpoint").Default("none")).
	Field(service.NewStringField("driver"))

type openAiProcessor struct {
	sourceField string
	targetField string
	client      AiProcessor
	model       string
	prompt      string
	apiEndpoint string
	// driver can be azure or openai
	driver        string
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
		model       string
		apiKey      string
		prompt      string
		driver      string
		apiEndpoint string
	)

	sourceField, err := conf.FieldString("source_field")

	if err != nil {
		return nil, err
	}

	apiEndpoint, err = conf.FieldString("api_endpoint")

	if err != nil {
		return nil, err
	}

	targetField, err = conf.FieldString("target_field")

	if err != nil {
		return nil, err
	}

	driver, err = conf.FieldString("driver")

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

	var aiDriver AiProcessor
	switch driver {
	case "azure":
		aiDriver = NewAzureProcessor(apiKey, apiEndpoint)
	case "openai":
		aiDriver = NewOpenAIProcessor(apiKey, model)
	}

	return &openAiProcessor{
		sourceField:   sourceField,
		targetField:   targetField,
		client:        aiDriver,
		model:         model,
		prompt:        prompt,
		driver:        driver,
		apiEndpoint:   apiEndpoint,
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

	prompt := fmt.Sprintf("Take the data: '%s' and respond after doing following: '%s' .", value, o.prompt)

	resp, err := o.client.Ask(prompt)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		if err != nil {
			return []*service.Message{m}, nil
		}
	}

	payload := make(map[string]interface{})

	payload[o.targetField] = resp

	for k, v := range content.(map[string]interface{}) {
		payload[k] = v
	}

	m.SetStructured(payload)

	return []*service.Message{m}, nil
}

func (o *openAiProcessor) Close(ctx context.Context) error {
	return nil
}
