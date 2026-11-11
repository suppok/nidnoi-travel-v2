// Reads the booking summary text made by the backoffice sales tool "สรุปยอดโอน"
// (frontend/backoffice/src/features/sales-tool/booking-summary/summary.ts) so sales can paste it
// here to prefill the form. It relies on that text's "• <label> : <value>" lines and its payment
// sentences: keep both files in sync.

export type PricedLine = { price: number; count: number }

export type ParsedSummary = {
  source: string | null
  refName: string | null
  refBookingNo: string | null
  tourCode: string | null
  tourName: string | null // without the "(5 วัน 3 คืน)" suffix
  startDate: string | null // YYYY-MM-DD
  endDate: string | null
  travellers: number | null
  adultPrice: number | null
  adultPriceBeforeDiscount: number | null
  child: PricedLine | null
  infant: PricedLine | null
  singleRoom: PricedLine | null // count = rooms
  visa: PricedLine | null
  foreignPassport: PricedLine | null
  joinLand: PricedLine | null
  fuelTax: PricedLine | null
  airportTax: PricedLine | null
  options: (PricedLine & { name: string })[]
  discountPerPerson: number | null
  creditFeePercent: number | null // set only when paying by card
  isDeposit: boolean | null // null = no payment line found
  depositPerPerson: number | null
  depositDueDate: string | null
  depositDueTime: string | null // HH:mm
  remainingDueDate: string | null
  fullDueDate: string | null
  fullDueTime: string | null
  total: number | null
  customerName: string | null // without the "คุณ" prefix
  customerPhone: string | null
  customerAddress: string | null
  salesName: string | null // "<name> <phone>" as written; match against the sales list by prefix
}

const THAI_MONTHS = [
  'มกราคม', 'กุมภาพันธ์', 'มีนาคม', 'เมษายน', 'พฤษภาคม', 'มิถุนายน',
  'กรกฎาคม', 'สิงหาคม', 'กันยายน', 'ตุลาคม', 'พฤศจิกายน', 'ธันวาคม',
]

// "15 มกราคม 2570" (Buddhist year) → "2027-01-15"
export function parseThaiDate(text: string): string | null {
  const m = /(\d{1,2})\s+(\S+)\s+(\d{4})/.exec(text)
  if (!m) return null
  const month = THAI_MONTHS.indexOf(m[2])
  if (month < 0) return null
  let year = Number(m[3])
  if (year > 2400) year -= 543
  return `${year}-${String(month + 1).padStart(2, '0')}-${m[1].padStart(2, '0')}`
}

const count = (value: string | undefined) => Number(/\(จำนวน\s+(\d+)/.exec(value ?? '')?.[1] ?? 0)

function amount(value: string | undefined): number | null {
  const m = /([\d,]+)\s*บาท/.exec(value ?? '')
  return m ? Number(m[1].replace(/,/g, '')) : null
}

function priced(value: string | undefined): PricedLine | null {
  const price = amount(value)
  return price === null ? null : { price, count: count(value) }
}

// "... ภายในวันที่ 15 มกราคม 2570 (ก่อน 18:00 น.) ..." → date and time
function deadline(line: string): { date: string | null; time: string | null } {
  const after = line.split('ภายในวันที่')[1] ?? ''
  return { date: parseThaiDate(after), time: /\(ก่อน\s+(\d{1,2}:\d{2})\s*น\.\)/.exec(after)?.[1] ?? null }
}

// Labels the summary writes itself; any other "<name>/ท่าน" priced line is an option.
const KNOWN_LABELS = new Set([
  'ราคาทัวร์ผู้ใหญ่/ท่าน', 'ราคาเด็กอายุต่ำกว่า 20 ปี/ท่าน', 'ราคาเด็กทารก/ท่าน', 'ค่า visa/ท่าน',
  'ค่า passport ต่างชาติ/ท่าน', 'ค่าห้องพักเดี่ยว/ห้อง', 'ค่า join land/ท่าน', 'ค่าภาษีน้ำมัน/ท่าน',
  'ค่าภาษีสนามบิน/ท่าน', 'ส่วนลด/ท่าน',
])

// Returns null when the text doesn't look like a booking summary at all.
export function parseSummary(text: string): ParsedSummary | null {
  const fields = new Map<string, string>()
  const options: ParsedSummary['options'] = []
  let refName: string | null = null
  let creditFeePercent: number | null = null
  const pay: Pick<ParsedSummary, 'isDeposit' | 'depositPerPerson' | 'depositDueDate' | 'depositDueTime' | 'remainingDueDate' | 'fullDueDate' | 'fullDueTime'> = {
    isDeposit: null, depositPerPerson: null, depositDueDate: null, depositDueTime: null, remainingDueDate: null, fullDueDate: null, fullDueTime: null,
  }

  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim()
    // Payment sentences have no "label : value" shape (and their times contain ":").
    if (/^•\s*กรุณาชำระเงิน/.test(line)) {
      const due = deadline(line)
      if (line.includes('มัดจำ')) {
        pay.isDeposit = true
        pay.depositPerPerson = amount(/\(ท่านละ\s*([^)]*)\)/.exec(line)?.[1])
        pay.depositDueDate = due.date
        pay.depositDueTime = due.time
      } else if (line.includes('ส่วนเหลือ')) {
        pay.remainingDueDate = due.date
      } else if (line.includes('เต็มจำนวน')) {
        pay.isDeposit = false
        pay.fullDueDate = due.date
        pay.fullDueTime = due.time
      }
      continue
    }
    const m = /^•\s*(.+?)\s*:\s*(.*?)$/.exec(line)
    if (!m) continue
    const [, label, rawValue] = m
    const value = rawValue === '-' ? '' : rawValue
    if (fields.has(label)) continue
    fields.set(label, value)

    if (/^ชื่อ(\s|$)/.test(label)) refName = value || null // "ชื่อ" or "ชื่อ <source short name>"
    const fee = /^ค่าธรรมเนียมบัตรเครดิต\s*\(([\d.]+)%\)/.exec(label)
    if (fee) creditFeePercent = Number(fee[1])
    if (label.endsWith('/ท่าน') && !KNOWN_LABELS.has(label) && value.includes('(จำนวน')) {
      const p = priced(value)
      if (p) options.push({ name: label.slice(0, -'/ท่าน'.length), ...p })
    }
  }
  if (!fields.has('ชื่อโปรแกรม') && !fields.has('ผู้จอง')) return null

  const get = (label: string) => fields.get(label) || null
  const [start, end] = (get('วันเดินทาง') ?? '').split(/\s+-\s+/)
  const adultLine = get('ราคาทัวร์ผู้ใหญ่/ท่าน')

  return {
    source: get('จองผ่าน'),
    refName,
    refBookingNo: get('เลข booking'),
    tourCode: get('รหัสทัวร์'),
    tourName: get('ชื่อโปรแกรม')?.replace(/\s*\(\d+\s*วัน\s*\d+\s*คืน\)\s*$/, '') ?? null,
    startDate: start ? parseThaiDate(start) : null,
    endDate: end ? parseThaiDate(end) : null,
    travellers: Number(/\d+/.exec(get('จำนวนผู้เดินทาง') ?? '')?.[0]) || null,
    adultPrice: amount(adultLine ?? undefined),
    adultPriceBeforeDiscount: amount(/\(จาก\s*([^)]*)\)/.exec(adultLine ?? '')?.[1]),
    child: priced(fields.get('ราคาเด็กอายุต่ำกว่า 20 ปี/ท่าน')),
    infant: priced(fields.get('ราคาเด็กทารก/ท่าน')),
    singleRoom: priced(fields.get('ค่าห้องพักเดี่ยว/ห้อง')),
    visa: priced(fields.get('ค่า visa/ท่าน')),
    foreignPassport: priced(fields.get('ค่า passport ต่างชาติ/ท่าน')),
    joinLand: priced(fields.get('ค่า join land/ท่าน')),
    fuelTax: priced(fields.get('ค่าภาษีน้ำมัน/ท่าน')),
    airportTax: priced(fields.get('ค่าภาษีสนามบิน/ท่าน')),
    options,
    discountPerPerson: amount(fields.get('ส่วนลด/ท่าน')),
    creditFeePercent,
    ...pay,
    total: amount(fields.get('รวมเป็นเงิน')),
    customerName: get('ผู้จอง')?.replace(/^คุณ\s*/, '') || null,
    customerPhone: get('เบอร์โทร'),
    customerAddress: get('ที่อยู่'),
    salesName: get('ผู้ดูแล'),
  }
}
