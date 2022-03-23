package cloudwatchrum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchrum "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchrum"
)

func TestAccCloudWatchRUMAppMonitor_basic(t *testing.T) {
	var am cloudwatchrum.AppMonitor
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatchrum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchrum.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRUMAppMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchRUMAppMonitorExists(resourceName, &am),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("MyTestAppMonitor-%d", rInt)),
					testAccCheckCloudWatchRUMAppMonitorName(&am, fmt.Sprintf("MyTestAppMonitor-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "domain", "amazondomains.com"),
					testAccCheckCloudWatchRUMAppMonitorDomain(&am, "amazondomains.com"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.allow_cookies", "true"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.enable_xray", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorModifiedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchRUMAppMonitorExists(resourceName, &am),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", "true"),
					testAccCheckCloudWatchRUMAppMonitorCWLogEnabled(&am, true),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("MyTestAppMonitor-%d", rInt)),
					testAccCheckCloudWatchRUMAppMonitorName(&am, fmt.Sprintf("MyTestAppMonitor-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "domain", "amazondomains.com"),
					testAccCheckCloudWatchRUMAppMonitorDomain(&am, "amazondomains.com"),
					testAccCheckCloudWatchRUMAppMonitorConfig(&am, &cloudwatchrum.AppMonitorConfiguration{
						AllowCookies: aws.Bool(true),
					}),
				),
			},
		},
	})
}

func TestAccCloudWatchRUMAppMonitor_disappears(t *testing.T) {
	var am cloudwatchrum.AppMonitor
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatchrum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchrum.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRUMAppMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchRUMAppMonitorExists(resourceName, &am),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudWatchRUMAppMonitorConfig(am *cloudwatchrum.AppMonitor,
	amc *cloudwatchrum.AppMonitorConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		given := am.AppMonitorConfiguration
		expected := amc

		if (given.AllowCookies != nil) != (expected.AllowCookies != nil) {
			return fmt.Errorf("Expected allow cookies to be present: %t, received: %t",
				expected.AllowCookies != nil, given.AllowCookies != nil)
		} else if *given.AllowCookies != *expected.AllowCookies {
			return fmt.Errorf("Expected allowed cookies: %t, received: %t",
				*expected.AllowCookies, *given.AllowCookies)
		}

		if (given.EnableXRay != nil) != (expected.EnableXRay != nil) {
			return fmt.Errorf("Expected enable xray to be present: %t, received: %t",
				expected.EnableXRay != nil, given.EnableXRay != nil)
		} else if *given.EnableXRay != *expected.EnableXRay {
			return fmt.Errorf("Expected enable xray: %t, received: %t",
				*expected.EnableXRay, *given.EnableXRay)
		}

		if len(expected.ExcludedPages) > 0 || len(given.ExcludedPages) > 0 {
			e, g := aws.StringValueSlice(expected.ExcludedPages), aws.StringValueSlice(given.ExcludedPages)

			if len(e) != len(g) {
				return fmt.Errorf("Expected %d excluded pages, received %d", len(e), len(g))
			}

			for ei, ev := range e {
				gv := g[ei]
				if gv != ev {
					return fmt.Errorf("Expected excluded page %d to be %s, received %s", ei, ev, gv)
				}
			}
		}

		if len(expected.FavoritePages) > 0 || len(given.FavoritePages) > 0 {
			e, g := aws.StringValueSlice(expected.FavoritePages), aws.StringValueSlice(given.FavoritePages)

			if len(e) != len(g) {
				return fmt.Errorf("Expected %d favorite pages, received %d", len(e), len(g))
			}

			for ei, ev := range e {
				gv := g[ei]
				if gv != ev {
					return fmt.Errorf("Expected favorite page %d to be %s, received %s", ei, ev, gv)
				}
			}
		}

		if (given.GuestRoleArn != nil) != (expected.GuestRoleArn != nil) {
			return fmt.Errorf("Expected guest role arn to be present: %t, received: %t",
				expected.GuestRoleArn != nil, given.GuestRoleArn != nil)
		} else if *given.GuestRoleArn != *expected.GuestRoleArn {
			return fmt.Errorf("Expected guest role arn: %s, received: %s",
				*expected.GuestRoleArn, *given.GuestRoleArn)
		}

		if (given.IdentityPoolId != nil) != (expected.IdentityPoolId != nil) {
			return fmt.Errorf("Expected identity pool id to be present: %t, received: %t",
				expected.IdentityPoolId != nil, given.IdentityPoolId != nil)
		} else if *given.IdentityPoolId != *expected.IdentityPoolId {
			return fmt.Errorf("Expected identity pool id: %s, received: %s",
				*expected.IdentityPoolId, *given.IdentityPoolId)
		}

		if len(expected.IncludedPages) > 0 || len(given.IncludedPages) > 0 {
			e, g := aws.StringValueSlice(expected.IncludedPages), aws.StringValueSlice(given.IncludedPages)

			if len(e) != len(g) {
				return fmt.Errorf("Expected %d included pages, received %d", len(e), len(g))
			}

			for ei, ev := range e {
				gv := g[ei]
				if gv != ev {
					return fmt.Errorf("Expected included page %d to be %s, received %s", ei, ev, gv)
				}
			}
		}

		if (given.SessionSampleRate != nil) != (expected.SessionSampleRate != nil) {
			return fmt.Errorf("Expected session sample rate to be present: %t, received: %t",
				expected.SessionSampleRate != nil, given.SessionSampleRate != nil)
		} else if (given.SessionSampleRate != nil) && *given.SessionSampleRate != *expected.SessionSampleRate {
			return fmt.Errorf("Expected session sample rate: %g, received: %g",
				*expected.SessionSampleRate, *given.SessionSampleRate)
		}

		if len(expected.Telemetries) > 0 || len(given.Telemetries) > 0 {
			e, g := aws.StringValueSlice(expected.Telemetries), aws.StringValueSlice(given.Telemetries)

			if len(e) != len(g) {
				return fmt.Errorf("Expected %d telemetries, received %d", len(e), len(g))
			}

			for ei, ev := range e {
				gv := g[ei]
				if gv != ev {
					return fmt.Errorf("Expected telemetry %d to be %s, received %s", ei, ev, gv)
				}
			}
		}

		return nil
	}
}

func testAccCheckCloudWatchRUMAppMonitorCWLogEnabled(am *cloudwatchrum.AppMonitor, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if enabled != *am.DataStorage.CwLog.CwLogEnabled {
			return fmt.Errorf("Expected app monitor CW log enabled: %t, given: %t", enabled, *am.DataStorage.CwLog.CwLogEnabled)
		}
		return nil
	}
}

func testAccCheckCloudWatchRUMAppMonitorDomain(am *cloudwatchrum.AppMonitor, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if am.Domain == nil {
			if domain != "" {
				return fmt.Errorf("Received empty domain, expected: %s", domain)
			}
			return nil
		}

		if domain != *am.Domain {
			return fmt.Errorf("Expected domain: %s, given: %s", domain, *am.Domain)
		}
		return nil
	}
}

func testAccCheckCloudWatchRUMAppMonitorName(am *cloudwatchrum.AppMonitor, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if name != *am.Name {
			return fmt.Errorf("Expected app monitor name: %s, given: %s", name, *am.Name)
		}
		return nil
	}
}

func testAccCheckCloudWatchRUMAppMonitorExists(n string, am *cloudwatchrum.AppMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchRUMConn
		appMonitor, err := tfcloudwatchrum.FindAppMonitorByName(context.Background(), conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*am = *appMonitor

		return nil
	}
}

func testAccCheckRUMAppMonitorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchRUMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatchrum_app_monitor" {
			continue
		}

		appMonitor, err := tfcloudwatchrum.FindAppMonitorByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error reading CloudWatch RUM App Monitor (%s): %w", rs.Primary.Attributes["name"], err)
		}

		if appMonitor != nil {
			return fmt.Errorf("CloudWatch RUM App Monitor (%s) still exists", rs.Primary.Attributes["name"])
		}
	}

	return nil
}

func testAccAppMonitorConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchrum_app_monitor" "test" {
  domain         = "amazondomains.com"
  name           = "MyTestAppMonitor-%d"
}
`, rInt)
}

func testAccAppMonitorModifiedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchrum_app_monitor" "test" {
  cw_log_enabled = true
  domain         = "amazondomains.com"
  name           = "MyTestAppMonitor-%d"

  app_monitor_configuration {
    allow_cookies = true
  }
}
`, rInt)
}
