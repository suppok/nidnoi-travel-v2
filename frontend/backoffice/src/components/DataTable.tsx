import { Skeleton, Table, type TableProps } from 'antd';
import type { CSSProperties } from 'react';

const SKELETON_ROWS = 8;
// Rows past this one share its delay, so a long list still finishes fading in quickly.
const MAX_STAGGER_ROWS = 12;
const SKELETON_KEY = '__skeleton';

// Bar widths vary by row so the placeholder reads as content, not a grid of identical blocks.
const BAR_WIDTHS = ['85%', '60%', '75%', '50%', '90%', '65%', '70%', '55%'];

type SkeletonRow = { [SKELETON_KEY]: string };

/**
 * antd `Table` that shows skeleton rows in the real columns while `loading`, instead of a spinner over the
 * table. Takes the same props. While loading, row callbacks (onRow, rowClassName, rowSelection, expandable,
 * summary, custom components) are dropped: the placeholder rows carry no record for them to read.
 */
export default function DataTable<T extends object>({ loading, columns, ...rest }: TableProps<T>) {
  const isLoading = typeof loading === 'object' ? !!loading.spinning : !!loading;
  // Rows fade in briefly when they mount (after the skeleton, or new rows), one after another from the top;
  // edited rows keep their key and don't. The row index feeds the stagger delay (--row, see global.css).
  if (!isLoading) {
    const { onRow, className } = rest;
    return (
      <Table<T>
        columns={columns}
        {...rest}
        className={['data-table-fade', className].filter(Boolean).join(' ')}
        onRow={(record, index) => {
          const props = onRow?.(record, index) ?? {};
          return { ...props, style: { ...props.style, '--row': Math.min(index ?? 0, MAX_STAGGER_ROWS) } as CSSProperties };
        }}
      />
    );
  }

  const rows: SkeletonRow[] = Array.from({ length: SKELETON_ROWS }, (_, i) => ({ [SKELETON_KEY]: `skeleton-${i}` }));
  const skeletonColumns: TableProps<SkeletonRow>['columns'] = (columns ?? []).map((col, c) => ({
    key: col.key ?? `col-${c}`,
    // A function title (sort-aware) needs the table's sort state; plain titles are all this app uses.
    title: typeof col.title === 'function' ? null : col.title,
    width: col.width,
    align: col.align,
    fixed: col.fixed,
    className: col.className,
    render: (_: unknown, __: SkeletonRow, r: number) => (
      <Skeleton.Input active size="small" block style={{ width: BAR_WIDTHS[(r + c) % BAR_WIDTHS.length], minWidth: 24, height: 14 }} />
    ),
  }));

  return (
    <Table<SkeletonRow>
      aria-busy="true"
      columns={skeletonColumns}
      dataSource={rows}
      rowKey={SKELETON_KEY}
      size={rest.size}
      bordered={rest.bordered}
      scroll={rest.scroll}
      tableLayout={rest.tableLayout}
      showHeader={rest.showHeader}
      className={rest.className}
      pagination={false}
    />
  );
}
