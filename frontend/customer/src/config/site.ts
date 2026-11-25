// Business facts shown across the site. Keep them here so marketing copy has one source of truth.
// Only verified values go live; anything unconfirmed stays empty and its UI is hidden.

export const LINE_ID = '@nidnoi_travel';

export const contact = {
  phones: [
    { display: '061 926 6611', tel: '+66619266611' },
    { display: '02 044 6161', tel: '+6620446161' },
  ],
  email: 'nidnoitravel@gmail.com',
  lineId: LINE_ID,
  lineUrl: `https://line.me/R/ti/p/${encodeURIComponent(LINE_ID)}`,
  hours: ['จันทร์–เสาร์', '09:00–19:00'],
} as const;

/** Opens a LINE chat with a prefilled message (used for "ask about this tour"). */
export function lineMessageUrl(text: string): string {
  return `https://line.me/R/oaMessage/${encodeURIComponent(LINE_ID)}/?${encodeURIComponent(text)}`;
}

export type SocialLink = { name: 'facebook' | 'instagram' | 'youtube' | 'tiktok'; label: string; url: string };

// TODO(content): add the real profile URLs. Empty list = icons hidden.
export const socialLinks: SocialLink[] = [];

// TODO(content): license number from the Department of Tourism (ใบอนุญาตนำเที่ยว). Empty = not shown.
export const tourLicenseNo = '';

export type Stat = { value: string; label: string };

// TODO(content): real, verifiable numbers only (travellers served, countries, rating). Empty = row hidden.
export const companyStats: Stat[] = [];

// Trust signals on the home page. Avoid claims that need a number we haven't confirmed.
export const promises = [
  {
    icon: 'medal',
    title: 'คุณภาพคุ้มราคา',
    text: 'คัดโรงแรม มื้ออาหาร และเส้นทางที่ไม่เร่งรีบ ให้ทุกวันของทริปคุ้มค่าที่สุด',
  },
  {
    icon: 'headset',
    title: 'ตอบไว คุยกับคนจริง',
    text: 'ทักแชท LINE หรือโทรหาเรา ทีมงานพร้อมช่วยตั้งแต่ก่อนจองจนจบทริป',
  },
  {
    icon: 'shield',
    title: 'เดินทางอย่างมั่นใจ',
    text: 'ราคาและวันเดินทางชัดเจน เช็กที่นั่งว่างได้ตลอดเวลา ไม่มีค่าใช้จ่ายแอบแฝง',
  },
] as const;

// Fallback hero for landing pages whose country/city has no heroImage yet (never ship a landing without one).
// TODO(content): replace with an R2 image (media.*) before launch.
export const DEFAULT_HERO_IMAGE = 'https://images.unsplash.com/photo-1488085061387-422e29b40080?w=2000&q=80';

// Placeholder photos for countries whose heroImage hasn't been uploaded in the backoffice yet, keyed by slug.
// A real heroImage always wins; countries not listed here fall back to DEFAULT_HERO_IMAGE where a photo is required.
// TODO(content): upload real hero images in the backoffice, then delete this map before launch.
const unsplash = (id: string) => `https://images.unsplash.com/photo-${id}?w=2000&q=80`;
export const PLACEHOLDER_COUNTRY_IMAGES: Readonly<Record<string, string>> = {
  croatia: unsplash('1555990793-da11153b2473'), // Dubrovnik
  georgia: unsplash('1565008447742-97f6f38c985c'),
  jordan: unsplash('1579606032821-4e6161c81bd3'), // Petra
  china: unsplash('1508804185872-d7badad00f7d'),
  denmark: unsplash('1513622470522-26c3c8a854bc'), // Nyhavn
  norway: unsplash('1601439678777-b2b3c56fa627'), // Geirangerfjord
  finland: unsplash('1507272931001-fc06c17e4f43'), // aurora
  france: unsplash('1499856871958-5b9627545d1a'), // Paris
  singapore: unsplash('1565967511849-76a60a516170'),
  japan: unsplash('1493976040374-85c8e12f0c0e'),
  korea: unsplash('1517154421773-0529f29ea451'),
  taiwan: unsplash('1470004914212-05527e49370b'),
  vietnam: unsplash('1528127269322-539801943592'),
  turkey: unsplash('1524231757912-21f4fe3a7200'),
};
