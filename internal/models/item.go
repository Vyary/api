package models

import "encoding/json"

type Item struct {
	ID              string          `json:"id"`
	Realm           string          `json:"realm"`
	W               int             `json:"w"`
	H               int             `json:"h"`
	Icon            string          `json:"icon"`
	Name            string          `json:"name"`
	BaseType        string          `json:"base_type"`
	Category        string          `json:"category"`
	SubCategory     string          `json:"sub_category"`
	Rarity          string          `json:"rarity"`
	Support         bool            `json:"support"`
	Desecrated      bool            `json:"desecrated"`
	Properties      json.RawMessage `json:"properties"`
	Requirements    json.RawMessage `json:"requirements"`
	EnchantMods     json.RawMessage `json:"enchantMods"`
	RuneMods        json.RawMessage `json:"runeMods"`
	ImplicitMods    json.RawMessage `json:"implicitMods"`
	ExplicitMods    json.RawMessage `json:"explicitMods"`
	FracturedMods   json.RawMessage `json:"fracturedMods"`
	DesecratedMods  json.RawMessage `json:"desecratedMods"`
	FlavourText     string          `json:"flavourText"`
	DescrText       string          `json:"descrText"`
	SecDescrText    string          `json:"secDescrText"`
	IconTierText    string          `json:"iconTierText"`
	GemSocketsCount int             `json:"gemSocketsCount"`
}
