package datadog

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
)

const (
	pluginName            = "datadog"
	incidentIdParam       = "Incident ID"
	customerImpactedParam = "Customer Impacted"
	titleParam            = "Title"
	fieldsParam           = "Fields"
	leaderIdParam         = "Leader User ID"
	impactStartParam      = "Impact Start"
	impactEndParam        = "Impact End"
	impactScopeParam      = "Impact Scope"
	detectedParam         = "Detected"
	resolvedParam         = "Resolved"
	messageParam          = "Message"
	linkParam             = "Link"
	fieldsDefault         = `{\n
\t\"state\":\n
\t{\n
\t\t\"type\": \"dropdown\",\n
\t\t\"value\": \"resolved\"\n
\t},\n
\t\"summary\":\n
\t{\n
\t\t\"type\": \"textbox\",\n
\t\t\"value\": \"TODO: put the summary of the incident here\"\n
\t}\n
}`
)

func createIncident(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	customerImpacted, fields, leaderId, err := validateIncidentParams(params)
	if err != nil {
		return nil, err
	}

	reqBody := IncidentRequestBody{
		Data: IncidentData{
			Type: "incidents",
			Attributes: IncidentAttributes{
				CustomerImpacted: customerImpacted,
				Fields:           fields,
				Title:            params[titleParam],
			},
			Relationships: IncidentRelationships{
				CommanderUser: CommanderUser{
					Data: CommanderUserData{
						ID:   leaderId,
						Type: "users",
					},
				},
			},
		},
	}

	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, requestUrl+"/api/v2/incidents", consts.DefaultTimeout, nil, nil, marshalledReqBody)
	return res.Body, err
}

func updateIncident(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	customerImpacted, fields, leaderId, err := validateIncidentParams(params)
	if err != nil {
		return nil, err
	}

	reqBody := IncidentRequestBody{
		Data: IncidentData{
			Id:   params[incidentIdParam],
			Type: "incidents",
			Attributes: IncidentAttributes{
				CustomerImpactEnd:   params[impactEndParam],
				CustomerImpactScope: params[impactScopeParam],
				CustomerImpactStart: params[impactStartParam],
				CustomerImpacted:    customerImpacted,
				Detected:            params[detectedParam],
				Fields:              fields,
				Resolved:            params[resolvedParam],
				Title:               params[titleParam],
			},
			Relationships: IncidentRelationships{
				CommanderUser: CommanderUser{
					Data: CommanderUserData{
						ID:   leaderId,
						Type: "users",
					},
				},
			},
		},
	}

	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal request body")
	}

	url := fmt.Sprintf("%s/api/v2/incidents/%s", requestUrl, params[incidentIdParam])
	res, err := requests.SendRequest(ctx, plugin, http.MethodPatch, url, consts.DefaultTimeout, nil, nil, marshalledReqBody)
	return res.Body, err
}

func updateIncidentTimeline(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	incidentID := params[incidentIdParam]
	message := params[messageParam]

	reqBody := json.RawMessage(`{"data":{"type":"incident_timeline_cells","attributes":{"cell_type":"markdown","content":{"content":"` + message + `"},"source":"unknown"} } }`)

	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return []byte("failed to marshal json"), nil
	}

	url := fmt.Sprintf("%s/api/v2/incidents/%s/timeline", requestUrl, incidentID)
	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, url, consts.DefaultTimeout, nil, nil, marshalledReqBody)
	return res.Body, err
}

func addLinkToIncident(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	incidentID := params[incidentIdParam]
	title := params[titleParam]
	link := params[linkParam]

	reqBody := json.RawMessage(`{"data":[{"type":"incident_attachments","attributes":{"attachment_type":"link","attachment":{"documentUrl":"` + link + `","title":"` + title + `"} } }]}`)

	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return []byte("failed to marshal json"), nil
	}

	url := fmt.Sprintf("%s/api/v2/incidents/%s/attachments", requestUrl, incidentID)
	res, err := requests.SendRequest(ctx, plugin, http.MethodPatch, url, consts.DefaultTimeout, nil, nil, marshalledReqBody)
	return res.Body, err
}
