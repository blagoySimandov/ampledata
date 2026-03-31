import type { UserResponse } from "@/api";

export type ContactFormVariant = "landing" | "account";

export type SubmitStatus = "idle" | "loading" | "success" | "error";

export interface ContactFormWidgetProps {
  variant: ContactFormVariant;
  user?: UserResponse;
}
