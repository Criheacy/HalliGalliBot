package main

import (
	"encoding/json"
)

type OpType int8
type IntentType string

const (
	Dispatch OpType = iota
	Heartbeat
	Identify
	_
	_
	_
	Resume
	Reconnect
	_
	InvalidSession
	Hello
	HeartbeatAck
	HttpCallback
)

const (
	Ready         IntentType = "READY"
	MessageCreate IntentType = "MESSAGE_CREATE"
)

type MessageModel struct {
	Op        OpType     `json:"op"`
	MessageId int        `json:"s"`
	Intent    IntentType `json:"t"`
	Body      any        `json:"d"`
}

type IdentifyBody struct {
	Token      string            `json:"token"`
	Intents    int32             `json:"intents"`
	Shard      [2]int            `json:"shard"`
	Properties map[string]string `json:"properties"`
}

const Intents int32 = 1 + (1 << 1) + (1 << 9) + (1 << 10) + (1 << 12)
const DefaultHeartbeatIntervalMillis int = 1000
const DefaultLastMessageId int = 0

type MessageModelRaw struct {
	Op        OpType          `json:"op"`
	MessageId int             `json:"s"`
	Intent    IntentType      `json:"t"`
	Body      json.RawMessage `json:"d"`
}

func GetOpType(source []byte) (MessageModelRaw, error) {
	var raw MessageModelRaw
	err := json.Unmarshal(source, &raw)
	if err != nil {
		return MessageModelRaw{}, err
	}
	return raw, nil
}

func BuildRequest(messageType OpType, body any) MessageModel {
	return MessageModel{Op: messageType, Body: body}
}

func (model MessageModel) GetString() ([]byte, error) {
	return json.Marshal(model)
}

func ParseHelloResponseBody(source json.RawMessage) (HelloBody, error) {
	var body HelloBody
	err := json.Unmarshal(source, &body)
	if err != nil {
		return HelloBody{}, err
	}
	return body, nil
}

func ParseReadyMessageResponseBody(source json.RawMessage) (ReadyMessageBody, error) {
	var body ReadyMessageBody
	err := json.Unmarshal(source, &body)
	if err != nil {
		return ReadyMessageBody{}, err
	}
	return body, nil
}

type HelloBody struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type HeartbeatBody struct {
	LastMessageId string `json:"s"`
}

type ReadyMessageBody struct {
	Version   int              `json:"version"`
	SessionId string           `json:"session_id"`
	User      ReadyMessageUser `json:"user"`
	Shard     [2]int           `json:"shard"`
}

type ReadyMessageUser struct {
	Id       string `json:"id"`
	UserName string `json:"username"`
	Bot      bool   `json:"bot"`
}
