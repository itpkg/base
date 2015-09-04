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
	return &Response{Ok: true, Title: "label.success", Data: make(map[string]interface{}, 0), Errors: make([]string, 0)}
}

type Response struct {
	Ok     bool                   `json:"ok"`
	Title  string                 `json:"title"`
	Data   map[string]interface{} `json:"data"`
	Errors []string               `json:"errors"`
}

func (p *Response) AddError(err string) {
	p.Ok = false
	p.Title = "label.failed"
	p.Errors = append(p.Errors, err)
}
func (p *Response) AddData(key string, val interface{}) {
	p.Data[key] = val
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
	var newB *Button
	if new {
		newB = &Button{
			Id:     "new",
			Label:  "form.button.new",
			Style:  "primary",
			Action: fmt.Sprintf("%s/new", action),
			Method: "GET",
		}
	} else {
		newB = nil
	}

	var bodyV [][]interface{}
	if view || edit || remove {
		header = append(header, &Th{Label: "form.button.manage", Width: "20%"})

		for _, row := range body() {
			bg := make([]*Button, 0)
			if view {
				bg = append(bg, &Button{
					Id:     "new",
					Label:  "form.button.view",
					Style:  "success",
					Action: fmt.Sprintf("%s/%v", action, row[0]),
					Method: "GET",
				})
			}
			if edit {
				bg = append(bg, &Button{
					Id:     "new",
					Label:  "form.button.edit",
					Style:  "warning",
					Action: fmt.Sprintf("%s/%v/edit", action, row[0]),
					Method: "GET",
				})
			}
			if remove {
				bg = append(bg, &Button{
					Id:     "new",
					Label:  "form.button.remove",
					Style:  "danger",
					Action: fmt.Sprintf("%s/%v", action, row[0]),
					Method: "DELETE",
				})
			}
			row = append(row, bg)
			bodyV = append(bodyV, row)
		}
	} else {
		bodyV = body()
	}

	return &Table{
		Id:     id,
		Title:  fmt.Sprintf("table.title.%s", id),
		Action: action,
		Header: header,
		Body:   bodyV,
		New:    newB,
		Pager:  pager,
	}
}

type Table struct {
	Id     string          `json:"id"`
	Action string          `json:"action"`
	Title  string          `json:"title"`
	Header []*Th           `json:"header"`
	Body   [][]interface{} `json:"body"`
	New    *Button         `json:"new"`
	Pager  *Pager          `json:"pager"`
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

func (p *Form) AddTextField(id string, value interface{}, required bool) {
	p.Fields = append(p.Fields, &TextField{
		Field: Field{
			Id:   id,
			Type: "text",
		},
		Label:       p.label(id),
		Value:       value,
		Size:        8,
		Required:    required,
		Placeholder: p.placeholder(id),
	})
}

func (p *Form) AddEmailField(id string, value interface{}, required bool) {
	p.Fields = append(p.Fields, &TextField{
		Field: Field{
			Id:   id,
			Type: "email",
		},
		Label:       "form.field.email",
		Value:       value,
		Size:        7,
		Required:    required,
		Placeholder: "form.placeholder.email",
	})
}

func (p *Form) AddPasswordField(id string, required, confirmed bool) {

	p.AddField(&PasswordField{
		Field: Field{
			Id:   id,
			Type: "password",
		},
		Label:       "form.field.password",
		Required:    required,
		Size:        6,
		Placeholder: "form.placeholder.password",
	})
	if confirmed {
		p.AddField(&PasswordField{
			Field: Field{
				Id:   "re_" + id,
				Type: "password",
			},
			Label:       "form.field.re_password",
			Required:    required,
			Size:        6,
			Placeholder: "form.placeholder.re_password",
		})
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

func (p *Form) AddSelectField(id string, ofn func() []Option) {
	p.AddField(&SelectField{
		Field: Field{
			Id:   id,
			Type: "select",
		},
		Label:   p.label(id),
		Options: ofn(),
	})
}

func (p *Form) AddCheckboxGroupField(id string, ofn func() []Option) {
	p.AddField(&GroupField{
		Field: Field{
			Id:   id,
			Type: "checkbox",
		},
		Label:   p.label(id),
		Options: ofn(),
	})
}

func (p *Form) AddRadioGroupField(id string, ofn func() []Option) {
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
	Label    string   `json:"label"`
	Options  []Option `json:"options"`
	Multiple bool     `json:"multi"`
}
type GroupField struct {
	Field
	Label   string   `json:"label"`
	Options []Option `json:"options"`
}

type Button struct {
	Id      string `json:"id"`
	Label   string `json:"label"`
	Action  string `json:"action"`
	Method  string `json:"method"`
	Confirm bool   `json:"confirm"`
	Style   string `json:"style"`
}
