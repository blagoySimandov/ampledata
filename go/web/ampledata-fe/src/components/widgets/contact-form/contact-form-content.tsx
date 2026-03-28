import { ContactFormFields } from "./contact-form-fields";
import { ContactFormSuccess } from "./contact-form-success";
import type { SubmitStatus } from "./types";

interface Props {
  variant: "landing" | "account";
  name: string;
  email: string;
  subject: string;
  message: string;
  status: SubmitStatus;
  errorMessage: string;
  onNameChange: (value: string) => void;
  onEmailChange: (value: string) => void;
  onSubjectChange: (value: string) => void;
  onMessageChange: (value: string) => void;
  onStatusReset: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

export function ContactFormContent({
  variant,
  name,
  email,
  subject,
  message,
  status,
  errorMessage,
  onNameChange,
  onEmailChange,
  onSubjectChange,
  onMessageChange,
  onStatusReset,
  onSubmit,
}: Props) {
  if (status === "success") {
    return <ContactFormSuccess onReset={onStatusReset} />;
  }

  return (
    <ContactFormFields
      variant={variant}
      name={name}
      email={email}
      subject={subject}
      message={message}
      status={status}
      errorMessage={errorMessage}
      onNameChange={onNameChange}
      onEmailChange={onEmailChange}
      onSubjectChange={onSubjectChange}
      onMessageChange={onMessageChange}
      onSubmit={onSubmit}
    />
  );
}
