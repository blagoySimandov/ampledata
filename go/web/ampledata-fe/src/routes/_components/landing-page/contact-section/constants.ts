/**
 * Formspree sends emails to ampledata.io
 */
export const FORMSPREE_ENDPOINT = "https://formspree.io/f/xeerypze";

export interface FormState {
  name: string;
  email: string;
  subject: string;
  message: string;
}

export const EMPTY_FORM: FormState = {
  name: "",
  email: "",
  subject: "",
  message: "",
};

export type SubmitStatus = "idle" | "loading" | "success" | "error";
