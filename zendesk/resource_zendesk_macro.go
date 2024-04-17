package zendesk

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	client "github.com/nukosuke/go-zendesk/zendesk"
)

// https://developer.zendesk.com/api-reference/ticketing/business-rules/macros/
func resourceZendeskMacro() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a macro resource.",
		CreateContext: resourceZendeskMacroCreate,
		ReadContext:   resourceZendeskMacroRead,
		UpdateContext: resourceZendeskMacroUpdate,
		DeleteContext: resourceZendeskMacroDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"url": {
				Description: "The URL for this macro.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"title": {
				Description: "The title of the macro.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the macro.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"position": {
				Description: "The position of the macro.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"restriction": {
				Description: "The restriction of the macro.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"active": {
				Description: "The active status of the macro.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"actions": {
				Description: "The actions of the macro.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Description: "The field of the action.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"value": {
							Description: "The value of the action.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// marshalMacro encodes the provided macro into the provided resource data
func marshalMacro(field client.Macro, d identifiableGetterSetter) error {
	fields := map[string]interface{}{
		"url":         field.URL,
		"title":       field.Title,
		"description": field.Description,
		"position":    field.Position,
		"restriction": field.Restriction,
		"active":      field.Active,
		"actions":     field.Actions,
	}

	err := setSchemaFields(d, fields)
	if err != nil {
		return err
	}

	return nil
}

// unmarshalMacro parses the provided ResourceData and returns a macro
func unmarshalMacro(d identifiableGetterSetter) (client.Macro, error) {
	m := client.Macro{}

	if v := d.Id(); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return m, fmt.Errorf("could not parse ticket field id %s: %v", v, err)
		}
		m.ID = id
	}

	if v, ok := d.GetOk("url"); ok {
		m.URL = v.(string)
	}

	if v, ok := d.GetOk("title"); ok {
		m.Title = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		m.Description = v.(string)
	}

	if v, ok := d.GetOk("position"); ok {
		m.Position = v.(int)
	}

	if v, ok := d.GetOk("active"); ok {
		m.Active = v.(bool)
	}

	if v, ok := d.GetOk("restriction"); ok {
		m.Restriction = v.(string)
	}

	if v, ok := d.GetOk("actions"); ok {
		actions := v.([]interface{})
		for _, action := range actions {
			actionMap := action.(map[string]interface{})
			m.Actions = append(m.Actions, client.MacroAction{
				Field: actionMap["field"].(string),
				Value: actionMap["value"].(string),
			})
		}
	}

	return m, nil
}

// resourceZendeskMacroCreate creates a new macro
func resourceZendeskMacroCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zd := meta.(*client.Client)
	return createMacro(ctx, d, zd)
}

// createMacro creates a new macro
func createMacro(ctx context.Context, d identifiableGetterSetter, zd client.MacroAPI) diag.Diagnostics {
	var diags diag.Diagnostics

	m, err := unmarshalMacro(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Actual API request
	m, err = zd.CreateMacro(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", m.ID))

	err = marshalMacro(m, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceZendeskMacroRead reads a macro
func resourceZendeskMacroRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zd := meta.(*client.Client)
	return readMacro(ctx, d, zd)
}

// readMacro reads a macro
func readMacro(ctx context.Context, d identifiableGetterSetter, zd client.MacroAPI) diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	macro, err := zd.GetMacro(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = marshalMacro(macro, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceZendeskMacroUpdate updates a macro
func resourceZendeskMacroUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zd := meta.(*client.Client)
	return updateMacro(ctx, d, zd)
}

// updateMacro updates a macro
func updateMacro(ctx context.Context, d identifiableGetterSetter, zd client.MacroAPI) diag.Diagnostics {
	var diags diag.Diagnostics

	m, err := unmarshalMacro(d)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	// Actual API request
	m, err = zd.UpdateMacro(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}

	err = marshalMacro(m, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceZendeskMacroDelete deletes a macro
func resourceZendeskMacroDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zd := meta.(*client.Client)
	return deleteMacro(ctx, d, zd)
}

// deleteMacro deletes a macro
func deleteMacro(ctx context.Context, d identifiable, zd client.MacroAPI) diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	err = zd.DeleteMacro(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
