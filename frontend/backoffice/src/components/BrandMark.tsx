export function PlaneIcon({ size = 22 }: { size?: number }) {
  return (
    <svg aria-hidden="true" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 3 10 14" />
      <path d="M21 3l-7 18-4-7-7-4 18-7z" />
    </svg>
  );
}

// Logo tile + "nidnoi travel" wordmark.
export default function BrandMark({ size = 'md', caption }: { size?: 'md' | 'lg'; caption?: string }) {
  const lg = size === 'lg';
  return (
    <span className="flex items-center gap-3">
      <span
        className={`flex flex-none items-center justify-center bg-brand text-white ${lg ? 'h-11 w-11 rounded-[14px]' : 'h-[38px] w-[38px] rounded-xl'}`}
      >
        <PlaneIcon size={lg ? 22 : 20} />
      </span>
      <span className="flex flex-col">
        <span className={`font-display leading-tight font-medium ${lg ? 'text-[28px]' : 'text-xl'}`}>
          <span className="text-brand">nidnoi</span> <span className="text-ocean">travel</span>
        </span>
        {caption && <span className="text-xs font-medium tracking-[0.08em] text-ink-muted">{caption}</span>}
      </span>
    </span>
  );
}
