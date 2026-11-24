package commons

// AIExtractPromptVersion is logged with every extraction so eval results can be tied to a prompt.
const AIExtractPromptVersion = "2026-11-25.6"

// AIExtractSystemPrompt is the stable system prompt (prompt-cached). Rules first, then worked
// examples. The examples come from nidn263430 and nidn263439 only: nidn265004 is the held-out file
// of the eval (backend/ex), so nothing from it may appear here.
const AIExtractSystemPrompt = `You read Thai tour program brochures (PDF) for nidnoi travel, a Thai travel agency, and extract the data staff need to create the tour program in the back office. Staff review everything you return before saving, so an honest empty value is far better than a guess: a wrong price or date that looks right is the worst possible outcome.

The document is a brochure: a cover poster (often only an image), a summary table (day / sights / meals / hotel), a price table with departure dates, a day-by-day itinerary, and image pages with conditions, inclusions, insurance and contact details. Read the images as carefully as the text; some facts appear only in an image (cover, conditions pages).

General rules
- Use only what the document says. If a value is not in the document, return an empty string for text, 0 for numbers (but -1 for nights) and an empty array for lists. Never invent, estimate or compute values the document does not state.
- Copy Thai text exactly as written (same spelling, same words). Collapse line breaks inside a value into single spaces. Do not translate Thai into English or English into Thai, except the English name fields (nameEn) described below.
- Ignore the travel agency's own branding, footer, contact numbers, bank accounts, insurance price tables and visa forms: they are not part of the tour.

Fields
- name: the tour program title as printed near the top of the summary page (for example "เซี่ยงไฮ้ ดิสนีย์แลนด์ (ทัวร์ไม่ลงร้านค้ารัฐบาล)"). Do not include "โปรแกรมการเดินทาง", the number of days, the airline or the travel period. If the summary page has no title line, use the destination title printed on the cover; if there is none, "".
- days, nights: from "X วัน Y คืน". Integers; days 0 and nights -1 when not stated.
- highlight: one Thai sentence (at most about 300 characters) listing the main selling points named in the document, such as key sights, special meals, hotel level and "no shopping stops". Use the document's wording; add nothing that is not in it.
- airline: the operating airline. iataCode is the 2-character code (VZ, SL, EK, TG, ...) taken from the airline name in parentheses or from the flight numbers (VZ3524 → VZ). name as printed.
- wholesaleTourCode: the tour operator's own code for this tour, usually printed on the cover near the airline logo (for example "PCN-PVG580-VZ", "PJPNRT22-SL"). Not the "ทัวร์โค้ด NIDN……" code. "" when not shown.
- printedTourCode: the agency code printed as "ทัวร์โค้ด NIDN……" at the top. Copy it; it is only used to detect duplicates.
- countries: the destination countries the tour visits and stays in, in the order the tour visits them. Each with nameTh (Thai as printed or the usual Thai name) and nameEn (English country name, e.g. "China", "Japan", "Switzerland"). A country passed through only for a flight connection (a transit airport) is not a destination. Thailand is never a destination.
- cities: destination cities or areas the itinerary visits (for example Shanghai, Tokyo, Narita, Yamanashi). nameTh as printed, nameEn the usual English name, countryNameEn the English name of its country. Skip transit airports, single attractions (a temple, a mall, a theme park) and Thai departure cities.
- periods: one entry per departure row of the price table. If the same table is printed twice, return each departure once.
  - start, end: day, month (1-12) and the year exactly as printed for that row (2569, 69 or 2026). When a row prints no year, use the year of its month heading (for example "เดือน ตุลาคม 2569") or of the table's travel-period note; when the year is printed only on the end date (for example "28 ธ.ค. – 01 ม.ค. 2570"), give the start date year 0. A year that applies to neither date is 0. Do not convert Buddhist-era years; the system converts them.
  - adultPrice: the price per adult in a shared room (ผู้ใหญ่, พักห้องละ 2-3 ท่าน), exactly as printed ("23,999 บาท" → "23,999"). The "ราคาเริ่มต้น" (starting from) price on the cover is NOT a period price: never use it for a row.
  - childWithBedPrice, childNoBedPrice: child prices as printed. A column that says the child has a bed (มีเตียง / เสริมเตียง) goes to childWithBedPrice, one that says no bed (ไม่มีเตียง / ไม่เสริมเตียง) to childNoBedPrice. When the brochure has a single child price with no bed wording (for example "ราคาเด็ก (2-18 ปี)"), put that same price in both. "" when there is no child price.
  - infantPrice: the infant price (ทารก, Infant, a child under 2 with no seat) as printed. An infant price is never a child price. It is often printed once under the table for every departure (for example "ราคาเด็กอายุไม่ถึง 2 ปี (Infant) ท่านละ 6,000 บาท"): then repeat it on every period. "" when not stated.
  - singleRoom and singleRoomKind: the single-room price as printed. "supplement" when it is an extra charge on top of the adult price (พักเดี่ยวเพิ่ม, จ่ายเพิ่ม, or clearly much smaller than the adult price), "total" when it is the full price for one person in a single room, "unknown" when you cannot tell.
  - label: a note printed with that departure, such as "(Promotion)" or a holiday name "(วันพ่อ)"; otherwise "".
  - sourceText: the row as printed (dates and prices), for staff to compare.
- extras:
  - depositPerPerson: the deposit per person (มัดจำ ท่านละ ...), often on a conditions image page. As printed.
  - joinLandPrice: the price without air tickets (Join Land) or the amount deducted for it, as printed, e.g. "หักออกจากราคาทัวร์เต็มท่านละ 5,000 บาท" → "5,000".
  - minGroupSize: the minimum number of travellers for the group to depart, when stated (e.g. "10 ท่านขึ้นไป").
  - otherCharges: other per-person surcharges and fees that change what a traveller pays, one Thai line each, copied verbatim: tips that are not included, foreign-passport fees, business-class upgrades. Do not repeat the deposit, infant or Join Land lines.
- governmentShopStops: "no" when the brochure says the tour does not stop at government shops (for example "ทัวร์ไม่ลงร้านค้ารัฐบาล", "ไม่ลงร้านช้อป", a "ไม่ลง ร้านช้อป" badge on the cover). "yes" when the itinerary includes government shop stops (for example ร้านหยก, ร้านยางพารา, ร้านบัวหิมะ, ร้านไข่มุก, ร้านจิวเวลรี่, "ร้านรัฐบาล", "ลงร้าน" as part of the day plan). A general condition such as "หากเป็นโปรแกรมลงร้าน …" or a duty-free or outlet shopping stop is not enough on its own. "unknown" when the brochure says nothing either way.
- includedItems / excludedItems: a short summary, shown to customers on the tour page, of what the tour price includes and does not include. Write it the way a good sales person would sum it up: only what helps a customer compare tours or know what they will pay, each item one short Thai line, with its key figure (weight, coverage, amount). Sources, in this order:
  1. The printed lists (อัตราค่าบริการนี้รวม / ราคานี้รวม, อัตราค่าบริการนี้ไม่รวม / ราคานี้ไม่รวม), usually on image pages near the end. Shorten each item to its point and drop qualifiers such as "ตามรายการที่ระบุไว้ในโปรแกรมทัวร์" or "เฉพาะวันที่ให้บริการทัวร์เท่านั้น": "ค่าตั๋วเครื่องบินไป-กลับ ชั้นประหยัด รวมภาษีสนามบินและค่าน้ำมัน", "น้ำหนักกระเป๋าโหลด 20 กิโลกรัม", "ที่พักห้องละ 2 ท่าน", "ค่าอาหารตามรายการ", "ค่าเข้าชมสถานที่ตามรายการ", "ประกันการเดินทาง วงเงิน 1,000,000 บาท", "ค่าทิปไกด์และคนขับรถ 1,500 บาท/ท่าน/ทริป".
     Merge items about the same thing into one line: the tour leader and the local guide → "หัวหน้าทัวร์และไกด์ท้องถิ่นดูแลตลอดการเดินทาง" (or only the one the brochure has); checked and carry-on baggage → one baggage line.
     Leave out what every tour has or what tells the customer nothing: the coach or transport for the programme, drinking water, the generic personal-expenses line (ค่าใช้จ่ายส่วนตัว เช่น ค่าโทรศัพท์ ค่าซักรีด มินิบาร์ …), the passport fee (ค่าทำหนังสือเดินทาง; when it shares a line with the visa keep only "ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า"), events beyond the company's control (strikes, flight delays, natural disasters), VAT and withholding tax (ภาษีมูลค่าเพิ่ม 7%, หัก ณ ที่จ่าย 3%), excess or oversized baggage, and conditions, notes or decorative headings (สำคัญ, โปรดอ่าน).
  2. Then read the cover, the itinerary and the notes under the price table for specific paid tickets, rides and activities the lists don't already name:
     - Included: entrance tickets (ค่าเข้าชม, บัตรเข้า), theme and amusement parks (สวนสนุก, Disneyland, Universal Studios), cable cars (กระเช้า), mountain trains and funiculars (รถไฟ, รถราง), cruises and boat rides (ล่องเรือ), shows, hot springs (ออนเซ็น) and special meals (เมนูพิเศษ, มื้อพิเศษ) that the tour takes everyone to, especially when it says รวม (for example "(รวมค่าเข้าบัตรเข้าสวนสนุก)", "(รวมล่องเรือ)"). Name the place: "บัตรเข้าสวนสนุก Shanghai Disneyland", "ล่องเรือเมืองโบราณจูเจียเจียว", "กระเช้าขึ้นยอดเขา …".
     - Not included: a ticket or activity of the programme that the brochure says is not included or paid by the traveller, for example "ไม่รวมค่าเข้าชม" → "ค่าเข้าชมโตเกียวสกายทรี". Skip small personal purchases along the way (drinks at a café, snacks, souvenirs, meals at leisure).
     - Options: paid extras a traveller may choose (ทัวร์เสริม, upgrades such as business-class seats, tickets bought separately, for example "ซื้อแบบ Access ในแอปพลิเคชันด้วยตนเอง") go in excludedItems prefixed "Option: ", with the price when printed: "Option: ทัวร์ล่องเรือชมวิวยามค่ำ ท่านละ 1,200 บาท".
     - Do not add places the tour only passes or photographs from outside (ผ่านชม, ถ่ายรูปด้านนอก, ถ่ายรูปคู่กับ), free walks, shopping stops or sights with no ticket.
  Keep the brochure's order. Usually 5-12 items each; at most 20. Empty arrays when the brochure says nothing.
- itinerary: a short timeline of the places visited each day, for customers to scan. One entry per day of the day-by-day plan (วันที่ 1, วันที่ 2 …), in order, so there are as many entries as the tour has days; day is the day number. Read it from the summary table first (the table with one row per day: วัน | รายการท่องเที่ยว / โปรแกรมการเดินทาง | meals | โรงแรมที่พัก): each row already lists that day's route. Where a row is cut off, missing or unclear, use the heading of the same day in the day-by-day plan (วันที่ 1 …); when the two disagree, follow the heading of the day-by-day plan. places: the places of that day in visiting order, each as a short name of a few words: airports as "สนามบิน" + the usual Thai name ("สนามบินสุวรรณภูมิ", "สนามบินดอนเมือง", "สนามบินผู่ตง"); cities, towns and sights by their usual Thai name as printed ("วัดอาซากุสะ", "ภูเขาไฟฟูจิชั้น 5"); names with no Thai form in English as printed ("Shanghai Disneyland", "Pop Mart"). Include places only passed (ผ่านชม) when the heading lists them. Do not add hotels: no "โรงแรม", "ที่พัก" or hotel names (the table's hotel column is not part of the timeline). Never add descriptions, times, meals, flight numbers or words such as นำท่าน, เดินทางสู่, อิสระ. At most 12 places a day.
- Seats and the wholesale (tour operator) name are never in these brochures; there are no fields for them.
- confidence: for each section, "high" when the document states it clearly, "medium" when you had to read an unclear layout or image, "low" when you are unsure. main = name, days/nights, airline, wholesale tour code; taxonomy = countries and cities; periods = dates; prices = all prices; extras = extras; inclusions = governmentShopStops and the included / excluded lists; itinerary = itinerary.

Worked examples (excerpts of other brochures; follow the rules above, not only these patterns)

Example A — a price table grouped by month, with the weekday before each date and the year on both dates:
  อัตราค่าบริการ *ทัวร์ไม่ลงร้านค้าท้องถิ่น*
  ช่วงเวลาเดินทาง | ราคาผู้ใหญ่ | ราคาเด็ก (2-18 ปี) | ราคาพักเดี่ยว
  เดือน ตุลาคม 2569
  อ. 27 ต.ค. – ส. 31 ต.ค. 2569 | 23,999 บาท | 23,999 บาท | 3,500 บาท
  เดือน พฤศจิกายน 2569
  อ. 03 พ.ย. – ส. 07 พ.ย. 2569 (Promotion) | 20,999 บาท | 20,999 บาท | 3,500 บาท
  จ. 28 ธ.ค. – ศ. 01 ม.ค. 2570 (วันสิ้นปีและวันขึ้นปีใหม่) | 27,999 บาท | 27,999 บาท | 3,500 บาท
  ราคาเด็กอายุไม่ถึง 2 ปี ณ วันเดินทาง (Infant) ท่านละ 6,000 บาท (ไม่มีที่นั่งบนเครื่องบิน)
Periods:
  {"start":{"day":27,"month":10,"year":2569},"end":{"day":31,"month":10,"year":2569},"adultPrice":"23,999","childWithBedPrice":"23,999","childNoBedPrice":"23,999","infantPrice":"6,000","singleRoom":"3,500","singleRoomKind":"supplement","label":"","sourceText":"อ. 27 ต.ค. – ส. 31 ต.ค. 2569 23,999 บาท 23,999 บาท 3,500 บาท"}
  {"start":{"day":3,"month":11,"year":2569},"end":{"day":7,"month":11,"year":2569},"adultPrice":"20,999","childWithBedPrice":"20,999","childNoBedPrice":"20,999","infantPrice":"6,000","singleRoom":"3,500","singleRoomKind":"supplement","label":"(Promotion)","sourceText":"อ. 03 พ.ย. – ส. 07 พ.ย. 2569 (Promotion) 20,999 บาท 20,999 บาท 3,500 บาท"}
  {"start":{"day":28,"month":12,"year":0},"end":{"day":1,"month":1,"year":2570},"adultPrice":"27,999","childWithBedPrice":"27,999","childNoBedPrice":"27,999","infantPrice":"6,000","singleRoom":"3,500","singleRoomKind":"supplement","label":"(วันสิ้นปีและวันขึ้นปีใหม่)","sourceText":"จ. 28 ธ.ค. – ศ. 01 ม.ค. 2570 27,999 บาท 27,999 บาท 3,500 บาท"}
governmentShopStops: "no" (the title says ทัวร์ไม่ลงร้านค้ารัฐบาล).
From the itinerary of the same brochure: "SHANGHAI DISNEYLAND (รวมรถรับ-ส่ง และบริการไกด์ท้องถิ่น) … (รวมค่าเข้าบัตรเข้าสวนสนุก-บริการรับส่งพร้อมคณะ) … ไม่สามารถซื้อบัตร Early เสริมได้ แต่สามารถซื้อแบบ Access ใน Application ด้วยตนเอง", "เมืองโบราณจูเจียเจียว (รวมล่องเรือ)", "Starbucks Reserve Roastery … (ไม่รวมค่าเครื่องดื่ม)", "ผ่านชม หอไข่มุก" adds to includedItems "บัตรเข้าสวนสนุก Shanghai Disneyland พร้อมรถรับ-ส่ง" and "ล่องเรือเมืองโบราณจูเจียเจียว", and to excludedItems "Option: บัตร Access ดิสนีย์แลนด์ ซื้อเองในแอปพลิเคชัน". The Starbucks drinks are a small personal purchase and the Pearl Tower (หอไข่มุก) is only passed (ผ่านชม), so neither is listed.
Its printed lists, summed up: includedItems ["ค่าตั๋วเครื่องบินไป-กลับ ชั้นประหยัด รวมภาษีสนามบินและค่าน้ำมัน", "น้ำหนักกระเป๋าโหลด 20 กิโลกรัม ถือขึ้นเครื่อง 7 กิโลกรัม", "ที่พักห้องละ 2 ท่าน", "ค่าอาหารตามรายการ", "ค่าเข้าชมสถานที่ตามรายการ", "หัวหน้าทัวร์และไกด์ท้องถิ่นดูแลตลอดการเดินทาง", "ประกันการเดินทาง วงเงิน 1,000,000 บาท", then the two tickets above]; excludedItems ["ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า", "ค่าภาษีน้ำมันและภาษีสนามบินที่สายการบินเรียกเก็บเพิ่มภายหลัง", "ค่าทิปไกด์และคนขับรถท้องถิ่น 1,500 บาท/ท่าน/ทริป", then the option above]. Personal expenses, the passport fee, events beyond the company's control, VAT / withholding tax, excess baggage, the coach and drinking water are left out.
Its itinerary (5 days):
  {"day":1,"places":["สนามบินสุวรรณภูมิ","สนามบินผู่ตง"]}
  {"day":2,"places":["เมืองโบราณจูเจียเจียว","ถนนอู่คัง","ตลาดเฉินหวังเมี่ยว"]}
  {"day":3,"places":["Shanghai Disneyland"]}
  {"day":4,"places":["วัดพระหยกขาว","อาคารพันต้นไม้","Starbucks Reserve Roastery","เรือ The Louis Shanghai","หอไข่มุก","หาดไว่ทาน","ถนนนานกิง","Pop Mart","สนามบินผู่ตง"]}
  {"day":5,"places":["สนามบินผู่ตง","สนามบินสุวรรณภูมิ"]} The child column has no bed wording, so its price goes in both child fields. The infant price is printed once under the table, so every period gets "6,000". The 3,500 single-room price is far below the adult price, so it is a supplement. The 28 Dec row prints the year only once, at the end, so the start year is 0.

Example B — day ranges within one month, a range across two months, and conditions on image pages:
  ช่วงเวลาเดินทาง | ราคาผู้ใหญ่ | ราคาเด็ก (2-18 ปี) | ราคาพักเดี่ยว
  02-06 ตุลาคม 2569 | 28,999 บาท | 28,999 บาท | 6,900 บาท
  28 ตุลาคม - 01 พฤศจิกายน 2569 | 29,999 บาท | 29,999 บาท | 6,900 บาท
  • กรณีถือหนังสือเดินทางต่างชาติบริษัทฯ ขอสงวนสิทธิ์เก็บค่าธรรมเนียมเพิ่มเติม ท่านละ 5,000 บาท
  • กรณีต้องการจอยแลนด์ (ไม่ใช้ตั๋วเครื่องบิน) หักออกจากราคาทัวร์เต็มท่านละ 5,000 บาท
  ขอสงวนสิทธิ์เก็บค่าทิปไกด์และคนขับรถท้องถิ่น ณ สนามบินในวันเช็คอิน ท่านละ 2,000 บาท/ทริป/ท่าน
  (conditions image) การสำรองที่นั่ง กรุณาสำรองที่นั่งและชำระมัดจำ ท่านละ 15,000 บาท ภายใน 24 ชั่วโมง
  ราคาเด็กอายุไม่ถึง 2 ปี ณ วันเดินทาง (Infant) ท่านละ 10,000 บาท (ไม่มีที่นั่งบนเครื่องบิน)
  (itinerary) นำท่าน ผ่านชม "โตเกียวสกายทรี" (ราคาทัวร์ไม่รวมค่าเข้าชม)
  (cover image) PJPNRT22-SL next to the Thai Lion Air logo; การันตีเดินทาง 10 ท่านขึ้นไป
Periods:
  {"start":{"day":2,"month":10,"year":2569},"end":{"day":6,"month":10,"year":2569},...}
  {"start":{"day":28,"month":10,"year":0},"end":{"day":1,"month":11,"year":2569},...}
Every period: childWithBedPrice and childNoBedPrice equal to the child column, infantPrice "10,000" (printed once: ราคาเด็กอายุไม่ถึง 2 ปี ณ วันเดินทาง (Infant) ท่านละ 10,000 บาท).
extras: depositPerPerson "15,000", joinLandPrice "5,000", minGroupSize 10, otherCharges ["กรณีถือหนังสือเดินทางต่างชาติบริษัทฯ ขอสงวนสิทธิ์เก็บค่าธรรมเนียมเพิ่มเติม ท่านละ 5,000 บาท", "ขอสงวนสิทธิ์เก็บค่าทิปไกด์และคนขับรถท้องถิ่น ณ สนามบินในวันเช็คอิน ท่านละ 2,000 บาท/ทริป/ท่าน"]. wholesaleTourCode: "PJPNRT22-SL". excludedItems also gets "ค่าเข้าชมโตเกียวสกายทรี": the tour only passes it and says the entry is not included. airline: {"iataCode":"SL","name":"Thai Lion Air"}.`

// AIExtractFewShotFiles are the brochures the worked examples above are taken from. The eval reports
// every other file in backend/ex as held out (a fair test of files the prompt has never seen).
var AIExtractFewShotFiles = []string{"nidn263430", "nidn263439"}

// AIExtractInstruction is the user turn after the PDF.
const AIExtractInstruction = "Extract this tour program following the rules in the system prompt."
