package agentphone

import(
	"net/url"
	"strconv"
)

// AgentPhone uses two pagination strategies, Reference -> docs.agentphone.ai/pagination):

type ListParams struct {
	Limit 	int
	Offset 	int
}

// toQuery converts non-zero fields into URL query parameters, e.g.
// "?limit=10&offset=20". Returns an empty string if no fields are set.
func (p *ListParams) toQuery() string {
	if p == nil {
		return ""
	}

	q := url.Values{}
	if p.Limit != 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset != 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}

	encoded := q.Encode()
	if encoded == "" {
		return ""
	}

	return "?" + encoded
}

// OffsetPageInfo is the pagination metadata on offset-paginated list
// responses. Embed it in each resource's ListXxxResponse type alongside a
// data slice tagged json:"data",

type OffsetPageInfo struct {
	HasMore		bool  `json:"hasMore"`
	Total   	int   `json:"total,omitempty"`
}


// CursorListParams holds cursor-based pagination parameters, used by
// time-ordered endpoints like GetMessages.
type CursorListParams struct {
	Limit 		int
	Before		string
	After		string
}

func (p *CursorListParams) toQuery() string {
	if p == nil {
		return ""
	}

	q := url.Values{}
	if p.Limit != 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}
	if p.After != "" {
		q.Set("after", p.After)
	}

	encoded := q.Encode()
	if encoded == "" {
		return ""
	}
	return "?" + encoded
}

type CursorPageInfo struct{
	HasMore bool `json:"hasMore"`
} 


// Float64, Int, and Bool return a pointer to the given value. Use them to
// set optional numeric/boolean request fields whose zero value is
// meaningful and must be distinguished from "not set, use the platform
// default" — e.g. CreateAgentParams.InterruptionSensitivity, where 0 means
// "never interrupted", not "unset". A plain (non-pointer) field with
// omitempty can't make that distinction, since it would omit an explicit
// zero the same way it omits an unset field.
func Float64(v float64) *float64 { return &v }
func Int(v int) *int             { return &v }
func Bool(v bool) *bool          { return &v }