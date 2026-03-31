import { useState } from "react";
import type { UserResponse } from "@/api";
import { FORMSPREE_ENDPOINT } from "./constants";
import type { ContactFormVariant, SubmitStatus } from "./types";

interface Props {
  variant: ContactFormVariant;
  user?: UserResponse;
}

export function useContactForm({ variant, user }: Props) {
  const [name, setName] = useState(
    user ? `${user.first_name} ${user.last_name}`.trim() : "",
  );
  const [email, setEmail] = useState(user?.email ?? "");
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [status, setStatus] = useState<SubmitStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");

  const isLanding = variant === "landing";

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setErrorMessage("");

    const payload = isLanding
      ? { name, email, subject, message }
      : { name, email, message };

    try {
      const res = await fetch(FORMSPREE_ENDPOINT, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setStatus("success");
        setSubject("");
        setMessage("");
      } else {
        const data = (await res.json().catch(() => ({}))) as { error?: string };
        setErrorMessage(
          data.error ?? "Something went wrong. Please try again.",
        );
        setStatus("error");
      }
    } catch {
      setErrorMessage(
        "Network error. Please check your connection and try again.",
      );
      setStatus("error");
    }
  }

  return {
    isLanding,
    name,
    setName,
    email,
    setEmail,
    subject,
    setSubject,
    message,
    setMessage,
    status,
    setStatus,
    errorMessage,
    handleSubmit,
  };
}
