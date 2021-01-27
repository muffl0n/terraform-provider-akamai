package appsec

import (
	"context"
	"errors"
	"strconv"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// appsec v1
//
// https://developer.akamai.com/api/cloud_security/application_security/v1.html
func resourceApiRequestConstraints() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiRequestConstraintsUpdate,
		ReadContext:   resourceApiRequestConstraintsRead,
		UpdateContext: resourceApiRequestConstraintsUpdate,
		DeleteContext: resourceApiRequestConstraintsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"security_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_endpoint_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					Alert,
					Deny,
					None,
				}, false),
			},
			"output_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Text Export representation",
			},
		},
	}
}

func resourceApiRequestConstraintsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceApiRequestConstraintsRead")

	getApiRequestConstraints := appsec.GetApiRequestConstraintsRequest{}

	configid, err := tools.GetIntValue("config_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	getApiRequestConstraints.ConfigID = configid

	version, err := tools.GetIntValue("version", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	getApiRequestConstraints.Version = version

	policyid, err := tools.GetStringValue("security_policy_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	getApiRequestConstraints.PolicyID = policyid

	ApiID, err := tools.GetIntValue("api_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	getApiRequestConstraints.ApiID = ApiID

	apirequestconstraints, err := client.GetApiRequestConstraints(ctx, getApiRequestConstraints)
	if err != nil {
		logger.Errorf("calling 'getApiRequestConstraints': %s", err.Error())
		return diag.FromErr(err)
	}

	ots := OutputTemplates{}
	InitTemplates(ots)

	outputtext, err := RenderTemplates(ots, "apirequestconstraintsDS", apirequestconstraints)
	if err == nil {
		d.Set("output_text", outputtext)
	}

	d.SetId(strconv.Itoa(getApiRequestConstraints.ConfigID))

	return nil
}

func resourceApiRequestConstraintsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceApiRequestConstraintsRemove")

	getPolicyProtections := appsec.GetPolicyProtectionsRequest{}
	removeApiRequestConstraints := appsec.RemoveApiRequestConstraintsRequest{}
	removePolicyProtections := appsec.RemovePolicyProtectionsRequest{}

	configid, err := tools.GetIntValue("config_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	version, err := tools.GetIntValue("version", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	policyid, err := tools.GetStringValue("security_policy_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	getPolicyProtections.ConfigID = configid
	removeApiRequestConstraints.ConfigID = configid
	removePolicyProtections.ConfigID = configid

	getPolicyProtections.Version = version
	removeApiRequestConstraints.Version = version
	removePolicyProtections.Version = version

	getPolicyProtections.PolicyID = policyid
	removeApiRequestConstraints.PolicyID = policyid
	removePolicyProtections.PolicyID = policyid

	policyprotections, err := client.GetPolicyProtections(ctx, getPolicyProtections)
	if err != nil {
		logger.Errorf("calling 'getPolicyProtections': %s", err.Error())
		return diag.FromErr(err)
	}

	apiEndpointID, err := tools.GetIntValue("api_endpoint_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	removeApiRequestConstraints.ApiID = apiEndpointID

	if removeApiRequestConstraints.ApiID == 0 {
		if policyprotections.ApplyAPIConstraints == true {
			removePolicyProtections.ApplyAPIConstraints = false
			removePolicyProtections.ApplyApplicationLayerControls = policyprotections.ApplyApplicationLayerControls
			removePolicyProtections.ApplyBotmanControls = policyprotections.ApplyBotmanControls
			removePolicyProtections.ApplyNetworkLayerControls = policyprotections.ApplyNetworkLayerControls
			removePolicyProtections.ApplyRateControls = policyprotections.ApplyRateControls
			removePolicyProtections.ApplyReputationControls = policyprotections.ApplyReputationControls
			removePolicyProtections.ApplySlowPostControls = policyprotections.ApplySlowPostControls

			_, errd := client.RemovePolicyProtections(ctx, removePolicyProtections)
			if errd != nil {
				logger.Errorf("calling 'removePolicyProtections': %s", errd.Error())
				return diag.FromErr(errd)
			}
		}
	} else {
		removeApiRequestConstraints.Action = "none"
		_, erru := client.RemoveApiRequestConstraints(ctx, removeApiRequestConstraints)
		if erru != nil {
			logger.Errorf("calling 'removeApiRequestConstraints': %s", erru.Error())
			return diag.FromErr(erru)
		}
	}

	d.SetId("")
	return nil
}

func resourceApiRequestConstraintsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceApiRequestConstraintsUpdate")

	updateApiRequestConstraints := appsec.UpdateApiRequestConstraintsRequest{}

	configid, err := tools.GetIntValue("config_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	updateApiRequestConstraints.ConfigID = configid

	version, err := tools.GetIntValue("version", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	updateApiRequestConstraints.Version = version

	policyid, err := tools.GetStringValue("security_policy_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	updateApiRequestConstraints.PolicyID = policyid

	apiEndpointID, err := tools.GetIntValue("api_endpoint_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	updateApiRequestConstraints.ApiID = apiEndpointID

	action, err := tools.GetStringValue("action", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	updateApiRequestConstraints.Action = action

	_, erru := client.UpdateApiRequestConstraints(ctx, updateApiRequestConstraints)
	if erru != nil {
		logger.Errorf("calling 'updateApiRequestConstraints': %s", erru.Error())
		return diag.FromErr(erru)
	}

	return resourceApiRequestConstraintsRead(ctx, d, m)
}
