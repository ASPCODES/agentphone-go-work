package agentphone

import(
	"net/http"
)

// setAuthHeader attaches the client's API key to an outgoing request as a Bearer token, per AgentPhone's authentication scheme:

//	Authorization: Bearer <api_key>

// Every request built in client.request() goes through here. Incase AgentPhone ever adds a second auth method (e.g. sub-account tokens), this is the only place that needs to change.
func (c *Client) setAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}
