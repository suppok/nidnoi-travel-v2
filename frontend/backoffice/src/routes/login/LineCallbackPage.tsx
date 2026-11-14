import { Alert, Button, Spin } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router';
import { ApiError } from '../../api/client.ts';
import { clearPendingLineLogin, lineRedirectUri, peekPendingLineLogin } from '../../auth/lineLogin.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { ErrorCodes } from '../../constants/errorCodes.ts';
import { errorMessage } from '../../lib/errors.ts';

type Callback = { code: string; pending: NonNullable<ReturnType<typeof peekPendingLineLogin>> } | { error: string };

// What LINE sent back (?code&state, or ?error when the user cancelled), checked against the login
// this browser started.
function readCallback(params: URLSearchParams): Callback {
  if (params.get('error')) return { error: 'ยกเลิกการเข้าสู่ระบบด้วย LINE' };
  const pending = peekPendingLineLogin(params.get('state'));
  const code = params.get('code');
  if (!pending || !code) return { error: 'ลิงก์เข้าสู่ระบบด้วย LINE หมดอายุหรือไม่ถูกต้อง กรุณาลองใหม่' };
  return { code, pending };
}

// LINE sends the browser back here after "ล็อกอินด้วย LINE". The code is exchanged once; a LINE
// account no user linked is denied by the backend.
export default function LineCallbackPage() {
  const { user, loginWithLine } = useAuth();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [callback] = useState(() => readCallback(params));
  const [failure, setFailure] = useState<string | null>(null);
  const sent = useRef(false); // StrictMode runs effects twice; a code works only once

  useEffect(() => {
    if (!('code' in callback) || sent.current) return;
    sent.current = true;
    const { code, pending } = callback;
    clearPendingLineLogin();
    loginWithLine({ code, redirectUri: lineRedirectUri(), codeVerifier: pending.verifier, nonce: pending.nonce, rememberMe: pending.rememberMe })
      .then(() => navigate(pending.returnTo || '/dashboard', { replace: true }))
      .catch((err: unknown) =>
        setFailure(
          err instanceof ApiError && err.code === ErrorCodes.LineNotLinked
            ? 'บัญชี LINE นี้ยังไม่ได้เชื่อมกับผู้ใช้ในระบบหลังบ้าน — เข้าสู่ระบบด้วยอีเมลก่อน แล้วเชื่อมบัญชี LINE ที่เมนูชื่อผู้ใช้ "เชื่อมบัญชี LINE"'
            : errorMessage(err),
        ),
      );
  }, [callback, loginWithLine, navigate]);

  const error = 'error' in callback ? callback.error : failure;
  // Already signed in and nothing to finish (e.g. the back button onto this page).
  if (user && 'error' in callback) return <Navigate to="/dashboard" replace />;

  return (
    <div className="flex w-full max-w-[480px] flex-col gap-5 rounded-[28px] bg-white px-10 py-10 shadow-[0_30px_60px_-30px_rgba(29,43,79,0.28)]">
      <h1 className="m-0 font-display text-[28px] font-normal">เข้าสู่ระบบด้วย LINE</h1>
      {error ? (
        <>
          <Alert type="error" showIcon title="เข้าสู่ระบบไม่สำเร็จ" description={error} />
          <Link to="/login">
            <Button type="primary" block size="large">
              กลับไปหน้าเข้าสู่ระบบ
            </Button>
          </Link>
        </>
      ) : (
        <div className="flex items-center gap-3 text-ink-muted">
          <Spin /> กำลังตรวจสอบบัญชี LINE…
        </div>
      )}
    </div>
  );
}
