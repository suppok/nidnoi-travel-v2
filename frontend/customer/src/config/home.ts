// Home hero campaign. Static (no carousel) for LCP; change it per season.
// The linked country must have active tours, or the build fails (see index.astro).
export const heroCampaign = {
  countrySlug: 'japan',
  eyebrow: 'ใบไม้แดงโคโยะ · ตุลาคม–ธันวาคม',
  titleLead: 'ทัวร์ญี่ปุ่น',
  titlePrefix: 'ฤดู',
  titleEmphasis: 'ใบไม้เปลี่ยนสี',
  sub: 'เก็บทุกวิวสีแดงสดตั้งแต่โตเกียวไปจนถึงเกียวโต ในจังหวะที่ใบไม้กำลังเปลี่ยนสี เลือกวันเดินทางและเช็กที่นั่งว่างได้ทันที',
  location: 'เกียวโต, ญี่ปุ่น',
  // TODO(content): replace with an R2 image (media.*) before launch.
  image: 'https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=2000&q=80',
  imageAlt: 'วัดในเกียวโตท่ามกลางใบไม้เปลี่ยนสีฤดูใบไม้ร่วง',
} as const;
