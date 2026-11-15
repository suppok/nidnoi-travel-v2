import { Alert, App, Button, Spin } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router';
import { ApiError } from '../../api/client.ts';
import { authApi } from '../../api/auth.ts';
import { clearPendingLineLogin, lineRedirectUri, peekPendingLineLogin } from '../../auth/lineLogin.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { ErrorCodes } from '../../constants/errorCodes.ts';
import { errorMessage } from '../../lib/errors.ts';

type Pending = NonNullable<ReturnType<typeof peekPendingLineLogin>>;
type Callback = { code: string; pending: Pending } | { error: string; link: boolean };

// What LINE sent back (?code&state, or ?error when the user cancelled), checked against the login
// or link this browser started.
function readCallback(params: URLSearchParams): Callback {
  const pending = peekPendingLineLogin(params.get('state'));
  const link = pending?.mode === 'link';
  if (params.get('error')) {
    clearPendingLineLogin();
    return { error: link ? 'ยกเลิกการเชื่อมบัญชี LINE' : 'ยกเลิกการเข้าสู่ระบบด้วย LINE', link };
  }
  const code = params.get('code');
  if (!pending || !code) return { error: 'ลิงก์จาก LINE หมดอายุหรือไม่ถูกต้อง กรุณาลองใหม่', link };
  return { code, pending };
}

// LINE sends the browser back here after "ล็อกอินด้วย LINE" or "เชื่อมบัญชี LINE". The code is used
// once: to sign in (a LINE account no user linked is denied) or to link the signed-in user's account.
export default function LineCallbackPage() {
  const { status, user, loginWithLine, reloadUser } = useAuth();
  const { message } = App.useApp();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [callback] = useState(() => readCallback(params));
  const [failure, setFailure] = useState<string | null>(null);
  const sent = useRef(false); // StrictMode runs effects twice; a code works only once
  const linking = 'pending' in callback ? callback.pending.mode === 'link' : callback.link;

  useEffect(() => {
    if (!('code' in callback) || sent.current) return;
    const { code, pending } = callback;
    const body = { code, redirectUri: lineRedirectUri(), codeVerifier: pending.verifier, nonce: pending.nonce };

    if (pending.mode === 'link') {
      // Linking needs the session, restored from the refresh cookie on page load.
      if (status === 'loading' || !user) return;
      sent.current = true;
      clearPendingLineLogin();
      authApi
        .lineLink(body)
        .then(async (res) => {
          await reloadUser();
          message.success(`เชื่อมบัญชี LINE "${res.line.displayName}" แล้ว`);
          navigate(pending.returnTo || '/dashboard', { replace: true });
        })
        .catch((err: unknown) => setFailure(errorMessage(err)));
      return;
    }

    sent.current = true;
    clearPendingLineLogin();
    loginWithLine({ ...body, rememberMe: pending.rememberMe })
      .then(() => navigate(pending.returnTo || '/dashboard', { replace: true }))
      .catch((err: unknown) =>
        setFailure(
          err instanceof ApiError && err.code === ErrorCodes.LineNotLinked
            ? 'บัญชี LINE นี้ยังไม่ได้เชื่อมกับผู้ใช้ในระบบหลังบ้าน — เข้าสู่ระบบด้วยอีเมลก่อน แล้วเชื่อมบัญชี LINE ที่เมนูชื่อผู้ใช้ "เชื่อมบัญชี LINE"'
            : errorMessage(err),
        ),
      );
  }, [callback, status, user, loginWithLine, reloadUser, navigate, message]);

  const signedOut = 'pending' in callback && callback.pending.mode === 'link' && status === 'anonymous';
  const error = 'error' in callback ? callback.error : signedOut ? 'กรุณาเข้าสู่ระบบก่อน แล้วเชื่อมบัญชี LINE อีกครั้งที่เมนูชื่อผู้ใช้' : failure;
  // Already signed in and nothing to finish (e.g. the back button onto this page).
  if (user && 'error' in callback && !callback.link) return <Navigate to="/dashboard" replace />;

  return (
    <div className="flex w-full max-w-[480px] flex-col gap-5 rounded-[28px] bg-white px-10 py-10 shadow-[0_30px_60px_-30px_rgba(29,43,79,0.28)]">
      <h1 className="m-0 font-display text-[28px] font-normal">{linking ? 'เชื่อมบัญชี LINE' : 'เข้าสู่ระบบด้วย LINE'}</h1>
      {error ? (
        <>
          <Alert type="error" showIcon title={linking ? 'เชื่อมบัญชีไม่สำเร็จ' : 'เข้าสู่ระบบไม่สำเร็จ'} description={error} />
          <Link to={linking && user ? '/dashboard' : '/login'}>
            <Button type="primary" block size="large">
              {linking && user ? 'กลับหน้าภาพรวม' : 'กลับไปหน้าเข้าสู่ระบบ'}
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
