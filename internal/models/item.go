package models

type Item struct {
	ID              string `json:"market_id"`
	Realm           string
	W               int
	H               int
	Icon            string
	Name            string `json:"name"`
	BaseType        string `json:"base_type"`
	Category        string `json:"category"`
	SubCategory     string
	Rarity          string
	Support         bool
	Desecrated      bool
	Properties      string `json:"properties,omitempty"`
	Requirements    string
	EnchantMods     string
	RuneMods        string
	ImplicitMods    string
	ExplicitMods    string
	FracturedMods   string
	DesecratedMods  string
	FlavourText     string
	DescrText       string
	SecDescrText    string
	IconTierText    string
	GemSocketsCount int
}
