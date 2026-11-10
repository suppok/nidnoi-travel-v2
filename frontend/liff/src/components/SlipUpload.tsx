import { useRef, useState } from 'react'
import { uploadSlip, ApiError } from '../api/client.ts'

type Props = {
  value: string
  onChange: (url: string) => void
}

export function SlipUpload({ value, onChange }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')

  async function handleFile(file: File) {
    if (!file.type.startsWith('image/')) {
      setError('กรุณาเลือกไฟล์รูปภาพเท่านั้น')
      return
    }

    setError('')
    setUploading(true)
    try {
      const url = await uploadSlip(file)
      onChange(url)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('อัพโหลดไม่สำเร็จ กรุณาลองอีกครั้ง')
      }
    } finally {
      setUploading(false)
    }
  }

  if (value) {
    return (
      <div className="space-y-2">
        <img
          src={value}
          alt="สลิปโอนเงิน"
          className="h-48 w-auto rounded-lg border border-gray-200 object-contain"
        />
        <button
          type="button"
          onClick={() => {
            onChange('')
            if (inputRef.current) inputRef.current.value = ''
          }}
          className="text-sm text-blue-600"
        >
          เปลี่ยนรูป
        </button>
      </div>
    )
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file) handleFile(file)
        }}
        className="hidden"
      />
      <button
        type="button"
        onClick={() => inputRef.current?.click()}
        disabled={uploading}
        className="flex w-full items-center justify-center rounded-lg border-2 border-dashed border-gray-300 py-8 text-sm text-gray-500 transition hover:border-blue-400 hover:text-blue-500 disabled:opacity-50"
      >
        {uploading ? 'กำลังอัพโหลด...' : 'แตะเพื่อเลือกรูปสลิป'}
      </button>
      {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
    </div>
  )
}
