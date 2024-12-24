package main

type Config struct {
	Namespace         string `json:"namespace"`
	ServiceName       string `json:"serviceName"`
	PrimaryLabelValue string `json:"primaryLabelValue"`
	StandbyLabelValue string `json:"standbyLabelValue"`
	LabelKey          string `json:"labelKey"`
	WecomWebhook      string `json:"wecomWebhook"`
}
