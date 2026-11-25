// Home hero campaign; change it per season. The background cycles through `images` every 5 s (right-to-left wipe);
// the first image is the LCP image and the og:image, so put the strongest photo first.
// The linked country must have active tours, or the build fails (see index.astro).
export const heroCampaign = {
  countrySlug: 'japan',
  titleLead: 'ทัวร์ญี่ปุ่น',
  titlePrefix: 'ฤดู',
  titleEmphasis: 'ใบไม้เปลี่ยนสี',
  sub: 'เก็บทุกวิวสีแดงสดตั้งแต่โตเกียวไปจนถึงเกียวโต ในจังหวะที่ใบไม้กำลังเปลี่ยนสี เลือกวันเดินทางและเช็กที่นั่งว่างได้ทันที',
  location: 'เกียวโต, ญี่ปุ่น',
  // TODO(content): replace with R2 images (media.*) before launch.
  images: [
    { src: 'https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=2000&q=80', alt: 'วัดในเกียวโตท่ามกลางใบไม้เปลี่ยนสีฤดูใบไม้ร่วง' },
    { src: 'https://images.unsplash.com/photo-1493976040374-85c8e12f0c0e?w=2000&q=80', alt: 'ถนนเก่าในเกียวโตกับเจดีย์ยาซากะ' },
    { src: 'https://images.unsplash.com/photo-1478436127897-769e1b3f0f36?w=2000&q=80', alt: 'อุโมงค์เสาโทริอิศาลเจ้าฟูชิมิอินาริ' },
    { src: 'https://images.unsplash.com/photo-1490806843957-31f4c9a91c65?w=2000&q=80', alt: 'ภูเขาไฟฟูจิ' },
  ],
} as const;
