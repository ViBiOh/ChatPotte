package slack

import "fmt"

// Block response for slack
type Block any

// Element response for slack
type Element any

// Text Slack's model
type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewText creates Text
func NewText(text string) Text {
	return Text{
		Type: "mrkdwn",
		Text: text,
	}
}

// NewPlainText creates PlainText
func NewPlainText(text string) Text {
	return Text{
		Type: "plain_text",
		Text: text,
	}
}

// Accessory Slack's model
type Accessory struct {
	Type  string `json:"type"`
	Image string `json:"image_url"`
	Alt   string `json:"alt_text"`
}

// NewAccessory creates Accessory
func NewAccessory(image, alt string) *Accessory {
	return &Accessory{
		Type:  "image",
		Image: image,
		Alt:   alt,
	}
}

// ButtonElement response for slack
type ButtonElement struct {
	Type     string `json:"type"`
	Text     Text   `json:"text"`
	ActionID string `json:"action_id"`
	Value    string `json:"value,omitempty"`
	Style    string `json:"style,omitempty"`
}

// NewButtonElement creates ButtonElement
func NewButtonElement(text string, actionID, value, style string) Element {
	return ButtonElement{
		Type:     "button",
		Text:     NewPlainText(text),
		ActionID: actionID,
		Value:    value,
		Style:    style,
	}
}

// Actions response for slack
type Actions struct {
	Type     string    `json:"type"`
	BlockID  string    `json:"block_id,omitempty"`
	Elements []Element `json:"elements"`
}

// NewActions creates Actions
func NewActions(blockID string, elements ...Element) Block {
	return Actions{
		Type:     "actions",
		Elements: elements,
		BlockID:  blockID,
	}
}

// Section response for slack
type Section struct {
	Accessory *Accessory `json:"accessory,omitempty"`
	Text      Text       `json:"text"`
	Type      string     `json:"type"`
	Fields    []Text     `json:"fields,omitempty"`
}

// NewSection creates Section
func NewSection(text Text) Section {
	return Section{
		Type: "section",
		Text: text,
	}
}

// IsZero check if instance is populated
func (s Section) IsZero() bool {
	return len(s.Type) == 0 && len(s.Fields) == 0 && s.Accessory == nil && len(s.Text.Text) == 0
}

// WithAccessory set accessory for section
func (s Section) WithAccessory(accessory *Accessory) Section {
	s.Accessory = accessory
	return s
}

// AddField add given field to section
func (s Section) AddField(field Text) Section {
	if s.Fields == nil {
		s.Fields = []Text{field}
	} else {
		s.Fields = append(s.Fields, field)
	}

	return s
}

// Context response for slack
type Context struct {
	Type     string    `json:"type"`
	Elements []Element `json:"elements"`
}

// NewContext creates Context
func NewContext() Context {
	return Context{
		Type: "context",
	}
}

// AddElement add given element to context
func (c Context) AddElement(element Element) Context {
	if c.Elements == nil {
		c.Elements = []Element{element}
	} else {
		c.Elements = append(c.Elements, element)
	}

	return c
}

// Response response content
type Response struct {
	ResponseType    string  `json:"response_type,omitempty"`
	Text            string  `json:"text,omitempty"`
	Blocks          []Block `json:"blocks,omitempty"`
	ReplaceOriginal bool    `json:"replace_original,omitempty"`
	DeleteOriginal  bool    `json:"delete_original,omitempty"`
	AsUser          bool    `json:"as_user,omitempty"`
}

// NewResponse creates text response
func NewResponse(message string) Response {
	return Response{
		Text:         message,
		ResponseType: "in_channel",
	}
}

// Ephemeral set type to ephemeral
func (r Response) Ephemeral() Response {
	r.ResponseType = "ephemeral"
	return r
}

// WithReplaceOriginal set replace original to true
func (r Response) WithReplaceOriginal() Response {
	r.ReplaceOriginal = true
	return r
}

// WithDeleteOriginal set delete original to true
func (r Response) WithDeleteOriginal() Response {
	r.DeleteOriginal = true
	return r
}

// WithAsUser set as user to true
func (r Response) WithAsUser() Response {
	r.AsUser = true
	return r
}

// AddBlock add given block to response
func (r Response) AddBlock(block Block) Response {
	if r.Blocks == nil {
		r.Blocks = []Block{block}
	} else {
		r.Blocks = append(r.Blocks, block)
	}

	return r
}

// SlashPayload receives by a slash command
type SlashPayload struct {
	ChannelID   string `json:"channel_id"`
	Command     string `json:"command"`
	ResponseURL string `json:"response_url"`
	Text        string `json:"text"`
	Token       string `json:"token"`
	UserID      string `json:"user_id"`
}

// InteractiveAction response from slack
type InteractiveAction struct {
	Type     string `json:"type"`
	BlockID  string `json:"block_id,omitempty"`
	ActionID string `json:"action_id,omitempty"`
	Value    string `json:"value,omitempty"`
}

// InteractivePayload response from slack
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

// NewError creates ephemeral error response
func NewError(err error) Response {
	return NewEphemeralMessage(fmt.Sprintf("Oh! It's broken 😱. Reason is: %s", err))
}

// NewEphemeralMessage creates ephemeral text response
func NewEphemeralMessage(message string) Response {
	return NewResponse(message).Ephemeral().WithReplaceOriginal()
}
