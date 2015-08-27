package base

import (
	"fmt"
)

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
	Locale string        `json:"locale"`
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
	Locale string          `json:"locale"`
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
func NewForm(id, resource, action string) *Form {
	return &Form{
		Id:        id,
		Resource:  resource,
		Title:     fmt.Sprintf("form.title.%s", id),
		Action:    action,
		Method:    "POST",
		Multipart: false,
		Fields:    make([]interface{}, 0),
		Buttons:   make([]interface{}, 0),
		Submit:    "form.button.submit",
		Reset:     "form.button.reset",
	}
}

type Form struct {
	Resource  string        `json:"-"`
	Locale    string        `json:"locale"`
	Id        string        `json:"id"`
	Title     string        `json:"title"`
	Method    string        `json:"method"`
	Action    string        `json:"action"`
	Multipart bool          `json:"multipart"`
	Fields    []interface{} `json:"fields"`
	Buttons   []interface{} `json:"buttons"`
	Submit    string        `json:"submit"`
	Reset     string        `json:"reset"`
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

func (p *Form) AddTextField(id string, value interface{}) {
	p.Fields = append(p.Fields, TextField{
		Field: Field{
			Id:   id,
			Type: "text",
		},
		Label: p.label(id),
		Value: value,
	})
}

func (p *Form) AddEmailField(id string, value interface{}) {
	p.Fields = append(p.Fields, TextField{
		Field: Field{
			Id:   id,
			Type: "email",
		},
		Label: "form.email",
		Value: value,
	})
}

func (p *Form) AddPasswordField(id string, confirm bool) {
	var cl interface{}
	if confirm {
		cl = "form.password_confirm"
	} else {
		cl = nil
	}
	p.AddField(&PasswordField{
		Field: Field{
			Id:   id,
			Type: "password",
		},
		Label:   "form.password",
		Confirm: cl,
	})
}

func (p *Form) AddTextareaField(id string, value interface{}) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "text",
		},
		Label: p.label(id),
		Value: value,
		Rows:  10,
	})
}

func (p *Form) AddMarkdownField(id string, value interface{}) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "markdown",
		},
		Label: p.label(id),
		Value: value,
		Rows:  10,
	})
}

func (p *Form) AddHtmlField(id string, value interface{}) {
	p.Fields = append(p.Fields, TextareaField{
		Field: Field{
			Id:   id,
			Type: "html",
		},
		Label: p.label(id),
		Value: value,
		Rows:  10,
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
	return fmt.Sprintf("form.%s.%s", p.Resource, id)
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
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}
type TextareaField struct {
	Field
	Label string      `json:"label"`
	Value interface{} `json:"value"`
	Rows  int         `json:"rows"`
}
type PasswordField struct {
	Field
	Label   string      `json:"label"`
	Confirm interface{} `json:"confirm"`
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
