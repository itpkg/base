package base_test

import (
	"encoding/json"
	"testing"

	"github.com/itpkg/base"
)

func TestForm(t *testing.T) {
	fm := base.NewForm("fmId", "demo", "/form.action")
	fm.AddButton("view", "btnAction", "PUT", true, "success")
	fm.AddHiddenField("hid", "hidVal")
	fm.AddTextField("tid", "txtVal")
	fm.AddEmailField("eid", "aaa@aaa.com")
	fm.AddPasswordField("pid", true)
	fm.AddPasswordField("pid", false)
	fm.AddTextareaField("taid", "text area")
	fm.AddMarkdownField("taid", "## 中文\n### subject")
	fm.AddHtmlField("taid", "<h1>title</h1><hr/>")

	items := []base.Option{
		base.Option{
			Id:      "item-1",
			Label:   "label-1",
			Checked: true,
		}, base.Option{
			Id:    "item-2",
			Label: "label-2",
		},
	}

	fm.AddSelectField("sid", func() []base.Option { return items })
	fm.AddRadioGroupField("rid", func() []base.Option { return items })
	fm.AddCheckboxGroupField("cid", func() []base.Option { return items })

	checkJson(t, &fm)
}

var pager = &base.Pager{Total: 127, Size: 30, Index: 3}

func TestTable(t *testing.T) {
	items := [][]interface{}{[]interface{}{1, "aaa1", "bbb1"}, []interface{}{2, "aaa2", "bbb2"}}
	tab := base.NewTable(
		"tid",
		"/users",
		[]*base.Th{&base.Th{Label: "H1", Width: "10"}, &base.Th{Label: "H2", Width: "20"}},
		func() [][]interface{} { return items },
		true, true, true, true,
		pager,
	)
	checkJson(t, &tab)
}

func TestList(t *testing.T) {
	items := []interface{}{"l1", 222, 444, true}
	lst := base.NewList(
		"lid",
		"/users",
		func() []interface{} { return items },
		pager,
	)
	checkJson(t, &lst)
}

func checkJson(t *testing.T, obj interface{}) {

	if js, err := json.MarshalIndent(obj, "", "  "); err == nil {
		t.Logf(string(js))
	} else {
		t.Errorf("Form json FAILED!")
	}
}
