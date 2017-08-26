package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const apiKey string = "136085"
const apiUserName string = "miniproject"
const apiPassword string = "Pr!nt123"

func main() {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/showProductList", showProductListHandler)
	http.HandleFunc("/fillInTemplatePage", fillInTemplatePageHandler)
	http.HandleFunc("/processOrder", processOrderHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return
	}
}

func processOrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	if r.Method == "POST" {
		r.ParseForm()
		fmt.Println(r.Form)
		orderObject := getOrderObject(r.Form)
		orderJson, err := json.Marshal(orderObject)
		fmt.Println("orderObject json: " + string(orderJson))
		fmt.Println("err: ", err)
	}
}

func getOrderObject(formData url.Values) CreateOrderObject {
	orderCustomer := OrderCustomer{
		FirstName:   formData["firstName"][0],
		LastName:    formData["lastName"][0],
		CompanyName: formData["companyName"][0],
		Address1:    formData["address1"][0],
		Address2:    formData["address2"][0],
		City:        formData["city"][0],
		State:       formData["state"][0],
		PostalCode:  formData["postalCode"][0],
		CountryCode: formData["countryCode"][0],
		Email:       formData["email"][0],
		Phone:       formData["phone"][0]}
	templateData := []TemplateData{}
	//Todo ^ populate template data from form data(any field taht starts with __template... should const that...
	prodID, _ := strconv.Atoi(formData["productID"][0])
	prodQuantity, _ := strconv.Atoi(formData["quantity"][0])
	item := Item{
		ItemSequenceNumber: 1,
		ProductID:          prodID,
		Quantity:           prodQuantity,
		TemplateData:       templateData}
	shipmentObject := Shipment{
		ShipmentSequenceNumber: 1,
		FirstName:              formData["firstName"][0],
		LastName:               formData["lastName"][0],
		CompanyName:            formData["companyName"][0],
		Address1:               formData["address1"][0],
		Address2:               formData["address2"][0],
		City:                   formData["city"][0],
		State:                  formData["state"][0],
		PostalCode:             formData["postalCode"][0],
		CountryCode:            formData["countryCode"][0],
		Phone:                  formData["phone"][0],
		ShippingMethod:         "FDXG"}
	orderObject := CreateOrderObject{
		PartnerOrderReference: formData["partnerOrderReference"][0],
		OrderCustomer:         orderCustomer,
		Items:                 []Item{item},
		Shipments:             []Shipment{shipmentObject}}

	fmt.Println("partnerOrderRef: " + orderObject.PartnerOrderReference)
	return orderObject
}

func fillInTemplatePageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	if r.Method == "POST" {
		r.ParseForm()
		fmt.Println(r.Form)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
		if len(r.Form) > 0 {
			productId := r.Form["productChoice"][0]
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, buildTemplate(productId))
		}
	}
}

func getStandardOrderFields() string {
	byteRead, _ := ioutil.ReadFile("orderBasics.txt")
	customerInfo := string(byteRead)

	return customerInfo
}

func buildTemplate(productId string) string {
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("Please fill out the fields below: \n\n")
	bodyBuffer.WriteString("<form action=\"processOrder\" method=\"post\">\n")
	data := getDataFromPFLAPI("products/" + productId)
	templateList := gjson.Get(data, "results.data.templateFields.fieldlist.field")

	//bodyBuffer.WriteString(data)
	bodyBuffer.WriteString(getStandardOrderFields())

	fmt.Println("templateList:" + templateList.String())
	templateList.ForEach(func(key, value gjson.Result) bool {
		fieldName := value.Get("fieldname").String()
		required := value.Get("required").String() == "Y"
		//visible := value.Get("visible").String() == "Y"
		//fieldType := value.Get("type").String()
		//defaultValue := value.Get("default").String()
		//orgValue := value.Get("orgvalue").String()

		bodyBuffer.WriteString("<br>Field: " + fieldName)
		bodyBuffer.WriteString("<input type=\"text\" name=\"templateField__" + fieldName + "\"")
		if required {
			bodyBuffer.WriteString(" required")
		}
		bodyBuffer.WriteString(">\n")

		bodyBuffer.WriteString("\n\n")
		return true
	})
	bodyBuffer.WriteString("<br><br><input type=\"submit\" value=\"Purchase\">")
	bodyBuffer.WriteString("<input type=\"hidden\" name=\"productID\" value=\"" + productId + "\" />")
	bodyBuffer.WriteString("</form>")
	return bodyBuffer.String()
}

func getDataFromPFLAPI(requestLocation string) string {
	req, _ := http.NewRequest("GET", "https://testapi.pfl.com/"+requestLocation+"?apikey="+apiKey, nil)
	req.SetBasicAuth(apiUserName, apiPassword)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return ""
}

func showProductListHandler(w http.ResponseWriter, r *http.Request) {
	jsonString := getDataFromPFLAPI("products")
	productList := gjson.Get(jsonString, "results.data")
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("<form action=\"/fillInTemplatePage\" method=\"post\" id=\"productSelection\">\n <input type=\"submit\">\n")
	bodyBuffer.WriteString("<select name=\"productChoice\">\n")

	productList.ForEach(func(key, value gjson.Result) bool {
		productName := value.Get("name")
		productId := value.Get("productID")
		bodyBuffer.WriteString("<option value=\"" + productId.String() + "\">" + productName.String() + "</option>\n")
		return true
	})

	bodyBuffer.WriteString("</select></form>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, bodyBuffer.String())
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Go web app test page 14423")
}

func testStuff() {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	req, err := http.NewRequest("GET", "https://testapi.pfl.com/products?apikey=136085", nil)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth("miniproject", "Pr!nt123")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		//fmt.Println(bodyString)
		testValue := gjson.Get(bodyString, "results.data.#.name")
		fmt.Println("am here: ", testValue)

	}
}
