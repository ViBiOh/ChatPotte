package slack

import "fmt"

type Block any

type Element any

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewText(text string) Text {
	return Text{
		Type: "mrkdwn",
		Text: text,
	}
}

func NewPlainText(text string) Text {
	return Text{
		Type: "plain_text",
		Text: text,
	}
}

type Accessory struct {
	Type  string `json:"type"`
	Image string `json:"image_url"`
	Alt   string `json:"alt_text"`
}

func NewAccessory(image, alt string) *Accessory {
	return &Accessory{
		Type:  "image",
		Image: image,
		Alt:   alt,
	}
}

type ButtonElement struct {
	Type     string `json:"type"`
	Text     Text   `json:"text"`
	ActionID string `json:"action_id"`
	Value    string `json:"value,omitempty"`
	Style    string `json:"style,omitempty"`
}

func NewButtonElement(text string, actionID, value, style string) Element {
	return ButtonElement{
		Type:     "button",
		Text:     NewPlainText(text),
		ActionID: actionID,
		Value:    value,
		Style:    style,
	}
}

type Image struct {
	Type  string `json:"type"`
	Image string `json:"image_url"`
	Alt   string `json:"alt_text"`
	Title Text   `json:"title"`
}

func NewImage(image, alt, title string) Image {
	return Image{
		Type:  "image",
		Image: image,
		Alt:   alt,
		Title: NewPlainText(title),
	}
}

type Actions struct {
	Type     string    `json:"type"`
	BlockID  string    `json:"block_id,omitempty"`
	Elements []Element `json:"elements"`
}

func NewActions(blockID string, elements ...Element) Actions {
	return Actions{
		Type:     "actions",
		Elements: elements,
		BlockID:  blockID,
	}
}

type Section struct {
	Accessory *Accessory `json:"accessory,omitempty"`
	Text      Text       `json:"text"`
	Type      string     `json:"type"`
	Fields    []Text     `json:"fields,omitempty"`
}

func NewSection(text Text) Section {
	return Section{
		Type: "section",
		Text: text,
	}
}

func (s Section) IsZero() bool {
	return len(s.Type) == 0 && len(s.Fields) == 0 && s.Accessory == nil && len(s.Text.Text) == 0
}

func (s Section) WithAccessory(accessory *Accessory) Section {
	s.Accessory = accessory
	return s
}

func (s Section) AddField(field Text) Section {
	if s.Fields == nil {
		s.Fields = []Text{field}
	} else {
		s.Fields = append(s.Fields, field)
	}

	return s
}

type Context struct {
	Type     string    `json:"type"`
	Elements []Element `json:"elements"`
}

func NewContext() Context {
	return Context{
		Type: "context",
	}
}

func (c Context) AddElement(element Element) Context {
	if c.Elements == nil {
		c.Elements = []Element{element}
	} else {
		c.Elements = append(c.Elements, element)
	}

	return c
}

type Response struct {
	ResponseType    string  `json:"response_type,omitempty"`
	Text            string  `json:"text,omitempty"`
	Blocks          []Block `json:"blocks,omitempty"`
	ReplaceOriginal bool    `json:"replace_original,omitempty"`
	DeleteOriginal  bool    `json:"delete_original,omitempty"`
}

func NewResponse(message string) Response {
	return Response{
		Text:         message,
		ResponseType: "in_channel",
	}
}

func (r Response) Ephemeral() Response {
	r.ResponseType = "ephemeral"
	return r
}

func (r Response) WithReplaceOriginal() Response {
	r.ReplaceOriginal = true
	return r
}

func (r Response) WithDeleteOriginal() Response {
	r.DeleteOriginal = true
	return r
}

func (r Response) AddBlock(block Block) Response {
	if r.Blocks == nil {
		r.Blocks = []Block{block}
	} else {
		r.Blocks = append(r.Blocks, block)
	}

	return r
}

type SlashPayload struct {
	ChannelID   string `json:"channel_id"`
	Command     string `json:"command"`
	ResponseURL string `json:"response_url"`
	Text        string `json:"text"`
	Token       string `json:"token"`
	UserID      string `json:"user_id"`
}

type InteractiveAction struct {
	Type     string `json:"type"`
	BlockID  string `json:"block_id,omitempty"`
	ActionID string `json:"action_id,omitempty"`
	Value    string `json:"value,omitempty"`
}

type InteractivePayload struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Container struct {
		ChannelID string `json:"channel_id"`
	} `json:"container"`
	Type        string              `json:"type"`
	ResponseURL string              `json:"response_url"`
	Actions     []InteractiveAction `json:"actions"`
}

func NewError(err error) Response {
	return NewEphemeralMessage(fmt.Sprintf("Oh! It's broken 😱. Reason is: %s", err))
}

func NewEphemeralMessage(message string) Response {
	return NewResponse(message).Ephemeral().WithReplaceOriginal()
}
