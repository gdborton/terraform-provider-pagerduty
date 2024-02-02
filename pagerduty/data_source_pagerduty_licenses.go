package pagerduty

import (
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyLicenses() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyLicensesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"licenses": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: licenseSchema,
				},
			},
		},
	}
}

func dataSourcePagerDutyLicensesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Fetching PagerDuty Licenses")

	return retry.Retry(5*time.Minute, func() *retry.RetryError {
		licenses, _, err := client.Licenses.List()
		if err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			// Delaying retry by 30s as recommended by PagerDuty
			// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
			time.Sleep(30 * time.Second)
			return retry.RetryableError(err)
		}

		newLicenses := flattenLicenses(licenses)
		d.Set("licenses", newLicenses)

		if idValue, ok := d.GetOk("id"); !ok {
			d.SetId(id.UniqueId())
		} else {
			d.SetId(idValue.(string))
		}
		return nil
	})
}

func flattenLicenses(licenses []*pagerduty.License) []map[string]interface{} {
	updated := make([]map[string]interface{}, len(licenses))
	for i, license := range licenses {
		updated[i] = flattenLicense(license)
	}

	return updated
}

func flattenLicense(l *pagerduty.License) map[string]interface{} {
	license := map[string]interface{}{
		"id":                    l.ID,
		"type":                  l.Type,
		"name":                  l.Name,
		"description":           l.Description,
		"summary":               l.Summary,
		"role_group":            l.RoleGroup,
		"allocations_available": l.AllocationsAvailable,
		"current_value":         l.CurrentValue,
		"self":                  l.Self,
		"html_url":              l.HTMLURL,
		"valid_roles":           l.ValidRoles,
	}

	return license
}
