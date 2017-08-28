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
	"strings"
)

const portValue string = ":23888"
const apiKey string = "136085"
const apiUserName string = "miniproject"
const apiPassword string = "Pr!nt123"
const pflAPIBaseLink string = "https://testapi.pfl.com/"
const templatePrefix string = "template."

func main() {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/showProductList", showProductListHandler)
	http.HandleFunc("/fillInTemplatePage", fillInTemplatePageHandler)
	http.HandleFunc("/processOrder", processOrderHandler)
	http.HandleFunc("/testStuff", testStuffHandler)
	err := http.ListenAndServe(portValue, nil)
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
		orderJsonBytes, err := json.Marshal(orderObject)
		orderJson := string(orderJsonBytes)
		fmt.Println("orderObject json: " + orderJson)
		fmt.Println("err: ", err)
		response := postDataToPFLAPI("orders", orderJson)
		fmt.Println("response: " + response)
		if len(response) > 0 {
			fmt.Fprintf(w, "Order Number:\n")
			//TODO handle error result(if one exists...)
			fmt.Fprintf(w, gjson.Get(response, "results.data.orderNumber").String())
		} else {
			prettyOrderJson, _ := json.MarshalIndent(orderObject, "", "	")
			fmt.Fprintf(w, "Order failed to get a response. Json sent:\n"+string(prettyOrderJson))
		}
	}
}

func postDataToPFLAPI(apiPath string, jsonData string) string {
	return pflAPIRequest("POST", apiPath, strings.NewReader(jsonData))
}

// Returns an object ready for json marshalling for the pfl create order api
func getOrderObject(formData url.Values) CreateOrderObject {
	//TODO handle missing fields
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
	templatePrefixLength := len(templatePrefix)
	for key, value := range formData {
		if strings.HasPrefix(key, templatePrefix) {
			templateData = append(templateData, TemplateData{key[templatePrefixLength:], value[0]})
		}
	}
	prodID, _ := strconv.Atoi(formData["productID"][0])
	prodQuantity, _ := strconv.Atoi(formData["quantity"][0])
	itemFile := ""
	if value, found := formData["itemFile"]; found {
		itemFile = value[0]
	}
	item := Item{
		ItemSequenceNumber: 1,
		ProductID:          prodID,
		Quantity:           prodQuantity,
		TemplateData:       templateData,
		ItemFile:           itemFile}
	// Note: shipmentObject mostly copies customer info at this time, to avoid excessive fields on page
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
		ShippingMethod:         formData["shippingMethod"][0]}
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
		} //TODO handle missing form data
	}
}

func getStandardOrderFields() string {
	basicCustomerInfoBytes, _ := ioutil.ReadFile("orderBasics.txt")
	//TODO catch and handle error
	return string(basicCustomerInfoBytes)
}

func buildTemplate(productId string) string {
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("Please fill out the fields below: \n\n")
	bodyBuffer.WriteString("<form action=\"processOrder\" method=\"post\">\n")
	data := getDataFromPFLAPI("products/" + productId)

	bodyBuffer.WriteString(getStandardOrderFields())
	bodyBuffer.WriteString(getShippingChoiceField(data))
	bodyBuffer.WriteString(getQuantityChoiceField(data))
	bodyBuffer.WriteString(getTemplateFieldsForPage(data))

	bodyBuffer.WriteString("<br><br><input type=\"submit\" value=\"Purchase\">")
	bodyBuffer.WriteString("<input type=\"hidden\" name=\"productID\" value=\"" + productId + "\" />")
	bodyBuffer.WriteString("</form>")
	return bodyBuffer.String()
}

func getQuantityChoiceField(jsonData string) string {
	var bodyBuffer bytes.Buffer
	defaultQuantity := gjson.Get(jsonData, "results.data.quantityDefault")
	minimumQuantity := gjson.Get(jsonData, "results.data.quantityMinimum")
	maximumQuantity := gjson.Get(jsonData, "results.data.quantityMaximum")
	quantityIncrement := gjson.Get(jsonData, "results.data.quantityIncrement")

	bodyBuffer.WriteString("<br>Quantity: <input type=\"number\" name=\"quantity\"")
	if len(minimumQuantity.String()) > 0 {
		bodyBuffer.WriteString(" min=\"" + minimumQuantity.String() + "\"")
	}
	if len(maximumQuantity.String()) > 0 {
		bodyBuffer.WriteString(" max=\"" + maximumQuantity.String() + "\"")
	}
	if len(defaultQuantity.String()) > 0 {
		bodyBuffer.WriteString(" value=\"" + defaultQuantity.String() + "\"")
	}
	if len(quantityIncrement.String()) > 0 {
		bodyBuffer.WriteString(" step=\"" + quantityIncrement.String() + "\"")
	}
	bodyBuffer.WriteString(">")
	return bodyBuffer.String()
}

func getShippingChoiceField(jsonData string) string {
	var bodyBuffer bytes.Buffer
	shippingOptionList := gjson.Get(jsonData, "results.data.deliveredPrices")

	bodyBuffer.WriteString("<br>Shipping Method:<select name=\"shippingMethod\">\n")

	shippingOptionList.ForEach(func(key, value gjson.Result) bool {
		deliveryMethodCode := value.Get("deliveryMethodCode").String()
		deliveryMethodDescription := value.Get("description").String()
		deliveryMethodPrice := value.Get("price").Float()
		bodyBuffer.WriteString("<option value=\"" + deliveryMethodCode + "\">" + deliveryMethodDescription + "--$" + strconv.FormatFloat(deliveryMethodPrice, 'f', 2, 64) + "</option>\n")
		return true
	})
	bodyBuffer.WriteString("\n</select>\n")
	return bodyBuffer.String()
}

func getTemplateFieldsForPage(jsonData string) string {
	var bodyBuffer bytes.Buffer
	templateList := gjson.Get(jsonData, "results.data.templateFields.fieldlist.field")

	if len(templateList.String()) <= 0 {
		imageURL := gjson.Get(jsonData, "results.data.imageURL").String()
		if len(imageURL) <= 0 {
			imageURL = "http://www.yourdomain.com/files/printReadyArtwork1.pdf"
		}
		bodyBuffer.WriteString("<br>ItemField: <input type=\"text\" name=\"itemFile\" value=\"" + imageURL + "\" />")
	}
	fmt.Println("\n\ntemplateList:" + templateList.String() + "\n\n\n")
	templateList.ForEach(func(key, value gjson.Result) bool {
		//TODO utilize additional information
		fieldName := value.Get("fieldname").String()
		required := value.Get("required").String() == "Y"
		//visible := value.Get("visible").String() == "Y"
		//fieldType := value.Get("type").String()
		defaultValue := value.Get("default").String()
		//orgValue := value.Get("orgvalue").String()
		//htmlFieldName := value.Get("htmlfieldname").String()
		bodyBuffer.WriteString("<br>Field: " + fieldName)
		bodyBuffer.WriteString("<input type=\"text\" name=\"" + templatePrefix + fieldName + "\" value=\"" + defaultValue + "\"")
		if required {
			bodyBuffer.WriteString(" required")
		}
		bodyBuffer.WriteString(">\n")
		return true
	})
	return bodyBuffer.String()
}

func getDataFromPFLAPI(requestLocation string) string {
	return pflAPIRequest("GET", requestLocation, strings.NewReader(""))
}

func pflAPIRequest(requestType string, requestLocation string, jsonData *strings.Reader) string {
	//TODO handle errors
	pflRequest, _ := http.NewRequest(requestType, pflAPIBaseLink+requestLocation+"?apikey="+apiKey, jsonData)
	pflRequest.SetBasicAuth(apiUserName, apiPassword)
	pflRequest.Header.Set("Content-Type", "application/json")

	pflResponse, _ := http.DefaultClient.Do(pflRequest)
	defer pflResponse.Body.Close()
	if pflResponse.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(pflResponse.Body)
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
	fmt.Fprintf(w, "Invalid page request")
}

func testStuffHandler(w http.ResponseWriter, r *http.Request) {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	req, err := http.NewRequest("GET", "https://testapi.pfl.com/products/22784?apikey=136085", nil)
	if err != nil {
		// TODO handle err
	}
	req.SetBasicAuth("miniproject", "Pr!nt123")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// TODO handle err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		fmt.Println("am here: ", bodyString)
	}
}
