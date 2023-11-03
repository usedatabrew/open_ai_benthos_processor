# Open AI Processor for Benthos

<img src='https://github.com/usedatabrew/pglogicalstream/blob/main/images/databrew-logo.png' width="200px" align="middle" >

Benthos processors are functions applied to messages passing through a pipeline. The function signature allows a processor to mutate or drop messages depending on the content of the message

Welcome to the Open AI Processor for Benthos! This processor allows you to enrich the context of your messages with requests to Open AI.

## Getting Started

To get started you have to run benthos with custom processor. Since this processor is not adopted by benthos itself 
you have to create a new benthos build with plugin registered

```go
package main

import (
		"context"
	"github.com/benthosdev/benthos/v4/public/service"

	_ "github.com/benthosdev/benthos/v4/public/components/all"
	// import open_ai processor
	_ "github.com/usedatabrew/open_ai_benthos_processor"
)

func main() {
	// here we initialize benthos
	service.RunCLI(context.Background())
}
```

### Create benthos configuration with processor

```yaml
pipeline:
  label: open_ai_processor
  processors:
    - open_ai:
        source_field: ""
        target_field: ""
        prompt: ""
        api_url: "https://api.openai.com/v1/completions"
        api_key: ""
        model: "gpt-3.5-turbo-instruct"
```
