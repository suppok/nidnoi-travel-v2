import type { ReactNode } from 'react';

export default function SectionTitle({ icon, title, subtitle, extra }: { icon?: ReactNode; title: string; subtitle?: ReactNode; extra?: ReactNode }) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h2 className="m-0 flex items-center gap-2 text-[17px] font-semibold text-ink">
          {icon && <span className="text-ocean">{icon}</span>}
          {title}
        </h2>
        {subtitle && <p className="m-0 mt-0.5 text-[13px] text-ink-muted">{subtitle}</p>}
      </div>
      {extra}
    </div>
  );
}
