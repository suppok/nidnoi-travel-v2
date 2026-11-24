import type { TourProgramState } from '../constants/tourProgramStates.ts';
import type { TourItineraryDay } from './tourProgram.ts';

// GET /tour-programs/ai-extract/config
export type AIExtractConfig = { ready: boolean; maxPdfSizeMb: number };

// POST /tour-programs/ai-extract (multipart "file") → poll GET /tour-programs/ai-extract/:jobId
export type AIExtractStart = { jobId: string; state: 'pending' };

export type AIConfidenceLevel = 'high' | 'medium' | 'low';
export type AIConfidence = {
  main: AIConfidenceLevel;
  taxonomy: AIConfidenceLevel;
  periods: AIConfidenceLevel;
  prices: AIConfidenceLevel;
  extras: AIConfidenceLevel;
  inclusions: AIConfidenceLevel; // government shops and the included / excluded lists
  itinerary: AIConfidenceLevel;
};

// field = the form field it is about ("name", "countryIds", "periods[3].price"), '' = the whole file.
export type AIWarning = { field: string; message: string };

// A period create body without tourProgramId. Seats are never in the PDF: 0 / 0 (not specified).
export type AIDraftPeriod = {
  startDate: string;
  endDate: string;
  price: number | null;
  priceBeforeDiscount: number | null;
  childWithBedPrice: number | null;
  childNoBedPrice: number | null;
  infantPrice: number | null;
  singleSupplement: number | null;
  deposit: number | null;
  seatsTotal: number | null;
  seatsAvailable: number | null;
  note: string;
};

// Shaped like TourProgramCreateRequest; null where the PDF had nothing usable. Never carries a code.
export type AIDraft = {
  refId: number | null;
  name: string;
  wholesaleId: string | null; // always null: staff pick the wholesale
  wholesaleTourCode: string;
  state: TourProgramState;
  days: number | null;
  nights: number | null;
  airlineId: string | null;
  highlight: string;
  description: string;
  bannerImage: string | null;
  cardImage: string | null;
  pdfFile: string | null;
  docxFile: string | null;
  internalNote: string;
  governmentShopStops: boolean | null;
  includedItems: string[];
  excludedItems: string[];
  itinerary: TourItineraryDay[];
  countryIds: string[];
  cityIds: string[];
  images: string[];
  periods: AIDraftPeriod[];
};

export type AIExtractResult = {
  draft: AIDraft;
  warnings: AIWarning[];
  confidence: AIConfidence;
  usage: { inputTokens: number; outputTokens: number; cacheReadInputTokens: number; cacheCreationInputTokens: number };
  model: string;
  fileName: string;
  durationMs: number;
};

export type AIExtractJob = {
  jobId: string;
  state: 'pending' | 'done' | 'failed';
  result: AIExtractResult | null;
  error: { code: string; message: string } | null;
};
