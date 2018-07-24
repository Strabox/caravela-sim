package metrics

type Node struct {
	ApiRequestsReceived int64 `json:"ApiRequestsReceived"`
	RequestsSubmitted   int64 `json:"RequestsSubmitted"`
}

func NewNode() *Node {
	return &Node{
		ApiRequestsReceived: 0,
		RequestsSubmitted:   0,
	}
}

func (node *Node) APIRequestReceived() {
	node.ApiRequestsReceived++
}

func (node *Node) APIRequestsReceived() int64 {
	return node.ApiRequestsReceived
}

func (node *Node) TotalRequestsSubmitted() int64 {
	return node.RequestsSubmitted
}
