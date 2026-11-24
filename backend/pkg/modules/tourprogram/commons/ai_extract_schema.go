// Package commons holds tour program helpers shared by its use cases: the AI extraction schema, prompt
// and post-processing.
package commons

import "encoding/json"

// AIExtractSchema is the structured-output JSON schema the model must answer with (Claude structured
// outputs: the reply is constrained to it). Every property is required. "Not found" is "" for text,
// 0 for numbers and -1 for nights — not null: structured outputs allow at most 16 union-typed
// (nullable) parameters, so the schema has none. Ranges, lengths and formats are checked in Go
// (NormalizeAIExtraction), which already treats those values as missing.
//
// Keep it in step with AIExtraction below.
var AIExtractSchema = json.RawMessage(`{
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "days", "nights", "highlight", "airline", "wholesaleTourCode", "printedTourCode", "countries", "cities", "periods", "extras", "governmentShopStops", "includedItems", "excludedItems", "itinerary", "confidence"],
  "properties": {
    "name": {"type": "string", "description": "Tour program title as printed, Thai text verbatim"},
    "days": {"type": "integer", "description": "0 when not stated"},
    "nights": {"type": "integer", "description": "-1 when not stated"},
    "highlight": {"type": "string", "description": "One short Thai summary of the tour's selling points, built only from text in the document"},
    "airline": {
      "type": "object", "additionalProperties": false, "required": ["iataCode", "name"],
      "properties": {
        "iataCode": {"type": "string", "description": "2-character IATA airline code, e.g. VZ, SL, EK"},
        "name": {"type": "string"}
      }
    },
    "wholesaleTourCode": {"type": "string", "description": "The tour operator's own tour code, e.g. PCN-PVG580-VZ"},
    "printedTourCode": {"type": "string", "description": "Our own tour code printed as ทัวร์โค้ด, e.g. NIDN263430"},
    "countries": {
      "type": "array",
      "items": {
        "type": "object", "additionalProperties": false, "required": ["nameTh", "nameEn"],
        "properties": {"nameTh": {"type": "string"}, "nameEn": {"type": "string"}}
      }
    },
    "cities": {
      "type": "array",
      "items": {
        "type": "object", "additionalProperties": false, "required": ["nameTh", "nameEn", "countryNameEn"],
        "properties": {
          "nameTh": {"type": "string"},
          "nameEn": {"type": "string"},
          "countryNameEn": {"type": "string"}
        }
      }
    },
    "periods": {
      "type": "array",
      "items": {
        "type": "object", "additionalProperties": false,
        "required": ["start", "end", "adultPrice", "childWithBedPrice", "childNoBedPrice", "infantPrice", "singleRoom", "singleRoomKind", "label", "sourceText"],
        "properties": {
          "start": {"$ref": "#/$defs/date"},
          "end": {"$ref": "#/$defs/date"},
          "adultPrice": {"type": "string", "description": "Adult price exactly as printed, e.g. \"23,999\" or \"57,900.-\""},
          "childWithBedPrice": {"type": "string", "description": "Child sharing with adults, with a bed, as printed"},
          "childNoBedPrice": {"type": "string", "description": "Child sharing with adults, without a bed, as printed"},
          "infantPrice": {"type": "string", "description": "Infant (under 2, no seat), as printed"},
          "singleRoom": {"type": "string"},
          "singleRoomKind": {"type": "string", "enum": ["supplement", "total", "unknown"]},
          "label": {"type": "string", "description": "Note printed with this departure, e.g. (Promotion), (วันพ่อ)"},
          "sourceText": {"type": "string", "description": "The row as printed, for review"}
        }
      }
    },
    "extras": {
      "type": "object", "additionalProperties": false,
      "required": ["depositPerPerson", "joinLandPrice", "minGroupSize", "otherCharges"],
      "properties": {
        "depositPerPerson": {"type": "string"},
        "joinLandPrice": {"type": "string", "description": "Price without air tickets (Join Land), or the amount deducted for it"},
        "minGroupSize": {"type": "integer", "description": "0 when not stated"},
        "otherCharges": {"type": "array", "items": {"type": "string"}, "description": "Other per-person charges or surcharges, one Thai line each, verbatim"}
      }
    },
    "governmentShopStops": {"type": "string", "enum": ["yes", "no", "unknown"], "description": "Does the tour stop at government shops (ลงร้านรัฐบาล)?"},
    "includedItems": {"type": "array", "items": {"type": "string"}, "description": "What the tour price includes, one short Thai line each"},
    "excludedItems": {"type": "array", "items": {"type": "string"}, "description": "What the tour price does not include, one short Thai line each"},
    "itinerary": {
      "type": "array",
      "description": "One entry per day of the day-by-day plan, in order",
      "items": {
        "type": "object", "additionalProperties": false, "required": ["day", "places"],
        "properties": {
          "day": {"type": "integer"},
          "places": {"type": "array", "items": {"type": "string"}, "description": "Short place names in visiting order, no hotels"}
        }
      }
    },
    "confidence": {
      "type": "object", "additionalProperties": false,
      "required": ["main", "taxonomy", "periods", "prices", "extras", "inclusions", "itinerary"],
      "properties": {
        "main": {"$ref": "#/$defs/level"},
        "taxonomy": {"$ref": "#/$defs/level"},
        "periods": {"$ref": "#/$defs/level"},
        "prices": {"$ref": "#/$defs/level"},
        "extras": {"$ref": "#/$defs/level"},
        "inclusions": {"$ref": "#/$defs/level"},
        "itinerary": {"$ref": "#/$defs/level"}
      }
    }
  },
  "$defs": {
    "date": {
      "type": "object", "additionalProperties": false, "required": ["day", "month", "year"],
      "properties": {
        "day": {"type": "integer"},
        "month": {"type": "integer", "description": "1-12"},
        "year": {"type": "integer", "description": "Exactly as printed for this date or its section heading (2569, 69 or 2026); 0 when no year applies"}
      }
    },
    "level": {"type": "string", "enum": ["high", "medium", "low"]}
  }
}`)

// AIExtraction is the model's answer (AIExtractSchema). Pointers are the schema's nullable values.
type AIExtraction struct {
	Name            *string      `json:"name"`
	Days            *int         `json:"days"`
	Nights          *int         `json:"nights"`
	Highlight       *string      `json:"highlight"`
	Airline         AIAirline    `json:"airline"`
	WholesaleTour   *string      `json:"wholesaleTourCode"`
	PrintedTourCode *string      `json:"printedTourCode"`
	Countries       []AIPlace    `json:"countries"`
	Cities          []AICity     `json:"cities"`
	Periods         []AIPeriod   `json:"periods"`
	Extras          AIExtras     `json:"extras"`
	GovernmentShops string       `json:"governmentShopStops"` // yes | no | unknown
	IncludedItems   []string     `json:"includedItems"`
	ExcludedItems   []string     `json:"excludedItems"`
	Itinerary       []AIDay      `json:"itinerary"`
	Confidence      AIConfidence `json:"confidence"`
}

type AIAirline struct {
	IATACode *string `json:"iataCode"`
	Name     *string `json:"name"`
}

type AIPlace struct {
	NameTh *string `json:"nameTh"`
	NameEn *string `json:"nameEn"`
}

type AICity struct {
	NameTh        *string `json:"nameTh"`
	NameEn        *string `json:"nameEn"`
	CountryNameEn *string `json:"countryNameEn"`
}

type AIDate struct {
	Day   int  `json:"day"`
	Month int  `json:"month"`
	Year  *int `json:"year"`
}

type AIPeriod struct {
	Start          AIDate  `json:"start"`
	End            AIDate  `json:"end"`
	AdultPrice     *string `json:"adultPrice"`
	ChildWithBed   *string `json:"childWithBedPrice"`
	ChildNoBed     *string `json:"childNoBedPrice"`
	Infant         *string `json:"infantPrice"`
	SingleRoom     *string `json:"singleRoom"`
	SingleRoomKind string  `json:"singleRoomKind"` // supplement | total | unknown
	Label          *string `json:"label"`
	SourceText     string  `json:"sourceText"`
}

type AIExtras struct {
	DepositPerPerson *string  `json:"depositPerPerson"`
	JoinLandPrice    *string  `json:"joinLandPrice"`
	MinGroupSize     *int     `json:"minGroupSize"`
	OtherCharges     []string `json:"otherCharges"`
}

// AIConfidence is the model's own confidence per section: high | medium | low.
type AIConfidence struct {
	Main       string `json:"main"`
	Taxonomy   string `json:"taxonomy"`
	Periods    string `json:"periods"`
	Prices     string `json:"prices"`
	Extras     string `json:"extras"`
	Inclusions string `json:"inclusions"` // government shops and the included / excluded lists
	Itinerary  string `json:"itinerary"`
}

// AIDay is one day of the short timeline (also the normalized and draft shape).
type AIDay struct {
	Day    int      `json:"day"`
	Places []string `json:"places"`
}
