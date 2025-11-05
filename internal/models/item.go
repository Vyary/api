package models

import (
	"encoding/json"
)

type ItemID string
type League string
type Currency string
type Prices map[Currency]PriceDTO

type Price struct {
	ItemID     ItemID
	Price      float64
	CurrencyID string
	Volume     float64
	Stock      float64
	League     League
	Timestamp  int64
}

type PriceDTO struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	Stock  float64 `json:"stock"`
}

type Item struct {
	ID             ItemID            `json:"id"`
	Realm          string            `json:"realm,omitempty"`
	Category       string            `json:"category,omitempty"`
	SubCategory    string            `json:"subCategory,omitempty"`
	Icon           string            `json:"icon,omitempty"`
	IconTierText   string            `json:"iconTierText,omitempty"`
	Name           string            `json:"name,omitempty"`
	BaseType       string            `json:"baseType,omitempty"`
	Rarity         string            `json:"rarity,omitempty"`
	W              int               `json:"w,omitempty"`
	H              int               `json:"h,omitempty"`
	Ilvl           int               `json:"ilvl,omitempty"`
	SocketedItems  *json.RawMessage  `json:"socketedItems,omitempty"`
	Properties     *json.RawMessage  `json:"properties,omitempty"`
	Requirements   *json.RawMessage  `json:"requirements,omitempty"`
	EnchantMods    *json.RawMessage  `json:"enchantMods,omitempty"`
	RuneMods       *json.RawMessage  `json:"runeMods,omitempty"`
	ImplicitMods   *json.RawMessage  `json:"implicitMods,omitempty"`
	ExplicitMods   *json.RawMessage  `json:"explicitMods,omitempty"`
	FracturedMods  *json.RawMessage  `json:"fracturedMods,omitempty"`
	DesecratedMods *json.RawMessage  `json:"desecratedMods,omitempty"`
	FlavourText    string            `json:"flavourText,omitempty"`
	DescrText      string            `json:"descrText,omitempty"`
	SecDescrText   string            `json:"secDescrText,omitempty"`
	Support        bool              `json:"support,omitempty"`
	Duplicated     bool              `json:"duplicated,omitempty"`
	Corrupted      bool              `json:"corrupted,omitempty"`
	Sanctified     bool              `json:"sanctified,omitempty"`
	Desecrated     bool              `json:"desecrated,omitempty"`
	Prices         map[League]Prices `json:"prices"`
}
