package discord

import (
	"context"
	"fmt"
	"log/slog"
)

const customIDMaxLen = 100

type interactionType uint

const (
	pingInteraction               interactionType = 1
	ApplicationCommandInteraction interactionType = 2
	MessageComponentInteraction   interactionType = 3
)

type InteractionCallbackType uint

const (
	pongCallback                     InteractionCallbackType = 1
	ChannelMessageWithSource         InteractionCallbackType = 4
	DeferredChannelMessageWithSource InteractionCallbackType = 5
	DeferredUpdateMessage            InteractionCallbackType = 6
	UpdateMessageCallback            InteractionCallbackType = 7
)

type componentType uint

const (
	ActionRowType componentType = 1
	buttonType    componentType = 2
)

type buttonStyle uint

const (
	PrimaryButton   buttonStyle = 1
	SecondaryButton buttonStyle = 2
	DangerButton    buttonStyle = 4
)

const (
	EphemeralMessage int = 1 << 6
)

type InteractionRequest struct {
	Member        Member `json:"member"`
	ID            string `json:"id"`
	GuildID       string `json:"guild_id"`
	Token         string `json:"token"`
	ApplicationID string `json:"application_id"`
	Message       struct {
		Interaction struct {
			Name string `json:"name"`
		} `json:"interaction"`
	} `json:"message"`
	Data struct {
		Name     string          `json:"name"`
		CustomID string          `json:"custom_id"`
		Options  []CommandOption `json:"options"`
	} `json:"data"`
	Type interactionType `json:"type"`
}

type Member struct {
	User struct {
		ID       string `json:"id,omitempty"`
		Username string `json:"username,omitempty"`
	} `json:"user,omitempty"`
}

type InteractionDataResponse struct {
	Content         string          `json:"content,omitempty"`
	AllowedMentions AllowedMentions `json:"allowed_mentions"`
	Embeds          []Embed         `json:"embeds"`      // no `omitempty` to pass empty array when cleared
	Components      []Component     `json:"components"`  // no `omitempty` to pass empty array when cleared
	Attachments     []Attachment    `json:"attachments"` // no `omitempty` to pass empty array when cleared
	Flags           int             `json:"flags"`
}

// NewDataResponse create a data response
func NewDataResponse(content string) InteractionDataResponse {
	return InteractionDataResponse{
		Content: content,
		AllowedMentions: AllowedMentions{
			Parse: []string{},
		},
	}
}

// AddEmbed add given embed to response
func (d InteractionDataResponse) AddEmbed(embed Embed) InteractionDataResponse {
	if d.Embeds == nil {
		d.Embeds = []Embed{embed}
	} else {
		d.Embeds = append(d.Embeds, embed)
	}

	return d
}

type InteractionResponse struct {
	Data InteractionDataResponse `json:"data,omitempty"`
	Type InteractionCallbackType `json:"type,omitempty"`
}

func NewResponse(iType InteractionCallbackType, content string) InteractionResponse {
	return InteractionResponse{
		Type: iType,
		Data: NewDataResponse(content),
	}
}

func (i InteractionResponse) Ephemeral() InteractionResponse {
	i.Data.Flags = EphemeralMessage
	return i
}

func (i InteractionResponse) AddEmbed(embed Embed) InteractionResponse {
	i.Data = i.Data.AddEmbed(embed)
	return i
}

func (i InteractionResponse) AddComponent(component Component) InteractionResponse {
	if i.Data.Components == nil {
		i.Data.Components = []Component{component}
	} else {
		i.Data.Components = append(i.Data.Components, component)
	}

	return i
}

func (i InteractionResponse) AddAttachment(filename, filepath string, size int64) InteractionResponse {
	i.Data.Attachments = append(i.Data.Attachments, newAttachment(len(i.Data.Attachments), size, filename, filepath, i.Data.Flags&EphemeralMessage != 0))
	return i
}

func AsyncResponse(replace, ephemeral bool) InteractionResponse {
	response := InteractionResponse{
		Type: DeferredChannelMessageWithSource,
	}

	if replace {
		response.Type = DeferredUpdateMessage
	}

	if ephemeral {
		response.Data.Flags = EphemeralMessage
	}

	return response
}

func NewError(replace bool, err error) InteractionResponse {
	return NewEphemeral(replace, fmt.Sprintf("Oh! It's broken 😱. Reason is: %s", err))
}

func NewEphemeral(replace bool, content string) InteractionResponse {
	callback := ChannelMessageWithSource
	if replace {
		callback = UpdateMessageCallback
	}

	instance := InteractionResponse{Type: callback}
	instance.Data.Content = content
	instance.Data.Flags = EphemeralMessage
	instance.Data.Embeds = []Embed{}
	instance.Data.Components = []Component{}
	instance.Data.Attachments = []Attachment{}

	return instance
}

type AllowedMentions struct {
	Parse []string `json:"parse"`
}

type Image struct {
	URL string `json:"url"`
}

func NewImage(url string) *Image {
	return &Image{
		URL: url,
	}
}

type Author struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

func NewAuthor(name, url string) *Author {
	return &Author{
		Name: name,
		URL:  url,
	}
}

type Embed struct {
	Thumbnail   *Image  `json:"thumbnail,omitempty"`
	Image       *Image  `json:"image,omitempty"`
	Author      *Author `json:"author,omitempty"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         string  `json:"url,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
	Color       int     `json:"color,omitempty"`
}

func (e Embed) SetColor(color int) Embed {
	e.Color = color
	return e
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

func NewField(name, value string) Field {
	return Field{
		Name:   name,
		Value:  value,
		Inline: true,
	}
}

type Component struct {
	Label      string        `json:"label,omitempty"`
	CustomID   string        `json:"custom_id,omitempty"`
	Components []Component   `json:"components,omitempty"`
	Type       componentType `json:"type,omitempty"`
	Style      buttonStyle   `json:"style,omitempty"`
}

func NewButton(style buttonStyle, label, customID string) Component {
	if len(customID) > customIDMaxLen {
		slog.LogAttrs(context.Background(), slog.LevelWarn, "`custom_id` exceeds max characters", slog.Int("max", customIDMaxLen))
	}

	return Component{
		Type:     buttonType,
		Style:    style,
		Label:    label,
		CustomID: customID,
	}
}

type Attachment struct {
	Filename  string `json:"filename"`
	filepath  string
	ID        int   `json:"id"`
	Size      int64 `json:"size,omitempty"`
	Ephemeral bool  `json:"ephemeral,omitempty"`
}

func newAttachment(id int, size int64, filename, filepath string, ephemeral bool) Attachment {
	return Attachment{
		ID:        id,
		Filename:  filename,
		Size:      size,
		filepath:  filepath,
		Ephemeral: ephemeral,
	}
}

type Command struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Options     []CommandOption `json:"options,omitempty"`
	Guilds      []string        `json:"-"`
}

type CommandOption struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
	Type        int    `json:"type,omitempty"`
	Required    bool   `json:"required,omitempty"`
}
