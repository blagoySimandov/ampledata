"use client";

import { useAuth } from "@workos-inc/authkit-react";
import { useEffect } from "react";

export function LoginClient({ returnTo }: { returnTo: string }) {
	const { signIn } = useAuth();

	useEffect(() => {
		signIn({ state: { returnTo } });
	}, [signIn, returnTo]);

	return (
		<div className="flex items-center justify-center min-h-screen">
			<div className="text-center">
				<p className="text-lg">Redirecting to sign in...</p>
			</div>
		</div>
	);
}
