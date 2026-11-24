import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import { TourProgramStates } from '../../constants/tourProgramStates.ts';
import type { AIDraftPeriod, AIExtractJob, AIExtractResult, AIWarning } from '../../types/aiExtract.ts';
import { mockAISetting } from './aiSettings.ts';
import { mockFindAirlineByCode } from './airlines.ts';
import { authorize, MockFail, type MockHandler, ok, validation } from './http.ts';
import { mockFindCityBySlug, mockFindCountryBySlug } from './taxonomy.ts';

// Mock of the AI prefill. Like the backend it starts a job and the page polls it; the mock "reads"
// any PDF as the same Shanghai brochure after a few seconds. A file name containing "fail" fails
// the job (AI_REQUEST_FAILED), so the error path can be tried.
const READ_MS = 6000;

type Job = { id: string; ownerId: string; startedAt: number; fileName: string; fail: boolean };
const jobs = new Map<string, Job>();

function period(start: string, end: string, price: number, note = ''): AIDraftPeriod {
  return {
    startDate: start,
    endDate: end,
    price,
    priceBeforeDiscount: null,
    childWithBedPrice: price,
    childNoBedPrice: price,
    infantPrice: 6000,
    singleSupplement: 3500,
    deposit: 10000,
    seatsTotal: 0,
    seatsAvailable: 0,
    note,
  };
}

function mockResult(fileName: string): AIExtractResult {
  const warnings: AIWarning[] = [];
  const china = mockFindCountryBySlug('china');
  const shanghai = mockFindCityBySlug('shanghai');
  const airline = mockFindAirlineByCode('VZ');
  if (!china) warnings.push({ field: 'countryIds', message: 'ไม่พบประเทศ "จีน / China" ในระบบ กรุณาเลือกเอง' });
  if (!shanghai) warnings.push({ field: 'cityIds', message: 'ไม่พบเมือง "เซี่ยงไฮ้ / Shanghai" ในระบบ (ไม่บังคับ — เพิ่มเองได้)' });
  if (!airline) warnings.push({ field: 'airlineId', message: 'ไม่พบสายการบิน "VIETJET AIR VZ" ในระบบ กรุณาเลือกเอง' });
  return {
    draft: {
      refId: null,
      name: 'เซี่ยงไฮ้ ดิสนีย์แลนด์ (ทัวร์ไม่ลงร้านค้ารัฐบาล)',
      wholesaleId: null,
      wholesaleTourCode: 'PCN-PVG580-VZ',
      state: TourProgramStates.Draft,
      days: 5,
      nights: 3,
      airlineId: airline?.id ?? null,
      highlight: 'เซี่ยงไฮ้ ดิสนีย์แลนด์ เมืองโบราณจูเจียเจียว (รวมล่องเรือ) เมนูพิเศษ เสี่ยวหลงเปา หมูน้ำแดง พักโรงแรม 4 ดาว ทัวร์ไม่ลงร้านค้ารัฐบาล',
      description: '',
      bannerImage: null,
      cardImage: null,
      pdfFile: null,
      docxFile: null,
      internalNote: 'จาก PDF (AI):\nกรณีถือหนังสือเดินทางต่างชาติบริษัทฯ ขอสงวนสิทธิ์เก็บค่าธรรมเนียมเพิ่มเติม ท่านละ 2,000 บาท',
      governmentShopStops: false,
      includedItems: [
        'ค่าตั๋วเครื่องบินไป-กลับ ชั้นประหยัด รวมภาษีสนามบินและค่าน้ำมัน',
        'น้ำหนักกระเป๋าโหลด 20 กิโลกรัม ถือขึ้นเครื่อง 7 กิโลกรัม',
        'ที่พักห้องละ 2 ท่าน',
        'ค่าอาหารตามรายการ',
        'ค่าเข้าชมสถานที่ตามรายการ',
        'หัวหน้าทัวร์และไกด์ท้องถิ่นดูแลตลอดการเดินทาง',
        'ประกันการเดินทาง วงเงิน 1,000,000 บาท',
        'บัตรเข้าสวนสนุก Shanghai Disneyland พร้อมรถรับ-ส่ง',
        'ล่องเรือเมืองโบราณจูเจียเจียว',
      ],
      excludedItems: [
        'ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า',
        'ค่าภาษีน้ำมันและภาษีสนามบินที่สายการบินเรียกเก็บเพิ่มภายหลัง',
        'ค่าทิปไกด์และคนขับรถท้องถิ่น 1,500 บาท/ท่าน/ทริป',
        'Option: บัตร Access ดิสนีย์แลนด์ ซื้อเองในแอปพลิเคชัน',
      ],
      itinerary: [
        { day: 1, places: ['สนามบินสุวรรณภูมิ', 'สนามบินผู่ตง'] },
        { day: 2, places: ['เมืองโบราณจูเจียเจียว', 'ถนนอู่คัง', 'ตลาดเฉินหวังเมี่ยว'] },
        { day: 3, places: ['Shanghai Disneyland'] },
        { day: 4, places: ['วัดพระหยกขาว', 'อาคารพันต้นไม้', 'Starbucks Reserve Roastery', 'หาดไว่ทาน', 'ถนนนานกิง', 'สนามบินผู่ตง'] },
        { day: 5, places: ['สนามบินผู่ตง', 'สนามบินสุวรรณภูมิ'] },
      ],
      countryIds: china ? [china.id] : [],
      cityIds: china && shanghai ? [shanghai.id] : [],
      images: [],
      periods: [
        period('2026-12-01', '2026-12-05', 23999, '(วันพ่อ)'),
        period('2026-12-14', '2026-12-18', 22999),
        period('2026-12-28', '2027-01-01', 27999, '(วันสิ้นปีและวันขึ้นปีใหม่)'),
        period('2027-03-02', '2027-03-06', 20999, '(Promotion)'),
      ],
    },
    warnings,
    confidence: { main: 'high', taxonomy: 'high', periods: 'high', prices: 'medium', extras: 'low', inclusions: 'medium', itinerary: 'high' },
    usage: { inputTokens: 68400, outputTokens: 5200, cacheReadInputTokens: 0, cacheCreationInputTokens: 2900 },
    model: mockAISetting().model,
    fileName,
    durationMs: READ_MS,
  };
}

export const aiExtractRoutes: Record<string, MockHandler> = {
  'GET /tour-programs/ai-extract/config': (req) => {
    authorize(req, 'GET /tour-programs/ai-extract/config', Permissions.TourProgramAIExtract);
    const s = mockAISetting();
    return ok({ ready: s.ready, maxPdfSizeMb: s.maxPdfSizeMb });
  },
  'POST /tour-programs/ai-extract': (req) => {
    const user = authorize(req, 'POST /tour-programs/ai-extract', Permissions.TourProgramAIExtract);
    const s = mockAISetting();
    if (!s.ready) throw new MockFail(503, ErrorCodes.AINotConfigured, 'AI is not configured');
    const file = req.body instanceof FormData ? req.body.get('file') : null;
    if (!(file instanceof File) || file.size === 0) throw validation({ file: 'a PDF file is required' });
    if (file.size > s.maxPdfSizeMb * 1024 * 1024) throw new MockFail(413, ErrorCodes.MediaTooLarge, 'The file is too large');
    if (file.type && file.type !== 'application/pdf') throw new MockFail(400, ErrorCodes.MediaTypeNotAllowed, 'This file type is not allowed for this purpose');
    const id = crypto.randomUUID();
    jobs.set(id, { id, ownerId: user.id, startedAt: Date.now(), fileName: file.name, fail: file.name.toLowerCase().includes('fail') });
    return ok({ jobId: id, state: 'pending' });
  },
  'GET /tour-programs/ai-extract/:jobId': (req) => {
    const user = authorize(req, 'GET /tour-programs/ai-extract/:jobId', Permissions.TourProgramAIExtract);
    const job = jobs.get(req.params.jobId);
    if (!job || job.ownerId !== user.id) throw new MockFail(404, ErrorCodes.NotFound, 'Not found');
    const body: AIExtractJob = { jobId: job.id, state: 'pending', result: null, error: null };
    if (Date.now() - job.startedAt >= READ_MS) {
      if (job.fail) {
        body.state = 'failed';
        body.error = { code: ErrorCodes.AIRequestFailed, message: 'The AI request failed: the AI declined to read this document' };
      } else {
        body.state = 'done';
        body.result = mockResult(job.fileName);
      }
    }
    return ok(body);
  },
};
