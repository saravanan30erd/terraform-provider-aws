package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
)

func resourceAwsLaunchConfigurationMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	log.Printf("[DEBUG] Performing Launch Config migration from %d", v)
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Launch Configuration State v0; migrating to v1")
		return migrateLaunchConfigurationStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateLaunchConfigurationStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is == nil || is.Attributes == nil {
		return nil, nil
	}

	newResource := resourceAwsLaunchConfiguration().Schema["ebs_block_device"].Elem.(*schema.Resource)

	oldSet := make(map[string]map[string]string, 0)
	for oldKey, v := range is.Attributes {
		if strings.HasPrefix(oldKey, "ebs_block_device.") && oldKey != "ebs_block_device.#" {
			keyParts := strings.Split(oldKey, ".")
			oldHash := keyParts[1]
			attribute := keyParts[2]

			if _, ok := oldSet[oldHash]; !ok {
				oldSet[oldHash] = make(map[string]string, 0)
			}

			oldSet[oldHash][attribute] = v
		}
	}

	log.Printf("[DEBUG] Migrating, OLD SET: %#v", oldSet)

	for oldHash, attributeDiff := range oldSet {
		f := schema.HashResource(newResource)

		attributeDiff["no_device"] = ""
		attributeDiff["delete_on_termination"] = ""

		attributeMap, err := convertDiffAttributesToMap(attributeDiff, newResource.Schema)
		if err != nil {
			return nil, err
		}

		log.Printf("[DEBUG] Migrating, Computing hash from %#v", attributeMap)
		newHash := f(attributeMap)

		for k, v := range attributeDiff {
			oldKey := fmt.Sprintf("ebs_block_device.%s.%s", oldHash, k)
			newKey := fmt.Sprintf("ebs_block_device.%d.%s", newHash, k)
			value := v

			log.Printf("[DEBUG] Migration adding %s: %s", newKey, value)
			is.Attributes[newKey] = value
			delete(is.Attributes, oldKey)
		}
	}

	log.Printf("[DEBUG] Migration of LC done: %#v", is.Attributes)

	return is, nil
}

func convertDiffAttributesToMap(diff map[string]string, s map[string]*schema.Schema) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 0)

	for k, rawValue := range diff {
		switch s[k].Type {
		case schema.TypeBool:
			// Verify that we can parse this as the correct type
			var n bool
			if err := mapstructure.WeakDecode(rawValue, &n); err != nil {
				return nil, fmt.Errorf("%s: %s", k, err)
			}
			m[k] = n
		case schema.TypeInt:
			// Verify that we can parse this as an int
			var n int
			if err := mapstructure.WeakDecode(rawValue, &n); err != nil {
				return nil, fmt.Errorf("%s: %s", k, err)
			}
			m[k] = n
		case schema.TypeFloat:
			// Verify that we can parse this as an int
			var n float64
			if err := mapstructure.WeakDecode(rawValue, &n); err != nil {
				return nil, fmt.Errorf("%s: %s", k, err)
			}
			m[k] = n
		case schema.TypeString:
			// Verify that we can parse this as a string
			var n string
			if err := mapstructure.WeakDecode(rawValue, &n); err != nil {
				return nil, fmt.Errorf("%s: %s", k, err)
			}
			m[k] = n
		default:
			panic(fmt.Sprintf("Unknown type: %#v", s[k].Type))
		}
	}
	return m, nil
}
