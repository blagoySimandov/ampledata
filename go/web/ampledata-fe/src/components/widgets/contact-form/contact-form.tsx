import { AccountContactCard } from "./account-contact-card";
import { ContactFormContent } from "./contact-form-content";
import { LandingContactSection } from "./landing-contact-section";
import { useContactForm } from "./use-contact-form";
import type { ContactFormWidgetProps } from "./types";

export function ContactFormWidget({ variant, user }: ContactFormWidgetProps) {
  const {
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
  } = useContactForm({ variant, user });

  const content = (
    <ContactFormContent
      variant={variant}
      name={name}
      email={email}
      subject={subject}
      message={message}
      status={status}
      errorMessage={errorMessage}
      onNameChange={setName}
      onEmailChange={setEmail}
      onSubjectChange={setSubject}
      onMessageChange={setMessage}
      onStatusReset={() => setStatus("idle")}
      onSubmit={handleSubmit}
    />
  );

  if (!isLanding) {
    return <AccountContactCard>{content}</AccountContactCard>;
  }

  return <LandingContactSection>{content}</LandingContactSection>;
}
