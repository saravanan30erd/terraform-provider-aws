package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAWSLaunchConfigurationMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"#1": {
			StateVersion: 0,
			Attributes: map[string]string{
				"ebs_block_device.#":                                "1",
				"ebs_block_device.1954862932.delete_on_termination": "true",
				"ebs_block_device.1954862932.device_name":           "/dev/sda2",
				"ebs_block_device.1954862932.encrypted":             "false",
				"ebs_block_device.1954862932.iops":                  "0",
				"ebs_block_device.1954862932.snapshot_id":           "",
				"ebs_block_device.1954862932.volume_size":           "0",
				"ebs_block_device.1954862932.volume_type":           "",
			},
			Expected: map[string]string{
				"ebs_block_device.2140691768.delete_on_termination": "",
				"ebs_block_device.2140691768.device_name":           "/dev/sda2",
				"ebs_block_device.2140691768.encrypted":             "false",
				"ebs_block_device.2140691768.iops":                  "0",
				"ebs_block_device.2140691768.no_device":             "",
				"ebs_block_device.2140691768.snapshot_id":           "",
				"ebs_block_device.2140691768.volume_size":           "0",
				"ebs_block_device.2140691768.volume_type":           "",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "lc-abc123",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsLaunchConfigurationMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestAWSLaunchConfigurationMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsLaunchConfigurationMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	is, err = resourceAwsLaunchConfigurationMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
