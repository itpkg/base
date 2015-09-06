package base

import (
	"fmt"
)

type Widget interface {
	T(i18n *I18n, locale string)
}

//----------navbar----------------------

func NewNavBar() *NavBar {
	return &NavBar{Links: make([]*DropDown, 0)}
}

type NavBar struct {
	Links []*DropDown `json:"links"`
}

func (p *NavBar) Add(dd *DropDown) {
	p.Links = append(p.Links, dd)
}
func (p *NavBar) T(i18n *I18n, locale string) {
	for _, v := range p.Links {
		//v.Label = i18n.T(locale, v.Label)
		v.T(i18n, locale)
	}
}

//------------response--------------------
func NewResponse() *Response {
	return &Response{Ok: true, Title: "label.success", Data: make([]interface{}, 0), Errors: make([]string, 0)}
}

type Response struct {
	Ok     bool          `json:"ok"`
	Title  string        `json:"title"`
	Data   []interface{} `json:"data"`
	Errors []string      `json:"errors"`
}

func (p *Response) AddError(err string) {
	p.Ok = false
	p.Title = "label.failed"
	p.Errors = append(p.Errors, err)
}
func (p *Response) AddData(val interface{}) {
	p.Data = append(p.Data, val)
}

func (p *Response) T(i18n *I18n, locale string) {
	p.Title = i18n.T(locale, p.Title)
	for k, v := range p.Errors {
		p.Errors[k] = i18n.T(locale, v)
	}

}

//------------dropdown--------------------
func NewDropDown(label string) *DropDown {
	return &DropDown{Label: label, Links: make([]*Link, 0)}
}

type DropDown struct {
	Label string  `json:"label"`
	Links []*Link `json:"links"`
}

func (p *DropDown) T(i18n *I18n, locale string) {
	p.Label = i18n.T(locale, p.Label)
	for _, v := range p.Links {
		v.Label = i18n.T(locale, v.Label)
	}

}

func (p *DropDown) Add(url, label string) {
	p.Links = append(p.Links, &Link{Label: label, Url: url})
}

func (p *DropDown) AddLinks(links []*Link) {
	p.Links = append(p.Links, links...)
}

type Link struct {
	Url   string `json:"url"`
	Label string `json:"label"`
}

func (p *Link) T(i18n *I18n, locale string) {
	p.Label = i18n.T(locale, p.Label)
}

//---------------------------List----------------------------------------------

func NewList(id, action string, body func() []interface{}, pager *Pager) *List {
	return &List{
		Id:     id,
		Title:  fmt.Sprintf("list.title.%s", id),
		Action: action,
		Body:   body(),
		Pager:  pager,
	}
}

type List struct {
	Id     string        `json:"id"`
	Action string        `json:"action"`
	Title  string        `json:"title"`
	Body   []interface{} `json:"body"`
	Pager  *Pager        `json:"pager"`
}

//---------------------------Table---------------------------------------------
func NewTable(id, action string, header []*Th, body func() [][]interface{}, new, view, edit, remove bool, pager *Pager) *Table {

	tab := Table{
		Id:       id,
		Title:    fmt.Sprintf("table.title.%s", id),
		Action:   action,
		Header:   header,
		Body:     body(),
		Refresh:  "form.button.refresh",
		Manage:   "form.button.manage",
		Pager:    pager,
		Messages: []string{"label.are_you_sure"},
	}
	if new {
		tab.New = "form.button.new"
	}
	if view {
		tab.View = "form.button.view"
	}
	if edit {
		tab.Edit = "form.button.edit"
	}
	if remove {
		tab.Remove = "form.button.remove"
	}

	return &tab
}

func (p *Table) T(i18n *I18n, locale string) {
	for _, th := range p.Header {
		th.Label = i18n.T(locale, th.Label)
	}
	p.View = i18n.T(locale, p.View)
	p.Remove = i18n.T(locale, p.Remove)
	p.Refresh = i18n.T(locale, p.Refresh)
	p.Edit = i18n.T(locale, p.Edit)
	p.New = i18n.T(locale, p.New)
	p.Manage = i18n.T(locale, p.Manage)
	p.Title = i18n.T(locale, p.Title)

	for k, v := range p.Messages {
		p.Messages[k] = i18n.T(locale, v)
	}
}

type Table struct {
	Id       string          `json:"id"`
	Action   string          `json:"action"`
	Title    string          `json:"title"`
	Header   []*Th           `json:"header"`
	Body     [][]interface{} `json:"body"`
	New      string          `json:"new"`
	Edit     string          `json:"edit"`
	Remove   string          `json:"remove"`
	Refresh  string          `json:"refresh"`
	View     string          `json:"view"`
	Manage   string          `json:"manage"`
	Pager    *Pager          `json:"pager"`
	Messages []string        `json:"messages"`
}

type Pager struct {
	Index int `json:"page"`
	Total int `json:"total"`
	Size  int `json:"size"`
}

type Th struct {
	Label string `json:"label"`
	Width string `json:"width"`
}

//---------------------------Form----------------------------------------------
func NewForm(id, action string) *Form {
	return &Form{
		Ok:        true,
		Id:        id,
		Title:     fmt.Sprintf("form.title.%s", id),
		Action:    action,
		Method:    "POST",
		Multipart: false,
		Fields:    make([]interface{}, 0),
		Buttons:   make([]interface{}, 0),
		Errors:    make([]string, 0),
		Submit:    "form.button.submit",
		Reset:     "form.button.reset",
	}
}

type Form struct {
	Ok        bool          `json:"ok"`
	Id        string        `json:"id"`
	Title     string        `json:"title"`
	Method    string        `json:"method"`
	Action    string        `json:"action"`
	Multipart bool          `json:"multipart"`
	Fields    []interface{} `json:"fields"`
	Buttons   []interface{} `json:"buttons"`
	Submit    string        `json:"submit"`
	Reset     string        `json:"reset"`
	Errors    []string      `json:"errors"`
}

func (p *Form) T(i18n *I18n, locale string) {
	p.Title = i18n.T(locale, p.Title)
	p.Submit = i18n.T(locale, p.Submit)
	p.Reset = i18n.T(locale, p.Reset)
	p.Action = Url(p.Action, locale, nil)
	for _, v := range p.Fields {
		switch v.(type) {
		case *TextField:
			v1 := v.(*TextField)
			v1.Label = i18n.T(locale, v1.Label)
			v1.Placeholder = i18n.T(locale, v1.Placeholder)
		case *PasswordField:
			v1 := v.(*PasswordField)
			v1.Label = i18n.T(locale, v1.Label)
			v1.Placeholder = i18n.T(locale, v1.Placeholder)
		case *SelectField:
			v1 := v.(*SelectField)
			v1.Label = i18n.T(locale, v1.Label)
		}

	}
	for _, v := range p.Buttons {
		switch v.(type) {
		case *Button:
			v1 := v.(*Button)
			v1.Label = i18n.T(locale, v1.Label)
		}

	}
}

func (p *Form) AddError(err string) {
	p.Ok = false
	p.Errors = append(p.Errors, err)
}

func (p *Form) AddButton(id, action, method string, confirm bool, style string) {
	p.Buttons = append(p.Buttons, &Button{
		Id:      id,
		Label:   fmt.Sprintf("form.button.%s", id),
		Action:  action,
		Method:  method,
		Confirm: confirm,
		Style:   style,
	})
}

func (p *Form) AddHiddenField(id string, value interface{}) {
	p.Fields = append(p.Fields, &HiddenField{
		Field: Field{
			Id:   id,
			Type: "hidden",
		},
		Value: value,
	})
}

func (p *Form) AddTextField(id string, value interface{}, required bool, readonly bool) {
	size := 8
	lbl := p.label(id)
	phl := p.placeholder(id)
	if id == "username" {
		size = 4
		lbl = "form.field.username"
		phl = "form.placeholder.username"
	}

	p.Fields = append(p.Fields, &TextField{
		Field: Field{
			Id:   id,
			Type: "text",
		},
		Label:       lbl,
		Value:       value,
		Size:        size,
		Required:    required,
		Readonly:    readonly,
		Placeholder: phl,
	})
}

func (p *Form) AddEmailField(id string, value interface{}, required bool, readonly bool) {
	p.Fields = append(p.Fields, &TextField{
		Field: Field{
			Id:   id,
			Type: "email",
		},
		Label:       "form.field.email",
		Value:       value,
		Size:        7,
		Required:    required,
		Readonly:    readonly,
		Placeholder: "form.placeholder.email",
	})
}

func (p *Form) AddPasswordField(id string, required, confirmed bool) {

	p1 := &PasswordField{
		Field: Field{
			Id:   id,
			Type: "password",
		},
		Label:       "form.field.password",
		Required:    required,
		Size:        6,
		Placeholder: "form.placeholder.password",
	}
	if id != "password" {
		p1.Placeholder = p.placeholder(id)
		p1.Label = p.label(id)
	}
	p.AddField(p1)

	if confirmed {
		p2 := &PasswordField{
			Field: Field{
				Id:   "re_" + id,
				Type: "password",
			},
			Label:       "form.field.re_password",
			Required:    required,
			Size:        6,
			Placeholder: "form.placeholder.re_password",
		}
		if id != "password" {
			p2.Placeholder = p.placeholder(p2.Id)
			p2.Label = p.label(p2.Id)
		}
		p.AddField(p2)
	}
}

func (p *Form) AddTextareaField(id string, value interface{}, required bool) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "text",
		},
		Label:    p.label(id),
		Value:    value,
		Rows:     10,
		Required: required,
	})
}

func (p *Form) AddMarkdownField(id string, value interface{}, required bool) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "markdown",
		},
		Label:    p.label(id),
		Value:    value,
		Rows:     10,
		Required: required,
	})
}

func (p *Form) AddHtmlField(id string, value interface{}, required bool) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "html",
		},
		Label:    p.label(id),
		Value:    value,
		Rows:     10,
		Required: required,
	})
}

func (p *Form) AddSelectField(id string, value interface{}, ofn func() []*Option, required bool) {
	p.AddField(&SelectField{
		Field: Field{
			Id:   id,
			Type: "select",
		},
		Label:    p.label(id),
		Options:  ofn(),
		Value:    value,
		Size:     5,
		Required: required,
	})
}

func (p *Form) AddCheckboxGroupField(id string, ofn func() []*Option) {
	p.AddField(&GroupField{
		Field: Field{
			Id:   id,
			Type: "checkbox",
		},
		Label:   p.label(id),
		Options: ofn(),
	})
}

func (p *Form) AddRadioGroupField(id string, ofn func() []*Option) {
	p.AddField(&GroupField{
		Field: Field{
			Id:   id,
			Type: "radio",
		},
		Label:   p.label(id),
		Options: ofn(),
	})
}

func (p *Form) AddField(f interface{}) {
	p.Fields = append(p.Fields, f)
}

func (p *Form) label(id string) string {
	return fmt.Sprintf("form.field.%s.%s", p.Id, id)
}

func (p *Form) placeholder(id string) string {
	return fmt.Sprintf("form.placeholder.%s.%s", p.Id, id)
}

type Field struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}
type HiddenField struct {
	Field
	Value interface{} `json:"value"`
}
type TextField struct {
	Field
	Label       string      `json:"label"`
	Value       interface{} `json:"value"`
	Size        int         `json:"size"`
	Required    bool        `json:"required"`
	Readonly    bool        `json:"readonly"`
	Placeholder string      `json:"placeholder"`
}
type TextareaField struct {
	Field
	Label    string      `json:"label"`
	Value    interface{} `json:"value"`
	Rows     int         `json:"rows"`
	Required bool        `json:"required"`
}
type PasswordField struct {
	Field
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Size        int    `json:"size"`
	Placeholder string `json:"placeholder"`
}
type Option struct {
	Id      interface{} `json:"id"`
	Label   string      `json:"label"`
	Checked bool        `json:"checked"`
}
type SelectField struct {
	Field
	Label    string      `json:"label"`
	Options  []*Option   `json:"options"`
	Multiple bool        `json:"multi"`
	Size     int         `json:"size"`
	Value    interface{} `json:"value"`
	Required bool        `json:"required"`
}
type GroupField struct {
	Field
	Label   string    `json:"label"`
	Options []*Option `json:"options"`
}

type Button struct {
	Id      string `json:"id"`
	Label   string `json:"label"`
	Action  string `json:"action"`
	Method  string `json:"method"`
	Confirm bool   `json:"confirm"`
	Style   string `json:"style"`
}
