import { ApiReferenceReact } from "@scalar/api-reference-react";
import "@scalar/api-reference-react/style.css";

const customCss = `
:root {
  --scalar-color-accent: oklch(0.553 0.195 38.402);
  --scalar-font: "Figtree Variable", ui-sans-serif, system-ui, sans-serif;
}
.dark-mode {
  --scalar-color-accent: oklch(0.72 0.17 41);
}
`;

function isDarkTheme() {
  return typeof document !== "undefined" && document.documentElement.classList.contains("dark");
}

export default function ApiDocs() {
  return (
    <ApiReferenceReact
      configuration={{
        url: "/openapi.json",
        theme: "default",
        customCss,
        hideDarkModeToggle: true,
        forceDarkModeState: isDarkTheme() ? "dark" : "light",
        metaData: { title: "AmpleData API" },
      }}
    />
  );
}
