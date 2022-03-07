package datadog

import (
	"encoding/json"
	"errors"
	"strconv"
)

func validateIncidentParams(params map[string]string) (bool, map[string]interface{}, string, error) {
	customerImpacted, err := strconv.ParseBool(params[customerImpactedParam])
	if err != nil {
		return false, nil, "", errors.New("failed to convert 'customer impacted' to boolean")
	}
	fields := map[string]interface{}{}
	if params[fieldsParam] != "" && params[fieldsParam] != fieldsDefault {
		err = json.Unmarshal([]byte(params[fieldsParam]), &fields)
		if err != nil {
			return false, nil, "", errors.New("failed to unmarshal 'fields' param")
		}
	}
	if params[leaderIdParam] == "" {
		params[leaderIdParam] = "00000000-0000-0000-0000-000000000000"
	}

	return customerImpacted, fields, params[leaderIdParam], nil
}
