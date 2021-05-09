package server

import "strings"

type Topic struct {
	Name   string
	Topics map[string]*Topic
}

func (t *Topic) String() string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString("name: \"" + t.Name + "\", ")
	sb.WriteString("topics: [")
	for _, v := range t.Topics {
		sb.WriteString(v.String())
	}
	sb.WriteString("]}, ")

	return sb.String()
}

func NewTopic(parent *Topic, name string) *Topic {
	topic := &Topic{Name: name}

	if parent != nil {
		if parent.Topics == nil {
			parent.Topics = make(map[string]*Topic)
		}
		parent.Topics[name] = topic
	}

	return topic
}
