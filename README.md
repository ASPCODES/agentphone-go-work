# AgentPhone Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/AgentPhone-AI/agentphone-go.svg)](https://pkg.go.dev/github.com/AgentPhone-AI/agentphone-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/AgentPhone-AI/agentphone-go)](https://goreportcard.com/report/github.com/AgentPhone-AI/agentphone-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **Official AgentPhone Go SDK.** Maintained by the AgentPhone team for building integrations with the [AgentPhone](https://agentphone.ai) platform.

Give your AI agents real phone numbers, SMS, and voice calls — from Go.

AgentPhone provides a REST API for provisioning phone numbers, managing AI agents, sending SMS, and placing voice calls. This SDK wraps that API in idiomatic Go.

## Installation

```bash
go get github.com/AgentPhone-AI/agentphone-go
```

Requires Go 1.21 or later.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	agentphone "github.com/AgentPhone-AI/agentphone-go"
)

func main() {
	client := agentphone.NewClient(os.Getenv("AGENTPHONE_API_KEY"))

	msg, err := client.Messages.Send(context.Background(), &agentphone.SendMessageParams{
		AgentID:  "agent_123",
		ToNumber: "+14155551234",
		Body:     "Hello from Go!",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Message sent:", msg.ID)
}
```

## Authentication

The SDK reads your API key from the `AGENTPHONE_API_KEY` environment variable by default, or you can pass it explicitly:

```go
client := agentphone.NewClient("your-api-key")
```

Get your API key from [agentphone.to](https://agentphone.to) under **Settings → API Keys**.

## Available Resources

| Resource        | Description                                                   |
| --------------- | ------------------------------------------------------------- |
| `Agents`        | Create and manage AI phone agents, voices, and agent webhooks |
| `Numbers`       | Buy, list, and manage phone numbers                           |
| `Calls`         | Place outbound calls, fetch recordings and transcripts        |
| `Messages`      | Send SMS/iMessage and reactions                               |
| `Conversations` | Manage threaded SMS conversations                             |
| `Contacts`      | Create and manage contacts                                    |
| `Contact Cards` | Manage per-number contact card info                           |
| `Webhooks`      | Configure inbound event webhooks and view delivery logs       |
| `Verification`  | Send and check verification codes                             |
| `Usage`         | Query usage stats by account, number, or agent                |
| `Sub-Accounts`  | Manage sub-accounts                                           |
| `SIP Trunks`    | Configure SIP trunk connections                               |
| `WhatsApp`      | Connect and manage WhatsApp numbers and templates             |
| `Registration`  | Manage A2P 10DLC registration                                 |
| `Location`      | Look up and refresh number location data                      |

Full endpoint-level reference: [docs.agentphone.ai/api-reference](https://docs.agentphone.ai/api-reference)

## Error Handling

API errors are returned as `*agentphone.APIError`, which includes the HTTP status code and the error message returned by the API:

```go
msg, err := client.Messages.Send(ctx, params)
if err != nil {
	var apiErr *agentphone.APIError
	if errors.As(err, &apiErr) {
		fmt.Println("status:", apiErr.StatusCode, "message:", apiErr.Message)
	}
}
```

## Contributing

Contributions are welcome! If you'd like to add a feature, fix a bug, or improve coverage of an endpoint, feel free to open an issue or pull request.

## License

MIT — see [LICENSE](LICENSE) for details.
