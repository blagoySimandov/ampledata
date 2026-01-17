import { LoginClient } from "./login-client"

export default async function LoginPage({
  searchParams,
}: {
  searchParams?: Promise<{ returnTo?: string }>
}) {
  const resolvedSearchParams = searchParams ? await searchParams : undefined
  const returnTo = resolvedSearchParams?.returnTo ?? "/enrich"
  return <LoginClient returnTo={returnTo} />
}

