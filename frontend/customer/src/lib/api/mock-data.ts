// Sample data in the shape of the planned public DTOs, used while USE_MOCK_DATA is true.
// Build-time only — never import this from an island.
// Content is adapted from the design mock (ex/nidnoi-travel-v2/data.jsx). Images are Unsplash placeholders.
import type { CityPublic, CityRef, CountryPublic, TaxonomyRef } from '../../types/taxonomy';
import type { AirlinePublic, TourPeriodPublic, TourProgramPublicDetail } from '../../types/tour-program';

const img = (id: string) => `https://images.unsplash.com/photo-${id}?w=2000&q=80`;
const UPDATED = '2026-11-01T09:00:00+07:00';

type CountrySeed = Omit<CountryPublic, 'tourCount' | 'startingPrice' | 'updateDate'>;
type CitySeed = Omit<CityPublic, 'tourCount' | 'startingPrice' | 'updateDate'>;

const COUNTRIES: CountrySeed[] = [
  {
    slug: 'japan',
    nameTh: 'ญี่ปุ่น',
    nameEn: 'Japan',
    description:
      'ดินแดนแห่งสี่ฤดูที่ทุกฤดูสวยไม่ซ้ำกัน ใบไม้เปลี่ยนสีในฤดูใบไม้ร่วง หิมะขาวในฤดูหนาว ดอกซากุระบานในฤดูใบไม้ผลิ และเทศกาลดอกไม้ไฟในฤดูร้อน รวมโปรแกรมทัวร์ญี่ปุ่นทั้งโตเกียว โอซาก้า ฮอกไกโด โทโฮคุ และโอกินาว่า เลือกวันเดินทางและเช็กที่นั่งว่างได้ทันที',
    heroImage: img('1493976040374-85c8e12f0c0e'),
    sortOrder: 1,
  },
  {
    slug: 'korea',
    nameTh: 'เกาหลีใต้',
    nameEn: 'Korea',
    description:
      'เที่ยวเกาหลีใต้ได้ทุกฤดู ตั้งแต่ใบไม้เปลี่ยนสีบนเกาะนามิ ไปจนถึงลานสกีในฤดูหนาว ช้อปปิ้งเมียงดง ชมพระราชวังเคียงบกกุง และป้อมฮวาซองมรดกโลก บินตรงจากกรุงเทพฯ เพียงห้าชั่วโมงครึ่ง',
    heroImage: img('1517154421773-0529f29ea451'),
    sortOrder: 2,
  },
  {
    slug: 'vietnam',
    nameTh: 'เวียดนาม',
    nameEn: 'Vietnam',
    description:
      'ใกล้ ไปง่าย ราคาดี เวียดนามกลางมีทั้งสะพานมือทองบนบานาฮิลล์ เมืองเก่าฮอยอันที่เป็นมรดกโลก และชายหาดดานัง เหมาะกับทริปสั้นสี่วันสามคืน',
    heroImage: img('1528127269322-539801943592'),
    sortOrder: 3,
  },
  {
    slug: 'china',
    nameTh: 'จีน',
    nameEn: 'China',
    description:
      'จีนกว้างใหญ่เกินกว่าจะเที่ยวครั้งเดียวหมด เริ่มจากเฉิงตู เมืองแพนด้า อาหารเสฉวนรสจัด และภูเขาหิมะซีหลิงที่ขึ้นกระเช้าไปสัมผัสหิมะได้ตลอดฤดูหนาว',
    heroImage: img('1508804185872-d7badad00f7d'),
    sortOrder: 4,
  },
  {
    slug: 'taiwan',
    nameTh: 'ไต้หวัน',
    nameEn: 'Taiwan',
    description:
      'ไต้หวันเที่ยวสบาย อาหารอร่อย ตั้งแต่ตึกไทเป 101 ตลาดกลางคืนซื่อหลิน ไปจนถึงทะเลสาบสุริยันจันทราและรถไฟสายเก่าบนเขาอาลีซาน',
    heroImage: img('1470004914212-05527e49370b'),
    sortOrder: 5,
  },
  {
    slug: 'europe',
    nameTh: 'ยุโรป',
    nameEn: 'Europe',
    description:
      'ทัวร์ยุโรปหลายประเทศในทริปเดียว อิตาลี สวิตเซอร์แลนด์ ฝรั่งเศส ขึ้นยอดเขาจุงเฟรา ล่องเรือแม่น้ำแซน และชมหอเอนเมืองปิซ่า',
    heroImage: img('1499856871958-5b9627545d1a'),
    sortOrder: 6,
  },
  {
    slug: 'turkey',
    nameTh: 'ตุรกี',
    nameEn: 'Turkey',
    description:
      'ตุรกีดินแดนสองทวีป นั่งบอลลูนชมพระอาทิตย์ขึ้นที่คัปปาโดเกีย เดินชมพระราชวังท็อปกาปึในอิสตันบูล และแช่น้ำแร่ที่ปามุคคาเล',
    heroImage: img('1524231757912-21f4fe3a7200'),
    sortOrder: 7,
  },
  {
    slug: 'georgia',
    nameTh: 'จอร์เจีย',
    nameEn: 'Georgia',
    description:
      'จอร์เจียประเทศเล็กแถบเทือกเขาคอเคซัส วิวภูเขาสูงที่คาซเบกี เมืองเก่าทบิลิซี และไวน์ที่ทำกันมากว่าแปดพันปี',
    heroImage: img('1565008447742-97f6f38c985c'),
    sortOrder: 8,
  },
  {
    // No tours: proves empty landings are not generated.
    slug: 'singapore',
    nameTh: 'สิงคโปร์',
    nameEn: 'Singapore',
    description: '',
    heroImage: img('1565967511849-76a60a516170'),
    sortOrder: 9,
  },
];

const CITIES: CitySeed[] = [
  { slug: 'tokyo', countrySlug: 'japan', nameTh: 'โตเกียว', nameEn: 'Tokyo', sortOrder: 1, heroImage: img('1540959733332-eab4deabeeaf'),
    description: 'โตเกียวเมืองหลวงที่ไม่เคยหลับ ช้อปปิ้งชินจูกุ ฮาราจูกุ ไหว้พระวัดอาซากุสะ และเดินทางต่อไปชมภูเขาไฟฟูจิได้ในวันเดียว' },
  { slug: 'osaka', countrySlug: 'japan', nameTh: 'โอซาก้า', nameEn: 'Osaka', sortOrder: 2, heroImage: img('1590559899731-a382839e5549'),
    description: 'โอซาก้าครัวของญี่ปุ่น ปราสาทโอซาก้า ย่านโดทงโบริ และเดินทางสะดวกไปเกียวโตกับนารา' },
  { slug: 'hokkaido', countrySlug: 'japan', nameTh: 'ฮอกไกโด', nameEn: 'Hokkaido', sortOrder: 3, heroImage: img('1542640244-7e672d6cef4e'),
    description: 'ฮอกไกโดเกาะเหนือสุดของญี่ปุ่น หิมะขาวในฤดูหนาว ทุ่งดอกไม้ในฤดูร้อน ซัปโปโร โอตารุ และอาหารทะเลสดจากฮาโกดาเตะ' },
  { slug: 'tohoku', countrySlug: 'japan', nameTh: 'โทโฮคุ', nameEn: 'Tohoku', sortOrder: 4, heroImage: img('1528360983277-13d401cdc186'),
    description: 'โทโฮคุภูมิภาคทางเหนือของเกาะฮอนชู ทะเลสาบโทวาดะ ออนเซ็นกลางป่า และใบไม้เปลี่ยนสีที่คนยังไม่พลุกพล่าน' },
  { slug: 'okinawa', countrySlug: 'japan', nameTh: 'โอกินาว่า', nameEn: 'Okinawa', sortOrder: 5, heroImage: img('1545569341-9eb8b30979d9'),
    description: 'โอกินาว่าหมู่เกาะทางใต้ของญี่ปุ่น ทะเลสีฟ้าใส แนวปะการัง และวัฒนธรรมริวกิวที่ไม่เหมือนที่ไหน' },
  { slug: 'seoul', countrySlug: 'korea', nameTh: 'โซล', nameEn: 'Seoul', sortOrder: 1, heroImage: img('1517154421773-0529f29ea451'),
    description: 'โซลเมืองหลวงของเกาหลีใต้ พระราชวังโบราณกลางเมืองสมัยใหม่ ช้อปเมียงดง และเดินทางต่อเกาะนามิได้ง่าย' },
  { slug: 'danang', countrySlug: 'vietnam', nameTh: 'ดานัง', nameEn: 'Da Nang', sortOrder: 1, heroImage: img('1528127269322-539801943592'),
    description: 'ดานังเมืองชายทะเลของเวียดนามกลาง ใกล้บานาฮิลล์และเมืองเก่าฮอยอัน' },
  { slug: 'chengdu', countrySlug: 'china', nameTh: 'เฉิงตู', nameEn: 'Chengdu', sortOrder: 1, heroImage: img('1547981609-4b6bfe67ca0b'),
    description: 'เฉิงตูเมืองหลวงของมณฑลเสฉวน บ้านของหมีแพนด้ายักษ์และอาหารรสจัด' },
  { slug: 'taipei', countrySlug: 'taiwan', nameTh: 'ไทเป', nameEn: 'Taipei', sortOrder: 1, heroImage: img('1470004914212-05527e49370b'),
    description: 'ไทเปเมืองหลวงของไต้หวัน ตึก 101 ตลาดกลางคืน และเมืองเก่าจิ่วเฟิ่น' },
  { slug: 'paris', countrySlug: 'europe', nameTh: 'ปารีส', nameEn: 'Paris', sortOrder: 1, heroImage: img('1499856871958-5b9627545d1a'),
    description: 'ปารีสเมืองแห่งศิลปะ หอไอเฟล พิพิธภัณฑ์ลูฟวร์ และล่องเรือแม่น้ำแซน' },
  { slug: 'istanbul', countrySlug: 'turkey', nameTh: 'อิสตันบูล', nameEn: 'Istanbul', sortOrder: 1, heroImage: img('1524231757912-21f4fe3a7200'),
    description: 'อิสตันบูลเมืองที่ตั้งอยู่บนสองทวีป มัสยิดสีน้ำเงิน ฮาเกียโซเฟีย และแกรนด์บาซาร์' },
  { slug: 'tbilisi', countrySlug: 'georgia', nameTh: 'ทบิลิซี', nameEn: 'Tbilisi', sortOrder: 1, heroImage: img('1565008447742-97f6f38c985c'),
    description: 'ทบิลิซีเมืองหลวงของจอร์เจีย เมืองเก่าริมแม่น้ำและโรงอาบน้ำกำมะถัน' },
];

const AIRLINES: Record<string, AirlinePublic> = {
  TG: { code: 'TG', name: 'Thai Airways', logo: null },
  JL: { code: 'JL', name: 'Japan Airlines', logo: null },
  NH: { code: 'NH', name: 'ANA', logo: null },
  MM: { code: 'MM', name: 'Peach Aviation', logo: null },
  KE: { code: 'KE', name: 'Korean Air', logo: null },
  VJ: { code: 'VJ', name: 'Vietjet Air', logo: null },
  '3U': { code: '3U', name: 'Sichuan Airlines', logo: null },
  BR: { code: 'BR', name: 'EVA Air', logo: null },
  EK: { code: 'EK', name: 'Emirates', logo: null },
  TK: { code: 'TK', name: 'Turkish Airlines', logo: null },
  QR: { code: 'QR', name: 'Qatar Airways', logo: null },
};

const GALLERY_POOL = [
  img('1480796927426-f609979314bd'),
  img('1528360983277-13d401cdc186'),
  img('1545569341-9eb8b30979d9'),
  img('1528164344705-47542687000d'),
  img('1554072675-66db59dba46f'),
];

// [start, end, price, priceBeforeDiscount, seatsAvailable]
type PeriodSeed = [string, string, number, number | null, number];

type TourSeed = {
  code: string;
  name: string;
  highlight: string;
  countries: string[];
  cities: string[];
  days: number;
  nights: number;
  airline: string;
  image: string;
  periods: PeriodSeed[];
};

const TOURS: TourSeed[] = [
  {
    code: 'nidn261690', name: 'ทัวร์ญี่ปุ่น โตเกียว ฟูจิ ฮาโกเน่', highlight: 'ใบไม้เปลี่ยนสี · ภูเขาไฟฟูจิ ชั้น 5 · ออนเซ็น',
    countries: ['japan'], cities: ['tokyo'], days: 6, nights: 4, airline: 'TG', image: img('1490806843957-31f4c9a91c65'),
    periods: [
      ['2026-11-22', '2026-11-27', 49900, 56900, 3],
      ['2026-12-06', '2026-12-11', 52900, 58900, 12],
      ['2026-12-20', '2026-12-25', 58900, 64900, 0],
      ['2027-01-10', '2027-01-15', 47900, 54900, 20],
      ['2027-01-24', '2027-01-29', 47900, 52900, 18],
    ],
  },
  {
    code: 'nidn261702', name: 'ทัวร์ญี่ปุ่น โอซาก้า เกียวโต นารา', highlight: 'วัดคิโยมิซุ · ปราสาทโอซาก้า · กวางนารา',
    countries: ['japan'], cities: ['osaka'], days: 5, nights: 3, airline: 'JL', image: img('1493976040374-85c8e12f0c0e'),
    periods: [
      ['2026-11-12', '2026-11-16', 39900, 45900, 5],
      ['2026-11-26', '2026-11-30', 41900, 47900, 14],
      ['2026-12-10', '2026-12-14', 43900, 49900, 20],
      ['2027-01-07', '2027-01-11', 38900, 44900, 25],
    ],
  },
  {
    code: 'nidn261755', name: 'ทัวร์ญี่ปุ่น ฮอกไกโด ซัปโปโร โอตารุ', highlight: 'หิมะแรกของฤดู · ตลาดปลาฮาโกดาเตะ · บ่อน้ำสีฟ้า',
    countries: ['japan'], cities: ['hokkaido'], days: 6, nights: 4, airline: 'TG', image: img('1542640244-7e672d6cef4e'),
    periods: [
      ['2026-12-18', '2026-12-23', 59900, 68900, 8],
      ['2027-01-08', '2027-01-13', 54900, 62900, 16],
      ['2027-01-22', '2027-01-27', 56900, 64900, 2],
      ['2027-02-05', '2027-02-10', 58900, 66900, 20],
    ],
  },
  {
    code: 'nidn261888', name: 'ทัวร์ญี่ปุ่น โตเกียว ดิสนีย์แลนด์ ชินจูกุ', highlight: 'ดิสนีย์แลนด์ · ดิสนีย์ซี · ช้อปชินจูกุ ฮาราจูกุ',
    countries: ['japan'], cities: ['tokyo'], days: 5, nights: 3, airline: 'NH', image: img('1480796927426-f609979314bd'),
    periods: [
      ['2026-12-04', '2026-12-08', 42900, 48900, 10],
      ['2026-12-11', '2026-12-15', 43900, 49900, 15],
      ['2026-12-25', '2026-12-29', 49900, 55900, 4],
      ['2027-01-15', '2027-01-19', 39900, 45900, 22],
    ],
  },
  {
    code: 'nidn261920', name: 'ทัวร์ญี่ปุ่น โทโฮคุ อาคิตะ อาโอโมริ', highlight: 'ทะเลสาบโทวาดะ · ทะเลสาบทาซาวะ · ภูเขาไฟอิวาเตะ',
    countries: ['japan'], cities: ['tohoku'], days: 7, nights: 5, airline: 'TG', image: img('1528360983277-13d401cdc186'),
    periods: [
      ['2026-11-14', '2026-11-20', 65900, 72900, 6],
      ['2026-11-28', '2026-12-04', 67900, 74900, 12],
      ['2026-12-12', '2026-12-18', 69900, null, 18],
    ],
  },
  {
    code: 'nidn261945', name: 'ทัวร์ญี่ปุ่น โอกินาว่า ชายหาด ดำน้ำ', highlight: 'หาดสีฟ้า · ปะการัง · พิพิธภัณฑ์สัตว์น้ำชูราอุมิ',
    countries: ['japan'], cities: ['okinawa'], days: 5, nights: 4, airline: 'MM', image: img('1545569341-9eb8b30979d9'),
    periods: [
      ['2027-03-10', '2027-03-14', 36900, null, 20],
      ['2027-03-24', '2027-03-28', 37900, null, 20],
      ['2027-04-07', '2027-04-11', 42900, null, 9],
    ],
  },
  {
    code: 'nidn255120', name: 'ทัวร์เกาหลี โซล นามิ พระราชวัง', highlight: 'เกาะนามิ · ป้อมฮวาซอง · ช้อปเมียงดง',
    countries: ['korea'], cities: ['seoul'], days: 5, nights: 3, airline: 'KE', image: img('1517154421773-0529f29ea451'),
    periods: [
      ['2026-11-15', '2026-11-19', 22900, 27900, 4],
      ['2026-11-29', '2026-12-03', 23900, 27900, 18],
      ['2026-12-13', '2026-12-17', 24900, 28900, 20],
      ['2026-12-27', '2026-12-31', 28900, 32900, 7],
    ],
  },
  {
    code: 'nidn248801', name: 'ทัวร์เวียดนาม ดานัง ฮอยอัน บานาฮิลล์', highlight: 'สะพานมือทอง · เมืองมรดกโลก · กระเช้าบานาฮิลล์',
    countries: ['vietnam'], cities: ['danang'], days: 4, nights: 3, airline: 'VJ', image: img('1528127269322-539801943592'),
    periods: [
      ['2026-11-15', '2026-11-18', 15990, 18990, 12],
      ['2026-11-29', '2026-12-02', 16990, 19990, 20],
      ['2026-12-13', '2026-12-16', 14990, 16990, 25],
    ],
  },
  {
    code: 'nidn271110', name: 'ทัวร์จีน เฉิงตู ภูเขาหิมะซีหลิง', highlight: 'หมีแพนด้ายักษ์ · ภูเขาหิมะซีหลิง · ถนนคนเดินชุนซี',
    countries: ['china'], cities: ['chengdu'], days: 5, nights: 3, airline: '3U', image: img('1547981609-4b6bfe67ca0b'),
    periods: [
      ['2026-12-05', '2026-12-09', 18900, 22900, 14],
      ['2026-12-19', '2026-12-23', 19900, 23900, 20],
      ['2027-01-09', '2027-01-13', 17900, 21900, 25],
    ],
  },
  {
    code: 'nidn264420', name: 'ทัวร์ไต้หวัน ไทเป อาลีซาน เย่หลิว', highlight: 'ทะเลสาบสุริยันจันทรา · ตึก 101 · ตลาดซื่อหลิน',
    countries: ['taiwan'], cities: ['taipei'], days: 5, nights: 4, airline: 'BR', image: img('1470004914212-05527e49370b'),
    periods: [
      ['2026-11-08', '2026-11-12', 21900, 25900, 6],
      ['2026-11-22', '2026-11-26', 22900, 26900, 15],
      ['2026-12-06', '2026-12-10', 23900, 27900, 20],
    ],
  },
  {
    code: 'nidn239905', name: 'ทัวร์ยุโรป อิตาลี สวิตเซอร์แลนด์ ฝรั่งเศส', highlight: 'ยอดเขาจุงเฟรา · ล่องเรือแม่น้ำแซน · หอเอนเมืองปิซ่า',
    countries: ['europe'], cities: ['paris'], days: 9, nights: 6, airline: 'EK', image: img('1499856871958-5b9627545d1a'),
    periods: [
      ['2027-03-07', '2027-03-15', 89900, 99900, 10],
      ['2027-03-21', '2027-03-29', 92900, 102900, 20],
      ['2027-04-04', '2027-04-12', 99900, 109900, 3],
    ],
  },
  {
    code: 'nidn287701', name: 'ทัวร์ตุรกี อิสตันบูล คัปปาโดเกีย', highlight: 'บอลลูนคัปปาโดเกีย · พระราชวังท็อปกาปึ · ปามุคคาเล',
    countries: ['turkey'], cities: ['istanbul'], days: 8, nights: 5, airline: 'TK', image: img('1524231757912-21f4fe3a7200'),
    periods: [
      ['2027-02-06', '2027-02-13', 39900, 45900, 12],
      ['2027-02-20', '2027-02-27', 41900, 47900, 20],
      ['2027-03-06', '2027-03-13', 42900, 48900, 20],
    ],
  },
  {
    code: 'nidn251212', name: 'ทัวร์จอร์เจีย ทบิลิซี คาซเบกี', highlight: 'โบสถ์เกอร์เกติ · เทือกเขาคอเคซัส · ไวน์จอร์เจีย',
    countries: ['georgia'], cities: ['tbilisi'], days: 7, nights: 5, airline: 'QR', image: img('1565008447742-97f6f38c985c'),
    periods: [
      ['2027-01-15', '2027-01-21', 49900, 56900, 15],
      ['2027-02-05', '2027-02-11', 51900, 58900, 20],
    ],
  },
];

const countryRef = (slug: string): TaxonomyRef => {
  const c = COUNTRIES.find((x) => x.slug === slug);
  if (!c) throw new Error(`mock: unknown country ${slug}`);
  return { slug: c.slug, nameTh: c.nameTh, nameEn: c.nameEn };
};

const cityRef = (slug: string): CityRef => {
  const c = CITIES.find((x) => x.slug === slug);
  if (!c) throw new Error(`mock: unknown city ${slug}`);
  return { slug: c.slug, countrySlug: c.countrySlug, nameTh: c.nameTh, nameEn: c.nameEn };
};

// รวม / ไม่รวม samples: most tours have both lists; one has none (section hidden) and the government
// shop flag varies, so every state of TourIncludes renders in a mock build.
const INCLUDED = [
  'ตั๋วเครื่องบินไป–กลับ ชั้นประหยัด พร้อมคณะ รวมภาษีสนามบินและค่าธรรมเนียมเชื้อเพลิง',
  'น้ำหนักกระเป๋าโหลดใต้ท้องเครื่องท่านละ 20 กิโลกรัม',
  'โรงแรมที่พักตามรายการ (พักห้องละ 2–3 ท่าน)',
  'อาหารตามที่ระบุในรายการ',
  'ค่าเข้าชมสถานที่ตามรายการ',
  'รถโค้ชปรับอากาศนำเที่ยว และหัวหน้าทัวร์ดูแลตลอดการเดินทาง',
  'ประกันอุบัติเหตุระหว่างการเดินทาง วงเงิน 1,000,000 บาท',
];
const EXCLUDED = [
  'ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า',
  'ค่าทิปไกด์และคนขับรถท้องถิ่น 1,500 บาท/ท่าน/ทริป',
  'ค่าภาษีน้ำมันที่สายการบินเรียกเก็บเพิ่มภายหลัง',
  'Option: ทัวร์ล่องเรือชมวิวยามค่ำ ท่านละ 1,200 บาท',
];
function inclusions(t: TourSeed, i: number): Pick<TourProgramPublicDetail, 'governmentShopStops' | 'includedItems' | 'excludedItems' | 'itinerary'> {
  if (i === TOURS.length - 1) return { governmentShopStops: null, includedItems: [], excludedItems: [], itinerary: [] };
  const china = t.countries.includes('china');
  return {
    governmentShopStops: china ? i % 2 === 0 : t.countries.includes('vietnam') ? false : null,
    includedItems: INCLUDED,
    excludedItems: i % 3 === 0 ? EXCLUDED : EXCLUDED.slice(0, 3),
    itinerary: timeline(t),
  };
}

// A plausible timeline from the seed: fly out, the highlights one per day, fly back.
function timeline(t: TourSeed): TourProgramPublicDetail['itinerary'] {
  const city = t.cities[0] ? cityRef(t.cities[0]).nameTh : countryRef(t.countries[0]).nameTh;
  const sights = t.highlight.split(' · ');
  return Array.from({ length: t.days }, (_, i) => {
    if (i === 0) return { day: 1, places: ['สนามบินสุวรรณภูมิ', `สนามบิน${city}`] };
    if (i === t.days - 1) return { day: i + 1, places: [`สนามบิน${city}`, 'สนามบินสุวรรณภูมิ'] };
    return { day: i + 1, places: [city, sights[(i - 1) % sights.length]] };
  });
}

// โปรไฟไหม้ samples: period id → staff order. Five upcoming ones, so the home page shows the first four.
const FIRE_SALES: Record<string, number> = {
  'nidn261690-p2': 1,
  'nidn255120-p2': 2,
  'nidn248801-p3': 3,
  'nidn261702-p3': 4,
  'nidn271110-p1': 5,
};

function toDetail(t: TourSeed, index: number): TourProgramPublicDetail {
  const periods: TourPeriodPublic[] = t.periods.map(([startDate, endDate, price, before, seats], i) => ({
    id: `${t.code}-p${i + 1}`,
    startDate,
    endDate,
    price,
    priceBeforeDiscount: before,
    childWithBedPrice: price - 2000,
    childNoBedPrice: price - 4000,
    singleSupplement: Math.round(price * 0.18 / 100) * 100,
    deposit: price >= 60000 ? 20000 : 10000,
    seatsAvailable: seats,
    soldOut: false, // like the backend: only a closed tour is sold out; 0 seats = not specified
    fireSale: `${t.code}-p${i + 1}` in FIRE_SALES && before !== null && before > price,
  }));
  // Backend rule: starting price = lowest active period price (sold-out periods still count as active).
  const cheapest = periods.reduce((a, b) => (b.price < a.price ? b : a));
  return {
    code: t.code,
    name: t.name,
    highlight: t.highlight,
    days: t.days,
    nights: t.nights,
    airline: AIRLINES[t.airline] ?? null,
    cardImage: t.image,
    countries: t.countries.map(countryRef),
    cities: t.cities.map(cityRef),
    startingPrice: cheapest.price,
    startingPriceBeforeDiscount: cheapest.priceBeforeDiscount,
    nextDepartureDate: periods[0]?.startDate ?? null,
    periodCount: periods.length,
    fireSales: periods.filter((p) => p.fireSale).map((p) => ({ ...p, order: FIRE_SALES[p.id] })),
    updateDate: UPDATED,
    description: [
      `${t.name} ${t.days} วัน ${t.nights} คืน บินกับ ${AIRLINES[t.airline]?.name ?? 'สายการบินชั้นนำ'} ไฮไลท์ของทริป ${t.highlight.replaceAll(' · ', ' ')}`,
      'โปรแกรมนี้จัดเส้นทางให้มีเวลาเที่ยวแต่ละจุดอย่างเต็มที่ ไม่เร่งรีบ พักโรงแรมมาตรฐาน มีหัวหน้าทัวร์คนไทยดูแลตลอดการเดินทาง',
      'ดาวน์โหลดรายละเอียดโปรแกรมแบบรายวันได้จากไฟล์ PDF หรือทัก LINE เพื่อสอบถามรายละเอียดเพิ่มเติมและสำรองที่นั่ง',
    ].join('\n\n'),
    bannerImage: t.image,
    images: GALLERY_POOL.filter((src) => src !== t.image).slice(0, 4),
    pdfFile: null,
    periods,
    ...inclusions(t, index),
  };
}

export const mockTours: TourProgramPublicDetail[] = TOURS.map(toDetail);

function stats(match: (t: TourProgramPublicDetail) => boolean) {
  const tours = mockTours.filter(match);
  const prices = tours.map((t) => t.startingPrice).filter((p): p is number => p !== null);
  return { tourCount: tours.length, startingPrice: prices.length ? Math.min(...prices) : null };
}

export const mockCountries: CountryPublic[] = COUNTRIES.map((c) => ({
  ...c,
  ...stats((t) => t.countries.some((x) => x.slug === c.slug)),
  updateDate: UPDATED,
}));

export const mockCities: CityPublic[] = CITIES.map((c) => ({
  ...c,
  ...stats((t) => t.cities.some((x) => x.slug === c.slug)),
  updateDate: UPDATED,
}));
