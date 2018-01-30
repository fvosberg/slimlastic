package example

//go:generate go run ../../cmds/generate/main.go -out elasticsearch_client.go -indexDefinition example.json Example

// Example is an example struct. It's just testdata for the generation of the client
type Example struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}
