import { render } from "../dist-ssr/entry-server.js";
import fs from "node:fs";

//Uses node to prerender and compile the landing page for SEO improvements
const html = await render();
const template = fs.readFileSync("dist/index.html", "utf-8");
const result = template.replace(
  '<div id="root"></div>',
  `<div id="root">${html}</div>`,
);
fs.writeFileSync("dist/index.html", result);
console.log("Pre-rendered /");
