package money

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type Currency string

var _ json.Unmarshaler = (*Currency)(nil)

func (c *Currency) UnmarshalJSON(data []byte) error {
	data = bytes.TrimPrefix(data, []byte(`"`))
	data = bytes.TrimSuffix(data, []byte(`"`))

	currency, ok := ValidateCurrency(string(data))
	if !ok {
		return errors.Errorf("invalid currency %q", string(data))
	}
	*c = currency
	return nil
}

type CurrencyInfo struct {
	Name string `json:"name"`
	// Prec is a maximum number of digits after the decimal point
	Prec int `json:"precision"`
}

//nolint:gochecknoglobals
var currenciesInfo = map[Currency]CurrencyInfo{
	// Fiat currencies, source: https://en.wikipedia.org/wiki/List_of_circulating_currencies
	"AED": {Name: "United Arab Emirates dirham", Prec: 2},
	"AFN": {Name: "Afghan afghani", Prec: 2},
	"ALL": {Name: "Albanian lek", Prec: 2},
	"AMD": {Name: "Armenian dram", Prec: 2},
	"ANG": {Name: "Netherlands Antillean guilder", Prec: 2},
	"AOA": {Name: "Angolan kwanza", Prec: 2},
	"ARS": {Name: "Argentine peso", Prec: 2},
	"AUD": {Name: "Australian dollar", Prec: 2},
	"AWG": {Name: "Aruban florin", Prec: 2},
	"AZN": {Name: "Azerbaijani manat", Prec: 2},
	"BAM": {Name: "Bosnia and Herzegovina convertible mark", Prec: 2},
	"BBD": {Name: "Barbadian dollar", Prec: 2},
	"BDT": {Name: "Bangladeshi taka", Prec: 2},
	"BGN": {Name: "Bulgarian lev", Prec: 2},
	"BHD": {Name: "Bahraini dinar", Prec: 3},
	"BIF": {Name: "Burundian franc", Prec: 2},
	"BMD": {Name: "Bermudian dollar", Prec: 2},
	"BND": {Name: "Brunei dollar", Prec: 2},
	"BOB": {Name: "Bolivian boliviano", Prec: 2},
	"BRL": {Name: "Brazilian real", Prec: 2},
	"BSD": {Name: "Bahamian dollar", Prec: 2},
	"BTN": {Name: "Bhutanese ngultrum", Prec: 2},
	"BWP": {Name: "Botswana pula", Prec: 2},
	"BYN": {Name: "Belarusian ruble", Prec: 2},
	"BZD": {Name: "Belize dollar", Prec: 2},
	"CAD": {Name: "Canadian dollar", Prec: 2},
	"CDF": {Name: "Congolese franc", Prec: 2},
	"CHF": {Name: "Swiss franc", Prec: 2},
	"CLP": {Name: "Chilean peso", Prec: 2},
	"CNY": {Name: "Renminbi", Prec: 1},
	"COP": {Name: "Colombian peso", Prec: 2},
	"CRC": {Name: "Costa Rican colón", Prec: 2},
	"CUP": {Name: "Cuban peso", Prec: 2},
	"CVE": {Name: "Cape Verdean escudo", Prec: 2},
	"CZK": {Name: "Czech koruna", Prec: 2},
	"DJF": {Name: "Djiboutian franc", Prec: 2},
	"DKK": {Name: "Danish krone", Prec: 2},
	"DOP": {Name: "Dominican peso", Prec: 2},
	"DZD": {Name: "Algerian dinar", Prec: 2},
	"EGP": {Name: "Egyptian pound", Prec: 2},
	"ERN": {Name: "Eritrean nakfa", Prec: 2},
	"ETB": {Name: "Ethiopian birr", Prec: 2},
	"EUR": {Name: "Euro", Prec: 2},
	"FJD": {Name: "Fijian dollar", Prec: 2},
	"FKP": {Name: "Falkland Islands pound", Prec: 2},
	"GBP": {Name: "Sterling", Prec: 2},
	"GEL": {Name: "Georgian lari", Prec: 2},
	"GHS": {Name: "Ghanaian cedi", Prec: 2},
	"GIP": {Name: "Gibraltar pound", Prec: 2},
	"GMD": {Name: "Gambian dalasi", Prec: 2},
	"GNF": {Name: "Guinean franc", Prec: 2},
	"GTQ": {Name: "Guatemalan quetzal", Prec: 2},
	"GYD": {Name: "Guyanese dollar", Prec: 2},
	"HKD": {Name: "Hong Kong dollar", Prec: 2},
	"HNL": {Name: "Honduran lempira", Prec: 2},
	"HRK": {Name: "Croatian kuna", Prec: 2},
	"HTG": {Name: "Haitian gourde", Prec: 2},
	"HUF": {Name: "Hungarian forint", Prec: 2},
	"IDR": {Name: "Indonesian rupiah", Prec: 2},
	"ILS": {Name: "Israeli new shekel", Prec: 2},
	"INR": {Name: "Indian rupee", Prec: 2},
	"IQD": {Name: "Iraqi dinar", Prec: 3},
	"IRR": {Name: "Iranian rial", Prec: 0},
	"ISK": {Name: "Icelandic króna", Prec: 2},
	"JMD": {Name: "Jamaican dollar", Prec: 2},
	"JOD": {Name: "Jordanian dinar", Prec: 2},
	"JPY": {Name: "Japanese yen", Prec: 2},
	"KES": {Name: "Kenyan shilling", Prec: 2},
	"KGS": {Name: "Kyrgyz som", Prec: 2},
	"KHR": {Name: "Cambodian riel", Prec: 2},
	"KMF": {Name: "Comorian franc", Prec: 2},
	"KPW": {Name: "North Korean won", Prec: 2},
	"KRW": {Name: "South Korean won", Prec: 2},
	"KWD": {Name: "Kuwaiti dinar", Prec: 3},
	"KYD": {Name: "Cayman Islands dollar", Prec: 2},
	"KZT": {Name: "Kazakhstani tenge", Prec: 2},
	"LAK": {Name: "Lao kip", Prec: 2},
	"LBP": {Name: "Lebanese pound", Prec: 2},
	"LKR": {Name: "Sri Lankan rupee", Prec: 2},
	"LRD": {Name: "Liberian dollar", Prec: 2},
	"LSL": {Name: "Lesotho loti", Prec: 2},
	"LYD": {Name: "Libyan dinar", Prec: 3},
	"MAD": {Name: "Moroccan dirham", Prec: 2},
	"MDL": {Name: "Moldovan leu", Prec: 2},
	"MKD": {Name: "Macedonian denar", Prec: 2},
	"MMK": {Name: "Burmese kyat", Prec: 2},
	"MNT": {Name: "Mongolian tögrög", Prec: 2},
	"MOP": {Name: "Macanese pataca", Prec: 2},
	"MUR": {Name: "Mauritian rupee", Prec: 2},
	"MVR": {Name: "Maldivian rufiyaa", Prec: 2},
	"MWK": {Name: "Malawian kwacha", Prec: 2},
	"MXN": {Name: "Mexican peso", Prec: 2},
	"MYR": {Name: "Malaysian ringgit", Prec: 2},
	"MZN": {Name: "Mozambican metical", Prec: 2},
	"NAD": {Name: "Namibian dollar", Prec: 2},
	"NGN": {Name: "Nigerian naira", Prec: 2},
	"NIO": {Name: "Nicaraguan córdoba", Prec: 2},
	"NOK": {Name: "Norwegian krone", Prec: 2},
	"NPR": {Name: "Nepalese rupee", Prec: 2},
	"NZD": {Name: "New Zealand dollar", Prec: 2},
	"OMR": {Name: "Omani rial", Prec: 3},
	"PAB": {Name: "Panamanian balboa", Prec: 2},
	"PEN": {Name: "Peruvian sol", Prec: 2},
	"PGK": {Name: "Papua New Guinean kina", Prec: 2},
	"PHP": {Name: "Philippine peso", Prec: 2},
	"PKR": {Name: "Pakistani rupee", Prec: 2},
	"PLN": {Name: "Polish złoty", Prec: 2},
	"PYG": {Name: "Paraguayan guaraní", Prec: 2},
	"QAR": {Name: "Qatari riyal", Prec: 2},
	"RON": {Name: "Romanian leu", Prec: 2},
	"RSD": {Name: "Serbian dinar", Prec: 2},
	"RUB": {Name: "Russian ruble", Prec: 2},
	"RWF": {Name: "Rwandan franc", Prec: 2},
	"SAR": {Name: "Saudi riyal", Prec: 2},
	"SBD": {Name: "Solomon Islands dollar", Prec: 2},
	"SCR": {Name: "Seychellois rupee", Prec: 2},
	"SDG": {Name: "Sudanese pound", Prec: 2},
	"SEK": {Name: "Swedish krona", Prec: 2},
	"SGD": {Name: "Singapore dollar", Prec: 2},
	"SHP": {Name: "Saint Helena pound", Prec: 2},
	"SLL": {Name: "Sierra Leonean leone", Prec: 2},
	"SOS": {Name: "Somali shilling", Prec: 2},
	"SRD": {Name: "Surinamese dollar", Prec: 2},
	"SSP": {Name: "South Sudanese pound", Prec: 2},
	"STN": {Name: "São Tomé and Príncipe dobra", Prec: 2},
	"SYP": {Name: "Syrian pound", Prec: 2},
	"SZL": {Name: "Swazi lilangeni", Prec: 2},
	"THB": {Name: "Thai baht", Prec: 2},
	"TJS": {Name: "Tajikistani somoni", Prec: 2},
	"TMT": {Name: "Turkmenistan manat", Prec: 2},
	"TND": {Name: "Tunisian dinar", Prec: 3},
	"TOP": {Name: "Tongan paʻanga", Prec: 2},
	"TRY": {Name: "Turkish lira", Prec: 2},
	"TTD": {Name: "Trinidad and Tobago dollar", Prec: 2},
	"TWD": {Name: "New Taiwan dollar", Prec: 2},
	"TZS": {Name: "Tanzanian shilling", Prec: 2},
	"UAH": {Name: "Ukrainian hryvnia", Prec: 2},
	"USD": {Name: "United States dollar", Prec: 2},
	"UYU": {Name: "Uruguayan peso", Prec: 2},
	"UZS": {Name: "Uzbekistani soʻm", Prec: 2},
	"VED": {Name: "Venezuelan bolívar digital", Prec: 2},
	"VES": {Name: "Venezuelan sovereign bolívar", Prec: 2},
	"VND": {Name: "Vietnamese đồng", Prec: 1},
	"VUV": {Name: "Vanuatu vatu", Prec: 2},
	"WST": {Name: "Samoan tālā", Prec: 2},
	"XAF": {Name: "Central African CFA franc", Prec: 2},
	"XCD": {Name: "Eastern Caribbean dollar", Prec: 2},
	"XOF": {Name: "West African CFA franc", Prec: 2},
	"XPF": {Name: "CFP franc", Prec: 2},
	"YER": {Name: "Yemeni rial", Prec: 2},
	"ZAR": {Name: "South African rand", Prec: 2},
	"ZMW": {Name: "Zambian kwacha", Prec: 2},
	// Cryptocurrencies
	"BTC":  {Name: "Bitcoin", Prec: 8},
	"USDT": {Name: "Tether", Prec: 4},
	"USDC": {Name: "USD Coin", Prec: 4},
	"BUSD": {Name: "Binance USD", Prec: 4},
	"XRP":  {Name: "XRP", Prec: 6},
	"ADA":  {Name: "Cardano", Prec: 6},
	"DOGE": {Name: "Dogecoin", Prec: 8},
	"LTC":  {Name: "Litecoin", Prec: 8},
	"XMR":  {Name: "Monero", Prec: 12},
	"TRX":  {Name: "TRON", Prec: 6},
	"BCH":  {Name: "Bitcoin Cash", Prec: 8},
	"ETH":  {Name: "Ethereum", Prec: 18},
	"ETC":  {Name: "Ethereum Classic", Prec: 18},
}

func ValidateCurrency(s string) (Currency, bool) {
	currency := Currency(strings.ToUpper(s))

	if _, ok := currenciesInfo[currency]; !ok {
		return "", false
	}
	return currency, true
}
