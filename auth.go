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


// setHeaders applies every header every outgoing request needs: the auth header, plus X-Sub-Account-Id when the client was scoped to a sub-account via WithSubAccount. 
func (c *Client) setHeaders(req *http.Request) {
	c.setAuthHeader(req)
	if c.subAccountID != "" {
		req.Header.Set("X-Sub-Account-Id", c.subAccountID)
	}
}
