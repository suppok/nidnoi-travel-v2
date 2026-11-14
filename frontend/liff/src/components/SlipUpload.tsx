import { useRef, useState } from 'react'
import { ApiError, uploadSlip } from '../api/client.ts'

// Slips per payment; MUST match the backend (maxSlips).
export const MAX_SLIPS = 10

type Props = {
  value: string[]
  onChange: (urls: string[]) => void
}

// One or more slip images (a payment split over several transfers). Picking several files uploads
// them one by one; each thumbnail can be removed.
export function SlipUpload({ value, onChange }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(0) // files still uploading
  const [error, setError] = useState('')
  const room = MAX_SLIPS - value.length

  async function handleFiles(files: File[]) {
    setError('')
    const images = files.filter((f) => f.type.startsWith('image/'))
    if (images.length < files.length) setError('เลือกได้เฉพาะไฟล์รูปภาพ')
    const picked = images.slice(0, room)
    if (images.length > room) setError(`แนบสลิปได้ไม่เกิน ${MAX_SLIPS} รูป`)
    if (picked.length === 0) return

    setUploading(picked.length)
    let urls = value
    for (const file of picked) {
      try {
        const url = await uploadSlip(file)
        urls = [...urls, url]
        onChange(urls)
      } catch (err) {
        setError(err instanceof ApiError ? err.message : 'อัพโหลดไม่สำเร็จ กรุณาลองอีกครั้ง')
      } finally {
        setUploading((n) => n - 1)
      }
    }
  }

  return (
    <div className="space-y-2">
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        multiple
        onChange={(e) => {
          const files = Array.from(e.target.files ?? [])
          e.target.value = '' // the same file can be picked again after removing it
          if (files.length) void handleFiles(files)
        }}
        className="hidden"
      />

      {value.length > 0 && (
        <div className="grid grid-cols-3 gap-2">
          {value.map((url, i) => (
            <div key={url} className="relative">
              <a href={url} target="_blank" rel="noreferrer">
                <img src={url} alt={`สลิป ${i + 1}`} className="aspect-[3/4] w-full rounded-lg border border-gray-200 bg-gray-50 object-cover" />
              </a>
              <button
                type="button"
                onClick={() => onChange(value.filter((u) => u !== url))}
                disabled={uploading > 0} // the upload loop appends to the list it started with
                aria-label={`ลบสลิป ${i + 1}`}
                className="absolute right-1 top-1 flex h-6 w-6 items-center justify-center rounded-full bg-black/60 text-sm leading-none text-white disabled:opacity-40"
              >
                ×
              </button>
            </div>
          ))}
          {room > 0 && (
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              disabled={uploading > 0}
              className="flex aspect-[3/4] w-full flex-col items-center justify-center rounded-lg border-2 border-dashed border-gray-300 text-xs text-gray-500 disabled:opacity-50"
            >
              <span className="text-2xl leading-none">+</span>
              {uploading > 0 ? `กำลังอัพโหลด (${uploading})` : 'เพิ่มสลิป'}
            </button>
          )}
        </div>
      )}

      {value.length === 0 && (
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          disabled={uploading > 0}
          className="flex w-full flex-col items-center justify-center gap-1 rounded-lg border-2 border-dashed border-gray-300 py-8 text-sm text-gray-500 transition hover:border-blue-400 hover:text-blue-500 disabled:opacity-50"
        >
          {uploading > 0 ? `กำลังอัพโหลด (${uploading})...` : 'แตะเพื่อเลือกรูปสลิป'}
          {uploading === 0 && <span className="text-xs text-gray-400">เลือกได้หลายรูป (ไม่เกิน {MAX_SLIPS} รูป)</span>}
        </button>
      )}

      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
