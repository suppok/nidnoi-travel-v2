import { useEffect, useState } from 'react'
import { ApiError, getLiffMe, linkLineAccount } from '../api/client.ts'
import type { LIFFMe } from '../types/index.ts'

// Why the app cannot be used by this LINE account.
export type StaffBlock = 'not-linked' | 'no-permission' | 'password-change' | 'link-invalid' | 'already-linked' | 'error'

type StaffState =
  | { status: 'loading' }
  | { status: 'ready'; me: LIFFMe; justLinked: boolean }
  | { status: 'blocked'; reason: StaffBlock; message?: string }

const LINK_PARAM = 'lineLink'

// The one-time code from the backoffice QR code (https://liff.line.me/<id>?lineLink=…). LINE may
// still carry it inside liff.state on the first load.
function readLinkCode(): string | null {
  const params = new URLSearchParams(window.location.search)
  const direct = params.get(LINK_PARAM)
  if (direct) return direct
  const state = params.get('liff.state')
  if (!state) return null
  const query = state.includes('?') ? state.slice(state.indexOf('?') + 1) : state
  return new URLSearchParams(query).get(LINK_PARAM)
}

function clearLinkCode() {
  const url = new URL(window.location.href)
  url.searchParams.delete(LINK_PARAM)
  url.searchParams.delete('liff.state')
  window.history.replaceState(null, '', url.toString())
}

function blockFor(err: unknown): StaffState {
  if (err instanceof ApiError) {
    switch (err.code) {
      case 'LINE_NOT_LINKED':
        return { status: 'blocked', reason: 'not-linked' }
      case 'PERMISSION_DENIED':
        return { status: 'blocked', reason: 'no-permission' }
      case 'PASSWORD_CHANGE_REQUIRED':
        return { status: 'blocked', reason: 'password-change' }
      case 'LINE_LINK_CODE_INVALID':
        return { status: 'blocked', reason: 'link-invalid' }
      case 'LINE_ALREADY_LINKED':
        return { status: 'blocked', reason: 'already-linked' }
    }
    return { status: 'blocked', reason: 'error', message: err.message }
  }
  return { status: 'blocked', reason: 'error', message: err instanceof Error ? err.message : undefined }
}

// Resolves the backoffice user behind this LINE account, finishing a QR-code link first when the
// URL carries one. Call only after LIFF is ready and the token is set on the API client.
export function useStaff(enabled: boolean): StaffState {
  const [state, setState] = useState<StaffState>({ status: 'loading' })

  useEffect(() => {
    if (!enabled) return
    let cancelled = false

    async function run() {
      const code = readLinkCode()
      let justLinked = false
      if (code) {
        clearLinkCode()
        try {
          await linkLineAccount(code)
          justLinked = true
        } catch (err) {
          if (!cancelled) setState(blockFor(err))
          return
        }
      }
      try {
        const me = await getLiffMe()
        if (!cancelled) setState({ status: 'ready', me, justLinked })
      } catch (err) {
        if (!cancelled) setState(blockFor(err))
      }
    }

    void run()
    return () => {
      cancelled = true
    }
  }, [enabled])

  return state
}
