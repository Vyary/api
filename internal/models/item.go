package models

import "encoding/json"

type Item struct {
	ID             string          `json:"id"`
	Realm          string          `json:"realm"`
	Category       string          `json:"category"`
	SubCategory    string          `json:"subCategory"`
	Icon           string          `json:"icon"`
	IconTierText   string          `json:"iconTierText"`
	Name           string          `json:"name"`
	BaseType       string          `json:"baseType"`
	Rarity         string          `json:"rarity"`
	W              int             `json:"w"`
	H              int             `json:"h"`
	Ilvl           int             `json:"ilvl"`
	SocketsCount   int             `json:"socketsCount"`
	Properties     json.RawMessage `json:"properties"`
	Requirements   json.RawMessage `json:"requirements"`
	EnchantMods    json.RawMessage `json:"enchantMods"`
	RuneMods       json.RawMessage `json:"runeMods"`
	ImplicitMods   json.RawMessage `json:"implicitMods"`
	ExplicitMods   json.RawMessage `json:"explicitMods"`
	FracturedMods  json.RawMessage `json:"fracturedMods"`
	DesecratedMods json.RawMessage `json:"desecratedMods"`
	FlavourText    string          `json:"flavourText"`
	DescrText      string          `json:"descrText"`
	SecDescrText   string          `json:"secDescrText"`
	Support        bool            `json:"support"`
	Duplicated     bool            `json:"duplicated"`
	Corrupted      bool            `json:"corrupted"`
	Sanctified     bool            `json:"sanctified"`
	Desecrated     bool            `json:"desecrated"`
	Buy            json.RawMessage `json:"buy"`
	Sell           json.RawMessage `json:"sell"`
}
