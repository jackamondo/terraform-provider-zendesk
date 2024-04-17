package zendesk

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/nukosuke/go-zendesk/zendesk/mock"
)

func TestReadMacro(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	id := 1234
	gs := &identifiableMapGetterSetter{
		mapGetterSetter: make(mapGetterSetter),
		id:              strconv.Itoa(id),
	}

	now := time.Now()
	field := zendesk.Macro{
		ID:          int64(id),
		URL:         "foobar",
		Title:       "foobar",
		Description: "foobar",
		Position:    int(50),
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Restriction: "restriction",
		Actions: []zendesk.MacroAction{
			{
				Field: "status",
				Value: "open",
			},
		},
	}

	m.EXPECT().GetMacro(Any(), Any()).Return(field, nil)
	if diags := readMacro(context.Background(), gs, m); len(diags) != 0 {
		t.Fatal("readMacro returned an error")
	}

	if v := gs.mapGetterSetter["url"]; v != field.URL {
		t.Fatalf("url field %v does not have expected value %v", v, field.URL)
	}

	if v := gs.mapGetterSetter["title"]; v != field.Title {
		t.Fatalf("type field %v does not have expected value %v", v, field.Title)
	}

}

func TestDeleteMacro(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := &identifiableMapGetterSetter{
		id: "12345",
	}

	m.EXPECT().DeleteMacro(Any(), Eq(int64(12345))).Return(nil)
	if diags := deleteMacro(context.Background(), i, m); len(diags) != 0 {
		t.Fatal("readMacro returned an error")
	}
}

func TestUpdateMacro(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := &identifiableMapGetterSetter{
		id:              "12345",
		mapGetterSetter: make(mapGetterSetter),
	}

	m.EXPECT().UpdateMacro(Any(), Eq(int64(12345)), Any()).Return(zendesk.Macro{}, nil)
	if diags := updateMacro(context.Background(), i, m); len(diags) != 0 {
		t.Fatal("readMacro returned an error")
	}
}

func TestCreateMacro(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := &identifiableMapGetterSetter{
		mapGetterSetter: make(mapGetterSetter),
	}

	out := zendesk.Macro{
		ID: 12345,
	}

	m.EXPECT().CreateMacro(Any(), Any()).Return(out, nil)
	if diags := createMacro(context.Background(), i, m); len(diags) != 0 {
		t.Fatal("create macro returned an error")
	}

	if v := i.Id(); v != "12345" {
		t.Fatalf("Create did not set resource id. Id was %s", v)
	}
}

func TestMarshalMacro(t *testing.T) {
	m := &identifiableMapGetterSetter{
		id: "1234",
		mapGetterSetter: mapGetterSetter{
			"url":         "https://example.zendesk.com/api/v2/macros/360011737434.json",
			"title":       "title",
			"description": "description",
			"position":    int(12),
			"active":      true,
			"restriction": "restriction",
			"actions": []interface{}{
				map[string]interface{}{
					"field": "status",
					"value": "open",
				},
			},
		},
	}

	macro, err := unmarshalMacro(m)
	if err != nil {
		t.Fatalf("Could not unmarshal macro: %v", err)
	}

	if v, ok := m.Get("url").(string); !ok || macro.URL != v {
		t.Fatalf("macro had URL value %v. should have been %v", macro.URL, v)
	}

	if v, ok := m.Get("title").(string); !ok || macro.Title != v {
		t.Fatalf("macro had incorrect title value %v. should have been %v", macro.Title, v)
	}

	if v, ok := m.Get("description").(string); !ok || macro.Description != v {
		t.Fatalf("macro had incorrect description value %v. should have been %v", macro.Description, v)
	}

	if v, ok := m.Get("position").(int); !ok {
		t.Fatalf("macro had incorrect position value %d. should have been %v", macro.Position, v)
	}

	if v, ok := m.Get("active").(bool); !ok {
		t.Fatalf("macro had incorrect active value %t. should have been %v", macro.Active, v)
	}
}

func testMacroDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(zendesk.MacroAPI)

	for k, r := range s.RootModule().Resources {
		if strings.HasPrefix(k, "data") {
			continue
		}

		if r.Type != "zendesk_macro" {
			continue
		}

		id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		_, err = client.GetMacro(context.Background(), id)
		if err == nil {
			return fmt.Errorf("did not get error from zendesk when trying to fetch the destroyed macro named %s", k)
		}

		zd, ok := err.(zendesk.Error)
		if !ok {
			return fmt.Errorf("error %v cannot be asserted as a zendesk error", err)
		}

		if zd.Status() != http.StatusNotFound {
			return fmt.Errorf("did not get a not found error after destroy. error was %v", zd)
		}

	}

	return nil
}
