import liff from '@line/liff'
import { useEffect, useState } from 'react'
import { env } from '../config/env.ts'

type LiffState =
  | { status: 'loading' }
  | { status: 'ready'; userId: string; displayName: string; accessToken: string }
  | { status: 'error'; message: string }

export function useLiff(): LiffState {
  const [state, setState] = useState<LiffState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false

    async function init() {
      try {
        await liff.init({ liffId: env.LIFF_ID })

        if (!liff.isLoggedIn()) {
          liff.login()
          return
        }

        const profile = await liff.getProfile()
        const accessToken = liff.getAccessToken()

        if (cancelled) return

        if (!accessToken) {
          setState({ status: 'error', message: 'ไม่สามารถรับ access token ได้' })
          return
        }

        setState({
          status: 'ready',
          userId: profile.userId,
          displayName: profile.displayName,
          accessToken,
        })
      } catch (err) {
        if (cancelled) return
        setState({
          status: 'error',
          message: err instanceof Error ? err.message : 'LIFF init failed',
        })
      }
    }

    init()
    return () => {
      cancelled = true
    }
  }, [])

  return state
}
