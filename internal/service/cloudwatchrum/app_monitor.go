package cloudwatchrum

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAppMonitor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppMonitorCreate,
		ReadContext:   resourceAppMonitorRead,
		UpdateContext: resourceAppMonitorUpdate,
		DeleteContext: resourceAppMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"app_monitor_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"enable_xray": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"excluded_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"favorite_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"guest_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"included_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"session_sample_rate": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidTypeStringNullableFloat,
						},
						"telemetries": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cw_log_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchRUMConn
	name := d.Get("name").(string)

	input := expandCreateAppMonitorInput(d, meta)

	_, err := conn.CreateAppMonitorWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("failed creating CloudWatch RUM App Monitor (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceAppMonitorRead(ctx, d, meta)
}

func resourceAppMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchRUMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	name := d.Get("name").(string)

	monitor, err := FindAppMonitorByName(ctx, conn, name)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CloudWatch RUM App Monitor %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading CloudWatch RUM App Monitor (%s): %s", name, err)
	}

	if monitor == nil {
		if d.IsNewResource() {
			return diag.Errorf("error reading CloudWatch RUM App Monitor (%s): not found", name)
		}

		log.Printf("[WARN] CloudWatch RUM App Monitor %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	d.Set("domain", monitor.Domain)
	d.Set("name", monitor.Name)
	d.Set("cw_log_enabled", monitor.DataStorage.CwLog.CwLogEnabled)

	if monitor.AppMonitorConfiguration != nil {
		if err := d.Set("app_monitor_configuration", []interface{}{flattenAppMonitorConfig(monitor.AppMonitorConfiguration)}); err != nil {
			return diag.Errorf("error setting app_monitor_configuration: %s", err)
		}
	} else {
		d.Set("app_monitor_configuration", nil)
	}

	tags := KeyValueTags(monitor.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "rum",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/appmonitor/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceAppMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchRUMConn
	name := d.Id()

	input := expandUpdateAppMonitorInput(d, meta)

	_, err := conn.UpdateAppMonitorWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("error updating CloudWatch RUM App Monitor (%s): %s", name, err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAppMonitorRead(ctx, d, meta)
}

func resourceAppMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchRUMConn
	name := d.Id()

	input := cloudwatchrum.DeleteAppMonitorInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteAppMonitorWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error deleting CloudWatch RUM App Monitor (%s): %s", name, err)
	}

	return nil
}

func expandCreateAppMonitorInput(d *schema.ResourceData, meta interface{}) cloudwatchrum.CreateAppMonitorInput {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	out := cloudwatchrum.CreateAppMonitorInput{
		Domain: aws.String(d.Get("domain").(string)),
		Name:   aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("app_monitor_configuration"); ok {
		out.AppMonitorConfiguration = expandAppMonitorConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("cw_log_enabled"); ok {
		out.CwLogEnabled = aws.Bool(v.(bool))
	}

	if len(tags) > 0 {
		out.Tags = Tags(tags.IgnoreAWS())
	}

	return out
}

func expandAppMonitorConfig(l []interface{}) *cloudwatchrum.AppMonitorConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := cloudwatchrum.AppMonitorConfiguration{}

	if v, ok := m["allow_cookies"].(bool); ok {
		config.AllowCookies = aws.Bool(v)
	}

	if v, ok := m["enable_xray"].(bool); ok {
		config.EnableXRay = aws.Bool(v)
	}

	if v, ok := m["excluded_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.ExcludedPages = flex.ExpandStringSet(v)
	}

	if v, ok := m["favorite_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.FavoritePages = flex.ExpandStringSet(v)
	}

	if v, ok := m["guest_role_arn"].(string); ok && v != "" {
		config.GuestRoleArn = aws.String(v)
	}

	if v, ok := m["identity_pool_id"].(string); ok && v != "" {
		config.IdentityPoolId = aws.String(v)
	}

	if v, ok := m["included_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.IncludedPages = flex.ExpandStringSet(v)
	}

	if v, ok := m["session_sample_rate"].(string); ok && v != "" {
		value, _ := strconv.ParseFloat(v, 64)
		config.SessionSampleRate = aws.Float64(value)
	}

	if v, ok := m["telemetries"].(*schema.Set); ok && v.Len() > 0 {
		config.Telemetries = flex.ExpandStringSet(v)
	}

	return &config
}

func expandUpdateAppMonitorInput(d *schema.ResourceData, meta interface{}) cloudwatchrum.UpdateAppMonitorInput {
	other := expandCreateAppMonitorInput(d, meta)

	return cloudwatchrum.UpdateAppMonitorInput{
		AppMonitorConfiguration: other.AppMonitorConfiguration,
		CwLogEnabled:            other.CwLogEnabled,
		Domain:                  other.Domain,
		Name:                    other.Name,
	}
}

func flattenAppMonitorConfig(config *cloudwatchrum.AppMonitorConfiguration) map[string]interface{} {
	if config == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := config.AllowCookies; v != nil {
		tfMap["allow_cookies"] = aws.BoolValue(v)
	}

	if v := config.EnableXRay; v != nil {
		tfMap["enable_xray"] = aws.BoolValue(v)
	}

	if v := config.ExcludedPages; len(v) > 0 {
		tfMap["excluded_pages"] = flex.FlattenStringSet(v)
	} else {
		tfMap["excluded_pages"] = nil
	}

	if v := config.FavoritePages; len(v) > 0 {
		tfMap["favorite_pages"] = flex.FlattenStringSet(v)
	} else {
		tfMap["favorite_pages"] = nil
	}

	if v := config.GuestRoleArn; v != nil {
		tfMap["guest_role_arn"] = aws.StringValue(v)
	}

	if v := config.IdentityPoolId; v != nil {
		tfMap["identity_pool_id"] = aws.StringValue(v)
	}

	if v := config.IncludedPages; len(v) > 0 {
		tfMap["included_pages"] = flex.FlattenStringSet(v)
	} else {
		tfMap["included_pages"] = nil
	}

	if v := config.SessionSampleRate; v == nil {
		tfMap["session_sample_rate"] = ""
	} else {
		tfMap["session_sample_rate"] = strconv.FormatFloat(aws.Float64Value(v), 'f', -1, 64)
	}

	if v := config.Telemetries; len(v) > 0 {
		tfMap["telemetries"] = flex.FlattenStringSet(v)
	} else {
		tfMap["telemetries"] = nil
	}

	return tfMap
}
