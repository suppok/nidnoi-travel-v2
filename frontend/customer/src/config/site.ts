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
