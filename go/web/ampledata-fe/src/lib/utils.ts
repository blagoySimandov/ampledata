import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function userInitials(
  firstName: string | null,
  lastName: string | null,
  email: string,
) {
  if (firstName && lastName)
    return `${firstName[0]}${lastName[0]}`.toUpperCase();
  if (firstName) return firstName[0].toUpperCase();
  return email[0].toUpperCase();
}

export function userDisplayName(
  firstName: string | null,
  lastName: string | null,
  email: string,
) {
  if (firstName || lastName)
    return [firstName, lastName].filter(Boolean).join(" ");
  return email;
}
