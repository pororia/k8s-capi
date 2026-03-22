import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const PUBLIC_PATHS = ['/login']

export function middleware(request: NextRequest) {
  const path = request.nextUrl.pathname
  const isPublic = PUBLIC_PATHS.some((p) => path.startsWith(p))

  // We can't read localStorage in middleware, so we rely on the client-side redirect
  // This middleware just passes through — auth check is done in client components
  if (isPublic) {
    return NextResponse.next()
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|api).*)'],
}
