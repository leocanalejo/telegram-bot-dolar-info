package dolar_argentina

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Update is the type of request that telegram sends once u send message to the bot
type Update struct {
	Message Message `json:"message"`
}

// Message is the structure of the message sent to the bot
type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

// Chat indicates the conversation to which the message belongs.
type Chat struct {
	Id int `json:"id"`
}

type Cotizacion struct {
	Nombre 		string 	`xml:"nombre"`
	Compra 		string 	`xml:"compra"`
	Venta 		string 	`xml:"venta"`
	Variacion 	string 	`xml:"variacion"`
}

type CotizacionesDolar struct {
	DolarOficial 		Cotizacion `xml:"casa349"`
	DolarBlue 			Cotizacion `xml:"casa310"`
	DolarContadoLiqui 	Cotizacion `xml:"casa312"`
	DolarBolsa 			Cotizacion `xml:"casa313"`
	DolarTurista 		Cotizacion `xml:"casa406"`
}

type DolarOficialBancos struct {
	Nacion		Cotizacion	`xml:"casa6"`
	Provincia 	Cotizacion 	`xml:"casa411"`
	Bbva 		Cotizacion 	`xml:"casa336"`
	Galicia		Cotizacion 	`xml:"casa342"`
	Santander 	Cotizacion 	`xml:"casa401"`
	Icbc 		Cotizacion 	`xml:"casa412"`
	Hipotecario Cotizacion 	`xml:"casa217"`
	Ciudad 		Cotizacion 	`xml:"casa402"`
	Patagonia 	Cotizacion 	`xml:"casa404"`
	Supervielle	Cotizacion 	`xml:"casa403"`
	Chaco 		Cotizacion	`xml:"casa334"`
	Pampa		Cotizacion	`xml:"casa335"`
	Cordoba		Cotizacion	`xml:"casa341"`
	Comafi 		Cotizacion 	`xml:"casa405"`
	Piano		Cotizacion	`xml:"casa37"`
}

type CotizacionesMonedas struct {
	Dolar		Cotizacion	`xml:"casa302"`
	Euro		Cotizacion	`xml:"casa303"`
	Real 		Cotizacion	`xml:"casa304"`
	Libra		Cotizacion	`xml:"casa305"`
	Uruguayo	Cotizacion	`xml:"casa307"`
	Chileno		Cotizacion	`xml:"casa308"`
	Guarani		Cotizacion	`xml:"casa398"`
}

type DolarSiResponse struct {
	CotizacionesDolar 	CotizacionesDolar 	`xml:"valores_principales"`
	DolarOficialBancos 	DolarOficialBancos 	`xml:"Capital_Federal"`
	CotizationesMonedas	CotizacionesMonedas	`xml:"cotizador"`
}

const telegramAPIBaseURL string = "https://api.telegram.org/bot"
const telegramAPISendMessage string = "/sendMessage"
const telegramTokenEnv string = "TELEGRAM_TOKEN"

var keywords = []string{"/help", "/info", "/dolar", "/dólar", "/Dolar", "/Dólar", "/dollar", "/Dollar", "/DOLAR", "/DÓLAR", "/DOLLAR", "/dolarbancos", "/monedas"}

func parseTelegramRequest(r *http.Request) (*Update, error) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	return &update, nil
}

// HandleTelegramWebHook sends a message back to the chat with a punchline starting by the message provided by the user.
func HandleTelegramWebHook(w http.ResponseWriter, r *http.Request) {

	// Parse incoming request
	var update, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("error parsing update, %s", err.Error())
		return
	}

	if !stringInSlice(update.Message.Text, keywords) {
		return
	}

	dolarSiResponse, err := getDolarSiResponse()
	if err != nil {
		log.Printf("got error when calling Dolar API %s", err.Error())
		return
	}

	var response string
	switch update.Message.Text {
	case "/help", "/info":
		response = getHelpResponse()
	case "/dolar", "/dólar", "/Dolar", "/Dólar", "/dollar", "/Dollar", "/DOLAR", "/DÓLAR", "/DOLLAR":
		response = extractDolarInfo(dolarSiResponse)
	case "/dolarbancos":
		response = extractDolarBancosInfo(dolarSiResponse)
	case "/monedas":
		response = extractCotizacionesMonedasInfo(dolarSiResponse)
	default:
		response = getHelpResponse()
	}

	if telegramResponseBody, err := sendTextToTelegramChat(update.Message.Chat.Id, response); err != nil {
		log.Printf("got error %s from telegram, reponse body is %s", err.Error(), telegramResponseBody)
		return
	}

	log.Printf("Response successfuly distributed to chat id %d", update.Message.Chat.Id)
}

func getHelpResponse() string {
	response := fmt.Sprintf(
		"Utiliza los siguientes comandos:\n\n" +
		"/dolar = valores del dolar\n" +
		"/dolarbancos = valor del dolar oficial en los diferentes bancos\n" +
		"/monedas = valores promedio de diferentes monesas")

	return response
}

func extractCotizacionesMonedasInfo(dolarSiResponse DolarSiResponse) string {
	response := fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xBA\xF0\x9F\x87\xB8",
		dolarSiResponse.CotizationesMonedas.Dolar.Compra,
		dolarSiResponse.CotizationesMonedas.Dolar.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xAA\xF0\x9F\x87\xBA",
		dolarSiResponse.CotizationesMonedas.Euro.Compra,
		dolarSiResponse.CotizationesMonedas.Euro.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xA7\xF0\x9F\x87\xB7",
		dolarSiResponse.CotizationesMonedas.Real.Compra,
		dolarSiResponse.CotizationesMonedas.Real.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xBA\xF0\x9F\x87\xBE",
		dolarSiResponse.CotizationesMonedas.Uruguayo.Compra,
		dolarSiResponse.CotizationesMonedas.Uruguayo.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xA7\xF0\x9F\x87\xB4",
		dolarSiResponse.CotizationesMonedas.Guarani.Compra,
		dolarSiResponse.CotizationesMonedas.Guarani.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xA8\xF0\x9F\x87\xB1",
		dolarSiResponse.CotizationesMonedas.Chileno.Compra,
		dolarSiResponse.CotizationesMonedas.Chileno.Venta)
	response += fmt.Sprintf("%s %s / %s\n",
		"\xF0\x9F\x87\xAC\xF0\x9F\x87\xA7",
		dolarSiResponse.CotizationesMonedas.Libra.Compra,
		dolarSiResponse.CotizationesMonedas.Libra.Venta)

	return response
}

func extractDolarBancosInfo(dolarSiResponse DolarSiResponse) string {
	response := fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Nacion.Nombre),
		dolarSiResponse.DolarOficialBancos.Nacion.Compra,
		dolarSiResponse.DolarOficialBancos.Nacion.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Provincia.Nombre),
		dolarSiResponse.DolarOficialBancos.Provincia.Compra,
		dolarSiResponse.DolarOficialBancos.Provincia.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Bbva.Nombre),
		dolarSiResponse.DolarOficialBancos.Bbva.Compra,
		dolarSiResponse.DolarOficialBancos.Bbva.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Galicia.Nombre),
		dolarSiResponse.DolarOficialBancos.Galicia.Compra,
		dolarSiResponse.DolarOficialBancos.Galicia.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Santander.Nombre),
		dolarSiResponse.DolarOficialBancos.Santander.Compra,
		dolarSiResponse.DolarOficialBancos.Santander.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Icbc.Nombre),
		dolarSiResponse.DolarOficialBancos.Icbc.Compra,
		dolarSiResponse.DolarOficialBancos.Icbc.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Hipotecario.Nombre),
		dolarSiResponse.DolarOficialBancos.Hipotecario.Compra,
		dolarSiResponse.DolarOficialBancos.Hipotecario.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Ciudad.Nombre),
		dolarSiResponse.DolarOficialBancos.Ciudad.Compra,
		dolarSiResponse.DolarOficialBancos.Ciudad.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Patagonia.Nombre),
		dolarSiResponse.DolarOficialBancos.Patagonia.Compra,
		dolarSiResponse.DolarOficialBancos.Patagonia.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Supervielle.Nombre),
		dolarSiResponse.DolarOficialBancos.Supervielle.Compra,
		dolarSiResponse.DolarOficialBancos.Supervielle.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Chaco.Nombre),
		dolarSiResponse.DolarOficialBancos.Chaco.Compra,
		dolarSiResponse.DolarOficialBancos.Chaco.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Pampa.Nombre),
		dolarSiResponse.DolarOficialBancos.Pampa.Compra,
		dolarSiResponse.DolarOficialBancos.Pampa.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Cordoba.Nombre),
		dolarSiResponse.DolarOficialBancos.Cordoba.Compra,
		dolarSiResponse.DolarOficialBancos.Cordoba.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s\n\n",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Comafi.Nombre),
		dolarSiResponse.DolarOficialBancos.Comafi.Compra,
		dolarSiResponse.DolarOficialBancos.Comafi.Venta)
	response += fmt.Sprintf("%s\nCompra: %s\nVenta: %s",
		strings.ToUpper(dolarSiResponse.DolarOficialBancos.Piano.Nombre),
		dolarSiResponse.DolarOficialBancos.Piano.Compra,
		dolarSiResponse.DolarOficialBancos.Piano.Venta)

	return response
}

func extractDolarInfo(dolarSiResponse DolarSiResponse) string {
	response := fmt.Sprintf("Oficial   %s / %s %s\n",
		dolarSiResponse.CotizacionesDolar.DolarOficial.Compra,
		dolarSiResponse.CotizacionesDolar.DolarOficial.Venta,
		getVariacion(dolarSiResponse.CotizacionesDolar.DolarOficial.Variacion))
	response += fmt.Sprintf("Blue      %s / %s %s\n",
		dolarSiResponse.CotizacionesDolar.DolarBlue.Compra,
		dolarSiResponse.CotizacionesDolar.DolarBlue.Venta,
		getVariacion(dolarSiResponse.CotizacionesDolar.DolarBlue.Variacion))
	response += fmt.Sprintf("Turista  %s %s\n",
		dolarSiResponse.CotizacionesDolar.DolarTurista.Venta,
		getVariacion(dolarSiResponse.CotizacionesDolar.DolarTurista.Variacion))
	response += fmt.Sprintf("CCL      %s / %s %s\n",
		dolarSiResponse.CotizacionesDolar.DolarContadoLiqui.Compra,
		dolarSiResponse.CotizacionesDolar.DolarContadoLiqui.Venta,
		getVariacion(dolarSiResponse.CotizacionesDolar.DolarContadoLiqui.Variacion))
	response += fmt.Sprintf("Bolsa    %s / %s %s",
		dolarSiResponse.CotizacionesDolar.DolarBolsa.Compra,
		dolarSiResponse.CotizacionesDolar.DolarBolsa.Venta,
		getVariacion(dolarSiResponse.CotizacionesDolar.DolarBolsa.Variacion))

	return response
}

func getVariacion(variacion string) string {
	if variacion == "0" {
		return "\xF0\x9F\x91\x8C " + variacion + "%"
	} else if strings.HasPrefix(variacion, "-") {
		return "\xF0\x9F\x91\x87 " + variacion + "%"
	} else if strings.Contains(variacion, ",") {
		return "\xF0\x9F\x91\x86 +" + variacion + "%"
	} else {
		return "\xF0\x9F\x91\x8C 0%"
	}
}

func getDolarSiResponse() (DolarSiResponse, error) {
	response := &DolarSiResponse{}
	httpResponse, err := http.Get("https://www.dolarsi.com/api/dolarSiInfo.xml")
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return *response, err
	}
	xmlResponse, _ := ioutil.ReadAll(httpResponse.Body)
	_ = xml.Unmarshal(xmlResponse, response)

	return *response, nil
}

// sendTextToTelegramChat sends a text message to the Telegram chat identified by its chat Id
func sendTextToTelegramChat(chatId int, text string) (string, error) {

	log.Printf("Sending %s to chat_id: %d", text, chatId)
	var telegramApi = telegramAPIBaseURL + os.Getenv(telegramTokenEnv) + telegramAPISendMessage
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})

	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)

	return bodyString, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}