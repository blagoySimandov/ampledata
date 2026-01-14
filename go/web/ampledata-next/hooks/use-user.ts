"use client";

import { useAuth } from "@workos-inc/authkit-react";
import { useEffect } from "react";
import { usePathname } from "next/navigation";

type UserOrNull = ReturnType<typeof useAuth>["user"];

export const useUser = (): UserOrNull => {
  const { user, isLoading, signIn } = useAuth();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !user) {
      signIn({ state: { returnTo: pathname } });
    }
  }, [isLoading, user, signIn, pathname]);

  return user;
};
