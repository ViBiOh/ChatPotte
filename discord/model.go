package discord

import "fmt"

type interactionType uint

const (
	pingInteraction interactionType = 1
	// ApplicationCommandInteraction occurs when user enter a slash command
	ApplicationCommandInteraction interactionType = 2
	// MessageComponentInteraction occurs when user interact with an action
	MessageComponentInteraction interactionType = 3
)

// InteractionCallbackType defines types of possible answer
type InteractionCallbackType uint

const (
	pongCallback InteractionCallbackType = 1
	// ChannelMessageWithSource answer
	ChannelMessageWithSource InteractionCallbackType = 4
	// DeferredChannelMessageWithSource deferred answer
	DeferredChannelMessageWithSource InteractionCallbackType = 5
	// DeferredUpdateMessage deferred in-place
	DeferredUpdateMessage InteractionCallbackType = 6
	// UpdateMessageCallback in place
	UpdateMessageCallback InteractionCallbackType = 7
)

type componentType uint

const (
	// ActionRowType for row
	ActionRowType componentType = 1
	buttonType    componentType = 2
)

type buttonStyle uint

const (
	// PrimaryButton is green
	PrimaryButton buttonStyle = 1
	// SecondaryButton is grey
	SecondaryButton buttonStyle = 2
	// DangerButton is red
	DangerButton buttonStyle = 4
)

const (
	// EphemeralMessage int value
	EphemeralMessage int = 1 << 6
)

// InteractionRequest when user perform an action
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

// Member of discord
type Member struct {
	User struct {
		ID       string `json:"id,omitempty"`
		Username string `json:"username,omitempty"`
	} `json:"user,omitempty"`
}

// InteractionDataResponse for responding to user
type InteractionDataResponse struct {
	Content         string          `json:"content,omitempty"`
	AllowedMentions AllowedMentions `json:"allowed_mentions"`
	Embeds          []Embed         `json:"embeds"`     // no `omitempty` to pass empty array when cleared
	Components      []Component     `json:"components"` // no `omitempty` to pass empty array when cleared
	Attachments     []Attachment    `json:"attachments,omitempty"`
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

// InteractionResponse for responding to user
type InteractionResponse struct {
	Data InteractionDataResponse `json:"data,omitempty"`
	Type InteractionCallbackType `json:"type,omitempty"`
}

// NewResponse creates a response
func NewResponse(iType InteractionCallbackType, content string) InteractionResponse {
	return InteractionResponse{
		Type: iType,
		Data: NewDataResponse(content),
	}
}

// Ephemeral set response to ephemeral
func (i InteractionResponse) Ephemeral() InteractionResponse {
	i.Data.Flags = EphemeralMessage
	return i
}

// AddEmbed add given embed to response
func (i InteractionResponse) AddEmbed(embed Embed) InteractionResponse {
	i.Data = i.Data.AddEmbed(embed)
	return i
}

// AddComponent add given component to response
func (i InteractionResponse) AddComponent(component Component) InteractionResponse {
	if i.Data.Components == nil {
		i.Data.Components = []Component{component}
	} else {
		i.Data.Components = append(i.Data.Components, component)
	}

	return i
}

// AddAttachment add given attachment to response
func (i InteractionResponse) AddAttachment(filename, filepath string, size int64) InteractionResponse {
	i.Data.Attachments = append(i.Data.Attachments, newAttachment(len(i.Data.Attachments), size, filename, filepath, i.Data.Flags&EphemeralMessage != 0))
	return i
}

// AsyncResponse to the user
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

// NewError creates an error response
func NewError(replace bool, err error) InteractionResponse {
	return NewEphemeral(replace, fmt.Sprintf("Oh! It's broken 😱. Reason is: %s", err))
}

// NewEphemeral creates an ephemeral response
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

	return instance
}

// AllowedMentions list
type AllowedMentions struct {
	Parse []string `json:"parse"`
}

// Image content
type Image struct {
	URL string `json:"url"`
}

// NewImage create an image
func NewImage(url string) *Image {
	return &Image{
		URL: url,
	}
}

// Author content
type Author struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// NewAuthor create an author
func NewAuthor(name, url string) *Author {
	return &Author{
		Name: name,
		URL:  url,
	}
}

// Embed of content
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

// SetColor define color of embed
func (e Embed) SetColor(color int) Embed {
	e.Color = color
	return e
}

// Field for embed
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// NewField creates new field
func NewField(name, value string) Field {
	return Field{
		Name:   name,
		Value:  value,
		Inline: true,
	}
}

// Component describes an interactive component
type Component struct {
	Label      string        `json:"label,omitempty"`
	CustomID   string        `json:"custom_id,omitempty"`
	Components []Component   `json:"components,omitempty"`
	Type       componentType `json:"type,omitempty"`
	Style      buttonStyle   `json:"style,omitempty"`
}

// NewButton creates a new button
func NewButton(style buttonStyle, label, customID string) Component {
	return Component{
		Type:     buttonType,
		Style:    style,
		Label:    label,
		CustomID: customID,
	}
}

// Attachment for file upload
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

// Command configuration
type Command struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Options     []CommandOption `json:"options,omitempty"`
	Guilds      []string        `json:"-"`
}

// CommandOption configuration option
type CommandOption struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
	Type        int    `json:"type,omitempty"`
	Required    bool   `json:"required,omitempty"`
}
