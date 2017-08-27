package main

type OrderCustomer struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CompanyName string `json:"companyName"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2"`
	City        string `json:"City"`
	State       string `json:"State"`
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type Item struct {
	ItemSequenceNumber int            `json:"itemSequenceNumber"`
	ProductID          int            `json:"productID"`
	Quantity           int            `json:"quantity"`
	TemplateData       []TemplateData `json:"templateData,omitempty"`
	ItemFile           string         `json:"itemFile,omitempty"`
}

type TemplateData struct {
	TemplateDataName  string `json:"templateDataName"`
	TemplateDataValue string `json:"templateDataValue"`
}

type Shipment struct {
	ShipmentSequenceNumber int    `json:"shipmentSequenceNumber"`
	FirstName              string `json:"firstName"`
	LastName               string `json:"lastName"`
	CompanyName            string `json:"companyName"`
	Address1               string `json:"address1"`
	Address2               string `json:"address2"`
	City                   string `json:"city"`
	State                  string `json:"state"`
	PostalCode             string `json"postalCode"`
	CountryCode            string `json:"countryCode"`
	Phone                  string `json:"phone"`
	ShippingMethod         string `json:"shippingMethod"`
}
type CreateOrderObject struct {
	PartnerOrderReference string        `json:"partnerOrderReference"`
	OrderCustomer         OrderCustomer `json:"orderCustomer"`
	Items                 []Item        `json:"items"`
	Shipments             []Shipment    `json:"shipments"`
}
